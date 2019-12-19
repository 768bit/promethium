package images

import (
	"archive/tar"
	"compress/gzip"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"runtime"

	"github.com/768bit/vutils"
	gounits "github.com/docker/go-units"
	"gopkg.in/yaml.v2"
)

type ImageConfFile struct {
	OS        string    `yaml:"OS" json:"OS"`
	Version   string    `yaml:"Version" json:"Version"`
	Type      string    `yaml:"Type" json:"Type"`
	Class     ImageType `yaml:"Class" json:"Class"`
	Size      string    `yaml:"Size" json:"Size"`
	Source    string    `yaml:"Source" json:"Source"`
	sizeBytes int64
	rootPath  string
}

func (icf *ImageConfFile) runDockerBuild(workDir string, mountPoint string) error {
	//docker run --privileged --rm -v $WORK_DIR:/output -v $MOUNT_POINT:/rootfs ubuntu-firecracker
	//based ont he root path we need to do a build using a workspace that we cleanup
	//first we need to build the container...
	imagename := "prm-bootstrap-build-" + icf.OS + "-" + icf.Version
	cmd := vutils.Exec.CreateAsyncCommand("docker", false, "build", "-t", imagename, ".").BindToStdoutAndStdErr().SetWorkingDir(icf.rootPath)
	if err := cmd.StartAndWait(); err != nil {
		return err
	} else {
		//we successfully built the image.. now execute the container...
		dockerCmd := vutils.Exec.CreateAsyncCommand("docker", false, "run", "--privileged", "--rm", "-v", workDir+":/output", "-v", mountPoint+":/rootfs", imagename)
		return dockerCmd.BindToStdoutAndStdErr().SetWorkingDir(workDir).StartAndWait()
	}
}

type ImageArchitecture string

const (
	X86_64  ImageArchitecture = "x86_64"
	AARCH64 ImageArchitecture = "aarch64"
)

var ImageArchitectures = []ImageArchitecture{
	X86_64,
	AARCH64,
}

type ImageType string

const (
	StandardImage ImageType = "standard"
	OSvImage      ImageType = "osv"
	QemuImage     ImageType = "qemu"
)

type ImageFsType string

const (
	Zfs   ImageFsType = "zfs"
	Ext4  ImageFsType = "ext4"
	ExFat ImageFsType = "exfat"
	Ntfs  ImageFsType = "ntfs"
	Fat32 ImageFsType = "fat32"
)

type ImageSourceType string

const (
	PromethiumImage ImageSourceType = "promethium"
	DockerImage     ImageSourceType = "docker"
	TarImage        ImageSourceType = "tar"
	RawImage        ImageSourceType = "raw"
	Qcow2Image      ImageSourceType = "qcow2"
	CapstanImage    ImageSourceType = "capstan"
)

type ImageBackendType uint8

const (
	QcowImageBackend    ImageBackendType = 0x00
	BlockImageBackend   ImageBackendType = 0x01
	UnknownImageBackend ImageBackendType = 0xff
)

//func NewImage() (*Image, error) {
////
////}

type Image struct {
	ID           string            `json:"id"`
	Name         string            `json:"name"`
	Version      string            `json:"version"`
	Type         ImageType         `json:"type"`
	Size         uint64            `json:"size"`
	Source       ImageSourceType   `json:"source"`
	SourceURI    string            `json:"sourceUri"`
	ImageHash    string            `json:"imageHash"`
	KernelHash   string            `json:"kernelHash"`
	Architecture ImageArchitecture `json:"architecture"`
	isPrk        bool
	backendType  ImageBackendType
	imagePath    string
	kernelPath   string
	bootParams   string
	prkPath      string
}

func determineImageBackend(path string) (ImageBackendType, error) {
	fi, err := os.Stat(path)
	if err != nil {
		return UnknownImageBackend, err
	}
	if fi.Mode()&os.ModeDevice != 0 {
		return BlockImageBackend, nil
	} else if fi.IsDir() {
		//doesnt support directories
		return UnknownImageBackend, errors.New("Directories are not supported for image backends. Only use block devices or Qcow2 images.")
	} else {
		//is a file... test if its a qcow...
		_, err := LoadQcowImage(path)
		if err != nil {
			return UnknownImageBackend, err
		}
		return QcowImageBackend, nil
	}

}

func NewImageFromQcow(name string, version string, imgType ImageType, size uint64, source ImageSourceType, sourceURI string, imagePath string, kernelPath string) (*Image, error) {
	//lets figure out what backend we are using for the image
	imgBackend, err := determineImageBackend(imagePath)
	if err != nil {
		return nil, err
	}

	//ok we know what the image backend is, but, we need to load it as required..

	imgId, _ := vutils.UUID.MakeUUIDString()
	im := &Image{
		ID:          imgId,
		Name:        name,
		Version:     version,
		Type:        imgType,
		backendType: imgBackend,
		Source:      source,
		SourceURI:   sourceURI,
		imagePath:   imagePath,
		kernelPath:  kernelPath,
	}
	hash, err := im.calculateImageHash()
	if err != nil {
		return nil, err
	}
	im.ImageHash = hash
	hash, err = im.calculateKernelHash()
	if err != nil {
		return nil, err
	}
	im.KernelHash = hash
	return im, nil
}

func prepareQcowImage(workspace string, size uint64) (*QcowImage, string, error) {
	if err := vutils.Files.CreateDirIfNotExist(workspace); err != nil {
		return nil, "", err
	} else {
		//create the meta file with all the details needed...
		imgPath := filepath.Join(workspace, "root.qcow2")
		img, err := CreateNewQcowImage(imgPath, size)
		if err != nil {
			return nil, "", err
		}
		err = img.Connect()
		if err != nil {
			return nil, "", err
		}
		err = img.MakeFilesystem("root", Ext4)
		if err != nil {
			return nil, "", err
		}
		mp, err := img.Mount()
		if err != nil {
			return nil, "", err
		}
		return img, mp, nil
	}
}

func BuildPackageFrom(imageRootPath string, outputPath string) (*Image, error) {
	//see if there is a Imageconf file..
	imgConf, err := loadImageConfFile(imageRootPath)
	if err != nil {
		return nil, err
	} else if imgConf.Type == "docker" {
		//this is a docker build..
		//make workdir
		tdir, err := ioutil.TempDir("", "prmbuild")
		if err != nil {
			return nil, err
		} else {
			defer os.RemoveAll(tdir)
			qcimg, mp, err := prepareQcowImage(tdir, uint64(imgConf.sizeBytes))
			if err != nil {
				return nil, err
			}
			err = imgConf.runDockerBuild(tdir, mp)
			if err != nil {
				return nil, err
			}
			//now we have build the docker image now package it...
			err = qcimg.Unmount()
			if err != nil {
				return nil, err
			}
			err = qcimg.Disconnect()
			if err != nil {
				return nil, err
			}
			kernelPath := filepath.Join(tdir, "kernel.elf")
			imagePath := filepath.Join(tdir, "root.qcow2")
			img, err := NewImageFromQcow(imgConf.OS, imgConf.Version, StandardImage, uint64(imgConf.sizeBytes), DockerImage, imgConf.Source, imagePath, kernelPath)
			if err != nil {
				return nil, err
			}
			img.Architecture = getCurrentSysArch()
			println("Building for architecture: " + string(img.Architecture))
			err = img.CreatePackage(tdir, outputPath)
			if err != nil {
				return nil, err
			}
			return img, nil
		}
	}
	return nil, nil
}

func getCurrentSysArch() ImageArchitecture {
	switch runtime.GOARCH {
	case "amd64":
		return X86_64
	case "arm64":
		return AARCH64
	}
	return X86_64
}

func LoadImageFromPrk(path string) (*Image, error) {
	//we need to get the prk file and get the metadata item..
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	defer file.Close()
	archive, err := gzip.NewReader(file)

	if err != nil {
		return nil, err
	}
	defer archive.Close()
	tr := tar.NewReader(archive)
	var img Image
	imageHash := ""
	kernelHash := ""

	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		} else if hdr.Name == "_meta" {
			//load the image metadata..
			bs, _ := ioutil.ReadAll(tr)
			err = json.Unmarshal(bs, &img)
			if err != nil {
				return nil, err
			}
		} else if hdr.Name == "root.qcow2" {
			//calculate hash for image...
			if hash, err := CalculateHashForReader(tr); err != nil {
				return nil, err
			} else {
				imageHash = hash
			}

		} else if hdr.Name == "kernel.elf" {
			//calculate hash for image...
			if hash, err := CalculateHashForReader(tr); err != nil {
				return nil, err
			} else {
				kernelHash = hash
			}

		}

	}

	if imageHash == "" {
		return nil, errors.New("Unable to calculate hash for embedded image file")
	} else if img.ImageHash == "" {
		return nil, errors.New("The image metadata information doesnt contain an image hash")
	} else if img.ImageHash != imageHash {
		return nil, errors.New("The image metadata image hash doesnt match the calculated hash for the embedded image")
	} else if kernelHash == "" {
		return nil, errors.New("Unable to calculate hash for embedded kernel file")
	} else if img.KernelHash == "" {
		return nil, errors.New("The image metadata information doesnt contain a kernel hash")
	} else if img.KernelHash != kernelHash {
		return nil, errors.New("The image metadata kernel hash doesnt match the calculated hash for the embedded kernel")
	}

	img.imagePath = path + "#root.qcow2"
	img.kernelPath = path + "#kernel.elf"
	img.isPrk = true
	img.backendType = QcowImageBackend
	img.prkPath = path
	return &img, nil

}

var PrkImagePathRX = regexp.MustCompile("^(.+\\.prk)#root.qcow2$")

func (im *Image) CreatePackage(workspacePath string, outputRoot string) error {

	//need to establish if this is already an image file... we dont output packages for things that are already packages..

	if im.isPrk {
		return errors.New("Cannot create a package from a package (it would overwrite the package).")
	}

	vutils.Files.CreateDirIfNotExist(outputRoot)
	opath := filepath.Join(outputRoot, fmt.Sprintf("%s-%s.prk", im.Name, im.Version))
	os.Remove(opath)

	//create the image metadata file

	if err := im.writeMetadataForWorkspace(workspacePath); err != nil {
		return err
	}

	files, err := ioutil.ReadDir(workspacePath)
	if err != nil {
		return err
	}
	flist := make([]string, len(files))
	for index, fileInfo := range files {
		flist[index] = fileInfo.Name()
	}
	argsList := append([]string{"-czvf", opath}, flist...)
	buildCmd := vutils.Exec.CreateAsyncCommand("tar", false, argsList...)
	err = buildCmd.SetWorkingDir(workspacePath).BindToStdoutAndStdErr().StartAndWait()
	return nil
}

func (im *Image) writeMetadataForWorkspace(workspacePath string) error {
	metaPath := filepath.Join(workspacePath, "_meta")
	if encConf, err := json.Marshal(im); err != nil {

		return err

	} else if err := ioutil.WriteFile(metaPath, encConf, 0440); err != nil {

		return err

	}
	return nil
}

func (im *Image) calculateImageHash() (string, error) {
	f, err := os.Open(im.imagePath)
	if err != nil {
		return "", err
	}
	defer f.Close()
	return CalculateHashForReader(f)
}

func (im *Image) calculateKernelHash() (string, error) {
	f, err := os.Open(im.kernelPath)
	if err != nil {
		return "", err
	}
	defer f.Close()
	return CalculateHashForReader(f)
}

func CalculateHashForReader(src io.Reader) (string, error) {
	h := sha256.New()
	if _, err := io.Copy(h, src); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

var GitHubAddrRX = regexp.MustCompile("^github.com/(.*)$")

func establishPathForImageConf(inUri string) (string, string, error) {
	//here we need to see if we can get to the requred item...
	if len(inUri) == 0 {
		return "", "", errors.New("Unsupported Image Conf Path")
	} else if inUri[0] == '/' || inUri[0:1] == "./" {
		//is a directory
		fullPath := inUri
		if inUri[0] != '/' {
			cwd, _ := os.Getwd()
			fullPath = filepath.Join(cwd, inUri[2:])
		}
		imageConfPath := filepath.Join(fullPath, "Imageconf")
		if !vutils.Files.CheckPathExists(imageConfPath) {
			return "", "", errors.New("Unsupported Image Conf Path. File missing.")
		}
		return fullPath, imageConfPath, nil

	} else {
		//is potentially a url..
		//github etc..

		//build the github uri but only if it begins with github.com

		// matches := GitHubAddrRX.FindAllString(inUri, 1)

		// if len(matches) > 0 {
		// 	//lets figure out what we are going to get..
		// 	fullUri := "https://" + inUri
		// 	ctx, cancel := context.WithCancel(context.Background())
		// 	// Build the client
		// 	client := &getter.Client{
		// 		Ctx:     ctx,
		// 		Src:     fullUri,
		// 		Dst:     "output.",
		// 		Pwd:     "/tmp",
		// 		Mode:    mode,
		// 		Options: opts,
		// 	}
		// 	err = client.Get()
		// 	if err != nil {
		// 		return err
		// 	}
		// }

	}
	return "", "", errors.New("Unsupported Image Conf Path")
}

func loadImageConfFile(path string) (*ImageConfFile, error) {
	imgRoot, imgConfPath, err := establishPathForImageConf(path)
	if err != nil {
		return nil, err
	}
	t := ImageConfFile{}
	file, err := os.Open(imgConfPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	ba, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal(ba, &t)
	if err != nil {
		return nil, err
	}
	t.rootPath = imgRoot
	sb, err := gounits.FromHumanSize(t.Size)
	if err != nil {
		return nil, err
	}
	t.sizeBytes = sb
	return &t, nil
}

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
	"strings"
	"syscall"
	"time"

	"github.com/768bit/promethium/api/models"
	"github.com/768bit/promethium/lib/config"
	"github.com/768bit/vutils"
	gounits "github.com/docker/go-units"
	"github.com/go-openapi/strfmt"
	"gopkg.in/yaml.v2"
)

type ImageConfFile struct {
	OS        string         `yaml:"OS" json:"OS"`
	Version   string         `yaml:"Version" json:"Version"`
	Type      string         `yaml:"Type" json:"Type"`
	Class     config.VmmType `yaml:"Class" json:"Class"`
	Size      string         `yaml:"Size" json:"Size"`
	Source    string         `yaml:"Source" json:"Source"`
	sizeBytes int64
	rootPath  string
}

type ImageCacheFile struct {
	ID         string   `json:"id"`
	ImageHash  string   `json:"imageHash"`
	KernelHash string   `json:"kernelHash"`
	Files      []string `json:"files"`
	hash       string
}

func (icf *ImageCacheFile) AddFile(path string) {
	icf.Files = append(icf.Files, path)
	icf.calculateHash()
}

func (icf *ImageCacheFile) calculateHash() {
	h := sha256.New()
	h.Write([]byte(icf.ID))
	h.Write([]byte(icf.ImageHash))
	h.Write([]byte(icf.KernelHash))
	h.Write([]byte(strings.Join(icf.Files, ",")))

	icf.hash = fmt.Sprintf("%x", h.Sum(nil))
}

func (icf *ImageCacheFile) GetHash() string {
	if icf.hash == "" {
		icf.calculateHash()
	}
	return icf.hash
}

func (icf *ImageConfFile) runDockerBuild(workDir string, mountPoint string) error {
	//docker run --privileged --rm -v $WORK_DIR:/output -v $MOUNT_POINT:/rootfs ubuntu-firecracker
	//based ont he root path we need to do a build using a workspace that we cleanup
	//first we need to build the container...
	imagename := "prm-bootstrap-build-" + icf.OS + "-" + icf.Version
	cmd := vutils.Exec.CreateAsyncCommand("docker", false, "build", "--network", "host", "-t", imagename, ".").BindToStdoutAndStdErr().SetWorkingDir(icf.rootPath)
	if err := cmd.StartAndWait(); err != nil {
		return err
	} else {
		//we successfully built the image.. now execute the container...
		dockerCmd := vutils.Exec.CreateAsyncCommand("docker", false, "run", "--privileged", "--network", "host", "--rm", "-v", workDir+":/output", "-v", mountPoint+":/rootfs", imagename)
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
	models.Image
	createdAt   time.Time
	isPrk       bool
	backendType ImageBackendType
	imagePath   string
	kernelPath  string
	prkPath     string
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

func NewImageFromQcow(name string, version string, vmmType config.VmmType, size uint64, source ImageSourceType, sourceURI string, imagePath string, kernelPath string) (*Image, error) {
	//lets figure out what backend we are using for the image
	imgBackend, err := determineImageBackend(imagePath)
	if err != nil {
		return nil, err
	}

	//ok we know what the image backend is, but, we need to load it as required..

	imgId, _ := vutils.UUID.MakeUUIDString()
	im := &Image{
		Image: models.Image{
			ID:        imgId,
			Name:      name,
			Version:   version,
			Type:      string(vmmType),
			Source:    string(source),
			SourceURI: sourceURI,
		},
		backendType: imgBackend,
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
			img, err := NewImageFromQcow(imgConf.OS, imgConf.Version, config.FirecrackerVmm, uint64(imgConf.sizeBytes), DockerImage, imgConf.Source, imagePath, kernelPath)
			if err != nil {
				return nil, err
			}
			img.Image.Architecture = string(getCurrentSysArch())
			println("Building for architecture: " + string(img.Architecture))
			prkPath, err := img.CreatePackage(tdir, outputPath)
			if err != nil {
				return nil, err
			}
			return LoadImageFromPrk(prkPath, nil)
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

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func LoadImageFromPrk(path string, imagesCache map[string]*ImageCacheFile) (*Image, error) {
	//we need to get the prk file and get the metadata item..
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	archive, err := gzip.NewReader(file)

	if err != nil {
		file.Close()
		return nil, err
	}
	tr := tar.NewReader(archive)
	var img Image
	imageHash := ""
	kernelHash := ""
	calcHash := false
	makeCacheEntry := false
	addCacheFile := false
	var currentCacheEntry *ImageCacheFile
	count := 0
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
				archive.Close()
				file.Close()
				return nil, err
			}
			if imagesCache != nil {
				if v, ok := imagesCache[img.ID]; !ok || v == nil {
					println("no cache entry for " + img.ID)
					makeCacheEntry = true
					calcHash = true
				} else if !contains(v.Files, path) {
					println("no cache file entry for " + img.ID)
					currentCacheEntry = v
					calcHash = true
					addCacheFile = true
				} else {
					println("have cache entry for " + img.ID)
					currentCacheEntry = v
				}
			} else {
				calcHash = true
			}
			count++
			if count >= 2 {
				break
			}
		} else if hdr.Name == "boot" {
			//load the image metadata..
			bs, _ := ioutil.ReadAll(tr)
			img.BootParams = string(bs)
			count++
			if count >= 2 {
				break
			}
		}
	}
	fi, err := file.Stat()
	if err != nil {
		return nil, err
	}
	stat := fi.Sys().(*syscall.Stat_t)
	img.createdAt = time.Unix(int64(stat.Ctim.Sec), int64(stat.Ctim.Nsec))

	pd, _ := strfmt.ParseDateTime(img.createdAt.Format("2006-01-02T15:04:05"))
	//println(pd.String() + " - " + img.createdAt.Format("2006-01-02T15:04:05"))
	img.Image.CreatedAt = pd

	archive.Close()
	file.Close()

	if calcHash {
		println("Calculating hash for " + img.ID)
		file, err = os.Open(path)
		if err != nil {
			return nil, err
		}
		defer file.Close()
		archive, err = gzip.NewReader(file)
		if err != nil {
			return nil, err
		}
		defer archive.Close()
		tr = tar.NewReader(archive)

		for {
			hdr, err := tr.Next()
			if err == io.EOF {
				break
			} else if err != nil {
				return nil, err
			} else if calcHash && hdr.Name == "root.qcow2" {
				//calculate hash for image...
				if hash, err := CalculateHashForReader(tr); err != nil {
					return nil, err
				} else {
					imageHash = hash
				}

			} else if calcHash && hdr.Name == "kernel.elf" {
				//calculate hash for image...
				if hash, err := CalculateHashForReader(tr); err != nil {
					return nil, err
				} else {
					kernelHash = hash
				}

			}

		}
	}

	if currentCacheEntry != nil && !calcHash {
		if currentCacheEntry.ImageHash != img.ImageHash {
			return nil, errors.New("The cached image hash doesnt match the hash for the embedded image")
		} else if currentCacheEntry.KernelHash != img.KernelHash {
			return nil, errors.New("The cached kernel hash doesnt match the hash for the embedded kernel")
		}
	} else {

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

		//now check cache entry

		if currentCacheEntry != nil {
			if currentCacheEntry.ImageHash != imageHash {
				return nil, errors.New("The cached image hash doesnt match the calculated hash for the embedded image")
			} else if currentCacheEntry.KernelHash != kernelHash {
				return nil, errors.New("The cached kernel hash doesnt match the calculated hash for the embedded kernel")
			}
		}

	}

	if makeCacheEntry {
		imagesCache[img.ID] = &ImageCacheFile{
			ID:         img.ID,
			ImageHash:  img.ImageHash,
			KernelHash: img.KernelHash,
			Files:      []string{path},
		}
		imagesCache[img.ID].calculateHash()
	} else if addCacheFile {
		imagesCache[img.ID].AddFile(path)
	}

	img.imagePath = path + "#root.qcow2"
	img.kernelPath = path + "#kernel.elf"
	img.isPrk = true
	img.backendType = QcowImageBackend
	img.prkPath = path
	return &img, nil

}

var PrkImagePathRX = regexp.MustCompile("^(.+\\.prk)#root.qcow2$")

func (im *Image) CreatePackage(workspacePath string, outputRoot string) (string, error) {

	//need to establish if this is already an image file... we dont output packages for things that are already packages..

	opath := filepath.Join(outputRoot, fmt.Sprintf("%s-%s.prk", im.Name, im.Version))
	if im.isPrk && im.prkPath == opath { //allow the clone and rebuild process...
		return "", errors.New("Cannot create a package from a package (it would overwrite the package).")
	}

	vutils.Files.CreateDirIfNotExist(outputRoot)
	os.Remove(opath)

	//create the image metadata file

	if err := im.writeMetadataForWorkspace(workspacePath); err != nil {
		return opath, err
	}

	files, err := ioutil.ReadDir(workspacePath)
	if err != nil {
		return opath, err
	}
	flist := make([]string, len(files))
	for index, fileInfo := range files {
		flist[index] = fileInfo.Name()
	}
	argsList := append([]string{"-czvf", opath}, flist...)
	buildCmd := vutils.Exec.CreateAsyncCommand("tar", false, argsList...)
	err = buildCmd.SetWorkingDir(workspacePath).BindToStdoutAndStdErr().StartAndWait()
	return opath, err
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

func (im *Image) CloneRootDiskToPath(id string, outPath string, size uint64) (string, string, error) {

	imgOutPath := filepath.Join(outPath, id+".qcow2")
	kernelOutPath := filepath.Join(outPath, id+".elf")
	if im.isPrk {
		file, err := os.Open(im.prkPath)
		if err != nil {
			return "", "", err
		}
		defer file.Close()
		archive, err := gzip.NewReader(file)
		if err != nil {
			return "", "", err
		}
		defer archive.Close()
		tr := tar.NewReader(archive)

		for {
			hdr, err := tr.Next()
			if err == io.EOF {
				break
			} else if err != nil {
				return "", "", err
			} else if hdr.Name == "root.qcow2" {
				ofile, err := os.Create(imgOutPath)
				if err != nil {
					return "", "", err
				}

				_, err = io.Copy(ofile, tr)
				ofile.Close()
				if err != nil {
					return "", "", err
				}

				//now we have the file lets do the expansion as needed

				err = doImageResize(imgOutPath, size)
				if err != nil {
					return "", "", err
				}

			} else if hdr.Name == "kernel.elf" {
				ofile, err := os.Create(kernelOutPath)
				if err != nil {
					return "", "", err
				}

				_, err = io.Copy(ofile, tr)
				ofile.Close()
				if err != nil {
					return "", "", err
				}

			}

		}
	} else {
		//just copy the images...
		if err := vutils.Files.Copy(im.imagePath, imgOutPath); err != nil {
			return "", "", err
		} else if err := vutils.Files.Copy(im.kernelPath, kernelOutPath); err != nil {
			return "", "", err
		}
		err := doImageResize(imgOutPath, size)
		if err != nil {
			return "", "", err
		}

	}
	return imgOutPath, kernelOutPath, nil

}

func doImageResize(path string, size uint64) error {
	qcimg, err := LoadQcowImage(path)
	if err != nil {
		return err
	}

	if size < qcimg.VirtualSize() {
		return errors.New("Unable to resize the image as it will be smaller than the original")
	} else if size > qcimg.VirtualSize() {
		//resize the image...
		err = qcimg.Resize(size)
		if err != nil {
			return err
		}
		err = qcimg.Connect()
		if err != nil {
			return err
		}
		// err = qcimg.Mount()
		// if err != nil {
		//   qcimg.Disconnect()
		//   return "", err
		// }
		// defer qcimg.Unmount()
		defer qcimg.Disconnect()
		part, err := qcimg.GetPartition(1)
		if err != nil {
			println(err.Error())
			if err := vutils.Exec.CreateAsyncCommand("e2fsck", false, "-f", "-p", qcimg.connectedDevice).Sudo().BindToStdoutAndStdErr().StartAndWait(); err != nil {
				return err
			}
			return qcimg.GrowFullPart()
		} else {
			err = part.GrowPart()
			if err != nil {
				return err
			}
		}
	}
	return nil
}

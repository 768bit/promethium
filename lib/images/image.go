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
	"github.com/768bit/promethium/lib/common"
	"github.com/768bit/promethium/lib/config"
	"github.com/768bit/vutils"
	gounits "github.com/docker/go-units"
	"github.com/go-openapi/strfmt"
	"gopkg.in/yaml.v2"
)

type ImageConfFile struct {
	Name       string         `yaml:"Name" json:"Name"`
	Version    string         `yaml:"Version" json:"Version"`
	Type       string         `yaml:"Type" json:"Type"`
	Class      config.VmmType `yaml:"Class" json:"Class"`
	Size       string         `yaml:"Size" json:"Size"`
	Source     string         `yaml:"Source" json:"Source"`
	KernelOnly bool           `yaml:"KernelOnly" json:"KernelOnly"`
	sizeBytes  int64
	rootPath   string
}

func (icf *ImageConfFile) runDockerBuild(workDir string, mountPoint string) error {
	//docker run --privileged --rm -v $WORK_DIR:/output -v $MOUNT_POINT:/rootfs ubuntu-firecracker
	//based ont he root path we need to do a build using a workspace that we cleanup
	//first we need to build the container...
	imagename := "prm-bootstrap-build-" + icf.Name + "-" + icf.Version
	cmd := vutils.Exec.CreateAsyncCommand("docker", false, "build", "--network", "host", "-t", imagename, ".").BindToStdoutAndStdErr().SetWorkingDir(icf.rootPath)
	if err := cmd.StartAndWait(); err != nil {
		return err
	} else {
		//we successfully built the image.. now execute the container...
		dockerCmd := vutils.Exec.CreateAsyncCommand("docker", false, "run", "--privileged", "--network", "host", "--rm", "-v", workDir+":/output", "-v", mountPoint+":/rootfs", imagename)
		return dockerCmd.BindToStdoutAndStdErr().SetWorkingDir(workDir).StartAndWait()
	}
}

type ImageCacheFile struct {
	ID              string   `json:"id"`
	RootDiskHash    string   `json:"rootDiskHash"`
	KernelHash      string   `json:"kernelHash"`
	OtherDiskHashes []string `json:"otherDiskHashes"`
	Files           []string `json:"files"`
	hash            string
}

func (icf *ImageCacheFile) AddFile(path string) {
	icf.Files = append(icf.Files, path)
	icf.calculateHash()
}

func (icf *ImageCacheFile) calculateHash() {
	h := sha256.New()
	h.Write([]byte(icf.ID))
	h.Write([]byte(icf.RootDiskHash))
	h.Write([]byte(icf.KernelHash))
	h.Write([]byte(strings.Join(icf.OtherDiskHashes, ",")))
	h.Write([]byte(strings.Join(icf.Files, ",")))

	icf.hash = fmt.Sprintf("%x", h.Sum(nil))
}

func (icf *ImageCacheFile) GetHash() string {
	if icf.hash == "" {
		icf.calculateHash()
	}
	return icf.hash
}

type Image struct {
	models.Image
	createdAt     time.Time
	isPrk         bool
	backendType   common.ImageBackendType
	diskPath      string
	kernelPath    string
	prkPath       string
	otherDisks    []string
	contains      common.ImageContainsBits
	cloudInitPath string
}

func determineImageBackend(path string) (common.ImageBackendType, error) {
	fi, err := os.Stat(path)
	if err != nil {
		return common.UnknownImageBackend, err
	}
	if fi.Mode()&os.ModeDevice != 0 {
		return common.BlockImageBackend, nil
	} else if fi.IsDir() {
		//doesnt support directories
		return common.UnknownImageBackend, errors.New("Directories are not supported for image backends. Only use block devices or Qcow2 images.")
	} else {
		//is a file... test if its a qcow...
		_, err := LoadQemuImage(path)
		if err != nil {
			return common.UnknownImageBackend, err
		}
		return common.QcowImageBackend, nil
	}

}

func NewKernelOnlyImage(name string, version string, sourceURI string, kernelPath string) (*Image, error) {
	//lets figure out what backend we are using for the image

	//ok we know what the image backend is, but, we need to load it as required..

	imgId, _ := vutils.UUID.MakeUUIDString()
	im := &Image{
		Image: models.Image{
			ID:                imgId,
			Name:              name,
			Version:           version,
			Type:              "kernel",
			Source:            "kernel",
			SourceURI:         sourceURI,
			Kernel:            &models.KernelImage{},
			RootDisk:          nil,
			OtherDisks:        []*models.DiskImage{},
			CloudInitUserData: nil,
			Contains:          []models.ImageContains{models.ImageContainsKernel},
		},
		backendType: common.UnknownImageBackend,
		kernelPath:  kernelPath,
		otherDisks:  []string{},
	}
	im.setFlag(common.ImageHasKernel)
	hash, err := im.calculateKernelHash()
	if err != nil {
		return nil, err
	}
	im.Kernel.Hash = hash
	return im, nil
}

func NewImageFromQcow(name string, version string, vmmType config.VmmType, size uint64, source common.ImageSourceType, sourceURI string, diskPath string) (*Image, error) {
	//lets figure out what backend we are using for the image
	imgBackend, err := determineImageBackend(diskPath)
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
			Kernel:    nil,
			RootDisk: &models.DiskImage{
				IsRoot: true,
				Size:   int64(size),
			},
			OtherDisks:        []*models.DiskImage{},
			CloudInitUserData: nil,
			Contains:          []models.ImageContains{models.ImageContainsRootDisk},
		},
		backendType: imgBackend,
		diskPath:    diskPath,
		otherDisks:  []string{},
	}
	im.setFlag(common.ImageHasDisk)
	hash, err := im.calculateImageHash()
	if err != nil {
		return nil, err
	}
	im.RootDisk.Hash = hash
	return im, nil
}

func prepareQcowImage(workspace string, size uint64) (*QemuImage, string, error) {
	if err := vutils.Files.CreateDirIfNotExist(workspace); err != nil {
		return nil, "", err
	} else {
		//create the meta file with all the details needed...
		imgPath := filepath.Join(workspace, "root.qcow2")
		img, err := CreateNewQemuImage(imgPath, size)
		if err != nil {
			return nil, "", err
		}
		err = img.Connect()
		if err != nil {
			return nil, "", err
		}
		err = img.MakeFilesystem("root", common.Ext4)
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
	} else {
		//this is a docker build..
		//make workdir
		tdir, err := ioutil.TempDir("", "prmbuild")
		if err != nil {
			return nil, err
		} else {
			defer os.RemoveAll(tdir)
			if imgConf.KernelOnly {
				//we are only doing a kernel build...
				err = imgConf.runDockerBuild(tdir, "/dev/null")
				if err != nil {
					return nil, err
				}
				kernelMetaPath := filepath.Join(tdir, "_kernel_meta.json")
				kernelPath := filepath.Join(tdir, "kernel.elf")
				if vutils.Files.CheckPathExists(kernelPath) && vutils.Files.CheckPathExists(kernelMetaPath) {
					//load kernel meta...

					km, err := common.LoadKernelMeta(kernelMetaPath)
					if err != nil {
						return nil, err
					}

					//now load the kernel and ensure it is valid..

					_, err = common.LoadKernelElf(kernelPath)
					if err != nil {
						return nil, err
					}

					img, err := NewKernelOnlyImage(imgConf.Name, km.Version, km.From, kernelPath)
					if err != nil {
						return nil, err
					}
					img.Image.Architecture = string(km.Machine)
					println("Building for architecture: " + string(img.Architecture))
					prkPath, err := img.CreatePackage(tdir, outputPath)
					if err != nil {
						return nil, err
					}
					return LoadImageFromPrk(prkPath, nil)
				} else {
					return nil, errors.New("Unable to load kernel image or kernel meta data")
				}
			}
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
			kernelMetaPath := filepath.Join(tdir, "_kernel_meta.json")
			kernelPath := filepath.Join(tdir, "kernel.elf")
			rootPath := filepath.Join(tdir, "root.qcow2")
			img, err := NewImageFromQcow(imgConf.Name, imgConf.Version, config.FirecrackerVmm, uint64(imgConf.sizeBytes), common.DockerImage, imgConf.Source, rootPath)
			if err != nil {
				return nil, err
			}
			///try and load the kernel, does it exist, if not dont include in the package
			//there are a couple of meta files created so lets read them and they can indicate what needs to be loaded...
			if vutils.Files.CheckPathExists(kernelPath) && vutils.Files.CheckPathExists(kernelMetaPath) {
				//load kernel meta...
				km, err := common.LoadKernelMeta(kernelMetaPath)
				if err != nil {
					return nil, err
				}

				_, err = common.LoadKernelElf(kernelPath)
				if err != nil {
					return nil, err
				}

				img.Image.Architecture = string(km.Machine)
				println("Building for architecture: " + string(img.Architecture))
				img.setFlag(common.ImageHasKernel)
				//now add the kernel - only if it is valid...
			} else {
				img.Image.Architecture = string(getCurrentSysArch())
				println("Building for architecture: " + string(img.Architecture))
			}
			prkPath, err := img.CreatePackage(tdir, outputPath)
			if err != nil {
				return nil, err
			}
			return LoadImageFromPrk(prkPath, nil)
		}
	}
	return nil, nil
}

func getCurrentSysArch() common.ImageArchitecture {
	switch runtime.GOARCH {
	case "amd64":
		return common.X86_64
	case "arm64":
		return common.AARCH64
	}
	return common.X86_64
}

func img_contains(s []models.ImageContains, e models.ImageContains) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

var ADDITIONAL_DISK_NAME_RX = regexp.MustCompile(`^disk-(\d+)\.`)

type ImageHashMap struct {
	Disk       string
	Kernel     string
	CloudInit  string
	OtherDisks []string
}

func GetMetaFromPrk(path string) (*Image, error) {
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

	//set image flags based on cotnains field...
	if img_contains(img.Contains, models.ImageContainsRootDisk) {
		img.setFlag(common.ImageHasDisk)
	}
	if img_contains(img.Contains, models.ImageContainsKernel) {
		img.setFlag(common.ImageHasKernel)
	}
	if img_contains(img.Contains, models.ImageContainsAdditionalDisks) {
		img.setFlag(common.ImageHasAdditionalDisk)
	}
	if img_contains(img.Contains, models.ImageContainsCloudInitUserData) {
		img.setFlag(common.ImageHasCloudInit)
	}

	archive.Close()
	file.Close()

	img.isPrk = true
	img.backendType = common.QcowImageBackend
	img.prkPath = path
	return &img, nil

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

	hm := &ImageHashMap{
		OtherDisks: []string{},
	}

	//set image flags based on cotnains field...
	if img_contains(img.Contains, models.ImageContainsRootDisk) {
		img.setFlag(common.ImageHasDisk)
	}
	if img_contains(img.Contains, models.ImageContainsKernel) {
		img.setFlag(common.ImageHasKernel)
	}
	if img_contains(img.Contains, models.ImageContainsAdditionalDisks) {
		img.setFlag(common.ImageHasAdditionalDisk)
	}
	if img_contains(img.Contains, models.ImageContainsCloudInitUserData) {
		img.setFlag(common.ImageHasCloudInit)
	}

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
			} else if calcHash && img.HasRootDisk() && hdr.Name == "root.qcow2" {
				//calculate hash for image...
				if hash, err := CalculateHashForReader(tr); err != nil {
					return nil, err
				} else {
					hm.Disk = hash
					img.diskPath = path + "#root.qcow2"
					img.RootDisk = &models.DiskImage{
						IsRoot: true,
						Hash:   hash,
					}
				}

			} else if calcHash && img.HasKernel() && hdr.Name == "kernel.elf" {
				//calculate hash for image...
				if hash, err := CalculateHashForReader(tr); err != nil {
					return nil, err
				} else {
					hm.Kernel = hash
					img.kernelPath = path + "#kernel.elf"
					img.Kernel = &models.KernelImage{
						Hash: hash,
					}
				}

			}

		}
	}

	println("Ready...")

	if currentCacheEntry != nil && !calcHash {
		if img.RootDisk != nil && currentCacheEntry.RootDiskHash != img.RootDisk.Hash {
			return nil, errors.New("The cached image hash doesnt match the hash for the embedded image")
		}
		if img.Kernel != nil && currentCacheEntry.KernelHash != img.Kernel.Hash {
			return nil, errors.New("The cached kernel hash doesnt match the hash for the embedded kernel")
		}
	} else {

		//only if the image contains a certain target do we verify

		if img.HasRootDisk() {
			println("Processing Disk...")
			if hm.Disk == "" {
				return nil, errors.New("Unable to calculate hash for embedded image file")
			} else if img.RootDisk.Hash == "" {
				return nil, errors.New("The image metadata information doesnt contain an image hash")
			} else if img.RootDisk.Hash != hm.Disk {
				return nil, errors.New("The image metadata image hash doesnt match the calculated hash for the embedded image")
			}
		}

		if img.HasKernel() {
			println("Processing Kernel...")
			if hm.Kernel == "" {
				return nil, errors.New("Unable to calculate hash for embedded kernel file")
			} else if img.Kernel.Hash == "" {
				return nil, errors.New("The image metadata information doesnt contain a kernel hash")
			} else if img.Kernel.Hash != hm.Kernel {
				return nil, errors.New("The image metadata kernel hash doesnt match the calculated hash for the embedded kernel")
			}
		}

		//now check cache entry

		if currentCacheEntry != nil {
			if currentCacheEntry.RootDiskHash != hm.Disk {
				return nil, errors.New("The cached image hash doesnt match the calculated hash for the embedded image")
			} else if currentCacheEntry.KernelHash != hm.Kernel {
				return nil, errors.New("The cached kernel hash doesnt match the calculated hash for the embedded kernel")
			}
		}

	}

	println("Ready...")
	if imagesCache != nil {
		if makeCacheEntry {
			imagesCache[img.ID] = &ImageCacheFile{
				ID:    img.ID,
				Files: []string{},
			}
			if img.HasKernel() {
				imagesCache[img.ID].KernelHash = img.Kernel.Hash
			}
			if img.HasRootDisk() {
				imagesCache[img.ID].RootDiskHash = img.RootDisk.Hash
			}
			imagesCache[img.ID].calculateHash()
		} else if addCacheFile {
			imagesCache[img.ID].AddFile(path)
		}
	}

	img.isPrk = true
	img.backendType = common.QcowImageBackend
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

	//before we write the meta data we need to poulate the "contains" property

	im.Contains = []models.ImageContains{}
	if im.HasKernel() {
		im.Contains = append(im.Contains, models.ImageContainsKernel)
	}
	if im.HasRootDisk() {
		im.Contains = append(im.Contains, models.ImageContainsRootDisk)
	}
	if im.HasAdditionalDisks() {
		im.Contains = append(im.Contains, models.ImageContainsAdditionalDisks)
	}
	if im.HasCloudInit() {
		im.Contains = append(im.Contains, models.ImageContainsCloudInitUserData)
	}

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

func (im *Image) HasRootDisk() bool {
	return im.hasFlag(common.ImageHasDisk)
}

func (im *Image) HasAdditionalDisks() bool {
	return im.hasFlag(common.ImageHasAdditionalDisk)
}

func (im *Image) HasKernel() bool {
	return im.hasFlag(common.ImageHasKernel)
}

func (im *Image) HasCloudInit() bool {
	return im.hasFlag(common.ImageHasCloudInit)
}

func (im *Image) GetID() string {
	return im.ID
}

func (im *Image) GetType() string {
	return im.Type
}

func (im *Image) GetBootParams() string {
	return im.BootParams
}

func (im *Image) AddCloudInitUserData(cloudInitPath string) error {
	//add the kernel image into the image and set the flag!
	if im.hasFlag(common.ImageHasCloudInit) {
		return errors.New("This image already has a cloud init config, cannot add a new one")
	}
	//load the cloud init user data config at path...

	ba, err := ioutil.ReadFile(cloudInitPath)
	if err != nil {
		return err
	}

	var ci *models.CloudInitUserData

	err = yaml.Unmarshal(ba, ci)
	if err != nil {
		return err
	}

	im.cloudInitPath = cloudInitPath

	im.setFlag(common.ImageHasCloudInit)
	return nil
}

func (im *Image) AddKernel(kernelPath string) error {
	//add the kernel image into the image and set the flag!
	if im.hasFlag(common.ImageHasKernel) {
		return errors.New("This image already has a kernel, cannot add a new one")
	}
	kernStat, err := os.Stat(kernelPath)
	if err != nil {
		return err
	}
	im.kernelPath = kernelPath
	hash, err := im.calculateKernelHash()
	if err != nil {
		return err
	}
	im.Kernel = &models.KernelImage{
		Hash: hash,
		Size: kernStat.Size(),
	}
	im.setFlag(common.ImageHasKernel)
	return nil
}

func (im *Image) AddDisk(diskPath string) error {
	//add the disk drive into the image and set the flag!
	//if we already have a root disk we add this disk as an additional item...
	if im.HasRootDisk() {
		//add the item to the array if it doesnt exist
		if !contains(im.otherDisks, diskPath) {
			//doesnt exist add to the list but only after we have interrogated the image...
			qcimg, err := LoadQemuImage(diskPath)
			if err != nil {
				return err
			}
			f, err := os.Open(diskPath)
			if err != nil {
				return err
			}
			defer f.Close()
			//now calculate hash for the assigned path...
			hash, err := CalculateHashForReader(f)
			im.OtherDisks = append(im.OtherDisks, &models.DiskImage{
				Size:   int64(qcimg.VirtualSize()),
				IsRoot: false,
				Hash:   hash,
			})
			im.otherDisks = append(im.otherDisks, diskPath)
			im.setFlag(common.ImageHasAdditionalDisk)
			return nil
		} else {
			return errors.New("Disk already added to image")
		}
	} else {
		qcimg, err := LoadQemuImage(diskPath)
		if err != nil {
			return err
		}
		f, err := os.Open(diskPath)
		if err != nil {
			return err
		}
		defer f.Close()
		//now calculate hash for the assigned path...
		hash, err := CalculateHashForReader(f)
		im.RootDisk = &models.DiskImage{
			IsRoot: true,
			Hash:   hash,
			Size:   int64(qcimg.VirtualSize()),
		}
		im.setFlag(common.ImageHasDisk)
		return nil
	}

}

func (im *Image) GetRootDiskReader() (*os.File, io.Reader, error) {
	if im.HasRootDisk() {
		file, err := os.Open(im.prkPath)
		if err != nil {
			return nil, nil, err
		}
		//defer file.Close()
		archive, err := gzip.NewReader(file)
		if err != nil {
			return file, nil, err
		}
		//defer archive.Close()
		tr := tar.NewReader(archive)

		for {
			hdr, err := tr.Next()
			if err == io.EOF {
				break
			} else if err != nil {
				return file, nil, err
			} else if hdr.Name == "root.qcow2" {
				return file, tr, nil
			}

		}
	}
	return nil, nil, errors.New("This image doesnt contain a root disk")
}

func (im *Image) GetKernelReader() (*os.File, io.Reader, error) {
	if im.HasKernel() {
		file, err := os.Open(im.prkPath)
		if err != nil {
			return nil, nil, err
		}
		//defer file.Close()
		archive, err := gzip.NewReader(file)
		if err != nil {
			return file, nil, err
		}
		//defer archive.Close()
		tr := tar.NewReader(archive)

		for {
			hdr, err := tr.Next()
			if err == io.EOF {
				break
			} else if err != nil {
				return file, nil, err
			} else if hdr.Name == "kernel.elf" {
				return file, tr, nil
			}

		}
	}
	return nil, nil, errors.New("This image doesnt contain a kernel disk")
}

func (im *Image) GetCloudInitReader() (*os.File, io.Reader, error) {
	return nil, nil, nil

}

func (im *Image) GetAdditionalDiskReader(index int) (*os.File, io.Reader, error) {
	return nil, nil, nil

}

// func (im *Image) GetBootParams() (string, error) {
// 	file, err := os.Open(im.prkPath)
// 	if err != nil {
// 		return "", err
// 	}
// 	defer file.Close()
// 	archive, err := gzip.NewReader(file)
// 	if err != nil {
// 		return "", err
// 	}
// 	defer archive.Close()
// 	tr := tar.NewReader(archive)

// 	for {
// 		hdr, err := tr.Next()
// 		if err == io.EOF {
// 			break
// 		} else if err != nil {
// 			return "", err
// 		} else if hdr.Name == "boot" {
// 			//now lets read everything now...
// 			ba, err := ioutil.ReadAll(tr)
// 			if err != nil {
// 				return "", err
// 			}
// 			return string(ba), nil
// 		}

// 	}
// 	return "", errors.New("Unable to get boot params file")

// }

func (im *Image) setFlag(flag common.ImageContainsBits) {
	im.contains = flag | im.contains
}

func (im *Image) clearFlag(flag common.ImageContainsBits) {
	im.contains = flag &^ im.contains
}

func (im *Image) hasFlag(flag common.ImageContainsBits) bool {
	return im.contains&flag != 0
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
	f, err := os.Open(im.diskPath)
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
	//if this is a kernel only image we need to deal with that
	if t.KernelOnly {
		return &t, nil
	}
	sb, err := gounits.FromHumanSize(t.Size)
	if err != nil {
		return nil, err
	}
	t.sizeBytes = sb
	return &t, nil
}

func (im *Image) ExtractRootDiskToPath(id string, outPath string, size uint64) (string, string, error) {

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

				//now we have the qcow - does it need to be expaned to raw?

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

		//need different strategies for different target environments - qcow2 vs raw image...

		//just copy the images...
		if err := vutils.Files.Copy(im.diskPath, imgOutPath); err != nil {
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

func (im *Image) ExtractKernelToPath(id string, outPath string) (string, string, error) {
	return "", "", nil
}

func doImageResize(path string, size uint64) error {
	qcimg, err := LoadQemuImage(path)
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

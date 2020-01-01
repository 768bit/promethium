package images

import (
	"path/filepath"

	"github.com/768bit/promethium/lib/common"
	"github.com/768bit/vutils"
)

//management of images for firecracker...

type KernelImageMap map[common.ImageArchitecture]map[string]*Image

type ImagesManager struct {
	rootPath    string
	capstanRoot string
	vmsRoot     string
	kernelsRoot string
	kernels     KernelImageMap
}

func NewImageManager(root string) *ImagesManager {
	im := &ImagesManager{
		rootPath:    root,
		capstanRoot: filepath.Join(root, "capstan"),
		vmsRoot:     filepath.Join(root, "vms"),
		kernelsRoot: filepath.Join(root, "kernels"),
		kernels:     KernelImageMap{},
	}
	im.init()
	return im
}

func (im *ImagesManager) init() {
	//ensure directories exist...
	vutils.Files.CreateDirIfNotExist(im.rootPath)
	vutils.Files.CreateDirIfNotExist(im.capstanRoot)
	vutils.Files.CreateDirIfNotExist(im.vmsRoot)
	vutils.Files.CreateDirIfNotExist(im.kernelsRoot)
	//scan the root dir for all images and make them available...

}

func (im *ImagesManager) scanVmImages() {
	//scan the root dir for all images and make them available...

}

func (im *ImagesManager) scanCapstanImages() {
	//scan the root dir for all images and make them available...

}

func (im *ImagesManager) scanKernelImages() {
	//scan the root dir for all images and make them available...
	for _, arch := range common.ImageArchitectures {
		kernelArchPath := filepath.Join(im.kernelsRoot, string(arch))
		vutils.Files.CreateDirIfNotExist(kernelArchPath)
		im.kernels[arch] = map[string]*Image{}
		im.scanKernelArchImages(kernelArchPath, im.kernels[arch])
	}
}

func (im *ImagesManager) scanKernelArchImages(kernelArchPath string, kernelArchMap map[string]*Image) {
	//scan the root dir for all images and make them available...

}

func (im *ImagesManager) Create() {
	//scan the root dir for all images and make them available...

}

func (im *ImagesManager) CreateFrom(source common.ImageSourceType, sourceURI string) {
	//scan the root dir for all images and make them available...
	switch source {
	case common.PromethiumImage:
		//lookup the current manager - is it available..
	case common.DockerImage:
		//attempt to obtain the image from docker...
		break
	case common.TarImage:
		//get a tar image from somewhere...
		break
	case common.RawImage:
		//use a raw disk image..
		break
	case common.Qcow2Image:
		break
	case common.CapstanImage:
		break

	}
}

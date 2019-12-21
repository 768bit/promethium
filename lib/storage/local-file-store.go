package storage

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/768bit/promethium/lib/images"
	"github.com/768bit/vutils"
)

//this represents a local data storage driver - it is installed as the default storage driver when the node is first deployed
//and uses the location /opt/promethium/storage/default-local

type LocalFileStorage struct {
	id            string
	sm            *StorageManager
	rootFolder    string
	imagesFolder  string
	imagesEnabled bool
	disksFolder   string
	disksEnabled  bool
	kernelsFolder string
	imagesCache   map[string]string
}

func getRootFolderFromConfig(config map[string]interface{}) (string, error) {
	if config == nil {
		return "", errors.New("Unable to get root folder from config")
	} else if v, ok := config["rootFolder"]; !ok || v == nil {
		return "", errors.New("Unable to get root folder from config")
	} else if vs, ok := config["rootFolder"].(string); !ok || vs == "" {
		return "", errors.New("Unable to get root folder from config")
	} else {
		return vs, nil
	}
}

func LoadLocalFileStorage(sm *StorageManager, id string, config map[string]interface{}) (*LocalFileStorage, error) {
	rootFolder, err := getRootFolderFromConfig(config)
	if err != nil {
		return nil, err
	}
	if !vutils.Files.CheckPathExists(rootFolder) {
		return nil, errors.New("The specified local file storage folder: " + rootFolder + " doesn't exist")
	}
	imagesFolder := filepath.Join(rootFolder, "images")
	disksFolder := filepath.Join(rootFolder, "disks")
	kernelsFolder := filepath.Join(rootFolder, "kernels")
	//create the instance...
	lfs := &LocalFileStorage{
		id:            id,
		sm:            sm,
		rootFolder:    rootFolder,
		disksFolder:   disksFolder,
		imagesFolder:  imagesFolder,
		imagesCache:   map[string]string{},
		kernelsFolder: kernelsFolder,
	}
	if !vutils.Files.CheckPathExists(disksFolder) {
		lfs.disksEnabled = false
	} else {
		lfs.disksEnabled = true
	}
	if !vutils.Files.CheckPathExists(imagesFolder) {
		lfs.imagesEnabled = false
	} else {
		lfs.imagesEnabled = true
	}
	return lfs, err
}

func InitLocalFileStorage(id string, config map[string]interface{}) error {
	rootFolder, err := getRootFolderFromConfig(config)
	if err != nil {
		return err
	}
	vutils.Files.CreateDirIfNotExist(rootFolder)
	imagesPath := filepath.Join(rootFolder, "images")
	vutils.Files.CreateDirIfNotExist(imagesPath)
	disksFolder := filepath.Join(rootFolder, "disks")
	vutils.Files.CreateDirIfNotExist(disksFolder)
	kernelsFolder := filepath.Join(rootFolder, "kernels")
	vutils.Files.CreateDirIfNotExist(kernelsFolder)
	return nil
}

func (lfs *LocalFileStorage) GetImages() ([]*images.Image, error) {
	//get images needs to get a list of files from the directory
	lfs.imagesCache = map[string]string{}
	if !lfs.imagesEnabled {
		return nil, errors.New("Unable to get images from this storage medium as it doesnt support images")
	}
	files := vutils.Files.GetFilesInDirWithExtension(lfs.imagesFolder, ".prk")
	imagesList := []*images.Image{}
	for _, file := range files {
		fullPath := filepath.Join(lfs.imagesFolder, file)
		img, err := images.LoadImageFromPrk(fullPath, lfs.sm.imagesCache)
		if err != nil {
			return nil, err
		}
		lfs.imagesCache[img.ID] = file[:len(file)-4]
		imagesList = append(imagesList, img)
	}
	return imagesList, nil
}

func (lfs *LocalFileStorage) GetURI() string {
	return "local-file://" + lfs.id
}

func (lfs *LocalFileStorage) GetImage(name string) (*images.Image, error) {
	imageFilePath := filepath.Join(lfs.imagesFolder, name+".prk")
	if !lfs.imagesEnabled {
		return nil, errors.New("Unable to get images from this storage medium as it doesnt support images")
	} else if !vutils.Files.CheckPathExists(imageFilePath) {
		return nil, errors.New("Unable to find image with name " + name)
	} else {
		img, err := images.LoadImageFromPrk(imageFilePath, lfs.sm.imagesCache)
		if err != nil {
			return nil, err
		}
		return img, nil
	}
}

func (lfs *LocalFileStorage) GetImageById(id string) (*images.Image, error) {

	if !lfs.imagesEnabled {
		return nil, errors.New("Unable to get images from this storage medium as it doesnt support images")
	} else if v, ok := lfs.imagesCache[id]; !ok || v == "" {
		_, err := lfs.GetImages()
		if err != nil {
			return nil, err
		}
		if v, ok := lfs.imagesCache[id]; !ok || v == "" {
			return nil, errors.New("Unable to find image with id " + id)
		} else {
			return lfs.GetImage(lfs.imagesCache[id])
		}
	} else {
		imageFilePath := filepath.Join(lfs.imagesFolder, lfs.imagesCache[id]+".prk")
		img, err := images.LoadImageFromPrk(imageFilePath, lfs.sm.imagesCache)
		if err != nil {
			return nil, err
		}
		return img, nil
	}
}

func (lfs *LocalFileStorage) CreateDiskFromImage(id string, img *images.Image, size uint64) (*VmmStorageDisk, *VmmKernel, error) {
	//first lets setup the image in a templ location and resize it...
	tdir, err := ioutil.TempDir("", "prmvm")
	if err != nil {
		return nil, nil, err
	} else {
		defer os.RemoveAll(tdir)
		imgPath, kernPath, err := img.CloneRootDiskToPath(id, tdir, size)
		if err != nil {
			return nil, nil, err
		}

		//ok we have the built images - we need to put them where they need to be... for local this is a simple copy
		newImagePath := filepath.Join(lfs.disksFolder, id, "root.qcow2")
		newKernelPath := filepath.Join(lfs.kernelsFolder, id+".elf")
		if err := vutils.Files.Copy(imgPath, newImagePath); err != nil {
			return nil, nil, err
		} else if err := vutils.Files.Copy(kernPath, newKernelPath); err != nil {
			return nil, nil, err
		}

		//everything copied ok.. lets instantiate the Disk...

		dsk, err := NewStorageDisk(id, "root", newImagePath, lfs)
		if err != nil {
			return nil, nil, err
		}

		kern := NewKernel(id, newKernelPath, lfs)
		return dsk, kern, nil
	}

}

func (lfs *LocalFileStorage) LookupPath(path string) (string, bool, error) {
	//path lookups are all about finding different classes of stuff
	if strings.HasPrefix(path, "/") {
		path = path[1:]
	}
	splitPath := strings.Split(path, "/")
	storageClass := splitPath[0]
	switch storageClass {
	case "kernels":
		if !lfs.disksEnabled {
			return "", false, errors.New("This storage target is not enabled for disk/kernel storage")
		}
		if len(splitPath) != 2 {
			return "", false, errors.New("The supplied path is invalid")
		}
		return filepath.Join(lfs.kernelsFolder, splitPath[1]), false, nil
	case "disks":
		if !lfs.disksEnabled {
			return "", false, errors.New("This storage target is not enabled for disk/kernel storage")
		}
		if len(splitPath) != 3 {
			return "", false, errors.New("The supplied path is invalid")
		}
		return filepath.Join(lfs.disksFolder, splitPath[1], splitPath[2]+".qcow2"), false, nil
	}
	return "", false, nil
}

package storage

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/768bit/promethium/lib/images"
	"github.com/768bit/vutils"
	"github.com/gobuffalo/envy"
	gozfs "github.com/mistifyio/go-zfs"
)

var ZFS_ROOT_PATH = envy.Get("PROMETHIUM_ZFS_ROOT_PATH", "nvmepool0/promethium")

type ZfsStorage struct {
	sm            *StorageManager
	id            string
	rootZfsPath   string
	mountPoint    string
	ds            *gozfs.Dataset
	disksDs       *gozfs.Dataset
	imagesCache   map[string]string
	imagesFolder  string
	imagesEnabled bool
	disksEnabled  bool
}

func LoadZfsStorage(sm *StorageManager, id string, zfsPath string) (*ZfsStorage, error) {
	//check the dataset exists
	ds, err := gozfs.GetDataset(zfsPath)
	if err != nil {
		return nil, err
	}

	//it exists...

	imagesFolder := filepath.Join(ds.Mountpoint, "images")

	zfsStorage := &ZfsStorage{
		sm:           sm,
		id:           id,
		mountPoint:   ds.Mountpoint,
		imagesCache:  map[string]string{},
		rootZfsPath:  zfsPath,
		ds:           ds,
		imagesFolder: imagesFolder,
	}

	vmsDs, err := gozfs.GetDataset(zfsPath)
	if err != nil {
		zfsStorage.disksEnabled = false
	} else {
		zfsStorage.disksEnabled = true
		zfsStorage.disksDs = vmsDs
	}

	if !vutils.Files.CheckPathExists(imagesFolder) {
		zfsStorage.imagesEnabled = false
	} else {
		zfsStorage.imagesEnabled = true
	}

	return zfsStorage, nil

}

func (zfs *ZfsStorage) isMounted() bool {
	//checks if the zfs dataset is mounted
	if zfs.ds.Mountpoint == "" {
		//it isnt mounted..
		return false
	} else if !vutils.Files.CheckPathExists(zfs.ds.Mountpoint) {
		return false
	} else {
		return true
	}
}

func (zfs *ZfsStorage) GetURI() string {
	return "zfs://" + zfs.id
}

func (zfs *ZfsStorage) LookupPath(path string) (string, bool, error) {
	return "zfs://" + zfs.id, false, nil
}

func (zfs *ZfsStorage) GetImages() ([]*images.Image, error) {
	//in zfs images are stored under a filesystem path..
	//vm disks are block devices (volumes)
	zfs.imagesCache = map[string]string{}
	if !zfs.isMounted() {
		return nil, errors.New("The ZFS Dataset is not mounted")
	} else if !zfs.imagesEnabled {
		//lets get all the images from the folder... if it exists
		return nil, errors.New("Unable to get images from this storage medium as it doesnt support images")
	} else {
		files := vutils.Files.GetFilesInDirWithExtension(zfs.imagesFolder, "prk")
		imagesList := []*images.Image{}
		for _, file := range files {
			imageFilePath := filepath.Join(zfs.imagesFolder, file)
			img, err := images.LoadImageFromPrk(imageFilePath, zfs.sm.imagesCache)
			if err != nil {
				return nil, err
			}
			zfs.imagesCache[img.ID] = file[:len(file)-4]
			imagesList = append(imagesList, img)
		}
		return imagesList, nil
	}
}

func (zfs *ZfsStorage) GetImage(name string) (*images.Image, error) {
	//in zfs images are stored under a filesystem path..
	//vm disks are block devices (volumes)
	if !zfs.isMounted() {
		return nil, errors.New("The ZFS Dataset is not mounted")
	} else if !zfs.imagesEnabled {
		//lets get all the images from the folder... if it exists
		return nil, errors.New("Unable to get images from this storage medium as it doesnt support images")
	} else {
		imageFilePath := filepath.Join(zfs.imagesFolder, name+".prk")
		if !vutils.Files.CheckPathExists(imageFilePath) {
			return nil, errors.New("Unable to find image with name " + name)
		} else {
			img, err := images.LoadImageFromPrk(imageFilePath, zfs.sm.imagesCache)
			if err != nil {
				return nil, err
			}
			return img, nil
		}
	}
}

func (zfs *ZfsStorage) GetImageById(id string) (*images.Image, error) {
	//in zfs images are stored under a filesystem path..
	//vm disks are block devices (volumes)
	if !zfs.isMounted() {
		return nil, errors.New("The ZFS Dataset is not mounted")
	} else if !zfs.imagesEnabled {
		//lets get all the images from the folder... if it exists
		return nil, errors.New("Unable to get images from this storage medium as it doesnt support images")
	} else if v, ok := zfs.imagesCache[id]; !ok || v == "" {
		//not in cache - maybe we need to go get the list of items and rebuild cache..
		_, err := zfs.GetImages()
		if err != nil {
			return nil, err
		}
		if v, ok := zfs.imagesCache[id]; !ok || v == "" {
			return nil, errors.New("Unable to find image with id " + id)
		} else {
			return zfs.GetImage(zfs.imagesCache[id])
		}
	} else {
		imageFilePath := filepath.Join(zfs.imagesFolder, zfs.imagesCache[id]+".prk")
		if !vutils.Files.CheckPathExists(imageFilePath) {
			return nil, errors.New("Unable to find image with id " + id)
		} else {
			img, err := images.LoadImageFromPrk(imageFilePath, zfs.sm.imagesCache)
			if err != nil {
				return nil, err
			}
			return img, nil
		}
	}
}

func (zfs *ZfsStorage) CreateDiskFromImage(id string, img *images.Image, size uint64) (*VmmStorageDisk, *VmmKernel, error) {
	return nil, nil, nil
}

func SetZfsRootPath(newRootPath string) {
	ZFS_ROOT_PATH = newRootPath
}

func NewZfsStorageDrive(id string, sizeMb int) (*Zfs, error) {
	//make an instance of zfs and lets see if it exists or not...
	datasetPath := fmt.Sprintf("%s/%s", ZFS_ROOT_PATH, id)
	size := uint64(sizeMb) * 1024 * 1024
	zfs := &Zfs{
		id:          id,
		datasetPath: datasetPath,
		size:        size,
	}
	if err := zfs.createOrLoad(); err != nil {
		return nil, err
	}
	return zfs, nil
}

type Zfs struct {
	id          string
	datasetPath string
	size        uint64
	ds          *gozfs.Dataset
}

func (zfs *Zfs) createOrLoad() error {
	//if it exists load it up - check sizes - auto expand?
	if !zfs.exists() {
		//create the dataset...
		_, err := gozfs.CreateVolume(zfs.getDataSetPath(), zfs.size, map[string]string{})
		if err != nil {
			return err
		}
	}
	return zfs.load()
}

func (zfs *Zfs) getMbSize() int {
	val := zfs.size / 1024 / 1024
	return int(val)
}

func (zfs *Zfs) exists() bool {
	ds, err := gozfs.GetDataset(zfs.getDataSetPath())
	if err != nil || ds == nil {
		return false
	}
	return true
}

func (zfs *Zfs) getDataSetPath() string {
	return zfs.datasetPath
}

func (zfs *Zfs) load() error {
	ds, err := gozfs.GetDataset(zfs.getDataSetPath())
	if err != nil {
		return err
	}

	//process ds
	zfs.ds = ds
	return zfs.checkSize()
}

func (zfs *Zfs) checkSize() error {
	//check the size is correct...
	currMbSize := int(zfs.ds.Volsize / 1024 / 1024)
	selMbSize := zfs.getMbSize()
	if currMbSize < selMbSize {
		//attempt the resize
		return zfs.resize(selMbSize)
	} else if currMbSize == selMbSize {
		//all good...
	} else {
		//the requested size is less than image size - use the actual size instead..
	}
	return nil
}

func (zfs *Zfs) resize(newSizeMb int) error {
	//check the size is correct...
	if err := zfs.ds.SetProperty("volsize", fmt.Sprintf("%dM", newSizeMb)); err != nil {
		return err
	}
	return zfs.load()

}

func (zfs *Zfs) Destroy() error {
	return zfs.ds.Destroy(gozfs.DestroyDefault)
}

func (zfs *Zfs) DestroyRecursive() error {
	return zfs.ds.Destroy(gozfs.DestroyRecursive)
}

package storage

import (
	"errors"
	"path/filepath"

	"github.com/768bit/promethium/lib/images"
	"github.com/768bit/vutils"
)

//this represents a local data storage driver - it is installed as the default storage driver when the node is first deployed
//and uses the location /opt/promethium/storage/default-local

type LocalFileStorage struct {
	id            string
	rootFolder    string
	imagesFolder  string
	imagesEnabled bool
	disksFolder   string
	disksEnabled  bool
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

func LoadLocalFileStorage(id string, config map[string]interface{}) (*LocalFileStorage, error) {
	rootFolder, err := getRootFolderFromConfig(config)
	if err != nil {
		return nil, err
	}
	if !vutils.Files.CheckPathExists(rootFolder) {
		return nil, errors.New("The specified local file storage folder: " + rootFolder + " doesn't exist")
	}
	imagesFolder := filepath.Join(rootFolder, "images")
	disksFolder := filepath.Join(rootFolder, "vmdisks")
	//create the instance...
	lfs := &LocalFileStorage{
		rootFolder:   rootFolder,
		disksFolder:  disksFolder,
		imagesFolder: imagesFolder,
		imagesCache:  map[string]string{},
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

func InitLocalFileStorage(id string, config map[string]interface{}) (*LocalFileStorage, error) {
	rootFolder, err := getRootFolderFromConfig(config)
	if err != nil {
		return nil, err
	}
	vutils.Files.CreateDirIfNotExist(rootFolder)
	imagesPath := filepath.Join(rootFolder, "images")
	vutils.Files.CreateDirIfNotExist(imagesPath)
	disksFolder := filepath.Join(rootFolder, "disks")
	vutils.Files.CreateDirIfNotExist(disksFolder)
	return LoadLocalFileStorage(id, config)
}

func (lfs *LocalFileStorage) GetImages() ([]*images.Image, error) {
	//get images needs to get a list of files from the directory
	lfs.imagesCache = map[string]string{}
	if !lfs.imagesEnabled {
		return nil, errors.New("Unable to get images from this storage medium as it doesnt support images")
	}
	files := vutils.Files.GetFilesInDirWithExtension(lfs.imagesFolder, "prk")
	imagesList := []*images.Image{}
	for _, file := range files {
		img, err := images.LoadImageFromPrk(file)
		if err != nil {
			return nil, err
		}
		lfs.imagesCache[img.ID] = file[:len(file)-4]
		imagesList = append(imagesList, img)
	}
	return imagesList, nil
}

func (lfs *LocalFileStorage) GetImage(name string) (*images.Image, error) {
	imageFilePath := filepath.Join(lfs.imagesFolder, name+".prk")
	if !lfs.imagesEnabled {
		return nil, errors.New("Unable to get images from this storage medium as it doesnt support images")
	} else if !vutils.Files.CheckPathExists(imageFilePath) {
		return nil, errors.New("Unable to find image with name " + name)
	} else {
		img, err := images.LoadImageFromPrk(imageFilePath)
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
		img, err := images.LoadImageFromPrk(imageFilePath)
		if err != nil {
			return nil, err
		}
		return img, nil
	}
}

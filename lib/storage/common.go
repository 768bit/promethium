package storage

import "github.com/768bit/promethium/lib/images"

func NewVmmStorage(uri string) (VmmStorage, error) {
	return nil, nil
}

type VmmStorage interface {
	Create(name string)
}

type StorageDriver interface {
	GetImages() ([]*images.Image, error)
	ImportImage(srcPrkPath string) (*images.Image, error)
	GetImage(name string) (*images.Image, error)
	GetImageById(id string) (*images.Image, error)
	//GetDisks() ([]*VmmStorageDisk, error)
	//GetDisk(path string) (*VmmStorageDisk, error)
	//CreateDisk(id string) (*VmmStorageDisk, error)
	//CreateDiskFromImage(id string, img *images.Image) (*VmmStorageDisk, error)
}

type ImageSpec struct {
	Architecture string
	Type         string
}

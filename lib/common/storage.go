package common

import (
	"io"
)

func NewVmmStorage(uri string) (VmmStorage, error) {
	return nil, nil
}

type VmmStorage interface {
	Create(name string)
}

type StorageDriver interface {
	GetURI() string
	GetImages() ([]Image, error)
	//ImportImage(srcPrkPath string) (*images.Image, error)
	ImportImageFromRdr(stream io.ReadCloser) error
	GetImage(name string) (Image, error)
	GetImageById(id string) (Image, error)
	//GetDisks() ([]*VmmStorageDisk, error)
	//GetDisk(path string) (*VmmStorageDisk, error)
	//CreateDisk(id string) (*VmmStorageDisk, error)
	CreateDiskFromImage(id string, img Image, size uint64) (*VmmStorageDisk, *VmmKernel, error)
	LookupPath(path string) (string, bool, error)
	WriteKernel(id string, source io.Reader) (string, error)
	WriteCloudInit(id string, source io.Reader) (string, error)
	WriteRootDisk(id string, source io.Reader, newSize int64, sourceIsRaw bool, growPart bool) (string, error)
	WriteAdditionalDisk(id string, index int, source io.Reader, newSize int64, sourceIsRaw bool, growPart bool) (string, error)
}

type ImageSpec struct {
	Architecture string
	Type         string
}

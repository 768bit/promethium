package storage

import (
	"errors"
	"os"
)

type DiskFileFormat string

const (
	QCow2VmmDiskFileFormat   DiskFileFormat = "qcow2"
	UnknownVmmDiskFileFormat DiskFileFormat = "unknown"
)

type VmmStorageDisk struct {
	path                  string
	isBlockDevice         bool
	isNbdMounted          bool
	storageDiskFileFormat DiskFileFormat
	size                  uint64
	storageDriver         StorageDriver
}

func NewStorageDisk(path string) (*VmmStorageDisk, error) {
	vmd := &VmmStorageDisk{
		path:                  path,
		isBlockDevice:         false,
		isNbdMounted:          false,
		storageDiskFileFormat: UnknownVmmDiskFileFormat,
	}
	if err := vmd.init(); err != nil {
		return nil, err
	}
	return vmd, nil
}

func (vmd *VmmStorageDisk) init() error {
	//check what type of device this is..
	if fmt, err := vmd.establishStorageType(); err != nil {
		return err
	} else {
		vmd.storageDiskFileFormat = fmt
	}
	//now we have format and type established.. this really serves to prove existence..
	return nil
}

func (vmd *VmmStorageDisk) establishStorageType() (DiskFileFormat, error) {
	//check what type of device this is..
	fi, err := os.Stat(vmd.path)
	if err != nil {
		return UnknownVmmDiskFileFormat, err
	}
	if fi.Mode()&os.ModeDevice != 0 {
		vmd.isBlockDevice = true
		return UnknownVmmDiskFileFormat, nil
	} else if fi.IsDir() {
		//doesnt support directories
		return UnknownVmmDiskFileFormat, errors.New("Directories are not supported for image backends. Only use block devices or Qcow2 images.")
	} else {

		return QCow2VmmDiskFileFormat, nil
	}
}

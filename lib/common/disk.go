package common

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"

	"github.com/768bit/promethium/lib/config"
)

type DiskFileFormat string

const (
	QCow2VmmDiskFileFormat   DiskFileFormat = "qcow2"
	UnknownVmmDiskFileFormat DiskFileFormat = "unknown"
)

type DiskMeta struct {
	Machine  string `json:"machine" yaml:"machine"`
	Platform string `json:"platform" yaml:"platform"`
	From     string `json:"from" yaml:"from"`
}

func LoadDiskMeta(path string) (*DiskMeta, error) {
	var dm DiskMeta
	ba, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(ba, &dm)
	if err != nil {
		return nil, err
	}
	return &dm, nil
}

type VmmStorageDisk struct {
	id                    string
	name                  string
	path                  string
	isBlockDevice         bool
	isNbdMounted          bool
	storageDiskFileFormat DiskFileFormat
	size                  uint64
	storageDriver         StorageDriver
}

func NewStorageDisk(id string, name string, path string, driver StorageDriver) (*VmmStorageDisk, error) {
	vmd := &VmmStorageDisk{
		id:                    id,
		name:                  name,
		path:                  path,
		isBlockDevice:         false,
		isNbdMounted:          false,
		storageDiskFileFormat: UnknownVmmDiskFileFormat,
		storageDriver:         driver,
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

func (vmd *VmmStorageDisk) ToDiskConfig() *config.VmmDiskConfig {
	//check what type of device this is..
	fullStorageUri := vmd.storageDriver.GetURI() + "/disks/" + vmd.id + "/" + vmd.name
	isRoot := false
	if vmd.name == "root" {
		isRoot = true
	}
	return &config.VmmDiskConfig{
		StorageURI: fullStorageUri,
		IsRoot:     isRoot,
	}
}

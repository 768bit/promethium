package storage

import "github.com/768bit/promethium/lib/config"

type StorageManager struct {
	targets map[string]StorageDriver
}

func NewStorageManager(configs []*config.StorageConfig) *StorageManager {
	sm := &StorageManager{
		targets: map[string]StorageDriver{},
	}
	return sm
}

func (sm *StorageManager) init(configs []*config.StorageConfig) (error, *StorageManager) {
	//using the configs passed in we need to load the storage drivers with the supplied config...

	for i, config := range configs {
		switch config.Driver {
		case "zfs":
			//load the zfs storage driver with ID...

		}
	}

	return nil, sm

}

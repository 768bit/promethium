package storage

type StorageManager struct {
	targets map[string]StorageDriver
}

func NewStorageManager(configs []*StorageConfig) *StorageManager {
	sm := &StorageManager{
		targets: map[string]StorageDriver{},
	}
	return sm
}

func (sm *StorageManager) init(configs []*StorageConfig) (error, *StorageManager) {
	//using the configs passed in we need to load the storage drivers with the supplied config...

	for _, config := range configs {
		switch config.Driver {
		case "zfs":
			//load the zfs storage driver with ID...

		}
	}

	return nil, sm

}

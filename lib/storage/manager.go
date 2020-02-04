package storage

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"github.com/768bit/promethium/lib/common"
	"github.com/768bit/promethium/lib/config"
	"github.com/768bit/promethium/lib/images"
	"github.com/768bit/vutils"
	"github.com/fsnotify/fsnotify"
)

type StorageManager struct {
	rootPath        string
	targets         map[string]common.StorageDriver
	imagesCachePath string
	imagesCache     map[string]*images.ImageCacheFile
	imagesHashMap   map[string]string
}

func NewStorageManager(promethiumRootPath string, configs []*config.StorageConfig, uid int, gid int) (*StorageManager, error) {
	sm := &StorageManager{
		rootPath:        promethiumRootPath,
		targets:         map[string]common.StorageDriver{},
		imagesCachePath: filepath.Join(promethiumRootPath, "cache", "images"),
		imagesCache:     nil,
		imagesHashMap:   map[string]string{},
	}
	return sm.init(configs, uid, gid)
}

func (sm *StorageManager) init(configs []*config.StorageConfig, uid int, gid int) (*StorageManager, error) {
	//using the configs passed in we need to load the storage drivers with the supplied config...

	vutils.Files.CreateDirIfNotExist(sm.imagesCachePath)

	sm.loadImagesCache()
	// cacheRoot := filepath.Join(sm.rootPath, "cache")
	// config.DoChmod(cacheRoot, 0600, true)
	// config.DoChown(cacheRoot, uid, gid, true)

	for _, config := range configs {
		println("Loading Storage Driver: " + config.Driver + " with ID: " + config.ID)
		switch config.Driver {
		case "zfs":
			//load the zfs storage driver with ID...
		case "local-file":
			store, err := LoadLocalFileStorage(sm, config.ID, config.Config)
			if err != nil {
				println("Error loading storage driver with id " + config.ID + " and type " + config.Driver + " -> " + err.Error())
			} else {
				sm.targets[config.ID] = store
				//now populate the cache
			}
		default:
			println("The Storage driver " + config.Driver + " is not supported.")
		}
	}

	go func() {
		println("Getting initial Images")
		sm.GetImages()
		go func() {
			for {
				time.Sleep(10 * time.Second)
				sm.writeImagesCache()
			}
		}()
	}()

	println("Storage Manager Ready")

	return sm, nil

}

func (sm *StorageManager) loadImagesCache() {
	vutils.Files.CreateDirIfNotExist(sm.imagesCachePath)
	files := vutils.Files.GetFilesInDirWithExtension(sm.imagesCachePath, ".json")
	icache := map[string]*images.ImageCacheFile{}
	for _, file := range files {
		fullPath := filepath.Join(sm.imagesCachePath, file)
		//try and load the cache entry...
		file, err := os.Open(fullPath)
		if err != nil {
			println("Unable to open images cache file " + fullPath + " : " + err.Error())
		} else {
			bs, err := ioutil.ReadAll(file)
			file.Close()
			if err != nil {
				println("Unable to read images cache file " + fullPath + " : " + err.Error())
			} else {
				ic := &images.ImageCacheFile{}
				err = json.Unmarshal(bs, ic)
				if err != nil {
					println("Unable to unmarshal images cache file " + fullPath + " : " + err.Error())
				} else {
					icache[ic.ID] = ic
					sm.imagesHashMap[ic.ID] = ic.GetHash()
				}
			}
		}
	}
	sm.imagesCache = icache
}

func (sm *StorageManager) writeImagesCache() {

	//first calculate the hash for images cache...

	//os.RemoveAll(sm.imagesCachePath)
	vutils.Files.CreateDirIfNotExist(sm.imagesCachePath)

	for id, cacheEntry := range sm.imagesCache {
		fullPath := filepath.Join(sm.imagesCachePath, id+".json")
		if vutils.Files.CheckPathExists(fullPath) {
			//check to see if the hash differs from previous or if it even exists!
			if v, ok := sm.imagesHashMap[id]; !ok || v == "" {
				sm.imagesHashMap[id] = cacheEntry.GetHash()
			} else if v == cacheEntry.GetHash() {
				continue
			} else {
				sm.imagesHashMap[id] = cacheEntry.GetHash()
			}
		} else {
			sm.imagesHashMap[id] = cacheEntry.GetHash()
		}
		bs, err := json.Marshal(cacheEntry)
		if err != nil {
			println("Unable to marshal cache entry to JSON: " + err.Error())
		} else {
			println("Writing cache entry for " + id)
			err = ioutil.WriteFile(fullPath, bs, 0600)
			if err != nil {
				println("Error writing Cache Entry: " + err.Error())
			}
		}
	}
	files := vutils.Files.GetFilesInDirWithExtension(sm.imagesCachePath, ".json")
	for _, file := range files {
		fullPath := filepath.Join(sm.imagesCachePath, file)
		id := file[:len(file)-5]
		if _, ok := sm.imagesHashMap[id]; !ok {
			println("Removing invalidated cache entry for " + id)
			os.Remove(fullPath)
		}
	}
}

func (sm *StorageManager) GetStorage(name string) (common.StorageDriver, error) {
	if sm == nil || sm.targets == nil {
		return nil, errors.New("Target Map is NIL!")
	}
	if v, ok := sm.targets[name]; !ok || v == nil {
		return nil, errors.New("Unable to get storage target with name " + name)
	} else {
		return v, nil
	}
}

func (sm *StorageManager) GetImages() []common.Image {

	imagesList := []common.Image{}
	for _, storageDriver := range sm.targets {
		imgs, err := storageDriver.GetImages()
		if err == nil && imgs != nil && len(imgs) > 0 {
			imagesList = append(imagesList, imgs...)
		} else if err != nil {
			println(err.Error())
		}
	}
	return imagesList
}

func (sm *StorageManager) GetImageByID(id string) (common.Image, error) {

	for _, storageDriver := range sm.targets {
		img, err := storageDriver.GetImageById(id)
		if err != nil || img == nil {
			continue
		} else if err == nil {
			return img, nil
		}
	}

	return nil, errors.New("Unable to find image with ID " + id)
}

func (sm *StorageManager) GetImage(name string) (common.Image, error) {

	for _, storageDriver := range sm.targets {
		img, err := storageDriver.GetImage(name)
		if err != nil || img == nil {
			continue
		} else if err == nil {
			return img, nil
		}
	}

	return nil, errors.New("Unable to find image with name " + name)
}

func (sm *StorageManager) MakeNewVmDiskAndKernelFromImage(id string, targetStorage string, img common.Image, size uint64) (*common.VmmStorageDisk, *common.VmmKernel, error) {

	//get the storage target
	target, err := sm.GetStorage(targetStorage)
	if err != nil {
		return nil, nil, err
	}

	return target.CreateDiskFromImage(id, img, size)

}

func (sm *StorageManager) ResolveStorageURI(uri string) (string, bool, error) {

	//parse the inbound uri...
	outUrl, err := url.Parse(uri)
	if err != nil {
		return "", false, err
	} else {
		switch outUrl.Scheme {
		case "local-file":
			//lets now lookup the path...
			id := outUrl.Host
			path := outUrl.Path
			println(outUrl.Scheme + " -> " + id + " : " + path)
			if stgt, err := sm.GetStorage(id); err != nil {
				return "", false, err
			} else {
				return stgt.LookupPath(path)
			}
		}
	}
	return "", false, nil
}

func (sm *StorageManager) runDirectoryWatch() {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	done := make(chan bool)
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				log.Println("event:", event)
				if event.Op&fsnotify.Write == fsnotify.Write {
					log.Println("modified file:", event.Name)
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("error:", err)
			}
		}
	}()

	err = watcher.Add("/tmp/foo")
	if err != nil {
		log.Fatal(err)
	}
	<-done
}

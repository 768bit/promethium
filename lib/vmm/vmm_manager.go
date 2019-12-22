package vmm

import (
	"errors"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/768bit/promethium/api/models"
	"github.com/768bit/promethium/lib/assets"
	"github.com/768bit/promethium/lib/config"
	"github.com/768bit/promethium/lib/networking"
	"github.com/768bit/promethium/lib/storage"
	"github.com/768bit/vutils"
)

/* the vmm manager scans the INSTANCE_DIR directory for instances and begins loading them as appropriate if the config allows it...

It will also load its global config so it can initialise itself...
*/

var ROOT_PATH string = ""

func NewVmmManager(config *config.PromethiumDaemonConfig) (*VmmManager, error) {
	log.Printf("Initialising VmmManager...")
	vmmMgr := &VmmManager{
		config:                 config,
		appRootPath:            config.AppRoot,
		instances:              map[string]*Vmm{},
		clusterInstances:       map[string]map[string]*Vmm{},
		instanceConfigRootPath: filepath.Join(config.AppRoot, "instances"),
	}
	ROOT_PATH = config.AppRoot
	if err := vmmMgr.init(); err != nil {
		return nil, err
	}
	return vmmMgr, nil
}

type VmmManager struct {
	config *config.PromethiumDaemonConfig

	appRootPath            string
	storageManager         *storage.StorageManager
	instanceConfigRootPath string

	fcInstanceRootPath string

	instances        map[string]*Vmm
	clusterInstances map[string]map[string]*Vmm

	networks *networking.Manager

	runGroup  sync.WaitGroup
	stopGroup sync.WaitGroup
	killGroup sync.WaitGroup
}

func (vmmMgr *VmmManager) init() error {
	if err := vmmMgr.createFolders(); err != nil {
		return err
	}
	if err := vmmMgr.setupNetworking(); err != nil {
		return err
	}
	if vmmMgr.config.IsNewConfig() {
		println("Is new config")
		err := storage.InitLocalFileStorage(vmmMgr.config.Storage[0].ID, vmmMgr.config.Storage[0].Config)
		if err != nil {
			println(err.Error())
		}
	}
	if storageMgr, err := storage.NewStorageManager(vmmMgr.appRootPath, vmmMgr.config.Storage); err != nil {
		return err
	} else {
		vmmMgr.storageManager = storageMgr
	}
	if err := vmmMgr.scanInstanceConfigs(); err != nil {
		return err
	}
	return nil
}

func (vmmMgr *VmmManager) scanInstanceConfigs() error {
	log.Printf("Scanning Instance Configs...")
	configMap := map[string]bool{}
	for _, instanceConf := range vutils.Files.GetFilesInDirWithExtension(vmmMgr.instanceConfigRootPath, ".json") {
		log.Printf("Loading Vmm Config: %s", instanceConf)
		if vmm, err := vmmMgr.LoadVmm(filepath.Join(vmmMgr.instanceConfigRootPath, instanceConf)); err != nil {
			println("Error loading VMM: " + err.Error())
			return err
		} else {
			configMap[vmm.id] = true
		}
	}
	return vmmMgr.scanInstances(configMap)
}

func (vmmMgr *VmmManager) scanInstances(configMap map[string]bool) error {
	//check that the instance exists in the configs... will also mean any orphans are tracked too...
	//
	log.Printf("Scanning Instances...")
	instanceDirs := []string{}
	clusterInstanceDirs := []string{}
	orphanedInstanceDirs := []string{}
	err := filepath.Walk(vmmMgr.fcInstanceRootPath, func(path string, f os.FileInfo, _ error) error {
		if f.IsDir() {

			//check if this directory is valid...
			if _, ok := vmmMgr.instances[f.Name()]; !ok {
				foundItem := false
				for _, clusterInstances := range vmmMgr.clusterInstances {
					if _, ok := clusterInstances[f.Name()]; ok {
						foundItem = true
						delete(configMap, f.Name())
						clusterInstanceDirs = append(clusterInstanceDirs, filepath.Join(vmmMgr.fcInstanceRootPath, f.Name()))
						break
					}
				}
				if !foundItem {
					//is an orphaned item..
					orphanedInstanceDirs = append(orphanedInstanceDirs, filepath.Join(vmmMgr.fcInstanceRootPath, f.Name()))
				}
			} else {
				delete(configMap, f.Name())
				instanceDirs = append(instanceDirs, filepath.Join(vmmMgr.fcInstanceRootPath, f.Name()))
			}

		}
		return nil
	})
	return err
}

func (vmmMgr *VmmManager) createFolders() error {
	log.Printf("Creating VmmManager Folder Structure...")
	vmmMgr.instanceConfigRootPath = filepath.Join(vmmMgr.appRootPath, "instances")
	vmmMgr.fcInstanceRootPath = filepath.Join(vmmMgr.appRootPath, "firecracker")
	binPath := filepath.Join(vmmMgr.appRootPath, "bin")
	vutils.Files.CreateDirIfNotExist(binPath)
	fcBinary, err := assets.FireCrackerAssets.Open("firecracker")
	if err == nil {
		of, err := os.Create(filepath.Join(binPath, "firecracker"))
		of.Chmod(0740)
		if err != nil {
			println(err.Error())
		} else {
			io.Copy(of, fcBinary)
			of.Close()
		}
		fcBinary.Close()
	}
	jlrBinary, err := assets.FireCrackerAssets.Open("jailer")
	if err == nil {
		of, err := os.Create(filepath.Join(binPath, "jailer"))
		of.Chmod(0740)
		if err != nil {
			println(err.Error())
		} else {
			io.Copy(of, jlrBinary)
			of.Close()
		}
		jlrBinary.Close()
	}
	if err := vutils.Files.CreateDirIfNotExist(vmmMgr.instanceConfigRootPath); err != nil {
		return err
	} else {
		return vutils.Files.CreateDirIfNotExist(vmmMgr.fcInstanceRootPath)
	}
}

func (vmmMgr *VmmManager) setupNetworking() error {
	log.Printf("Initialising Networking...")
	netMgr, err := networking.NewManager(vmmMgr.config.Networks)
	if err != nil {
		return err
	}
	vmmMgr.networks = netMgr
	return err
}

func (vmmMgr *VmmManager) Wait() error {
	return nil
}

func (vmmMgr *VmmManager) Kill() error {
	//kill all instances IMMEDIATELY
	for _, vmm := range vmmMgr.instances {
		vmm.Kill()
	}
	for _, vmmColl := range vmmMgr.clusterInstances {
		for _, vmm := range vmmColl {
			vmm.Kill()
		}
	}
	return nil
}

func (vmmMgr *VmmManager) WaitKill() error {
	//kill all instances IMMEDIATELY
	vmmMgr.killGroup = sync.WaitGroup{}
	for _, vmm := range vmmMgr.instances {
		vmmMgr.killGroup.Add(1)
		go func() {
			vmm.WaitKill(30 * time.Second)
			vmmMgr.killGroup.Done()
		}()
	}
	for _, vmmColl := range vmmMgr.clusterInstances {
		for _, vmm := range vmmColl {
			vmmMgr.killGroup.Add(1)
			go func() {
				vmm.WaitKill(30 * time.Second)
				vmmMgr.killGroup.Done()
			}()
		}
	}
	vmmMgr.killGroup.Wait()
	return nil
}

func (vmmMgr *VmmManager) Storage() *storage.StorageManager {

	return vmmMgr.storageManager
}

func (vmmMgr *VmmManager) Create(newVmConf *models.NewVM) (*Vmm, error) {
	//need to create a templated VM..

	//get storage back end

	//get linux bridge

	return vmmMgr.NewVmmFromImage(newVmConf.Name, newVmConf.Cpus, newVmConf.Memory, newVmConf.FromImage, uint64(newVmConf.RootDiskSize), newVmConf.StorageName, newVmConf.PrimaryNetworkID)

}

func (vmmMgr *VmmManager) List(showAll bool) []*Vmm {
	instList := []*Vmm{}
	for _, vmm := range vmmMgr.instances {
		if showAll {
			instList = append(instList, vmm)
		} else {
			status := vmm.Status()
			if status == "Running" || status == "Starting" {
				instList = append(instList, vmm)
			}
		}
	}
	return instList
}

func (vmmMgr *VmmManager) Get(id string) (*Vmm, error) {
	if v, ok := vmmMgr.instances[id]; !ok || v == nil {
		return nil, errors.New("Unable to find instance with that id")
	} else {
		return v, nil
	}
}

func (vmmMgr *VmmManager) Update() error {
	return nil
}

func (vmmMgr *VmmManager) Destroy() error {
	return nil
}

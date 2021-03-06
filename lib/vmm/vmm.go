package vmm

import (
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"

	"github.com/768bit/promethium/lib/common"
	"github.com/768bit/promethium/lib/config"
	"github.com/768bit/vutils"
)

/*
VMM is an instance of firecracker essentially but handles the config and other security setup for the process..
It will manage bindings and folder structures and ensure volumes etc required are loaded...

Network interface setup via OpenVSwitch of standard tap interface are done now (via config)

They can be plain VMs... or OSv unikernel enabled applications

*/

//VMM consists of the process and other related items...
//The process is loaded and dependencies are tracked..

//when creating a new vmm we only care about
func (mgr *VmmManager) NewVmmFromImage(name string, vcpus int64, mem int64, image string, size uint64, targetStorage string, primaryNetwork string, kernelImage string) (*Vmm, error) {

	//the primary network selection consists of:
	// bridge name
	// vlan (if applicable)
	// custom mac
	// ipconfig...
	// device will be eth0

	//disk needs to be generated... based on... source image size and what storage back end to use

	//get the image..
	vmmId, _ := vutils.UUID.MakeUUIDString()
	vmmConfig := &config.VmmConfig{
		ID:        vmmId,
		Name:      name,
		Clustered: false,
		Memory:    mem,
		Cpus:      vcpus,
		Kernel:    "",
		Volumes:   []*config.VmmVolumeConfig{},
		Network:   &config.VmmNetworkConfig{},
	}

	img, err := mgr.Storage().GetImageByID(image)
	if err != nil {

		println(err.Error())
		img, err = mgr.Storage().GetImage(image)
		if err != nil {

			println(err.Error())
			return nil, err
		}
	}

	println("Got image " + img.GetID())

	tstr, err := mgr.Storage().GetStorage(targetStorage)
	if err != nil {

		println(err.Error())
		return nil, err
	}
	pkgFile, ds, err := img.GetRootDiskReader()
	if err != nil {
		if pkgFile != nil {
			pkgFile.Close()
		}
		println(err.Error())
		return nil, err
	}
	defer pkgFile.Close()

	vmmConfig.Type = config.VmmType(img.GetType())

	rootDiskPath, err := tstr.WriteRootDisk(vmmId, ds, int64(size), false, true)
	if err != nil {

		println(err.Error())
		return nil, err
	}

	rootDisk, err := common.NewStorageDisk(vmmId, "root.img", rootDiskPath, tstr)
	if err != nil {

		println(err.Error())
		return nil, err
	}
	vmmConfig.Disks = []*config.VmmDiskConfig{
		rootDisk.ToDiskConfig(true),
	}

	//does the image have a kernel? have we selected a different one?

	if !img.HasKernel() && kernelImage == "" {
		fmt.Println("Kernel not specified for VM")
	} else if !img.HasKernel() {
		//we dont have a kernel but be can seek for one... one was specified...
		println("Seeking Kernel with ID: " + kernelImage)
		kimg, err := mgr.Storage().GetImageByID(kernelImage)
		if err != nil {

			println(err.Error())
			kimg, err = mgr.Storage().GetImage(kernelImage)
			if err != nil {

				println(err.Error())
				return nil, err
			}
		}
		kernFile, ds, err := kimg.GetKernelReader()
		if err != nil {
			if pkgFile != nil {
				pkgFile.Close()
			}
			println(err.Error())
			return nil, err
		}
		defer kernFile.Close()
		kernelDiskPath, err := tstr.WriteKernel(vmmId, ds)
		if err != nil {

			println(err.Error())
			return nil, err
		}
		kernelImg := common.NewKernel(vmmId, kernelDiskPath, tstr)
		vmmConfig.Kernel = kernelImg.GetURI()
	} else {
		//the main image contains a kernel
		kernFile, ds, err := img.GetKernelReader()
		if err != nil {
			if pkgFile != nil {
				pkgFile.Close()
			}
			println(err.Error())
			return nil, err
		}
		defer kernFile.Close()
		kernelDiskPath, err := tstr.WriteKernel(vmmId, ds)
		if err != nil {

			println(err.Error())
			return nil, err
		}
		kernelImg := common.NewKernel(vmmId, kernelDiskPath, tstr)
		vmmConfig.Kernel = kernelImg.GetURI()
	}

	vmmConfig.BootCmd = img.GetBootParams()

	//we have an image now we need to instantiate it in the storage target (and ther kernel)
	// disk, kernel, err := mgr.Storage().MakeNewVmDiskAndKernelFromImage(vmmId, targetStorage, img, size)
	// if err != nil {

	// 	println(err.Error())
	// 	return nil, err
	// }

	//save the config...

	vmmConfigPath := filepath.Join(mgr.instanceConfigRootPath, vmmId+".json")

	err, _ = vutils.Config.SaveConfigToFile("", vmmConfigPath, vmmConfig)
	if err != nil {
		return nil, err
	}

	vmm := &Vmm{
		mgr:        mgr,
		id:         vmmId,
		configPath: vmmConfigPath,
		config:     vmmConfig,
	}

	mgr.instances[vmmId] = vmm

	return vmm.init(vmmConfig)
}

func (mgr *VmmManager) LoadVmm(vmmConfigPath string) (*Vmm, error) {

	//vmmConfigPath := filepath.Join(mgr.instanceConfigRootPath, vmmId + ".json")

	vmmConfig := &config.VmmConfig{}

	if err := vutils.Config.LoadConfigFromFile(vmmConfigPath, vmmConfig); err != nil {
		return nil, err
	} else if vmmConfig == nil {
		return nil, errors.New("Config is nil")
	}

	vmm := &Vmm{
		mgr:        mgr,
		id:         vmmConfig.ID,
		configPath: vmmConfigPath,
		config:     vmmConfig,
	}

	mgr.instances[vmmConfig.ID] = vmm

	return vmm.init(vmmConfig)
}

type Vmm struct {
	mgr          *VmmManager
	id           string
	configPath   string
	config       *config.VmmConfig
	kernelPath   string
	rootDiskPath string

	fcInstancePath string
	instance       common.VmmProcess
}

func (vmm *Vmm) init(cfg *config.VmmConfig) (*Vmm, error) {
	//when it is initialised we need to look at the type...
	//for OSv we need to look at the build process.. is the image ready and available...
	// standard VMs need a kernel just like OSv...
	// in addition any resources that are needed for the execution of the VM need to be considered now...
	// the folder structure for the instance will be created now and used by the process (firecracker/qemu)

	vmm.config = cfg

	//establish kernel and root image paths

	kernelPath, _, err := vmm.mgr.Storage().ResolveStorageURI(cfg.Kernel)
	if err != nil {
		return vmm, err
	}
	drvList := []string{}
	foundRoot := false
	if cfg.Disks != nil && len(cfg.Disks) > 0 {
		drvList = append(drvList, "FOR_ROOT")
		for _, dsk := range cfg.Disks {

			path, _, err := vmm.mgr.Storage().ResolveStorageURI(dsk.StorageURI)
			if err != nil {
				return vmm, err
			}
			if dsk.IsRoot && !foundRoot {
				drvList[0] = path
				foundRoot = true
			} else {
				drvList = append(drvList, path)
			}

		}

	}

	switch cfg.Type {
	case config.FirecrackerVmm:
		fcp, err := NewFireCrackerProcessImg(vmm.id, vmm.config.Name, strings.TrimSpace(vmm.config.BootCmd), vmm.config.Cpus, vmm.config.Memory,
			kernelPath, drvList, nil, vmm.config.AutoStart)
		if err != nil {
			return vmm, err
		}
		vmm.instance = fcp
		return vmm, nil
	case config.OSvFirecrackerVmm:
		//fcp, err := NewFireCrackerProcess(vmm.id, vmm.config.Name, vmm.config.BootCmd, vmm.config.EntryPoint, vmm.config.Cpus, vmm.config.Memory,
		//  "", )
		return vmm, nil
	default:
		return vmm, errors.New("Unknown type: " + string(cfg.Type))
	}

}

func (vmm *Vmm) GetModel() string {
	return vmm.config.Name
}

func (vmm *Vmm) Name() string {
	return vmm.config.Name
}

func (vmm *Vmm) ID() string {
	return vmm.id
}

func (vmm *Vmm) Start() error {
	if vmm.instance == nil {
		return errors.New("Unable to start as instance isnt setup")
	} else {
		err := vmm.instance.Start()
		if err != nil {
			return err
		}

		return nil
	}
}

func (vmm *Vmm) Stop() error {
	if vmm.instance == nil {
		return errors.New("Unable to stop as instance isnt setup")
	} else {
		err := vmm.instance.Stop()
		if err != nil {
			return err
		}

		return nil
	}
}

func (vmm *Vmm) Shutdown() error {
	if vmm.instance == nil {
		return errors.New("Unable to shutdown as instance isnt setup")
	} else {
		err := vmm.instance.Shutdown()
		if err != nil {
			return err
		}

		return nil
	}
}

func (vmm *Vmm) Restart() error {
	if vmm.instance == nil {
		return errors.New("Unable to restart as instance isnt setup")
	} else {
		err := vmm.instance.Restart()
		if err != nil {
			return err
		}

		return nil
	}
}

func (vmm *Vmm) Reset() error {
	if vmm.instance == nil {
		return errors.New("Unable to reset as instance isnt setup")
	} else {
		err := vmm.instance.Reset()
		if err != nil {
			return err
		}

		return nil
	}
}

func (vmm *Vmm) Console() (io.ReadCloser, io.ReadCloser, io.WriteCloser, error) {

	return vmm.instance.Console()
}

func (vmm *Vmm) Status() string {
	return vmm.instance.GetStatus()
}

func (vmm *Vmm) Kill() error {
	return vmm.instance.Stop()
}

func (vmm *Vmm) WaitKill(timeout time.Duration) error {
	return vmm.instance.ShutdownTimeout(timeout)
}

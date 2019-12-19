package vmm

import (
  "errors"
  "github.com/768bit/promethium/lib/interfaces"
  "github.com/768bit/vutils"
  "path/filepath"
  "time"
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
func (mgr *VmmManager) NewVmmFromImage(name string, vcpus int64, mem int64, network interface{}, disk interface{}, vmmType VmmType, image string) (*Vmm, error) {

  //the primary network selection consists of:
  // bridge name
  // vlan (if applicable)
  // custom mac
  // ipconfig...
  // device will be eth0

  //disk needs to be generated... based on... source image size and what storage back end to use

  vmmId, _ := vutils.UUID.MakeUUIDString()
  vmmConfig := &VmmConfig{
    ID:         vmmId,
    Name:       name,
    Clustered:  false,
    Memory:     mem,
    Cpus:       vcpus,
    Type:       vmmType,
    Volumes:    []*VmmVolumeConfig{},
    Interfaces: []*VmmNetworkInterfaceConfig{},
    Disks:      []*VmmDiskConfig{},
  }

  //save the config...

  vmmConfigPath := filepath.Join(mgr.instanceConfigRootPath, vmmId + ".json")

  err, _ := vutils.Config.SaveConfigToFile("", vmmConfigPath, vmmConfig)
  if err != nil {
    return nil, err
  }

  vmm := &Vmm{
    mgr: mgr,
    id: vmmId,
    configPath:vmmConfigPath,
  }

  mgr.instances[vmmId] = vmm

  return vmm.init(vmmConfig)
}

func (mgr *VmmManager) LoadVmm(vmmConfigPath string) (*Vmm, error) {

  //vmmConfigPath := filepath.Join(mgr.instanceConfigRootPath, vmmId + ".json")

  var vmmConfig *VmmConfig

  if err := vutils.Config.LoadConfigFromFile(vmmConfigPath, vmmConfig); err != nil {
    return nil, err
  } else if vmmConfig == nil {
    return nil, errors.New("")
  }

  vmm := &Vmm{
    mgr: mgr,
    id: vmmConfig.ID,
    configPath:vmmConfigPath,
  }

  mgr.instances[vmmConfig.ID] = vmm

  return vmm.init(vmmConfig)
}

type Vmm struct {

  mgr *VmmManager
  id string
  configPath string
  config *VmmConfig

  fcInstancePath string
  instance interfaces.VmmProcess

}

func (vmm *Vmm) init(config *VmmConfig) (*Vmm, error) {
  //when it is initialised we need to look at the type...
  //for OSv we need to look at the build process.. is the image ready and available...
  // standard VMs need a kernel just like OSv...
  // in addition any resources that are needed for the execution of the VM need to be considered now...
  // the folder structure for the instance will be created now and used by the process (firecracker/qemu)

  vmm.config = config

  //establish instance path

  switch config.Type {
  case FirecrackerVmm:
    fcp, err := NewFireCrackerProcessImg(vmm.id, vmm.config.Name, vmm.config.BootCmd, vmm.config.Cpus, vmm.config.Memory,
      "", nil, nil, vmm.config.AutoStart)
    if err != nil {
      return vmm, err
    }
    vmm.instance = fcp
    return vmm, nil
  case OSvFirecrackerVmm:
    //fcp, err := NewFireCrackerProcess(vmm.id, vmm.config.Name, vmm.config.BootCmd, vmm.config.EntryPoint, vmm.config.Cpus, vmm.config.Memory,
    //  "", )
    return vmm, nil
  default:
    return vmm, errors.New("")
  }

}

func (vmm *Vmm) Name() (string) {
  return vmm.config.Name
}

func (vmm *Vmm) ID() (string) {
  return vmm.id
}

func (vmm *Vmm) Status() (string) {
  return vmm.instance.GetStatus()
}

func (vmm *Vmm) Kill() (error) {
  return vmm.instance.Stop()
}

func (vmm *Vmm) WaitKill(timeout time.Duration) (error) {
  return vmm.instance.ShutdownTimeout(timeout)
}

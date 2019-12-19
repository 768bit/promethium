package vmm

import "github.com/768bit/promethium/lib/cloudconfig"

type VmmType string
const (
  FirecrackerVmm VmmType = "firecracker-standard"
  OSvFirecrackerVmm VmmType = "firecracker-osv"
  QemuVmm VmmType = "qemu-standard"
)

type VmmConfig struct {
  ID string `json:"id"`
  Name string `json:"name"`
  Clustered bool `json:"clustered,omitempty"`
  ClusterID string `json:"clusterID,omitempty"`
  Memory int64 `json:"memory"`
  Cpus  int64 `json:"cpus"`
  Type VmmType `json:"type"`//OSv VMs have no mout points and the volume is the key bit.. cloud init is missing - but - ip information can be passed as required...
  Volumes []*VmmVolumeConfig `json:"volumes"`//volumes can be accessed over relevant sharing protocols...
  Interfaces []*VmmNetworkInterfaceConfig `json:"interfaces"`
  Disks []*VmmDiskConfig `json:"disks"`
  BootCmd string `json:"bootCmd,omitempty"`
  EntryPoint string `json:"entryPoint,omitempty"`
  AutoStart bool `json:"autoStart"`
}

type VmmDiskConfig struct {
  Size uint64 `json:"size"`
  IsRoot bool `json:"isRoot"`
  Storage *VmmDiskStorageConfig `json:"storage"`
}

type VmmDiskStorageConfig struct {
  Driver string `json:"driver"`
  ID string `json:"id"`
}

type VmmNetworkInterfaceConfig struct {
  Network string `json:"network"`
  Config *cloudconfig.MetaDataNetworkEthernetsConfig `json:"config"`
}

type VmmVolumeConfig struct {

}

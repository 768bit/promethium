package config

import "github.com/768bit/promethium/lib/cloudconfig"

type VmmType string

const (
	FirecrackerVmm    VmmType = "firecracker-standard"
	OSvFirecrackerVmm VmmType = "firecracker-osv"
	QemuVmm           VmmType = "qemu-standard"
)

type VmmConfig struct {
	ID        string             `json:"id"`
	Name      string             `json:"name"`
	Clustered bool               `json:"clustered,omitempty"`
	ClusterID string             `json:"clusterID,omitempty"`
	Memory    int64              `json:"memory"`
	Cpus      int64              `json:"cpus"`
	Type      VmmType            `json:"type"`    //OSv VMs have no mout points and the volume is the key bit.. cloud init is missing - but - ip information can be passed as required...
	Volumes   []*VmmVolumeConfig `json:"volumes"` //volumes can be accessed over relevant sharing protocols...
	Kernel    string             `json:"kernel"`
	//	Interfaces []*VmmNetworkInterfaceConfig `json:"interfaces"`
	Network    *VmmNetworkConfig `json:"network"`
	Disks      []*VmmDiskConfig  `json:"disks"`
	BootCmd    string            `json:"bootCmd,omitempty"`
	EntryPoint string            `json:"entryPoint,omitempty"`
	AutoStart  bool              `json:"autoStart"`
}

type VmmDiskConfig struct {
	IsRoot     bool   `json:"isRoot"`
	StorageURI string `json:"storageUri"`
}

type VmmNetworkConfig struct {
	Interfaces []*VmmNetworkInterfaceConfig                `json:"interfaces"`
	Bonds      cloudconfig.MetaDataNetworkBondsConfigMap   `json:"bonds"`
	Bridges    cloudconfig.MetaDataNetworkBridgesConfigMap `json:"bridges"`
	Vlans      cloudconfig.MetaDataNetworkVlansConfigMap   `json:"vlans"`
	Routes     []cloudconfig.MetaDataNetworkRoutesConfig   `json:"routes"`
}

type VmmNetworkInterfaceConfig struct {
	ID         string                                      `json:"id"`
	NetworkID  string                                      `json:"network"`
	MacAddress string                                      `json:"macAddress"`
	Config     *cloudconfig.MetaDataNetworkEthernetsConfig `json:"config"`
}

type VmmVolumeConfig struct {
}

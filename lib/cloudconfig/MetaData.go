package cloudconfig

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/teris-io/shortid"
	"gopkg.in/yaml.v2"
)

func NewMetaData(hostname string) *MetaData {
	sid, _ := shortid.Generate()
	id := fmt.Sprintf("vm-%s", sid)
	return &MetaData{
		InstanceID: id,
		Hostname:   hostname,
	}
}

func NewMetaDataWithNetworking(hostname string, networkConfig *MetaDataNetworkConfig) *MetaData {
	sid, _ := shortid.Generate()
	id := fmt.Sprintf("vm-%s", sid)
	networkConfig.Version = 2
	return &MetaData{
		InstanceID: id,
		Hostname:   hostname,
		Network:    networkConfig,
	}
}

type MetaData struct {
	InstanceID string                 `yaml:"instance-id" json:"instance-id"`
	Hostname   string                 `yaml:"hostname" json:"hostname"`
	Network    *MetaDataNetworkConfig `yaml:"network,omitempty,flow" json:"network,omitempty"`
}

type MetaDataNetworkConfig struct {
	Version   uint8                             `yaml:"version"`
	Ethernets MetaDataNetworkEthernetsConfigMap `yaml:"ethernets,omitempty,flow" json:"ethernets,omitempty"`
	Bonds     MetaDataNetworkBondsConfigMap     `yaml:"bonds,omitempty,flow" json:"bonds,omitempty"`
	Bridges   MetaDataNetworkBridgesConfigMap   `yaml:"bridges,omitempty,flow" json:"bridges,omitempty"`
	Vlans     MetaDataNetworkVlansConfigMap     `yaml:"vlans,omitempty,flow" json:"vlans,omitempty"`
	Routes    []MetaDataNetworkRoutesConfig     `yaml:"routes,omitempty,flow" json:"routes,omitempty"`
}

type MetaDataNetworkEthernetsConfigMap map[string]*MetaDataNetworkEthernetsConfig

type MetaDataNetworkEthernetsConfig struct {
	Match       *MetaDataNetworkEthernetsMatchConfig `yaml:"match,omitempty" json:"match,omitempty"`
	SetName     string                               `yaml:"set-name,omitempty" json:"set-name,omitempty"`
	Dhcp4       bool                                 `yaml:"dhcp4,omitempty" json:"dhcp4,omitempty"`
	Dhcp6       bool                                 `yaml:"dhcp6,omitempty" json:"dhcp6,omitempty"`
	Addresses   []string                             `yaml:"addresses,omitempty" json:"addresses,omitempty"`
	Gateway4    string                               `yaml:"gateway4,omitempty" json:"gateway4,omitempty"`
	Gateway6    string                               `yaml:"gateway6,omitempty" json:"gateway6,omitempty"`
	Nameservers *MetaDataNetworkEthernetsDNSConfig   `yaml:"nameservers,omitempty,flow" json:"nameservers,omitempty"`
	MTU         int32                                `yaml:"mtu,omitempty" json:"mtu,omitempty"`
	Routes      []MetaDataNetworkRoutesConfig        `yaml:"routes,omitempty" json:"routes,omitempty"`
}

type MetaDataNetworkEthernetsMatchConfig struct {
	MacAddress string `yaml:"macaddress,omitempty"`
	Driver     string `yaml:"driver,omitempty"`
	Name       string `yaml:"name,omitempty"`
}

type MetaDataNetworkEthernetsDNSConfig struct {
	Search    []string `yaml:"search,omitempty"`
	Addresses []string `yaml:"addresses,omitempty"`
}

type MetaDataNetworkBondsConfigMap map[string]*MetaDataNetworkBondsConfig

type MetaDataNetworkBondMode string

const (
	BondModeBalanceRoundRobin MetaDataNetworkBondMode = "balance-rr"
	BondModeActiveBackup      MetaDataNetworkBondMode = "active-backup"
	BondModeBalanceXOR        MetaDataNetworkBondMode = "balance-xor"
	BondModeBroadcast         MetaDataNetworkBondMode = "broadcast"
	BondModeLACP              MetaDataNetworkBondMode = "802.3ad"
	BondModeBalanceTLB        MetaDataNetworkBondMode = "balance-tlb"
	BondModeBalanceALB        MetaDataNetworkBondMode = "balance-alb"
)

type MetaDataNetworkBondLACPRate string

const (
	BondLACPRateSlow MetaDataNetworkBondLACPRate = "slow"
	BondLACPRateFast MetaDataNetworkBondLACPRate = "fast"
)

type MetaDataNetworkBondTransmitHashPolicy string

const (
	BondTransmitHashPolicyLayer2     MetaDataNetworkBondTransmitHashPolicy = "layer2"
	BondTransmitHashPolicyLayer2and3 MetaDataNetworkBondTransmitHashPolicy = "layer2+3"
	BondTransmitHashPolicyLayer3and4 MetaDataNetworkBondTransmitHashPolicy = "layer3+4"
	BondTransmitHashPolicyEncap2and3 MetaDataNetworkBondTransmitHashPolicy = "encap2+3"
	BondTransmitHashPolicyEncap3and4 MetaDataNetworkBondTransmitHashPolicy = "encap3+4"
)

type MetaDataNetworkBondAggregationSelectMode string

const (
	BondAggregationSelectModeStable    MetaDataNetworkBondAggregationSelectMode = "stable"
	BondAggregationSelectModeBandwidth MetaDataNetworkBondAggregationSelectMode = "bandwidth"
	BondAggregationSelectModeCount     MetaDataNetworkBondAggregationSelectMode = "count"
)

type MetaDataNetworkBondArpValidateMode string

const (
	BondArpValidateModeNone   MetaDataNetworkBondArpValidateMode = "none"
	BondArpValidateModeActive MetaDataNetworkBondArpValidateMode = "active"
	BondArpValidateModeBackup MetaDataNetworkBondArpValidateMode = "backup"
	BondArpValidateModeAll    MetaDataNetworkBondArpValidateMode = "all"
)

type MetaDataNetworkBondArpAllTargetsMode string

const (
	BondArpAllTargetsAny MetaDataNetworkBondArpAllTargetsMode = "any"
	BondArpAllTargetsAll MetaDataNetworkBondArpAllTargetsMode = "all"
)

type MetaDataNetworkBondFailoverMacPolicy string

const (
	BondFailoverMacPolicyNone   MetaDataNetworkBondFailoverMacPolicy = "none"
	BondFailoverMacPolicyActive MetaDataNetworkBondFailoverMacPolicy = "active"
	BondFailoverMacPolicyFollow MetaDataNetworkBondFailoverMacPolicy = "follow"
)

type MetaDataNetworkBondPrimaryReselectPolicy string

const (
	BondPrimaryReselectPolicyAlways  MetaDataNetworkBondPrimaryReselectPolicy = "always"
	BondPrimaryReselectPolicyBetter  MetaDataNetworkBondPrimaryReselectPolicy = "better"
	BondPrimaryReselectPolicyFailure MetaDataNetworkBondPrimaryReselectPolicy = "failure"
)

type MetaDataNetworkBondsConfig struct {
	Interfaces               []string                                 `yaml:"interfaces,omitempty,flow" json:"interfaces,omitempty"`
	Dhcp4                    bool                                     `yaml:"dhcp4,omitempty" json:"dhcp4,omitempty"`
	Dhcp6                    bool                                     `yaml:"dhcp6,omitempty" json:"dhcp6,omitempty"`
	Addresses                []string                                 `yaml:"addresses,omitempty" json:"addresses,omitempty"`
	Gateway4                 string                                   `yaml:"gateway4,omitempty" json:"gateway4,omitempty"`
	Gateway6                 string                                   `yaml:"gateway6,omitempty" json:"gateway6,omitempty"`
	Nameservers              *MetaDataNetworkEthernetsDNSConfig       `yaml:"nameservers,omitempty,flow" json:"nameservers,omitempty"`
	MTU                      int32                                    `yaml:"mtu,omitempty" json:"mtu,omitempty"`
	Routes                   []MetaDataNetworkRoutesConfig            `yaml:"routes,omitempty" json:"routes,omitempty"`
	Mode                     MetaDataNetworkBondMode                  `yaml:"mode,omitempty" json:"mode,omitempty"`
	LACPRate                 MetaDataNetworkBondLACPRate              `yaml:"lacp-rate,omitempty" json:"lacp-rate,omitempty"`
	MiiMonitorInterval       int32                                    `yaml:"mii-monitor-interval,omitempty" json:"mii-monitor-interval,omitempty"`
	MinLinks                 int32                                    `yaml:"min-links,omitempty" json:"min-links,omitempty"`
	TransmitHashPolicy       MetaDataNetworkBondTransmitHashPolicy    `yaml:"transmit-hash-policy,omitempty" json:"transmit-hash-policy,omitempty"`
	AggregationSelectionMode MetaDataNetworkBondAggregationSelectMode `yaml:"ad-select,omitempty" json:"ad-select,omitempty"`
	AllSlavesActive          bool                                     `yaml:"all-slaves-active,omitempty" json:"all-slaves-active,omitempty"`
	ArpInterval              int32                                    `yaml:"arp-interval,omitempty" json:"arp-interval,omitempty"`
	ArpIpTargets             []string                                 `yaml:"arp-ip-targets,omitempty" json:"arp-ip-targets,omitempty"`
	ArpValidate              MetaDataNetworkBondArpValidateMode       `yaml:"arp-validate,omitempty" json:"arp-validate,omitempty"`
	ArpAllTargets            MetaDataNetworkBondArpAllTargetsMode     `yaml:"arp-all-targets,omitempty" json:"arp-all-targets,omitempty"`
	UpDelay                  int32                                    `yaml:"up-delay,omitempty" json:"up-delay,omitempty"`
	DownDelay                int32                                    `yaml:"down-delay,omitempty" json:"down-delay,omitempty"`
	FailoverMacPolicy        MetaDataNetworkBondFailoverMacPolicy     `yaml:"fail-over-mac-policy,omitempty" json:"fail-over-mac-policy,omitempty"`
	GratuitousArp            int32                                    `yaml:"gratuitous-arp,omitempty" json:"gratuitous-arp,omitempty"`
	PacketsPerSlave          int32                                    `yaml:"packets-per-slave,omitempty" json:"packets-per-slave,omitempty"`
	PrimaryReselectPolicy    MetaDataNetworkBondPrimaryReselectPolicy `yaml:"primary-reselect-policy,omitempty" json:"primary-reselect-policy,omitempty"`
	LearnPacketInterval      int32                                    `yaml:"learn-packet-interval,omitempty" json:"learn-packet-interval,omitempty"`
}

type MetaDataNetworkBridgesConfigMap map[string]*MetaDataNetworkBridgesConfig

type MetaDataNetworkBridgesConfig struct {
	Interfaces   []string                           `yaml:"interfaces,omitempty,flow" json:"interfaces,omitempty"`
	Dhcp4        bool                               `yaml:"dhcp4,omitempty" json:"dhcp4,omitempty"`
	Dhcp6        bool                               `yaml:"dhcp6,omitempty" json:"dhcp6,omitempty"`
	Addresses    []string                           `yaml:"addresses,omitempty" json:"addresses,omitempty"`
	Gateway4     string                             `yaml:"gateway4,omitempty" json:"gateway4,omitempty"`
	Gateway6     string                             `yaml:"gateway6,omitempty" json:"gateway6,omitempty"`
	Nameservers  *MetaDataNetworkEthernetsDNSConfig `yaml:"nameservers,omitempty,flow" json:"nameservers,omitempty"`
	MTU          int32                              `yaml:"mtu,omitempty" json:"mtu,omitempty"`
	Routes       []MetaDataNetworkRoutesConfig      `yaml:"routes,omitempty" json:"routes,omitempty"`
	AgeingTime   int32                              `yaml:"ageing-time,omitempty" json:"ageing-time,omitempty"`
	Priority     int32                              `yaml:"priority,omitempty" json:"priority,omitempty"`
	ForwardDelay int32                              `yaml:"forward-delay,omitempty" json:"forward-delay,omitempty"`
	HelloTime    int32                              `yaml:"hello-time,omitempty" json:"hello-time,omitempty"`
	MaxAge       int32                              `yaml:"max-age,omitempty" json:"max-age,omitempty"`
	PathCost     int32                              `yaml:"path-cost,omitempty" json:"path-cost,omitempty"`
	STP          bool                               `yaml:"stp,omitempty" json:"stp,omitempty"`
}

type MetaDataNetworkVlansConfigMap map[string]*MetaDataNetworkVlansConfig

type MetaDataNetworkVlansConfig struct {
	ID          uint16                             `yaml:"id,omitempty" json:"id,omitempty"`
	Link        string                             `yaml:"link,omitempty" json:"link,omitempty"`
	Dhcp4       bool                               `yaml:"dhcp4,omitempty" json:"dhcp4,omitempty"`
	Dhcp6       bool                               `yaml:"dhcp6,omitempty" json:"dhcp6,omitempty"`
	Addresses   []string                           `yaml:"addresses,omitempty" json:"addresses,omitempty"`
	Gateway4    string                             `yaml:"gateway4,omitempty" json:"gateway4,omitempty"`
	Gateway6    string                             `yaml:"gateway6,omitempty" json:"gateway6,omitempty"`
	Nameservers *MetaDataNetworkEthernetsDNSConfig `yaml:"nameservers,omitempty,flow" json:"nameservers,omitempty"`
	MTU         int32                              `yaml:"mtu,omitempty" json:"mtu,omitempty"`
	Routes      []MetaDataNetworkRoutesConfig      `yaml:"routes,omitempty" json:"routes,omitempty"`
}

type MetaDataNetworkRoutesConfig struct {
	To     string `yaml:"to,omitempty" json:"to,omitempty"`
	Via    string `yaml:"via,omitempty" json:"via,omitempty"`
	Metric int32  `yaml:"metric,omitempty" json:"metric,omitempty"`
}

func (md *MetaData) WriteMetaData(dest string) error {
	x, err := yaml.Marshal(md)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(dest, x, 0660)
}

func (md *MetaData) WriteMetaDataJSON(dest string) error {
	x, err := json.Marshal(md)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(dest, x, 0660)
}

func (md *MetaData) Ethernets(networks MetaDataNetworkEthernetsConfigMap) {
	if md.Network == nil {
		md.Network = &MetaDataNetworkConfig{
			Version: 2,
		}
	}
	md.Network.Ethernets = networks
}

func (md *MetaData) Bonds(bonds MetaDataNetworkBondsConfigMap) {
	if md.Network == nil {
		md.Network = &MetaDataNetworkConfig{
			Version: 2,
		}
	}
	md.Network.Bonds = bonds
}

func (md *MetaData) Bridges(bridges MetaDataNetworkBridgesConfigMap) {
	if md.Network == nil {
		md.Network = &MetaDataNetworkConfig{
			Version: 2,
		}
	}
	md.Network.Bridges = bridges
}

func (md *MetaData) Vlans(vlans MetaDataNetworkVlansConfigMap) {
	if md.Network == nil {
		md.Network = &MetaDataNetworkConfig{
			Version: 2,
		}
	}
	md.Network.Vlans = vlans
}

func (md *MetaData) Routes(routes []MetaDataNetworkRoutesConfig) {
	if md.Network == nil {
		md.Network = &MetaDataNetworkConfig{
			Version: 2,
		}
	}
	md.Network.Routes = routes
}

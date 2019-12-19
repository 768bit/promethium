package cloudconfig

import (
  "encoding/json"
  "fmt"
  "github.com/teris-io/shortid"
  "gopkg.in/yaml.v2"
  "io/ioutil"
)

func NewMetaData(hostname string) (*MetaData) {
  sid, _ := shortid.Generate()
  id := fmt.Sprintf("vm-%s", sid)
  return &MetaData{
    InstanceID:id,
    Hostname: hostname,
  }
}

func NewMetaDataWithNetworking(hostname string, networkConfig *MetaDataNetworkConfig) (*MetaData) {
  sid, _ := shortid.Generate()
  id := fmt.Sprintf("vm-%s", sid)
  networkConfig.Version = 2
  return &MetaData{
    InstanceID:id,
    Hostname: hostname,
    Network:networkConfig,
  }
}

type MetaData struct {
  InstanceID string `yaml:"instance-id" json:"instance-id"`
  Hostname   string `yaml:"hostname" json:"hostname"`
  Network    *MetaDataNetworkConfig `yaml:"network,omitempty,flow" json:"network,omitempty"`
}

type MetaDataNetworkConfig struct {
  Version uint8 `yaml:"version"`
  Ethernets MetaDataNetworkEthernetsConfigMap `yaml:"ethernets,omitempty,flow" json:"ethernets,omitempty"`
  Bonds    MetaDataNetworkBondsConfigMap `yaml:"bonds,omitempty,flow" json:"bonds,omitempty"`
  Bridges    MetaDataNetworkBridgesConfigMap `yaml:"bridges,omitempty,flow" json:"bridges,omitempty"`
  Vlans    MetaDataNetworkVlansConfigMap `yaml:"vlans,omitempty,flow" json:"vlans,omitempty"`
  Routes    []MetaDataNetworkRoutesConfig `yaml:"routes,omitempty,flow" json:"routes,omitempty"`
}

type MetaDataNetworkEthernetsConfigMap map[string]*MetaDataNetworkEthernetsConfig

type MetaDataNetworkEthernetsConfig struct {
  Match *MetaDataNetworkEthernetsMatchConfig `yaml:"match,omitempty" json:"match,omitempty"`
  SetName string `yaml:"set-name,omitempty" json:"set-name,omitempty"`
  Dhcp4 bool `yaml:"dhcp4,omitempty" json:"dhcp4,omitempty"`
  Dhcp6 bool `yaml:"dhcp6,omitempty" json:"dhcp6,omitempty"`
  Addresses []string `yaml:"addresses,omitempty" json:"addresses,omitempty"`
  Gateway4 string `yaml:"gateway4,omitempty" json:"gateway4,omitempty"`
  Gateway6 string `yaml:"gateway6,omitempty" json:"gateway6,omitempty"`
  Nameservers *MetaDataNetworkEthernetsDNSConfig `yaml:"nameservers,omitempty,flow" json:"nameservers,omitempty"`
  MTU uint16 `yaml:"mtu,omitempty" json:"mtu,omitempty"`
}

type MetaDataNetworkEthernetsMatchConfig struct {
  MacAddress string `yaml:"macaddress,omitempty"`
  Driver string `yaml:"driver,omitempty"`
  Name string   `yaml:"name,omitempty"`
}

type MetaDataNetworkEthernetsDNSConfig struct {
  Search []string    `yaml:"search,omitempty"`
  Addresses []string `yaml:"addresses,omitempty"`
}

type MetaDataNetworkBondsConfigMap map[string]*MetaDataNetworkBondsConfig

type MetaDataNetworkBondsConfig struct {
  Interfaces []string `yaml:"interfaces,omitempty,flow" json:"interfaces,omitempty"`
  Dhcp4 bool `yaml:"dhcp4,omitempty" json:"dhcp4,omitempty"`
  Dhcp6 bool `yaml:"dhcp6,omitempty" json:"dhcp6,omitempty"`
  Addresses []string `yaml:"addresses,omitempty" json:"addresses,omitempty"`
  Gateway4 string `yaml:"gateway4,omitempty" json:"gateway4,omitempty"`
  Gateway6 string `yaml:"gateway6,omitempty" json:"gateway6,omitempty"`
  Nameservers *MetaDataNetworkEthernetsDNSConfig `yaml:"nameservers,omitempty,flow" json:"nameservers,omitempty"`
}

type MetaDataNetworkBridgesConfigMap map[string]*MetaDataNetworkBridgesConfig

type MetaDataNetworkBridgesConfig struct {
  Interfaces []string `yaml:"interfaces,omitempty,flow" json:"interfaces,omitempty"`
  Dhcp4 bool `yaml:"dhcp4,omitempty" json:"dhcp4,omitempty"`
  Dhcp6 bool `yaml:"dhcp6,omitempty" json:"dhcp6,omitempty"`
  Addresses []string `yaml:"addresses,omitempty" json:"addresses,omitempty"`
  Gateway4 string `yaml:"gateway4,omitempty" json:"gateway4,omitempty"`
  Gateway6 string `yaml:"gateway6,omitempty" json:"gateway6,omitempty"`
  Nameservers *MetaDataNetworkEthernetsDNSConfig `yaml:"nameservers,omitempty,flow" json:"nameservers,omitempty"`
}

type MetaDataNetworkVlansConfigMap map[string]*MetaDataNetworkVlansConfig

type MetaDataNetworkVlansConfig struct {
  ID uint16 `yaml:"id,omitempty" json:"id,omitempty"`
  Link string `yaml:"link,omitempty" json:"link,omitempty"`
  Dhcp4 bool `yaml:"dhcp4,omitempty" json:"dhcp4,omitempty"`
  Dhcp6 bool `yaml:"dhcp6,omitempty" json:"dhcp6,omitempty"`
  Addresses []string `yaml:"addresses,omitempty" json:"addresses,omitempty"`
  Gateway4 string `yaml:"gateway4,omitempty" json:"gateway4,omitempty"`
  Gateway6 string `yaml:"gateway6,omitempty" json:"gateway6,omitempty"`
  Nameservers *MetaDataNetworkEthernetsDNSConfig `yaml:"nameservers,omitempty,flow" json:"nameservers,omitempty"`
  MTU uint16 `yaml:"mtu,omitempty" json:"mtu,omitempty"`
}

type MetaDataNetworkRoutesConfig struct {
  To string `yaml:"to,omitempty" json:"to,omitempty"`
  Via string `yaml:"via,omitempty" json:"via,omitempty"`
  Metric uint `yaml:"metric,omitempty" json:"metric,omitempty"`
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

package networking

import (
  "fmt"
  "github.com/milosgajdos83/tenus"
  "net"
)

func NewLinuxBridge(config *NetworkConfig) (*LinuxBridge, error) {
  br := &LinuxBridge{
    id:                 config.ID,
    name:               config.Name,
    enabled:config.Enabled,
    config:config,
    interfaces:         map[string]*LinuxTapInterface{},
    masterInterfaceConfig: config.MasterInterface,
  }
  return br.init()
}

type LinuxBridge struct {
  id string
  name string
  config *NetworkConfig
  enabled bool
  interfaces map[string]*LinuxTapInterface
  masterInterfaceConfig *BridgeMasterInterfaceConfig
  masterInterface tenus.Linker
  bridgeVlanInterface tenus.Linker
  bridgeInterface tenus.Bridger
}

func (bridge *LinuxBridge) init() (*LinuxBridge, error) {
  //initialise the bridge witht he provided physical interfaces - does it already exist etc..
  //if we cant get the bridge create it...
  _, err := bridge.getBridge()
  if err != nil {
    return nil, err
  }

  //bring the bridge and its master interface up now...

  if err := bridge.applyIpConfig(); err != nil {
    return nil, err
  } else if err := bridge.bringUpInterfaces(); err != nil {
    return nil, err
  }
  return bridge, nil

}

func (bridge *LinuxBridge) getBridge() (tenus.Bridger, error) {
  //so we have an id - lets retrieve the bridge device with the id
  br, err := tenus.BridgeFromName(bridge.id)
  if err != nil {
    return bridge.createBridge()
  }
  bridge.bridgeInterface = br
  if err = bridge.assignMasterInterface(); err != nil {
    return nil, err
  }
  return br, err

}

func (bridge *LinuxBridge) createBridge() (tenus.Bridger, error) {
  //so we have an id - lets retrieve the bridge device with the id
  br, err := tenus.NewBridgeWithName(bridge.id)
  if err != nil {
    return nil, err
  }
  //check the physical interfaces are assigned correctly...
  bridge.bridgeInterface = br
  if err = bridge.assignMasterInterface(); err != nil {
    return nil, err
  }
  return br, err

}

func (bridge *LinuxBridge) assignMasterInterface() (error) {

  if bridge.masterInterfaceConfig != nil {
    //get the master interface...
    dl, err := tenus.NewLinkFrom(bridge.masterInterfaceConfig.Device)
    if err != nil {
      return err
    }
    bridge.masterInterface = dl
    return tenus.AddToBridge(bridge.masterInterface.NetInterface(), bridge.bridgeInterface.NetInterface())
  }

  return nil

}

func (bridge *LinuxBridge) applyIpConfig() (error) {
  //if there is an IP config for this bridge lets process it..
  //if vlan is set then a new vlan interface is created to split off from the bridge
  var targetInterface tenus.Linker = bridge.bridgeInterface
  if bridge.config != nil && bridge.enabled && bridge.config.IPV4 != nil && bridge.config.IPV4.Enabled {
    if bridge.config.IPV4.Vlan > 0 && bridge.config.IPV4.Vlan <= 4096 {
      //vlan enabled bridge interface.. need a new interface to add as slave to bridge
      vlan, err := tenus.NewVlanLinkWithOptions(bridge.id, tenus.VlanOptions{Id: bridge.config.IPV4.Vlan, Dev: fmt.Sprintf("%s-vlan%d", bridge.id, bridge.config.IPV4.Vlan)})
      if err != nil {
        return err
      }
      bridge.bridgeVlanInterface = vlan
      targetInterface = bridge.bridgeVlanInterface
    }

    //now we apply the selected config...

    if !bridge.config.IPV4.DHCP {
      //check the supplied ip stuff..
      ip, ipNet, err := net.ParseCIDR(bridge.config.IPV4.Address)
      if err != nil {
        return err
      }

      if err := targetInterface.SetLinkIp(ip, ipNet); err != nil {
        return err
      } else if bridge.config.IPV4.Gateway != "" {
        gw := net.ParseIP(bridge.config.IPV4.Gateway)
        err := targetInterface.SetLinkDefaultGw(&gw)
        if err != nil {
          return err
        }
      }
    }

    return targetInterface.SetLinkUp()

  }

  return nil

}

func (bridge *LinuxBridge) bringUpInterfaces() (error) {
  if bridge.masterInterface != nil {
    if err := bridge.masterInterface.SetLinkUp(); err != nil {
      return err
    }
  }
  err := bridge.bridgeInterface.SetLinkUp()
  if err != nil {
    return err
  } else if bridge.bridgeVlanInterface != nil {
    return bridge.bridgeVlanInterface.SetLinkUp()
  }
  return nil
}

func (bridge *LinuxBridge) bringDownInterfaces() (error) {
  if bridge.masterInterface != nil {
    if err := bridge.masterInterface.SetLinkDown(); err != nil {
      return err
    }
  }
  err := bridge.bridgeInterface.SetLinkDown()
  if err != nil {
    return err
  } else if bridge.bridgeVlanInterface != nil {
    return bridge.bridgeVlanInterface.SetLinkDown()
  }
  return nil
}

func (bridge *LinuxBridge) GetId() string {
  return bridge.id
}

func (bridge *LinuxBridge) GetName() string {
  return bridge.name
}

func (bridge *LinuxBridge) SetName(name string) {
  bridge.name = name
}

func (bridge *LinuxBridge) CreateInterface(vmid string, index uint, vlan uint16) (NetworkInterface, error) {
  //create an interface that will be used by a VM
  if vlan > 0 {
    iface, err := NewLinuxTapInterfaceVlan(vmid, index, true, bridge, vlan)
    if err != nil {
      return nil, err
    }
    bridge.interfaces[iface.interfaceName] = iface
    return iface, nil
  }
  iface, err := NewLinuxTapInterface(vmid, index, true, bridge)
  if err != nil {
    return nil, err
  }
  bridge.interfaces[iface.interfaceName] = iface
  err = bridge.bridgeInterface.AddSlaveIfc(iface.link.NetInterface())
  if err != nil {
    return nil, err
  }

  return iface, nil
}

func (bridge *LinuxBridge) GetInterface(interfaceId string) (NetworkInterface, error) {
  //create an interface that will be used by a VM
  return nil, nil
}

func (bridge *LinuxBridge) DestroyInterface(interfaceId string) {
  //create an interface that will be used by a VM
}

package networking

import (
	"fmt"

	"github.com/768bit/vutils"
	"github.com/milosgajdos83/tenus"
)

//create tap interfaces that can be used by firecracker/osv.. tap interfaces need to be bound to a bridge...

func NewLinuxTapInterface(vmID string, tapID uint, disableIPV6 bool, bridge *LinuxBridge) (*LinuxTapInterface, error) {
	iface := &LinuxTapInterface{
		vmID:         vmID,
		tapID:        tapID,
		ipV6Disabled: disableIPV6,
		bridge:       bridge,
	}
	if err := iface.init(); err != nil {
		return nil, err
	}
	return iface, nil
}
func NewLinuxTapInterfaceVlan(vmID string, tapID uint, disableIPV6 bool, bridge *LinuxBridge, vlan uint16) (*LinuxTapInterface, error) {
	iface := &LinuxTapInterface{
		vmID:         vmID,
		tapID:        tapID,
		ipV6Disabled: disableIPV6,
		bridge:       bridge,
	}
	if err := iface.initVlan(vlan); err != nil {
		return nil, err
	}
	return iface, nil
}

type LinuxTapInterface struct {
	tapID         uint
	vmID          string
	bridge        *LinuxBridge
	ipV6Disabled  bool
	interfaceName string
	link          tenus.Linker
	vlan          uint16
}

func (iface *LinuxTapInterface) GetId() string {
	return iface.interfaceName
}

func (iface *LinuxTapInterface) Enable() error {
	panic("implement me")
}

func (iface *LinuxTapInterface) Disable() error {
	panic("implement me")
}

func (iface *LinuxTapInterface) init() error {
	iface.interfaceName = fmt.Sprintf("%s%d", iface.vmID, iface.tapID)

	dl, err := tenus.NewLinkFrom(iface.interfaceName)
	if err != nil {
		cmd := vutils.Exec.CreateAsyncCommand("ip", false, "tuntap", "add", "dev", iface.interfaceName, "mode", "tap").Sudo()
		err := cmd.StartAndWait()
		if err != nil {
			return err
		}
		dl, err = tenus.NewLinkFrom(iface.interfaceName)
		if err != nil {
			return err
		}
	}
	iface.link = dl
	return nil
}

func (iface *LinuxTapInterface) initVlan(vlan uint16) error {
	iface.interfaceName = fmt.Sprintf("%s-%d", iface.vmID, iface.tapID)
	iface.vlan = vlan
	dl, err := tenus.NewVlanLinkWithOptions(iface.bridge.id, tenus.VlanOptions{Id: vlan, Dev: iface.interfaceName})
	if err != nil {
		return err
	}
	iface.link = dl
	return nil
}

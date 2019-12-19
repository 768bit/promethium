package networking

import "testing"

func TestLinuxBridge(t *testing.T) {
  br, err := NewLinuxBridge(&NetworkConfig{
    ID:              "extbr0",
    Name:            "extbr0",
    Type:            LinuxBridgeDriver,
    Enabled:         true,
    MasterInterface: nil,
    IPV4:            nil,
  })

  if err != nil {
    t.Errorf("Error creating linux bridge %s", err.Error())
    return
  }

  _, err = br.CreateInterface("test", 0, 0)

  if err != nil {
    t.Errorf("Error creating linux tap interface %s", err.Error())
    return
  }

}

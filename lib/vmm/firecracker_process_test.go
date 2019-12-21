package vmm

import (
	"io/ioutil"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/768bit/promethium/lib/cloudconfig"
	"github.com/768bit/promethium/lib/images"
	"github.com/768bit/promethium/lib/networking"
)

func TestProcessCreate(t *testing.T) {
	pkgEntry := filepath.Join(getRootPath(), "pkg")
	fcp, err := NewFireCrackerProcess("TESTER", "test4", "/cli/cli.so -h", pkgEntry, 1, 512, 64000000, false)
	if err != nil {
		t.Errorf("Error creating new process: %s", err.Error())
		return
	}
	err = fcp.Start()
	if err != nil {
		t.Errorf("Error starting jailed VM %s", err.Error())
	}
}

func TestProcessCreateImg(t *testing.T) {
	kernelPath := filepath.Join(getRootPath(), "workspace", "ubuntu", "bionic", "kernel.elf")
	imagePath := filepath.Join(getRootPath(), "workspace", "ubuntu", "bionic", "root.qcow2")
	bootPath := filepath.Join(getRootPath(), "workspace", "ubuntu", "bionic", "boot")
	ba, err := ioutil.ReadFile(bootPath)
	if err != nil {
		t.Errorf("Error reading boot: %s", err.Error())
		return
	}
	bootParams := strings.TrimSpace(string(ba))
	fcp, err := NewFireCrackerProcessImg("TESTER", "test4", bootParams, 1, 512, kernelPath, []string{imagePath}, nil, false)
	if err != nil {
		t.Errorf("Error creating new process: %s", err.Error())
		return
	}
	err = fcp.Start()
	if err != nil {
		t.Errorf("Error starting jailed VM %s", err.Error())
	}
	go func() {
		time.Sleep(15 * time.Second)
		fcp.Send("reboot\n")
	}()
	fcp.Wait()
}

func TestProcessCreateImg2(t *testing.T) {
	kernelPath := filepath.Join(getRootPath(), "workspace", "ubuntu", "bionic", "kernel.elf")
	imagePath := filepath.Join(getRootPath(), "workspace", "ubuntu", "bionic", "root.qcow2")
	bootPath := filepath.Join(getRootPath(), "workspace", "ubuntu", "bionic", "boot")
	ba, err := ioutil.ReadFile(bootPath)
	if err != nil {
		t.Errorf("Error reading boot: %s", err.Error())
		return
	}
	bootParams := strings.TrimSpace(string(ba))
	fcp, err := NewFireCrackerProcessImg("TESTER2", "test4", bootParams, 1, 512, kernelPath, []string{imagePath}, nil, false)
	if err != nil {
		t.Errorf("Error creating new process: %s", err.Error())
		return
	}
	err = fcp.Start()
	if err != nil {
		t.Errorf("Error starting jailed VM %s", err.Error())
	}
	go func() {
		time.Sleep(15 * time.Second)
		err := fcp.Stop()
		if err != nil {
			t.Errorf("Error stopping jailed VM %s", err.Error())
		}
	}()
	fcp.Wait()
}

func TestProcessCreateImg3(t *testing.T) {
	kernelPath := filepath.Join(getRootPath(), "workspace", "ubuntu", "bionic", "kernel.elf")
	imagePath := filepath.Join(getRootPath(), "workspace", "ubuntu", "bionic", "root.qcow2")
	bootPath := filepath.Join(getRootPath(), "workspace", "ubuntu", "bionic", "boot")
	ba, err := ioutil.ReadFile(bootPath)
	if err != nil {
		t.Errorf("Error reading boot: %s", err.Error())
		return
	}
	bootParams := strings.TrimSpace(string(ba))
	fcp, err := NewFireCrackerProcessImg("TESTER3", "test4", bootParams, 1, 512, kernelPath, []string{imagePath}, nil, false)
	if err != nil {
		t.Errorf("Error creating new process: %s", err.Error())
		return
	}
	err = fcp.Start()
	if err != nil {
		t.Errorf("Error starting jailed VM %s", err.Error())
	}
	go func() {
		time.Sleep(15 * time.Second)
		err := fcp.Shutdown()
		if err != nil {
			t.Errorf("Error shuttingdown jailed VM %s", err.Error())
		}
	}()
	fcp.Wait()
}

func TestProcessCreateImg4(t *testing.T) {
	kernelPath := filepath.Join(getRootPath(), "workspace", "ubuntu", "bionic", "kernel.elf")
	imagePath := filepath.Join(getRootPath(), "workspace", "ubuntu", "bionic", "root.qcow2")
	bootPath := filepath.Join(getRootPath(), "workspace", "ubuntu", "bionic", "boot")
	ba, err := ioutil.ReadFile(bootPath)
	if err != nil {
		t.Errorf("Error reading boot: %s", err.Error())
		return
	}
	bootParams := strings.TrimSpace(string(ba))
	br, err := networking.NewLinuxBridge(&networking.NetworkConfig{
		ID:              "extbr0",
		Name:            "extbr0",
		Type:            networking.LinuxBridgeDriver,
		Enabled:         true,
		MasterInterface: nil,
		IPV4:            nil,
	})

	if err != nil {
		t.Errorf("Error creating linux bridge %s", err.Error())
		return
	}

	iface, err := br.CreateInterface("testf", 0, 0)

	if err != nil {
		t.Errorf("Error creating linux tap interface %s", err.Error())
		return
	}
	fcp, err := NewFireCrackerProcessImg("TESTER4", "test4", bootParams, 1, 512, kernelPath, []string{imagePath}, []string{iface.GetId()}, false)
	if err != nil {
		t.Errorf("Error creating new process: %s", err.Error())
		return
	}
	err = fcp.Start()
	if err != nil {
		t.Errorf("Error starting jailed VM %s", err.Error())
	}
	go func() {
		time.Sleep(15 * time.Second)
		err := fcp.Send("ping www.google.com\n")
		if err != nil {
			t.Errorf("Error sending commandd to jailed VM %s", err.Error())
		}
		time.Sleep(15 * time.Second)
		err = fcp.Shutdown()
		if err != nil {
			t.Errorf("Error shuttingdown jailed VM %s", err.Error())
		}
	}()
	fcp.Wait()
}

func TestProcessCreateImg5(t *testing.T) {
	//these are the paths to, kernel, the qcow2 image with the root partition, and the boot paramaters
	// used.. VMM.Start() will make use of a lot of this
	kernelPath := filepath.Join(getRootPath(), "workspace", "ubuntu", "bionic-cloud", "kernel.elf")
	imagePath := filepath.Join(getRootPath(), "workspace", "ubuntu", "bionic-cloud", "root.qcow2")
	bootPath := filepath.Join(getRootPath(), "workspace", "ubuntu", "bionic-cloud", "boot")
	//read boot params fikle into a string
	ba, err := ioutil.ReadFile(bootPath)
	if err != nil {
		t.Errorf("Error reading boot: %s", err.Error())
		return
	}
	bootParams := strings.TrimSpace(string(ba))
	//create or bind to an existing linux bridge
	br, err := networking.NewLinuxBridge(&networking.NetworkConfig{
		ID:              "extbr1",
		Name:            "extbr1",
		Type:            networking.LinuxBridgeDriver,
		Enabled:         true,
		MasterInterface: nil,
		IPV4:            nil,
	})

	if err != nil {
		t.Errorf("Error creating linux bridge %s", err.Error())
		return
	}
	netConf := &cloudconfig.MetaDataNetworkConfig{
		Ethernets: map[string]*cloudconfig.MetaDataNetworkEthernetsConfig{
			"eth0": {
				Match: &cloudconfig.MetaDataNetworkEthernetsMatchConfig{
					Name: "eth0",
				},
				Addresses: []string{"192.168.197.124/24"},
				Gateway4:  "192.168.197.1",
				Nameservers: &cloudconfig.MetaDataNetworkEthernetsDNSConfig{
					Addresses: []string{"8.8.8.8", "8.8.4.4"},
				},
			},
		},
	}
	udatas := &cloudconfig.UserData{
		PackageUpdate:  true,
		PackageUpgrade: true,
		Users: []*cloudconfig.UserDataUserConfig{
			{
				Name: "craig",
				SshAuthorisedKeys: []string{
					"ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQDdi3LLLHu7ZFUG5PDAlwDQgMYHbG+vbjBGMwVr6E3foeIiVaa5EFQa/nWTb1f86DV2aOV2fmSj36QKWho84QcbwV67d/WTtGlPYHfeMEffdRPFx32dEC9CH3XxtZmMNDWDi/IgE8ZdEiF8EFbzbXuHwG2Et/606jP549tsyUSnfrDp+uZaAxFSLkHwDitm2Heoc1ur+rTo3PrkkF7Z6GZDE/vJs+k/TpRuEhUAaTOgLzX0met7iyJcP7/sQkR/F1keUC+s2/sFeFvATLWNVkyOZYvulYQUdk4ObnR51V1sXRD9AVfy7f6PYj5bNHt4mXN0PsSfSe6uLDjIPknYR3LF craig@skylaker",
				},
				Sudo:  "ALL=(ALL) NOPASSWD:ALL",
				Shell: "/bin/bash",
			},
		},
	}
	ud, err := images.MakeCloudInitImageBuilt("testing", netConf, udatas)
	if err != nil {
		t.Error(err)
	}
	iface, err := br.CreateInterface("testg", 0, 0)

	if err != nil {
		t.Errorf("Error creating linux tap interface %s", err.Error())
		return
	}
	fcp, err := NewFireCrackerProcessImg("TESTER5", "test4", bootParams, 1, 512, kernelPath, []string{imagePath}, []string{iface.GetId()}, false)
	if err != nil {
		t.Errorf("Error creating new process: %s", err.Error())
		return
	}
	fcp.SetCloudInit(ud)
	err = fcp.Start()
	if err != nil {
		t.Errorf("Error starting jailed VM %s", err.Error())
	}
	//go func() {
	//  time.Sleep(15 * time.Second)
	//  err := fcp.Send("ping www.google.com\n")
	//  if err != nil {
	//    t.Errorf("Error sending commandd to jailed VM %s", err.Error())
	//  }
	//  time.Sleep(15 * time.Second)
	//  err = fcp.Shutdown()
	//  if err != nil {
	//    t.Errorf("Error shuttingdown jailed VM %s", err.Error())
	//  }
	//}()
	fcp.Wait()
}

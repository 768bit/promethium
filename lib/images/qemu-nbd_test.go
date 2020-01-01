package images

import (
	"fmt"
	"github.com/768bit/vutils"
	"os"
	"path/filepath"
	"testing"
)

func TestQcowCreateDestroy(t *testing.T) {
	ni, _ := vutils.UUID.MakeUUIDString()
	tpath := filepath.Join(os.TempDir(), fmt.Sprintf("%s.qcow2", ni))
	fmt.Println("Creating " + tpath)
	qi, err := CreateNewQemuImage(tpath, 400*1024*1024)
	if err != nil {
		t.Error(err)
	} else {
		err = qi.Destroy()
		if err != nil {
			t.Error(err)
		}
	}
}

func TestQcowCreateResize(t *testing.T) {
	ni, _ := vutils.UUID.MakeUUIDString()
	tpath := filepath.Join(os.TempDir(), fmt.Sprintf("%s.qcow2", ni))
	fmt.Println("Creating " + tpath)
	qi, err := CreateNewQemuImage(tpath, 400*1024*1024)
	if err != nil {
		t.Error(err)
	} else {
		defer qi.Destroy()
		fmt.Println("Resizing " + tpath)
		newSize := uint64(110 * 1024 * 1024)
		if err := qi.Resize(newSize); err != nil {
			t.Error(err)
		} else if qi.VirtualSize() != newSize {
			t.Error("The new size of the image is incorrect")
		}
	}
}

func TestQcowCreateConnect(t *testing.T) {
	ni, _ := vutils.UUID.MakeUUIDString()
	tpath := filepath.Join(os.TempDir(), fmt.Sprintf("%s.qcow2", ni))
	fmt.Println("Creating " + tpath)
	qi, err := CreateNewQemuImage(tpath, 400*1024*1024)
	if err != nil {
		t.Error(err)
	} else {
		defer qi.Destroy()
		err := qi.Connect()
		if err != nil {
			t.Error(err)
		}
		err = qi.Disconnect()
		if err != nil {
			t.Error(err)
		}
	}
}

func TestQcowCreateNewGptSingle(t *testing.T) {
	ni, _ := vutils.UUID.MakeUUIDString()
	tpath := filepath.Join(os.TempDir(), fmt.Sprintf("%s.qcow2", ni))
	fmt.Println("Creating " + tpath)
	qi, err := CreateNewQemuImage(tpath, 400*1024*1024)
	if err != nil {
		t.Error(err)
	} else {
		defer qi.Destroy()
		err := qi.Connect()
		if err != nil {
			t.Error(err)
			return
		}
		err = qi.MakeGpt()
		if err != nil {
			t.Error(err)
			return
		}
		_, err = qi.CreatePartition("full", "100%", LinuxFilesystem, true)
		if err != nil {
			t.Error(err)
			return
		}
		err = qi.WriteTable()
		if err != nil {
			t.Error(err)
			return
		}
		prt, err := qi.GetPartition(1)
		if err != nil {
			t.Error(err)
			return
		}
		err = prt.MakeFilesystem(FS_EXT4)
		if err != nil {
			t.Error(err)
			return
		}

		err = qi.Disconnect()
		if err != nil {
			t.Error(err)
			return
		}
	}
}

func TestQcowCreateNewGptSingleWipeDisk(t *testing.T) {
	ni, _ := vutils.UUID.MakeUUIDString()
	tpath := filepath.Join(os.TempDir(), fmt.Sprintf("%s.qcow2", ni))
	fmt.Println("Creating " + tpath)
	qi, err := CreateNewQemuImage(tpath, 400*1024*1024)
	if err != nil {
		t.Error(err)
	} else {
		defer qi.Destroy()
		err := qi.Connect()
		if err != nil {
			t.Error(err)
			return
		}
		err = qi.MakeGpt()
		if err != nil {
			t.Error(err)
			return
		}
		_, err = qi.CreatePartition("full", "100%", LinuxFilesystem, true)
		if err != nil {
			t.Error(err)
			return
		}
		err = qi.WriteTable()
		if err != nil {
			t.Error(err)
			return
		}
		prt, err := qi.GetPartition(1)
		if err != nil {
			t.Error(err)
			return
		}
		err = prt.MakeFilesystem(FS_EXT4)
		if err != nil {
			t.Error(err)
			return
		}
		err = qi.SecureWipe(WipeImageSinglePassVerify)
		if err != nil {
			qi.Disconnect()
			t.Error(err)
			return
		}
		sz, _ := qi.SizeOnDisk()
		fmt.Println("Wiped. Disk Size:", sz)
		err = qi.Disconnect()
		if err != nil {
			t.Error(err)
			return
		}
	}
}

func TestQcowCreateNewRawGptSingleWipeDisk(t *testing.T) {
	ni, _ := vutils.UUID.MakeUUIDString()
	tpath := filepath.Join(os.TempDir(), fmt.Sprintf("%s.img", ni))
	fmt.Println("Creating " + tpath)
	qi, err := CreateNewQemuImageFormat(tpath, 400*1024*1024, "raw")
	if err != nil {
		t.Error(err)
	} else {
		defer qi.Destroy()
		err := qi.Connect()
		if err != nil {
			t.Error(err)
			return
		}
		err = qi.MakeGpt()
		if err != nil {
			t.Error(err)
			return
		}
		_, err = qi.CreatePartition("full", "100%", LinuxFilesystem, true)
		if err != nil {
			t.Error(err)
			return
		}
		err = qi.WriteTable()
		if err != nil {
			t.Error(err)
			return
		}
		prt, err := qi.GetPartition(1)
		if err != nil {
			t.Error(err)
			return
		}
		err = prt.MakeFilesystem(FS_EXT4)
		if err != nil {
			t.Error(err)
			return
		}
		err = qi.SecureWipe(WipeImageSinglePassVerify)
		if err != nil {
			qi.Disconnect()
			t.Error(err)
			return
		}
		sz, _ := qi.SizeOnDisk()
		fmt.Println("Wiped. Disk Size:", sz)
		err = qi.Disconnect()
		if err != nil {
			t.Error(err)
			return
		}
	}
}

func TestQcowCreateNewGptSingleWipePart(t *testing.T) {
	ni, _ := vutils.UUID.MakeUUIDString()
	tpath := filepath.Join(os.TempDir(), fmt.Sprintf("%s.qcow2", ni))
	fmt.Println("Creating " + tpath)
	qi, err := CreateNewQemuImage(tpath, 400*1024*1024)
	if err != nil {
		t.Error(err)
	} else {
		defer qi.Destroy()
		err := qi.Connect()
		if err != nil {
			t.Error(err)
			return
		}
		err = qi.MakeGpt()
		if err != nil {
			t.Error(err)
			return
		}
		_, err = qi.CreatePartition("full", "100%", LinuxFilesystem, true)
		if err != nil {
			t.Error(err)
			return
		}
		err = qi.WriteTable()
		if err != nil {
			t.Error(err)
			return
		}
		prt, err := qi.GetPartition(1)
		if err != nil {
			t.Error(err)
			return
		}
		err = prt.MakeFilesystem(FS_EXT4)
		if err != nil {
			t.Error(err)
			return
		}
		//lets now wipe the partition
		err = prt.SecureWipe(WipeImageSinglePassVerify)
		if err != nil {
			qi.Disconnect()
			t.Error(err)
			return
		}
		sz, _ := qi.SizeOnDisk()
		fmt.Println("Wiped. Disk Size:", sz)
		err = qi.Disconnect()
		if err != nil {
			t.Error(err)
			return
		}
	}
}

func TestQcowCreateNewRawGptSingleWipePart(t *testing.T) {
	ni, _ := vutils.UUID.MakeUUIDString()
	tpath := filepath.Join(os.TempDir(), fmt.Sprintf("%s.img", ni))
	fmt.Println("Creating " + tpath)
	qi, err := CreateNewQemuImageFormat(tpath, 400*1024*1024, "raw")
	if err != nil {
		t.Error(err)
	} else {
		defer qi.Destroy()
		err := qi.Connect()
		if err != nil {
			t.Error(err)
			return
		}
		err = qi.MakeGpt()
		if err != nil {
			t.Error(err)
			return
		}
		_, err = qi.CreatePartition("full", "100%", LinuxFilesystem, true)
		if err != nil {
			t.Error(err)
			return
		}
		err = qi.WriteTable()
		if err != nil {
			t.Error(err)
			return
		}
		prt, err := qi.GetPartition(1)
		if err != nil {
			t.Error(err)
			return
		}
		err = prt.MakeFilesystem(FS_EXT4)
		if err != nil {
			t.Error(err)
			return
		}
		//lets now wipe the partition
		err = prt.SecureWipe(WipeImageSinglePassVerify)
		if err != nil {
			qi.Disconnect()
			t.Error(err)
			return
		}
		sz, _ := qi.SizeOnDisk()
		fmt.Println("Wiped. Disk Size:", sz)
		err = qi.Disconnect()
		if err != nil {
			t.Error(err)
			return
		}
	}
}

func TestQcowCreateNewGptMultiple(t *testing.T) {
	ni, _ := vutils.UUID.MakeUUIDString()
	tpath := filepath.Join(os.TempDir(), fmt.Sprintf("%s.qcow2", ni))
	fmt.Println("Creating " + tpath)
	qi, err := CreateNewQemuImage(tpath, 400*1024*1024)
	if err != nil {
		t.Error(err)
	} else {
		defer qi.Destroy()
		err := qi.Connect()
		if err != nil {
			t.Error(err)
			return
		}
		err = qi.MakeGpt()
		if err != nil {
			t.Error(err)
			return
		}
		_, err = qi.CreatePartition("one", "33%", LinuxFilesystem, true)
		if err != nil {
			t.Error(err)
			return
		}
		_, err = qi.CreatePartition("two", "33%", LinuxFilesystem, false)
		if err != nil {
			t.Error(err)
			return
		}
		_, err = qi.CreatePartition("three", "34%", LinuxFilesystem, false)
		if err != nil {
			t.Error(err)
			return
		}
		err = qi.WriteTable()
		if err != nil {
			t.Error(err)
			return
		}
		prt, err := qi.GetPartition(1)
		if err != nil {
			t.Error(err)
			return
		}
		err = prt.MakeFilesystem(FS_EXT4)
		if err != nil {
			t.Error(err)
			return
		}
		prt, err = qi.GetPartition(2)
		if err != nil {
			t.Error(err)
			return
		}
		err = prt.MakeFilesystem(FS_EXT4)
		if err != nil {
			t.Error(err)
			return
		}
		prt, err = qi.GetPartition(3)
		if err != nil {
			t.Error(err)
			return
		}
		err = prt.MakeFilesystem(FS_EXFAT)
		if err != nil {
			t.Error(err)
			return
		}
		err = qi.Disconnect()
		if err != nil {
			t.Error(err)
			return
		}
	}
}

func TestQcowCreateNewGptMultipleSize(t *testing.T) {
	ni, _ := vutils.UUID.MakeUUIDString()
	tpath := filepath.Join(os.TempDir(), fmt.Sprintf("%s.qcow2", ni))
	fmt.Println("Creating " + tpath)
	qi, err := CreateNewQemuImage(tpath, 400*1024*1024)
	if err != nil {
		t.Error(err)
	} else {
		defer qi.Destroy()
		err := qi.Connect()
		if err != nil {
			t.Error(err)
			return
		}
		err = qi.MakeGpt()
		if err != nil {
			t.Error(err)
			return
		}
		_, err = qi.CreatePartition("one", "33M", LinuxFilesystem, true)
		if err != nil {
			t.Error(err)
			return
		}
		_, err = qi.CreatePartition("two", "33M", LinuxFilesystem, false)
		if err != nil {
			t.Error(err)
			return
		}
		_, err = qi.CreatePartition("three", "34M", LinuxFilesystem, false)
		if err != nil {
			t.Error(err)
			return
		}
		err = qi.WriteTable()
		if err != nil {
			t.Error(err)
			return
		}
		prt, err := qi.GetPartition(1)
		if err != nil {
			t.Error(err)
			return
		}
		err = prt.MakeFilesystem(FS_EXT4)
		if err != nil {
			t.Error(err)
			return
		}
		prt, err = qi.GetPartition(2)
		if err != nil {
			t.Error(err)
			return
		}
		err = prt.MakeFilesystem(FS_EXT4)
		if err != nil {
			t.Error(err)
			return
		}
		prt, err = qi.GetPartition(3)
		if err != nil {
			t.Error(err)
			return
		}
		err = prt.MakeFilesystem(FS_EXFAT)
		if err != nil {
			t.Error(err)
			return
		}
		p, err := prt.Mount()
		if err != nil {
			t.Error(err)
			return
		}
		println("Mounted at " + p)
		err = prt.Unmount()
		if err != nil {
			t.Error(err)
			return
		}
		err = qi.Disconnect()
		if err != nil {
			t.Error(err)
			return
		}
	}
}

func TestQcowCreateNewMbr(t *testing.T) {
	ni, _ := vutils.UUID.MakeUUIDString()
	tpath := filepath.Join(os.TempDir(), fmt.Sprintf("%s.qcow2", ni))
	fmt.Println("Creating " + tpath)
	qi, err := CreateNewQemuImage(tpath, 400*1024*1024)
	if err != nil {
		t.Error(err)
	} else {
		defer qi.Destroy()
		err := qi.Connect()
		if err != nil {
			t.Error(err)
			return
		}
		err = qi.MakeMbr()
		if err != nil {
			t.Error(err)
			return
		}
		_, err = qi.CreatePartitionAt("", 2048, 90*1024*1024, Linux, true)
		if err != nil {
			t.Error(err)
			return
		}
		err = qi.WriteTable()
		if err != nil {
			t.Error(err)
			return
		}
		err = qi.Disconnect()
		if err != nil {
			t.Error(err)
			return
		}
	}
}

func TestQcowCreateNewMbrSingle(t *testing.T) {
	ni, _ := vutils.UUID.MakeUUIDString()
	tpath := filepath.Join(os.TempDir(), fmt.Sprintf("%s.qcow2", ni))
	fmt.Println("Creating " + tpath)
	qi, err := CreateNewQemuImage(tpath, 400*1024*1024)
	if err != nil {
		t.Error(err)
	} else {
		defer qi.Destroy()
		err := qi.Connect()
		if err != nil {
			t.Error(err)
			return
		}
		err = qi.MakeMbr()
		if err != nil {
			t.Error(err)
			return
		}
		_, err = qi.CreatePartition("full", "100%", Linux, true)
		if err != nil {
			t.Error(err)
			return
		}
		err = qi.WriteTable()
		if err != nil {
			t.Error(err)
			return
		}
		prt, err := qi.GetPartition(0)
		if err != nil {
			t.Error(err)
			return
		}
		err = prt.MakeFilesystem(FS_EXT4)
		if err != nil {
			t.Error(err)
			return
		}
		p, err := prt.Mount()
		if err != nil {
			t.Error(err)
			return
		}
		println("Mounted at " + p)
		err = prt.Unmount()
		if err != nil {
			t.Error(err)
			return
		}
		err = qi.Disconnect()
		if err != nil {
			t.Error(err)
			return
		}
	}
}

func TestQcowCreateNewMbrMultiple(t *testing.T) {
	ni, _ := vutils.UUID.MakeUUIDString()
	tpath := filepath.Join(os.TempDir(), fmt.Sprintf("%s.qcow2", ni))
	fmt.Println("Creating " + tpath)
	qi, err := CreateNewQemuImage(tpath, 400*1024*1024)
	if err != nil {
		t.Error(err)
	} else {
		defer qi.Destroy()
		err := qi.Connect()
		if err != nil {
			t.Error(err)
			return
		}
		err = qi.MakeMbr()
		if err != nil {
			t.Error(err)
			return
		}
		_, err = qi.CreatePartition("", "33%", Linux, true)
		if err != nil {
			t.Error(err)
			return
		}
		_, err = qi.CreatePartition("", "33%", Linux, false)
		if err != nil {
			t.Error(err)
			return
		}
		_, err = qi.CreatePartition("", "33%", Linux, false)
		if err != nil {
			t.Error(err)
			return
		}
		err = qi.WriteTable()
		if err != nil {
			t.Error(err)
			return
		}
		err = qi.Disconnect()
		if err != nil {
			t.Error(err)
			return
		}
	}
}

package images

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/768bit/promethium/lib/common"
	"github.com/768bit/promethium/lib/images/diskfs/disk"
	"github.com/768bit/promethium/lib/images/diskfs/partition"
	"github.com/768bit/promethium/lib/images/diskfs/partition/gpt"
	"github.com/768bit/promethium/lib/images/diskfs/partition/mbr"
	"github.com/768bit/vutils"
	"github.com/palantir/stacktrace"
)

type QemuImageWipeValue string

const (
	WipeImageZeroValue   QemuImageWipeValue = "zero"
	WipeImageOneValue    QemuImageWipeValue = "one"
	WipeImageRandomValue QemuImageWipeValue = "random"
)

type QemuImageWipeLevel string

const (
	WipeImageSinglePass       QemuImageWipeLevel = "single"
	WipeImageSinglePassVerify QemuImageWipeLevel = "single-verify"
	WipeImageBSA              QemuImageWipeLevel = "bsa"
	WipeImageNCSC_TG_025      QemuImageWipeLevel = "NCSC-TG-025"
)

//QemuMbd amnages instances of created, connected and nounted Qcow2 images...
//this requires superuser access

func getNbdDeviceList() ([]string, error) {
	if nbdEnabled, _ := checkNbdModule(); !nbdEnabled {
		err := loadNbd()
		if err != nil {
			return nil, err
		}
	}
	cmd := exec.Command("bash", "-c", "ls /dev/nbd*")
	oba, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}
	ostr := string(oba)

	//now split the results on space...

	oarr := strings.Split(ostr, "\n")

	resArr := []string{}
	for _, item := range oarr {
		if len(item) > 0 {
			matched, _ := regexp.Match(`^/dev/nbd\d+$`, []byte(item))
			if matched {
				resArr = append(resArr, item)
				//quick chown...
				cmd := vutils.Exec.CreateAsyncCommand("chown", false, fmt.Sprintf("%d", UID), item).Sudo()
				if err := cmd.StartAndWait(); err != nil {
					return nil, err
				}
			}
		}

	}

	return resArr, nil
}

func checkNbdModule() (bool, error) {
	cmd := exec.Command("bash", "-c", "\"lsmod | grep nbd\"")
	oba, err := cmd.CombinedOutput()
	if err != nil {
		return false, err
	}
	ostr := string(oba)
	return strings.Contains(ostr, "nbd"), nil
}

func loadNbd() error {
	cmd := vutils.Exec.CreateAsyncCommand("modprobe", false, "nbd").Sudo()
	return cmd.StartAndWait()
}

func runPartProbe(dev string) error {
	cmd := vutils.Exec.CreateAsyncCommand("partprobe", false, dev).Sudo().BindToStdoutAndStdErr()
	return cmd.StartAndWait()
}

var TYPE_RX, _ = regexp.Compile(`TYPE="(\S+)"`)

func getFilesystemType(dev string) (string, error) {
	cmd := exec.Command("blkid", dev)
	oba, err := cmd.CombinedOutput()
	ostr := string(oba)
	if err != nil {
		return ostr, err
	}
	fstr := TYPE_RX.FindString(ostr)
	return fstr, nil
}

func newQemuNbdDaemon() *qemuNbd {
	qn := &qemuNbd{}
	return qn.init()
}

type qemuNbd struct {
	devList []string
	devMap  map[string]*QemuImage
}

var NoQemuNbdDeviceAvailable error = errors.New("No Free Qemu NBD Devices are available")

func (qn *qemuNbd) getFirstDev() string {
	for _, dev := range qn.devList {
		v, ok := qn.devMap[dev]
		if !ok || v == nil {
			return dev
		}
	}

	return ""
}

func (qn *qemuNbd) connect(image *QemuImage) error {
	for {
		dev := qn.getFirstDev()
		if dev == "" {
			return NoQemuNbdDeviceAvailable
		}
		cmd := vutils.Exec.CreateAsyncCommand("qemu-nbd", false, "-c", dev, "-f", image.sourceFormat, image.path).Sudo().BindToStdoutAndStdErr()
		err := cmd.StartAndWait()
		if err == nil {
			image.connected = true
			image.connectedDevice = dev
			qn.devMap[dev] = image
			return nil
		} else {
			qn.devMap[dev] = nil
		}
	}
}

func (qn *qemuNbd) disconnect(image *QemuImage) error {
	cmd := vutils.Exec.CreateAsyncCommand("qemu-nbd", false, "-d", image.connectedDevice).Sudo().BindToStdoutAndStdErr()
	err := cmd.StartAndWait()
	if err == nil {
		delete(qn.devMap, image.connectedDevice)
		image.connected = false
		image.connectedDevice = ""
		return nil
	}
	return err
}

func (qn *qemuNbd) init() *qemuNbd {
	dl, err := getNbdDeviceList()
	if dl != nil && err == nil && len(dl) > 0 {
		qn.devList = dl
	} else {
		qn.devList = []string{}
	}
	qn.devMap = map[string]*QemuImage{}
	return qn
}

var QemuNbd *qemuNbd
var UID int
var GID int

func StartQemuNbd(uid int, gid int) {
	if QemuNbd != nil {
		return
	}
	UID = uid
	GID = gid
	QemuNbd = newQemuNbdDaemon()
}

func CreateNewQemuImage(path string, size uint64) (*QemuImage, error) {
	qi := &QemuImage{
		path: path,
	}
	if err := qi.create(size, "qcow2"); err != nil {
		return nil, err
	} else if err := qi.init(); err != nil {
		return nil, err
	}
	return qi, nil
}

func CreateNewQemuImageFormat(path string, size uint64, format string) (*QemuImage, error) {
	qi := &QemuImage{
		path: path,
	}
	if err := qi.create(size, format); err != nil {
		return nil, err
	} else if err := qi.init(); err != nil {
		return nil, err
	}
	return qi, nil
}

func LoadQemuImage(path string) (*QemuImage, error) {
	println("Loading qcow image at path: " + path)
	qi := &QemuImage{
		path: path,
	}
	if err := qi.init(); err != nil {
		return nil, err
	}
	return qi, nil
}

var ImageConnectedError error = errors.New("Image is currently connected. Please unmount and disconnect it.")
var ImageMountedError error = errors.New("Image is currently mounted. Please unmount first.")
var ImageNotConnected error = errors.New("Image is not currently connected. Please connect it.")
var ImageNotMountedError error = errors.New("Image is not currently mounted. Please connect and mount it first.")
var ImageContainsPartitionTableError error = errors.New("Image has a partition table so unable to mount/dismount it.")
var ImageDoesntContainPartitionTableError error = errors.New("Image doesnt have a partition table so unable to create/get/modify partitions.")

type QemuImage struct {
	fd                 *os.File
	sourceFormat       string
	path               string
	size               uint64
	actualSize         uint64
	sizeOnDisk         uint64
	clusterSize        uint32
	connected          bool
	mounted            bool
	mountPoint         string
	isDirty            bool
	connectedDevice    string
	table              partition.Table
	isGpt              bool
	gpt                *gpt.Table
	mbr                *mbr.Table
	initialised        bool
	physicalSectorSize int
	logicalSectorSize  int
	partitions         []*QemuImagePartition
	disk               *disk.Disk
	tableSaved         bool
	hasTable           bool
}

func (qi *QemuImage) create(size uint64, format string) error {
	//create the new image at path...
	cmd := vutils.Exec.CreateAsyncCommand("qemu-img", false, "create", "-f", format, qi.path, fmt.Sprintf("%d", size)).BindToStdoutAndStdErr()
	return cmd.StartAndWait()
}

func (qi *QemuImage) init() error {
	//get status info etc...
	qi.connected = false
	qi.mounted = false
	qi.mountPoint = ""
	qi.isDirty = false
	qi.isGpt = false
	qi.gpt = nil
	qi.mbr = nil
	qi.initialised = false
	qi.tableSaved = false
	qi.hasTable = false
	return qi.getInfo()
}

func (qi *QemuImage) getInfo() error {
	//get status info etc...
	cmd := vutils.Exec.CreateAsyncCommand("qemu-img", false, "info", "--output=json", qi.path).CaptureStdoutAndStdErr(false, true)
	err := cmd.StartAndWait()
	if err != nil {
		return stacktrace.Propagate(err, "Failed to get qemu-img info")
	}
	outStr := cmd.GetStdoutBuffer()
	var outObj map[string]interface{}
	err = json.Unmarshal(outStr, &outObj)
	if err != nil {
		return err
	}
	//set the items

	if v, ok := outObj["virtual-size"]; ok {
		nv, ok := v.(float64)
		if ok {
			qi.size = uint64(nv)
		}
	}
	if v, ok := outObj["cluster-size"]; ok {
		nv, ok := v.(float64)
		if ok {
			qi.clusterSize = uint32(nv)
		}
	}
	if v, ok := outObj["dirty-flag"]; ok {
		nv, ok := v.(bool)
		if ok {
			qi.isDirty = nv
		}
	}
	if v, ok := outObj["actual-size"]; ok {
		nv, ok := v.(float64)
		if ok {
			qi.actualSize = uint64(nv)
		}
	}
	if v, ok := outObj["format"]; ok {
		nv, ok := v.(string)
		if ok {
			qi.sourceFormat = nv
		}
	}
	if v, err := qi.SizeOnDisk(); err != nil {
		return err
	} else {
		qi.sizeOnDisk = v
	}

	return nil
}

func (qi *QemuImage) isBlockDevice() (bool, error) {
	fi, err := os.Stat(qi.path)
	if err != nil {
		return false, err
	}
	if fi.Mode()&os.ModeDevice != 0 {
		return true, nil
	} else if fi.IsDir() {
		//doesnt support directories
		return false, errors.New("Directories are not supported for image backends. Only use block devices or image files.")
	} else {
		return false, nil
	}
}

func (qi *QemuImage) getImageDevicePath() (string, error) {
	if isBlk, err := qi.isBlockDevice(); err != nil {
		return "", err
	} else if isBlk {
		return qi.path, nil
	} else {
		return qi.connectedDevice, nil
	}
}

func (qi *QemuImage) IsMounted() bool {
	if qi.mounted {
		return true
	} else if qi.hasTable {
		partLen := len(qi.partitions)
		//lets see if any of the parts are mounted...
		if partLen > 0 {
			for _, part := range qi.partitions {
				if part != nil {
					if part.IsMounted() {
						return true
					}
				}
			}
		}
	}
	return false
}

func (qi *QemuImage) ConvertImgRaw(dest string) error {
	if qi.connected {
		err := qi.Disconnect()
		if err != nil {
			return err
		}
		err = vutils.Exec.CreateAsyncCommand("qemu-img", false, "convert", "-O", "raw", qi.path, dest).BindToStdoutAndStdErr().StartAndWait()
		if err != nil {
			return err
		}
		return qi.Connect()
	}

	return vutils.Exec.CreateAsyncCommand("qemu-img", false, "convert", "-O", "raw", qi.path, dest).BindToStdoutAndStdErr().StartAndWait()
}

func (qi *QemuImage) ConvertImgRawDevice(dest string) error {
	if qi.connected {
		err := qi.Disconnect()
		if err != nil {
			return err
		}
		err = vutils.Exec.CreateAsyncCommand("qemu-img", false, "convert", "-O", "host_device", qi.path, dest).BindToStdoutAndStdErr().StartAndWait()

		if err != nil {
			return err
		}
		return qi.Connect()
	}

	return vutils.Exec.CreateAsyncCommand("qemu-img", false, "convert", "-O", "host_device", qi.path, dest).BindToStdoutAndStdErr().StartAndWait()
}

func (qi *QemuImage) VirtualSize() uint64 {
	return qi.size
}

func (qi *QemuImage) GetFilesystem() (string, error) {
	if qi.hasTable {
		return "", ImageContainsPartitionTableError
	} else if !qi.connected {
		return "", ImageNotConnected
	}
	typ, err := getFilesystemType(qi.connectedDevice)
	if err != nil {
		return "", err
	}
	return typ, nil
}

func (qi *QemuImage) SizeOnDisk() (uint64, error) {
	return vutils.Files.FileSize(qi.path)
}

func (qi *QemuImage) Resize(newSize uint64) error {
	if isBlk, err := qi.isBlockDevice(); err != nil {
		return err
	} else if isBlk {
		return errors.New("Cannot resize a block device")
	}
	if qi.mounted {
		return ImageMountedError
	} else if qi.connected {
		return ImageConnectedError
	} else if err := qi.getInfo(); err != nil {
		return err
	} else if qi.size >= newSize {
		return errors.New("Cannot resize the image to a size smaller than its current size")
	} else {
		cmd := vutils.Exec.CreateAsyncCommand("qemu-img", false, "resize", qi.path, fmt.Sprintf("%d", newSize)).BindToStdoutAndStdErr()
		if err := cmd.StartAndWait(); err != nil {
			return err
		} else {
			return qi.getInfo()
		}
	}
}

func (qi *QemuImage) Connect() error {
	if qi.connected {
		return nil
	} else if isBlk, err := qi.isBlockDevice(); err != nil {
		return err
	} else if isBlk {
		qi.connected = true
		return qi.loadDisk()
	}
	err := QemuNbd.connect(qi)
	if err != nil {
		return err
	}
	return qi.loadDisk()
}

func (qi *QemuImage) Disconnect() error {
	if !qi.connected {
		return ImageNotConnected
	} else if qi.IsMounted() {
		return ImageMountedError
	} else if isBlk, err := qi.isBlockDevice(); err != nil {
		return err
	} else if isBlk {
		if err := qi.closeDisk(); err != nil {
			fmt.Println(err)
		}
		qi.connected = false
		return err
	}
	if err := qi.closeDisk(); err != nil {
		fmt.Println(err)
	}
	return QemuNbd.disconnect(qi)
}

func (qi *QemuImage) GrowFullPart() error {
	//this will unmount the temporary mountpoint
	if !qi.connected {
		return ImageNotConnected
	} else if qi.mounted {
		return ImageMountedError
	} else if hasParts, _ := qi.HasPartitions(); hasParts {
		return ImageContainsPartitionTableError
	}
	//we are good, lets do this!
	// cmd := vutils.Exec.CreateAsyncCommand("growpart", false, qp.img.connectedDevice, fmt.Sprintf("%d", qp.index)).Sudo()
	// err = cmd.BindToStdoutAndStdErr().StartAndWait()
	// if err != nil {
	// 	return err
	// }
	//println(err.Error())
	if err := vutils.Exec.CreateAsyncCommand("e2fsck", false, "-f", "-p", qi.connectedDevice).BindToStdoutAndStdErr().StartAndWait(); err != nil {
		return err
	}
	cmd := vutils.Exec.CreateAsyncCommand("resize2fs", false, qi.connectedDevice).Sudo()
	err := cmd.BindToStdoutAndStdErr().StartAndWait()
	if err != nil {
		return err
	}
	return nil

}

func (qi *QemuImage) loadDisk() error {
	qi.tableSaved = true
	if isBlk, err := qi.isBlockDevice(); err != nil {
		return err
	} else if !isBlk {
		// usr, _ := user.Current()
		// //quick chown...
		// cmd := vutils.Exec.CreateAsyncCommand("chown", false, usr.Username, qi.connectedDevice).Sudo()
		// if err := cmd.StartAndWait(); err != nil {
		// 	return err
		// }
	}
	devPath, err := qi.getImageDevicePath()
	if err != nil {
		return err
	}
	f, err := os.OpenFile(devPath, os.O_RDWR, 0660)
	if err != nil {
		return err
	}
	qi.fd = f
	fi, err := qi.fd.Stat()
	if err != nil {
		return err
	}
	dsk := &disk.Disk{
		File:              qi.fd,
		LogicalBlocksize:  512,
		PhysicalBlocksize: 512,
		Info:              fi,
		Size:              int64(qi.size),
	}
	//dsk, err := diskfs.Open(qi.connectedDevice)
	//if err != nil {
	//  return err
	//} else {
	qi.disk = dsk
	tb, err := qi.disk.GetPartitionTable()
	if err != nil {
		println(err.Error())
		qi.isGpt = false
		qi.mbr = nil
		qi.gpt = nil
		qi.initialised = false
		qi.tableSaved = false
		qi.hasTable = false
	} else {
		qi.table = tb
		qi.hasTable = true
		if qi.table != nil && qi.table.Type() == "gpt" {
			qi.isGpt = true
			qi.mbr = nil
			gt, ok := qi.table.(*gpt.Table)
			if ok && gt != nil {
				qi.gpt = gt
				qi.initialised = true
				qi.physicalSectorSize = qi.gpt.PhysicalSectorSize
				qi.logicalSectorSize = qi.gpt.LogicalSectorSize
			}
		} else if qi.table != nil && qi.table.Type() == "mbr" {
			qi.isGpt = false
			qi.gpt = nil
			mt, ok := qi.table.(*mbr.Table)
			if ok && mt != nil {
				qi.mbr = mt
				qi.initialised = true
				qi.physicalSectorSize = qi.mbr.PhysicalSectorSize
				qi.logicalSectorSize = qi.mbr.LogicalSectorSize
			}
		} else {
			qi.isGpt = false
			qi.mbr = nil
			qi.gpt = nil
			qi.initialised = false
			qi.tableSaved = false
			qi.hasTable = false
		}
	}
	//}
	return qi.loadParts()
}

func (qi *QemuImage) closeDisk() error {
	qi.disk.File.Sync()
	err := qi.disk.File.Close()
	if err != nil {
		return err
	}
	qi.fd = nil
	if isBlk, err := qi.isBlockDevice(); err != nil {
		return err
	} else if !isBlk {
		// cmd := vutils.Exec.CreateAsyncCommand("chown", false, "root", qi.connectedDevice).Sudo()
		// if err := cmd.StartAndWait(); err != nil {
		// 	return err
		// }
	}
	return nil
}

func (qi *QemuImage) loadParts() error {
	qi.partitions = []*QemuImagePartition{}
	if qi.initialised {
		if qi.isGpt {
			if qi.gpt != nil && qi.gpt.Partitions != nil && len(qi.gpt.Partitions) > 0 {
				for index, part := range qi.gpt.Partitions {
					if part.Start == 0 && part.End == 0 {
						continue
					}
					np := NewQemuImagePartitionGpt(qi.logicalSectorSize, part)
					np.setImage(index, qi)
					qi.partitions = append(qi.partitions, np)
				}
			}
		} else {
			if qi.mbr != nil && qi.mbr.Partitions != nil && len(qi.mbr.Partitions) > 0 {
				for index, part := range qi.mbr.Partitions {
					if part.Start == 0 && part.Type == mbr.Empty {
						continue
					}
					np := NewQemuImagePartitionMbr(qi.logicalSectorSize, part)
					np.setImage(index, qi)
					qi.partitions = append(qi.partitions, np)
				}
			}
		}
	}
	return nil
}

var EFI_END_POS = uint64((100*1024*1024)/512) + (RESERVED_START_BYTES)

func (qi *QemuImage) MakeGpt() error {
	qi.gpt = &gpt.Table{
		LogicalSectorSize:  512,
		PhysicalSectorSize: 512,
		Partitions: []*gpt.Partition{
			{Start: 2048, End: EFI_END_POS, Type: EFISystemPartition.GPT, Name: "EFI System"},
		},
		ProtectiveMBR: true,
	}
	qi.physicalSectorSize = qi.gpt.PhysicalSectorSize
	qi.logicalSectorSize = qi.gpt.LogicalSectorSize
	qi.mbr = nil
	qi.isGpt = true
	qi.tableSaved = false
	qi.hasTable = true
	err := qi.WriteTable()
	if err != nil {
		return err
	}
	//make the EFI partition filesystem..
	if prt, err := qi.GetPartition(0); err != nil {
		return err
	} else {
		return prt.MakeFilesystem(FS_FAT32)
	}
}

func (qi *QemuImage) MakeMbr() error {
	qi.mbr = &mbr.Table{
		LogicalSectorSize:  512,
		PhysicalSectorSize: 512,
		Partitions:         []*mbr.Partition{},
	}
	qi.physicalSectorSize = qi.mbr.PhysicalSectorSize
	qi.logicalSectorSize = qi.mbr.LogicalSectorSize
	qi.gpt = nil
	qi.isGpt = false
	qi.tableSaved = false
	qi.hasTable = true
	return qi.WriteTable()
}

func (qi *QemuImage) WriteTable() error {
	if qi.tableSaved {
		return nil
	}
	if !qi.connected {
		return ImageNotConnected
	} else if qi.mounted {
		return ImageMountedError
	}
	if !qi.hasTable {
		return errors.New("Unable to write partition table to disk - this disk doesnt have a partition table intitialised")
	}
	if qi.isGpt {
		//if len(qi.partitions) == 0 {
		//  return nil
		//}
		if qi.initialised {
			qi.gpt.Partitions = []*gpt.Partition{}
			for _, item := range qi.partitions {
				qi.gpt.Partitions = append(qi.gpt.Partitions, item.gptPart)
			}
		}
		err := qi.disk.PartitionOfSize(qi.gpt, int64(qi.size))
		if err != nil {
			return err
		}
	} else {
		//if len(qi.partitions) == 0 {
		//  return nil
		//}
		if qi.initialised {
			qi.mbr.Partitions = []*mbr.Partition{}
			for _, item := range qi.partitions {
				qi.mbr.Partitions = append(qi.mbr.Partitions, item.mbrPart)
			}
		}
		err := qi.disk.PartitionOfSize(qi.mbr, int64(qi.size))
		if err != nil {
			return err
		}
	}
	qi.initialised = true
	qi.tableSaved = true
	err := qi.Disconnect()
	if err != nil {
		return err
	}
	err = qi.Connect()
	if err != nil {
		return err
	}
	dp, err := qi.getImageDevicePath()
	if err != nil {
		return err
	}
	err = runPartProbe(dp)
	if err != nil {
		//return err
	}
	return qi.loadParts()
}

var PartitionMissingError = errors.New("Cannot retrieve partition as it doesnt exist")

func (qi *QemuImage) HasPartitions() (bool, int) {
	partLen := len(qi.partitions)
	return qi.hasTable, partLen
}

func (qi *QemuImage) GetPartition(partitionNumber int) (*QemuImagePartition, error) {
	if !qi.hasTable {
		return nil, ImageDoesntContainPartitionTableError
	}
	partLen := len(qi.partitions)
	if partLen == 0 || partitionNumber >= partLen {
		return nil, PartitionMissingError
	}
	return qi.partitions[partitionNumber], nil
}

func (qi *QemuImage) MakeFilesystem(name string, fsType common.ImageFsType) error {
	var err error = nil
	switch fsType {
	case common.ExFat:
		err = MakeExfat(name, qi.connectedDevice)
		break
	case common.Ext4:
		err = MakeExt4(name, qi.connectedDevice)
		break
	case common.Fat32:
		err = MakeFat32(name, qi.connectedDevice)
		break
	}
	if err != nil {
		return err
	}
	return nil
}

func (qi *QemuImage) Mount() (string, error) {
	//mount will create temporary mount point and mount the system there...
	if qi.hasTable {
		return "", ImageContainsPartitionTableError
	}
	if qi.mounted {
		return "", ImageMountedError
	}
	tdir, err := ioutil.TempDir("", "prmnt")
	if err != nil {
		return "", err
	}
	qi.mountPoint = tdir
	cmd := vutils.Exec.CreateAsyncCommand("mount", false, qi.connectedDevice, qi.mountPoint)
	err = cmd.StartAndWait()
	if err != nil {
		return "", err
	}
	qi.mounted = true
	return qi.mountPoint, nil
}

func (qi *QemuImage) Unmount() error {
	//mount will create temporary mount point and mount the system there...
	if !qi.IsMounted() {
		return ImageNotMountedError
	}
	//we need to figure out if the whole image is mounted or a partition and unmount everything...
	if !qi.mounted {
		partLen := len(qi.partitions)
		//lets see if any of the parts are mounted...
		if partLen > 0 {
			for _, part := range qi.partitions {
				if part != nil {
					if part.IsMounted() {
						err := part.Unmount()
						if err != nil {
							println(err.Error())
						}
					}
				}
			}
		}
		return nil
	}
	cmd := vutils.Exec.CreateAsyncCommand("umount", false, qi.mountPoint)
	err := cmd.StartAndWait()
	if err != nil {
		return err
	}
	qi.mounted = false
	os.RemoveAll(qi.mountPoint)
	qi.mountPoint = ""
	return nil
}

var RESERVED_START_BYTES = uint64(2048)

func parseStringSize(strSize string, blkSize uint64, diskSize uint64, isGpt bool) (uint64, error) {
	strLen := len(strSize)
	sbytes := RESERVED_START_BYTES
	if isGpt {
		sbytes = EFI_END_POS
	}
	if matched, err := regexp.MatchString(`^\d{1,3}%$`, strSize); err == nil && matched {
		val, err := strconv.Atoi(strSize[:strLen-1])
		if err != nil {
			return 0, err
		}
		if val > 100 {
			val = 100
		} else if val <= 0 {
			return 0, errors.New("Cannot make a disk of size using an invalid percentage")
		}
		fpct := float64(val) / 100
		newSize := (uint64(fpct*float64(diskSize-(sbytes*blkSize)-(blkSize*2))) / blkSize) * blkSize
		return newSize, nil
	} else {
		lastChar := strings.ToUpper(strSize[strLen-1:])
		mult := uint64(0)
		val := 0
		switch lastChar {
		case "K":
			mult = 1
		case "M":
			mult = 2
		case "G":
			mult = 3
		case "T":
			mult = 4
		case "P":
			mult = 5
		default:
			strLen++
		}
		val, err := strconv.Atoi(strSize[:strLen-1])
		if err != nil {
			return 0, err
		}
		oval := uint64(val)
		if mult > 0 {
			mult = uint64(math.Pow(1024, float64(mult)))
			oval = oval * mult
		}
		fmt.Printf("Making Partition -> Requested Size: %s | Val: %d | Last Char: %s | Mult: %d | In Bytes: %d\n", strSize, val, lastChar, mult, oval)
		if oval >= diskSize-(sbytes*blkSize) {
			return 0, errors.New("THe specified size is larger than the available space")
		}
		return oval, nil
	}
}

func (qi *QemuImage) CreatePartition(name string, strSize string, partitionType *QemuImagePartitionType, bootable bool) (*QemuImagePartition, error) {
	//need to figure out where in the array we place the new part...
	//additionally we need to see if there is an overlap..
	//if end > qi.size {
	//  return nil, errors.New(fmt.Sprintf("Cannot create a new partition as it ends past the extents of the disk"))
	//}
	if !qi.hasTable {
		return nil, ImageDoesntContainPartitionTableError
	}
	sbytes := RESERVED_START_BYTES
	if qi.isGpt {
		sbytes = EFI_END_POS + 1
	}

	indexWatermark := 0
	var newPart *QemuImagePartition = nil
	blkSize := uint64(qi.logicalSectorSize)
	partLen := len(qi.partitions)

	//parse the size..
	reqSize, err := parseStringSize(strSize, blkSize, qi.size, qi.isGpt)
	if err != nil {
		return nil, err
	}
	diskSects := qi.size / blkSize
	startSec := sbytes
	endSec := startSec + (reqSize / blkSize)
	partSizeSec := (endSec - startSec)
	size := (partSizeSec + 1) * blkSize

	if partLen > 0 {
		for index, part := range qi.partitions {
			if (startSec >= part.start && startSec <= part.end) || (endSec >= part.start && endSec <= part.end) {
				//new start or end exists in current partition
				startSec = part.end + 1
				endSec = startSec + partSizeSec
				if endSec == diskSects {
					endSec--
					size = (endSec - startSec + 1) * blkSize
				}
				if endSec*blkSize > qi.size {
					return nil, errors.New("Cannot create partition as it will be larger than the extents")
				}
				if index+1 == partLen {
					//we are at the end so we just add it...
					np, err := NewQemuImagePartition(name, startSec, endSec, size, blkSize, qi.isGpt, partitionType, bootable)
					if err != nil {
						return nil, err
					}
					newPart = np
					newPart.setImage(partLen, qi)
					qi.tableSaved = false
					qi.partitions = append(qi.partitions, np)
					return newPart, nil
				}
				continue
			} else if endSec < part.start {
				//the partiton exists before this partition and it fits
				if index == 0 {
					if endSec == diskSects {
						endSec--
						size = (endSec - startSec + 1) * blkSize
					}
					if endSec*blkSize > qi.size {
						return nil, errors.New("Cannot create partition as it will be larger than the extents")
					}
					//we are at the end so we just add it...
					np, err := NewQemuImagePartition(name, startSec, endSec, size, blkSize, qi.isGpt, partitionType, bootable)
					if err != nil {
						return nil, err
					}
					newPart = np
					newPart.setImage(0, qi)
					for _, item := range qi.partitions {
						item.index++
					}
					qi.tableSaved = false
					qi.partitions = append([]*QemuImagePartition{np}, qi.partitions...)
					return newPart, nil
				} else {
					if endSec == diskSects {
						endSec--
						size = (endSec - startSec + 1) * blkSize
					}
					if endSec*blkSize > qi.size {
						return nil, errors.New("Cannot create partition as it will be larger than the extents")
					}
					np, err := NewQemuImagePartition(name, startSec, endSec, size, blkSize, qi.isGpt, partitionType, bootable)
					if err != nil {
						return nil, err
					}
					newPart = np
					newPart.setImage(index, qi)
					for _, item := range qi.partitions[indexWatermark:] {
						item.index++
					}
					qi.tableSaved = false
					qi.partitions = append(append(qi.partitions[:indexWatermark], np), qi.partitions[indexWatermark:]...)
					return newPart, nil
				}
			} else if part.end < startSec {
				//the partition exists after this partition
				if index+1 == partLen {
					if endSec == diskSects {
						endSec--
						size = (endSec - startSec + 1) * blkSize
					}
					if endSec*blkSize > qi.size {
						return nil, errors.New("Cannot create partition as it will be larger than the extents")
					}
					//we are at the end so we just add it...
					np, err := NewQemuImagePartition(name, startSec, endSec, size, blkSize, qi.isGpt, partitionType, bootable)
					if err != nil {
						return nil, err
					}
					newPart = np
					newPart.setImage(partLen, qi)
					qi.tableSaved = false
					qi.partitions = append(qi.partitions, np)
					return newPart, nil
				} else {
					indexWatermark = index
				}
			}
		}
	} else {
		if endSec == diskSects {
			endSec--
			size = (endSec - startSec + 1) * blkSize
		}
		if endSec*blkSize > qi.size {
			fmt.Printf("Size: %d | Req Size %d\n", qi.size, endSec*blkSize)
			return nil, errors.New("Cannot create partition as it will be larger than the extents")
		}
		np, err := NewQemuImagePartition(name, startSec, endSec, size, blkSize, qi.isGpt, partitionType, bootable)
		if err != nil {
			return nil, err
		}
		newPart = np
		newPart.setImage(0, qi)
		qi.tableSaved = false
		qi.partitions = []*QemuImagePartition{np}
		return newPart, nil
	}
	return newPart, errors.New("Error creating new partition")
}

func (qi *QemuImage) CreatePartitionAt(name string, start uint64, end uint64, partitionType *QemuImagePartitionType, bootable bool) (*QemuImagePartition, error) {
	//need to figure out where in the array we place the new part...
	//additionally we need to see if there is an overlap..
	if !qi.hasTable {
		return nil, ImageDoesntContainPartitionTableError
	}
	if end > qi.size {
		return nil, errors.New(fmt.Sprintf("Cannot create a new partition as it ends past the extents of the disk"))
	}
	indexWatermark := 0
	var newPart *QemuImagePartition = nil
	blkSize := uint64(qi.logicalSectorSize)
	partLen := len(qi.partitions)
	startSec := start / blkSize
	endSec := end / blkSize
	size := (endSec - startSec + 1) * blkSize
	if partLen > 0 {
		for index, part := range qi.partitions {
			if (startSec >= part.start && startSec <= part.end) || (endSec >= part.start && endSec <= part.end) {
				//new start or end exists in current partition
				return nil, errors.New(fmt.Sprintf("Cannot create a new partition as it overlaps an existing partiton %d", index))
			} else if endSec < part.startByte {
				//the partiton exists before this partition
				if index == 0 {
					//we are at the end so we just add it...
					np, err := NewQemuImagePartition(name, startSec, endSec, size, blkSize, qi.isGpt, partitionType, bootable)
					if err != nil {
						return nil, err
					}
					newPart = np
					newPart.setImage(0, qi)
					for _, item := range qi.partitions {
						item.index++
					}
					qi.tableSaved = false
					qi.partitions = append([]*QemuImagePartition{np}, qi.partitions...)
					return newPart, nil
				} else {
					np, err := NewQemuImagePartition(name, startSec, endSec, size, blkSize, qi.isGpt, partitionType, bootable)
					if err != nil {
						return nil, err
					}
					newPart = np
					newPart.setImage(index, qi)
					for _, item := range qi.partitions[indexWatermark:] {
						item.index++
					}
					qi.tableSaved = false
					qi.partitions = append(append(qi.partitions[:indexWatermark], np), qi.partitions[indexWatermark:]...)
					return newPart, nil
				}
			} else if part.end < startSec {
				//the partition exists after this partition
				if index+1 == partLen {
					//we are at the end so we just add it...
					np, err := NewQemuImagePartition(name, startSec, endSec, size, blkSize, qi.isGpt, partitionType, bootable)
					if err != nil {
						return nil, err
					}
					newPart = np
					newPart.setImage(partLen, qi)
					qi.tableSaved = false
					qi.partitions = append(qi.partitions, np)
					return newPart, nil
				} else {
					indexWatermark = index
				}
			}
		}
	} else {
		np, err := NewQemuImagePartition(name, startSec, endSec, size, blkSize, qi.isGpt, partitionType, bootable)
		if err != nil {
			return nil, err
		}
		newPart = np
		newPart.setImage(0, qi)
		qi.tableSaved = false
		qi.partitions = []*QemuImagePartition{np}
		return newPart, nil
	}
	return newPart, errors.New("Error creating new partition")
}

func (qi *QemuImage) SecureWipe(level QemuImageWipeLevel) error {
	if !qi.connected {
		return ImageNotConnected
	}
	if qi.mounted {
		return ImageMountedError
	}
	//using the defined level securely erase the device...
	if isBlk, err := qi.isBlockDevice(); err != nil {
		return err
	} else if isBlk {
		fileSize := qi.VirtualSize()

		// calculate total number of parts the file will be chunked into
		chunkParts := int64(math.Ceil(float64(fileSize) / float64(IMAGE_WIPE_CHUNK_SIZE)))

		fmt.Println("Size:", fileSize, "Chunks:", chunkParts)

		return runSecureErase(qi.fd, int64(fileSize), chunkParts, level, 0, int64(fileSize))

	} else {
		fileSize, err := qi.SizeOnDisk()
		if err != nil {
			return err
		}

		// calculate total number of parts the file will be chunked into
		chunkParts := int64(math.Ceil(float64(fileSize) / float64(IMAGE_WIPE_CHUNK_SIZE)))

		fmt.Println("Size:", fileSize, "Chunks:", chunkParts)

		return runSecureErase(qi.fd, int64(fileSize), chunkParts, level, 0, int64(fileSize))
	}
}

func (qi *QemuImage) Destroy() error {
	if isBlk, err := qi.isBlockDevice(); err != nil {
		return err
	} else if isBlk {
		return errors.New("Cannot use 'destroy' on a block device")
	}
	return os.Remove(qi.path)
}

func runSecureErase(file *os.File, fileSize int64, chunkParts int64, level QemuImageWipeLevel, start int64, end int64) error {
	switch level {
	case WipeImageSinglePass:
		//do single pass 0
		return scanBlocksAndRun(file, fileSize, chunkParts, start, end, fillBlockOperation(WipeImageZeroValue))
	case WipeImageSinglePassVerify:
		return scanBlocksAndRun(file, fileSize, chunkParts, start, end, fillBlockOperation(WipeImageZeroValue), verifyBlockOperation())
	case WipeImageBSA:
		//Bruce Schneier’s Algorithm
		return scanBlocksAndRun(file, fileSize, chunkParts, start, end, fillBlockOperation(WipeImageOneValue), fillBlockOperation(WipeImageZeroValue), fillBlockOperation(WipeImageRandomValue), fillBlockOperation(WipeImageRandomValue), fillBlockOperation(WipeImageRandomValue), fillBlockOperation(WipeImageRandomValue), fillBlockOperation(WipeImageRandomValue))
	case WipeImageNCSC_TG_025:
		return scanBlocksAndRun(file, fileSize, chunkParts, start, end, fillBlockOperation(WipeImageZeroValue), verifyBlockOperation(), fillBlockOperation(WipeImageOneValue), verifyBlockOperation(), fillBlockOperation(WipeImageRandomValue), verifyBlockOperation())
	}
	return errors.New("unsupported Image Wipe Level")
}

const IMAGE_WIPE_CHUNK_SIZE = 1 * (1 << 20)

func openTargetForErase(targetFile string) (file *os.File, chunkParts int64, err error) {
	file, err = os.OpenFile(targetFile, os.O_RDWR, 0600)
	if err != nil {
		return
	}
	fileSize, err := file.Seek(0, io.SeekEnd)
	//fileInfo, err := file.Stat()
	if err != nil {
		return
	}
	//var fileSize int64 = fileInfo.Size()
	println(fileSize)
	// calculate total number of parts the file will be chunked into
	chunkParts = int64(math.Ceil(float64(fileSize) / float64(IMAGE_WIPE_CHUNK_SIZE)))
	return
}

type ImageWipeOperation func(file *os.File, position int64, size int64, currBlock []byte, writtenBlock []byte) error

func scanBlocksAndRun(file *os.File, fileSize int64, totalPartsNum int64, start int64, end int64, ops ...ImageWipeOperation) error {
	lastPosition := start
	chunkSize := end - start
	opsLen := len(ops)
	if opsLen == 0 {
		return errors.New("Must specify secure erasure operations")
	} else if totalPartsNum == 0 {
		return errors.New("No chunks to process")
	}

	currBlock := make([]byte, IMAGE_WIPE_CHUNK_SIZE)
	writeBlock := make([]byte, IMAGE_WIPE_CHUNK_SIZE)

	bytesWritten := int64(0)

	for i := int64(0); i < totalPartsNum; i++ {

		partSize := int64(math.Min(IMAGE_WIPE_CHUNK_SIZE, float64(chunkSize-int64(i*IMAGE_WIPE_CHUNK_SIZE))))
		if partSize < IMAGE_WIPE_CHUNK_SIZE {
			currBlock = make([]byte, partSize)
			writeBlock = make([]byte, partSize)
		} else if partSize == IMAGE_WIPE_CHUNK_SIZE && (int64(len(currBlock)) < partSize || int64(len(writeBlock)) < partSize) {
			currBlock = make([]byte, partSize)
			writeBlock = make([]byte, partSize)
		}
		for _, op := range ops {
			_, err := file.ReadAt(currBlock, lastPosition)
			if err != nil {
				return err
			}
			err = op(file, lastPosition, partSize, currBlock, writeBlock)
			if err != nil {
				return err
			}
			fmt.Println("Completed Operation for Block", i)
		}
		bytesWritten += int64(len(writeBlock))
		// update last written position
		lastPosition = lastPosition + partSize
	}
	fmt.Println("Wrote:", bytesWritten)
	return nil
}

func fillBlockOperation(fillVal QemuImageWipeValue) ImageWipeOperation {
	return func(file *os.File, position int64, size int64, currentBlock []byte, writtenBlock []byte) error {

		switch fillVal {
		case WipeImageZeroValue:
			getFilledBlock(writtenBlock, size, '0')
			_, err := file.WriteAt(writtenBlock, position)
			return err
		case WipeImageOneValue:
			getFilledBlock(writtenBlock, size, '1')
			_, err := file.WriteAt(writtenBlock, position)
			return err
		case WipeImageRandomValue:
			err := getRandomBlock(writtenBlock, size)
			if err != nil {
				return err
			}
			_, err = file.WriteAt(writtenBlock, position)
			return err
		default:
			return errors.New("Unsupporsted FillValue Option")
		}
	}
}

func verifyBlockOperation() ImageWipeOperation {
	return func(file *os.File, position int64, size int64, currentBlock []byte, writtenBlock []byte) error {
		//check that the current block matches the written block...
		writeLen := int64(len(writtenBlock))
		currLen := int64(len(currentBlock))
		if writeLen != currLen || writeLen != size || currLen != size {
			return errors.New("Unable to verify as the blocks arent of equal size")
		}
		//the sizes are valid - lets check each byte is equal...
		for i := int64(0); i < size; i++ {
			if currentBlock[i] != writtenBlock[i] {
				return errors.New("Block write verification failed")
			}
		}
		return nil
	}
}

func getFilledBlock(target []byte, size int64, fillVal byte) {
	for i := int64(0); i < size; i++ {
		target[i] = fillVal
	}
}

func getRandomBlock(target []byte, size int64) error {
	_, err := rand.Read(target)
	return err
}

type QemuImagePartitionType struct {
	GPT gpt.Type
	MBR mbr.Type
}

var UNSUPPORTED_MBR_PARTITION_TYPE mbr.Type = 0xff
var UNSUPPORTED_GPT_PARTITION_TYPE gpt.Type = "UNSUPPORTED"

var (
	Unused                   = &QemuImagePartitionType{gpt.Unused, mbr.Empty}
	MbrBoot                  = &QemuImagePartitionType{gpt.MbrBoot, mbr.Empty}
	EFISystemPartition       = &QemuImagePartitionType{gpt.EFISystemPartition, mbr.EFISystem}
	BiosBoot                 = &QemuImagePartitionType{gpt.BiosBoot, mbr.Empty}
	MicrosoftReserved        = &QemuImagePartitionType{gpt.MicrosoftReserved, mbr.Empty}
	MicrosoftBasicData       = &QemuImagePartitionType{gpt.MicrosoftBasicData, mbr.Empty}
	MicrosoftLDMMetadata     = &QemuImagePartitionType{gpt.MicrosoftLDMMetadata, mbr.Empty}
	MicrosoftLDMData         = &QemuImagePartitionType{gpt.MicrosoftLDMData, mbr.Empty}
	MicrosoftWindowsRecovery = &QemuImagePartitionType{gpt.MicrosoftWindowsRecovery, mbr.Empty}
	LinuxFilesystem          = &QemuImagePartitionType{gpt.LinuxFilesystem, mbr.Linux}
	LinuxRaid                = &QemuImagePartitionType{gpt.LinuxRaid, mbr.Linux}
	LinuxRootX86             = &QemuImagePartitionType{gpt.LinuxRootX86, mbr.Linux}
	LinuxRootX86_64          = &QemuImagePartitionType{gpt.LinuxRootX86_64, mbr.Linux}
	LinuxRootArm32           = &QemuImagePartitionType{gpt.LinuxRootArm32, mbr.Linux}
	LinuxRootArm64           = &QemuImagePartitionType{gpt.LinuxRootArm64, mbr.Linux}
	LinuxSwap                = &QemuImagePartitionType{gpt.LinuxSwap, mbr.Linux}
	LinuxLVM                 = &QemuImagePartitionType{gpt.LinuxLVM, mbr.LinuxLVM}
	LinuxDMCrypt             = &QemuImagePartitionType{gpt.LinuxDMCrypt, mbr.Linux}
	LinuxLUKS                = &QemuImagePartitionType{gpt.LinuxLUKS, mbr.Linux}
	VMWareFilesystem         = &QemuImagePartitionType{gpt.VMWareFilesystem, mbr.VMWareFS}
	Fat12                    = &QemuImagePartitionType{UNSUPPORTED_GPT_PARTITION_TYPE, mbr.Fat12}
	XenixRoot                = &QemuImagePartitionType{UNSUPPORTED_GPT_PARTITION_TYPE, mbr.XenixRoot}
	XenixUsr                 = &QemuImagePartitionType{UNSUPPORTED_GPT_PARTITION_TYPE, mbr.XenixUsr}
	Fat16                    = &QemuImagePartitionType{UNSUPPORTED_GPT_PARTITION_TYPE, mbr.Fat16}
	ExtendedCHS              = &QemuImagePartitionType{UNSUPPORTED_GPT_PARTITION_TYPE, mbr.ExtendedCHS}
	Fat16b                   = &QemuImagePartitionType{UNSUPPORTED_GPT_PARTITION_TYPE, mbr.Fat16b}
	NTFS                     = &QemuImagePartitionType{UNSUPPORTED_GPT_PARTITION_TYPE, mbr.NTFS}
	CommodoreFAT             = &QemuImagePartitionType{UNSUPPORTED_GPT_PARTITION_TYPE, mbr.CommodoreFAT}
	Fat32CHS                 = &QemuImagePartitionType{UNSUPPORTED_GPT_PARTITION_TYPE, mbr.Fat32CHS}
	Fat32LBA                 = &QemuImagePartitionType{UNSUPPORTED_GPT_PARTITION_TYPE, mbr.Fat32LBA}
	Fat16bLBA                = &QemuImagePartitionType{UNSUPPORTED_GPT_PARTITION_TYPE, mbr.Fat16bLBA}
	ExtendedLBA              = &QemuImagePartitionType{UNSUPPORTED_GPT_PARTITION_TYPE, mbr.ExtendedLBA}
	Linux                    = &QemuImagePartitionType{gpt.LinuxFilesystem, mbr.Linux}
	LinuxExtended            = &QemuImagePartitionType{gpt.LinuxFilesystem, mbr.LinuxExtended}
	Iso9660                  = &QemuImagePartitionType{UNSUPPORTED_GPT_PARTITION_TYPE, mbr.Iso9660}
	MacOSXUFS                = &QemuImagePartitionType{UNSUPPORTED_GPT_PARTITION_TYPE, mbr.MacOSXUFS}
	MacOSXBoot               = &QemuImagePartitionType{UNSUPPORTED_GPT_PARTITION_TYPE, mbr.MacOSXBoot}
	HFS                      = &QemuImagePartitionType{UNSUPPORTED_GPT_PARTITION_TYPE, mbr.HFS}
	Solaris8Boot             = &QemuImagePartitionType{UNSUPPORTED_GPT_PARTITION_TYPE, mbr.Solaris8Boot}
	GPTProtective            = &QemuImagePartitionType{gpt.EFISystemPartition, mbr.EFISystem}
	EFISystem                = &QemuImagePartitionType{gpt.EFISystemPartition, mbr.EFISystem}
	VMWareFS                 = &QemuImagePartitionType{gpt.VMWareFilesystem, mbr.VMWareFS}
	VMWareSwap               = &QemuImagePartitionType{UNSUPPORTED_GPT_PARTITION_TYPE, mbr.VMWareSwap}
)

func getMbrPartitionTypeCheck(inType mbr.Type) *QemuImagePartitionType {
	switch inType {
	case mbr.Empty:
		return Unused
	case mbr.Fat12:
		return Fat12
	case mbr.XenixRoot:
		return XenixRoot
	case mbr.XenixUsr:
		return XenixUsr
	case mbr.Fat16:
		return Fat16
	case mbr.ExtendedCHS:
		return ExtendedCHS
	case mbr.Fat16b:
		return Fat16b
	case mbr.NTFS:
		return NTFS
	case mbr.CommodoreFAT:
		return CommodoreFAT
	case mbr.Fat32CHS:
		return Fat32CHS
	case mbr.Fat32LBA:
		return Fat32LBA
	case mbr.Fat16bLBA:
		return Fat16bLBA
	case mbr.ExtendedLBA:
		return ExtendedLBA
	case mbr.Linux:
		return Linux
	case mbr.LinuxExtended:
		return LinuxExtended
	case mbr.LinuxLVM:
		return LinuxLVM
	case mbr.Iso9660:
		return Iso9660
	case mbr.MacOSXUFS:
		return MacOSXUFS
	case mbr.MacOSXBoot:
		return MacOSXBoot
	case mbr.HFS:
		return HFS
	case mbr.Solaris8Boot:
		return Solaris8Boot
	case mbr.EFISystem:
		return EFISystem
	case mbr.VMWareFS:
		return VMWareFS
	case mbr.VMWareSwap:
		return VMWareSwap
	}
	return nil
}

func getGptPartitionTypeCheck(inType gpt.Type) *QemuImagePartitionType {
	switch inType {
	case gpt.Unused:
		return Unused
	case gpt.MbrBoot:
		return MbrBoot
	case gpt.EFISystemPartition:
		return EFISystemPartition
	case gpt.BiosBoot:
		return BiosBoot
	case gpt.MicrosoftReserved:
		return MicrosoftReserved
	case gpt.MicrosoftBasicData:
		return MicrosoftBasicData
	case gpt.MicrosoftLDMMetadata:
		return MicrosoftLDMMetadata
	case gpt.MicrosoftLDMData:
		return MicrosoftLDMData
	case gpt.MicrosoftWindowsRecovery:
		return MicrosoftWindowsRecovery
	case gpt.LinuxFilesystem:
		return LinuxFilesystem
	case gpt.LinuxRaid:
		return LinuxRaid
	case gpt.LinuxRootX86:
		return LinuxRootX86
	case gpt.LinuxRootX86_64:
		return LinuxRootX86_64
	case gpt.LinuxRootArm32:
		return LinuxRootArm32
	case gpt.LinuxRootArm64:
		return LinuxRootArm64
	case gpt.LinuxSwap:
		return LinuxSwap
	case gpt.LinuxLVM:
		return LinuxLVM
	case gpt.LinuxDMCrypt:
		return LinuxDMCrypt
	case gpt.LinuxLUKS:
		return LinuxLUKS
	case gpt.VMWareFilesystem:
		return VMWareFilesystem
	}
	return nil
}

var UnsupportedGptPartitionType error = errors.New("The requested partition type is not supported on GPT partition tables.")
var UnsupportedMbrPartitionType error = errors.New("The requested partition type is not supported on MBR partition tables.")

func NewQemuImagePartition(name string, start uint64, end uint64, size uint64, blkSize uint64, isGpt bool, partType *QemuImagePartitionType, bootable bool) (*QemuImagePartition, error) {
	prt := &QemuImagePartition{
		start:     start,
		startByte: start / blkSize,
		end:       end,
		endByte:   end / blkSize,
		size:      size,
		isGpt:     isGpt,
		partType:  partType,
		name:      name,
		bootable:  bootable,
		mountMap:  map[string]bool{},
	}
	if isGpt {
		if prt.partType.GPT == UNSUPPORTED_GPT_PARTITION_TYPE {
			return nil, UnsupportedGptPartitionType
		}
		uuid, _ := vutils.UUID.MakeUUIDString()
		prt.gptPart = &gpt.Partition{
			Start: start,
			End:   end,
			Size:  prt.size,
			Type:  prt.partType.GPT,
			GUID:  uuid,
		}
		if prt.partType.GPT == gpt.EFISystemPartition {
			prt.name = "EFI System"
			prt.gptPart.Name = prt.name
		} else {
			prt.SetName(name)
		}
	} else {
		if prt.partType.MBR == UNSUPPORTED_MBR_PARTITION_TYPE {
			return nil, UnsupportedMbrPartitionType
		}
		prt.mbrPart = &mbr.Partition{
			Start:    uint32(prt.start),
			Size:     uint32(prt.size/blkSize) + 1,
			Bootable: bootable,
			Type:     prt.partType.MBR,
		}
	}
	return prt, nil
}

func NewQemuImagePartitionMbr(sectorSize int, part *mbr.Partition) *QemuImagePartition {
	mbrType := getMbrPartitionTypeCheck(part.Type)
	if mbrType == nil {
		return nil
	}
	blkSize := uint64(sectorSize)
	prt := &QemuImagePartition{
		start:     uint64(part.Start),
		startByte: uint64(part.Start) * blkSize,
		size:      uint64(part.Size),
		isGpt:     false,
		mbrPart:   part,
		partType:  mbrType,
		bootable:  part.Bootable,
		mountMap:  map[string]bool{},
	}
	prt.size = prt.start + prt.size
	prt.endByte = prt.startByte + prt.size
	return prt
}

func NewQemuImagePartitionGpt(sectorSize int, part *gpt.Partition) *QemuImagePartition {
	gptType := getGptPartitionTypeCheck(part.Type)
	if gptType == nil {
		return nil
	}
	blkSize := uint64(sectorSize)
	prt := &QemuImagePartition{
		start:     part.Start,
		startByte: part.Start * blkSize,
		size:      part.Size,
		end:       part.End,
		endByte:   part.End * blkSize,
		isGpt:     true,
		gptPart:   part,
		partType:  gptType,
		name:      part.Name,
		guid:      part.GUID,
		mountMap:  map[string]bool{},
	}
	return prt
}

type QemuPartitionFsType string

const (
	FS_EXFAT QemuPartitionFsType = "exfat"
	FS_EXT4  QemuPartitionFsType = "ext4"
	FS_FAT32 QemuPartitionFsType = "fat32"
	FS_NTFS  QemuPartitionFsType = "ntfs"
	FS_ZFS   QemuPartitionFsType = "zfs"
)

type QemuImagePartition struct {
	index      int
	start      uint64
	startByte  uint64
	size       uint64
	end        uint64
	endByte    uint64
	partType   *QemuImagePartitionType
	mbrPart    *mbr.Partition
	gptPart    *gpt.Partition
	isGpt      bool
	isFree     bool
	name       string
	bootable   bool
	guid       string
	mounted    bool
	mountPoint string
	mountCount uint
	mountMap   map[string]bool
	img        *QemuImage
	dev        string
}

func (qp *QemuImagePartition) setImage(index int, img *QemuImage) {
	qp.img = img
	qp.index = index
	qp.dev = fmt.Sprintf("%sp%d", qp.img.connectedDevice, qp.index+1)
	if !vutils.Files.PathExists(qp.dev) && qp.img.tableSaved {
		//run part probe..
		runPartProbe(qp.img.connectedDevice)
	}
	qp.GetFilesystem()
}

func (qp *QemuImagePartition) SetName(name string) {
	if !qp.isGpt || qp.partType.GPT == gpt.EFISystemPartition {
		return
	}
	qp.name = name
	qp.gptPart.Name = name
}

func (qp *QemuImagePartition) SetBootable(isBootable bool) {
	if qp.isGpt {
		return
	}
	qp.bootable = isBootable
	qp.mbrPart.Bootable = true
}

func (qp *QemuImagePartition) GetFilesystem() (string, error) {
	if !qp.img.tableSaved {
		return "", errors.New("Partition table ot saved cannot probe")
	}
	typ, err := getFilesystemType(qp.dev)
	if err != nil {
		return "", err
	}
	fmt.Println("FS TYPE: " + typ)
	return typ, nil
}

func (qp *QemuImagePartition) MakeFilesystem(fsType QemuPartitionFsType) error {
	var err error = nil
	switch fsType {
	case FS_EXFAT:
		err = MakeExfat(qp.name, qp.dev)
		break
	case FS_EXT4:
		err = MakeExt4(qp.name, qp.dev)
		break
	case FS_FAT32:
		err = MakeFat32(qp.name, qp.dev)
		break
	}
	if err != nil {
		return err
	}
	qp.GetFilesystem()
	return nil
}

func (qp *QemuImagePartition) Mount() (string, error) {
	//mount will create temporary mount point and mount the system there...
	if qp.mounted {
		return "", ImageMountedError
	}
	tdir, err := ioutil.TempDir("", "prmnt")
	if err != nil {
		return "", err
	}
	qp.mountPoint = tdir
	if err := qp.MountAt(qp.mountPoint); err != nil {
		os.RemoveAll(qp.mountPoint)
		return "", err
	}
	qp.mounted = true
	return qp.mountPoint, nil
}

func (qp *QemuImagePartition) MountAt(path string) error {
	if v, ok := qp.mountMap[path]; !ok || (ok && !v) {
		if !vutils.Files.PathExists(path) {
			return errors.New("Mount point does not exist")
		}
		cmd := vutils.Exec.CreateAsyncCommand("mount", false, qp.dev, path)
		err := cmd.StartAndWait()
		if err != nil {
			return err
		}
		qp.mountMap[path] = true
		qp.mountCount++
		return nil
	} else {
		return ImageMountedError
	}
}

func (qp *QemuImagePartition) IsMounted() bool {
	//this will unmount the temporary mountpoint
	return qp.mounted
}

func (qp *QemuImagePartition) Unmount() error {
	//this will unmount the temporary mountpoint
	if !qp.mounted {
		return ImageNotMountedError
	}
	if err := qp.UnmountAt(qp.mountPoint); err != nil {
		return err
	}
	qp.mounted = false
	qp.mountPoint = ""
	return nil
}

func (qp *QemuImagePartition) UnmountAt(path string) error {
	if v, ok := qp.mountMap[path]; ok && v {
		cmd := vutils.Exec.CreateAsyncCommand("umount", false, path)
		err := cmd.StartAndWait()
		if err != nil {
			return err
		}
		delete(qp.mountMap, path)
		qp.mountCount--
		return nil
	} else {
		return ImageNotMountedError
	}
}

func (qp *QemuImagePartition) SecureWipe(level QemuImageWipeLevel) error {
	if qp.mounted {
		return ImageMountedError
	}

	chunkSize := int64(qp.endByte - qp.startByte)

	// calculate total number of parts the file will be chunked into
	chunkParts := int64(math.Ceil(float64(chunkSize) / float64(IMAGE_WIPE_CHUNK_SIZE)))

	//if this is a compressed/sparse image then there is an issue - we need to overwrite the contents for the block device...
	if isBlk, err := qp.img.isBlockDevice(); err != nil {
		return err
	} else if isBlk {
		return runSecureErase(qp.img.fd, chunkSize, chunkParts, level, int64(qp.startByte), int64(qp.endByte))
	} else {
		//this isnt a block device so its format dependant, but, weve used ndb and there should be a device - does it exist?
		if qp.img.connected && qp.dev != "" {
			isBlk, err := qp.isBlockDevice()
			if err != nil {
				return err
			} else if isBlk {
				f, chunkParts, err := openTargetForErase(qp.dev)
				if err != nil {
					return err
				}
				defer f.Close()
				return runSecureErase(f, chunkSize, chunkParts, level, 0, chunkSize)
			}
		}
	}
	return errors.New("There was an error whilst erasing")
}

func (qp *QemuImagePartition) isBlockDevice() (bool, error) {
	fi, err := os.Stat(qp.dev)
	if err != nil {
		return false, err
	}
	if fi.Mode()&os.ModeDevice != 0 {
		return true, nil
	} else if fi.IsDir() {
		//doesnt support directories
		return false, errors.New("Directories are not supported for image backends. Only use block devices or image files.")
	} else {
		return false, nil
	}
}

func (qp *QemuImagePartition) GrowPart() error {
	//this will unmount the temporary mountpoint
	if qp.mounted {
		return ImageMountedError
	}
	//lets check there are no partitions after this one...
	_, err := qp.img.GetPartition(qp.index + 1)
	if err == PartitionMissingError {
		//we are good, lets do this!
		cmd := vutils.Exec.CreateAsyncCommand("growpart", false, qp.img.connectedDevice, fmt.Sprintf("%d", qp.index))
		err = cmd.BindToStdoutAndStdErr().StartAndWait()
		if err != nil {
			return err
		}
		cmd = vutils.Exec.CreateAsyncCommand("resize2fs", false, qp.dev)
		err = cmd.BindToStdoutAndStdErr().StartAndWait()
		if err != nil {
			return err
		}
		return nil

	} else {
		return errors.New("Unable to grow this partition as there are other partions after this one")
	}

}

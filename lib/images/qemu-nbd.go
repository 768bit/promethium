package images

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"os/exec"
	"os/user"
	"regexp"
	"strconv"
	"strings"

	"github.com/768bit/promethium/lib/images/diskfs/disk"
	"github.com/768bit/promethium/lib/images/diskfs/partition"
	"github.com/768bit/promethium/lib/images/diskfs/partition/gpt"
	"github.com/768bit/promethium/lib/images/diskfs/partition/mbr"
	"github.com/768bit/vutils"
	"github.com/palantir/stacktrace"
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
	devMap  map[string]*QcowImage
}

var NoQemuNbdDeviceAvailable error = errors.New("No Free Qemu NBD Devices are available")

func (qn *qemuNbd) getFirstDev() string {
	for _, dev := range qn.devList {
		_, ok := qn.devMap[dev]
		if !ok {
			return dev
		}
	}
	return ""
}

func (qn *qemuNbd) connect(image *QcowImage) error {
	for {
		dev := qn.getFirstDev()
		if dev == "" {
			return NoQemuNbdDeviceAvailable
		}
		cmd := vutils.Exec.CreateAsyncCommand("qemu-nbd", false, "-c", dev, "-f", "qcow2", image.path).Sudo().BindToStdoutAndStdErr()
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

func (qn *qemuNbd) disconnect(image *QcowImage) error {
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
	qn.devMap = map[string]*QcowImage{}
	return qn
}

var QemuNbd *qemuNbd = newQemuNbdDaemon()

func CreateNewQcowImage(path string, size uint64) (*QcowImage, error) {
	qi := &QcowImage{
		path: path,
	}
	if err := qi.create(size); err != nil {
		return nil, err
	} else if err := qi.init(); err != nil {
		return nil, err
	}
	return qi, nil
}

func LoadQcowImage(path string) (*QcowImage, error) {
	println("Loading qcow image at path: " + path)
	qi := &QcowImage{
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

type QcowImage struct {
	fd                 *os.File
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
	partitions         []*QcowImagePartition
	disk               *disk.Disk
	tableSaved         bool
}

func (qi *QcowImage) create(size uint64) error {
	//create the new image at path...
	cmd := vutils.Exec.CreateAsyncCommand("qemu-img", false, "create", "-f", "qcow2", qi.path, fmt.Sprintf("%d", size)).BindToStdoutAndStdErr()
	return cmd.StartAndWait()
}

func (qi *QcowImage) init() error {
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
	return qi.getInfo()
}

func (qi *QcowImage) getInfo() error {
	//get status info etc...
	cmd := vutils.Exec.CreateAsyncCommand("qemu-img", false, "info", "--output=json", "-f", "qcow2", qi.path).
		CaptureStdoutAndStdErr(false, true)
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
	if v, err := qi.SizeOnDisk(); err != nil {
		return err
	} else {
		qi.sizeOnDisk = v
	}

	return nil
}

func (qi *QcowImage) VirtualSize() uint64 {
	return qi.size
}

func (qi *QcowImage) SizeOnDisk() (uint64, error) {
	return vutils.Files.FileSize(qi.path)
}

func (qi *QcowImage) Resize(newSize uint64) error {
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

func (qi *QcowImage) Connect() error {
	if qi.connected {
		return nil
	}
	err := QemuNbd.connect(qi)
	if err != nil {
		return err
	}
	return qi.loadDisk()
}

func (qi *QcowImage) Disconnect() error {
	if !qi.connected {
		return ImageNotConnected
	} else if qi.mounted {
		return ImageMountedError
	}
	if err := qi.closeDisk(); err != nil {
		fmt.Println(err)
	}
	return QemuNbd.disconnect(qi)
}

func (qi *QcowImage) loadDisk() error {
	qi.tableSaved = true
	//quick chown...
	cmd := vutils.Exec.CreateAsyncCommand("chown", false, user.Current(), qi.connectedDevice).Sudo()
	if err := cmd.StartAndWait(); err != nil {
		return err
	}
	f, err := os.OpenFile(qi.connectedDevice, os.O_RDWR, 0660)
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
		qi.isGpt = false
		qi.mbr = nil
		qi.gpt = nil
		qi.initialised = false
		qi.tableSaved = false
	} else {
		qi.table = tb
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
		}
	}
	//}
	return qi.loadParts()
}

func (qi *QcowImage) closeDisk() error {
	qi.disk.File.Sync()
	err := qi.disk.File.Close()
	if err != nil {
		return err
	}
	qi.fd = nil
	cmd := vutils.Exec.CreateAsyncCommand("chown", false, "root", qi.connectedDevice).Sudo()
	if err := cmd.StartAndWait(); err != nil {
		return err
	}
	return nil
}

func (qi *QcowImage) loadParts() error {
	qi.partitions = []*QcowImagePartition{}
	if qi.initialised {
		if qi.isGpt {
			if qi.gpt != nil && qi.gpt.Partitions != nil && len(qi.gpt.Partitions) > 0 {
				for index, part := range qi.gpt.Partitions {
					if part.Start == 0 && part.End == 0 {
						continue
					}
					np := NewQcowImagePartitionGpt(qi.logicalSectorSize, part)
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
					np := NewQcowImagePartitionMbr(qi.logicalSectorSize, part)
					np.setImage(index, qi)
					qi.partitions = append(qi.partitions, np)
				}
			}
		}
	}
	return nil
}

var EFI_END_POS = uint64((100*1024*1024)/512) + (RESERVED_START_BYTES)

func (qi *QcowImage) MakeGpt() error {
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

func (qi *QcowImage) MakeMbr() error {
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
	return qi.WriteTable()
}

func (qi *QcowImage) WriteTable() error {
	if qi.tableSaved {
		return nil
	}
	if !qi.connected {
		return ImageNotConnected
	} else if qi.mounted {
		return ImageMountedError
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
	err = runPartProbe(qi.connectedDevice)
	if err != nil {
		//return err
	}
	return qi.loadParts()
}

var PartitionMissingError = errors.New("Cannot retrieve partition as it doesnt exist")

func (qi *QcowImage) GetPartition(partitionNumber int) (*QcowImagePartition, error) {
	partLen := len(qi.partitions)
	if partLen == 0 || partitionNumber >= partLen {
		return nil, PartitionMissingError
	}
	return qi.partitions[partitionNumber], nil
}

func (qi *QcowImage) MakeFilesystem(name string, fsType ImageFsType) error {
	var err error = nil
	switch fsType {
	case ExFat:
		err = MakeExfat(name, qi.connectedDevice)
		break
	case Ext4:
		err = MakeExt4(name, qi.connectedDevice)
		break
	case Fat32:
		err = MakeFat32(name, qi.connectedDevice)
		break
	}
	if err != nil {
		return err
	}
	return nil
}

func (qi *QcowImage) Mount() (string, error) {
	//mount will create temporary mount point and mount the system there...
	if qi.mounted {
		return "", ImageMountedError
	}
	tdir, err := ioutil.TempDir("", "prmnt")
	if err != nil {
		return "", err
	}
	qi.mountPoint = tdir
	cmd := vutils.Exec.CreateAsyncCommand("mount", false, qi.connectedDevice, qi.mountPoint).Sudo()
	err = cmd.StartAndWait()
	if err != nil {
		return "", err
	}
	qi.mounted = true
	return qi.mountPoint, nil
}

func (qi *QcowImage) Unmount() error {
	//mount will create temporary mount point and mount the system there...
	if !qi.mounted {
		return ImageNotMountedError
	}
	cmd := vutils.Exec.CreateAsyncCommand("umount", false, qi.mountPoint).Sudo()
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

func (qi *QcowImage) CreatePartition(name string, strSize string, partitionType *QcowImagePartitionType, bootable bool) (*QcowImagePartition, error) {
	//need to figure out where in the array we place the new part...
	//additionally we need to see if there is an overlap..
	//if end > qi.size {
	//  return nil, errors.New(fmt.Sprintf("Cannot create a new partition as it ends past the extents of the disk"))
	//}

	sbytes := RESERVED_START_BYTES
	if qi.isGpt {
		sbytes = EFI_END_POS + 1
	}

	indexWatermark := 0
	var newPart *QcowImagePartition = nil
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
					np, err := NewQcowImagePartition(name, startSec, endSec, size, blkSize, qi.isGpt, partitionType, bootable)
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
					np, err := NewQcowImagePartition(name, startSec, endSec, size, blkSize, qi.isGpt, partitionType, bootable)
					if err != nil {
						return nil, err
					}
					newPart = np
					newPart.setImage(0, qi)
					for _, item := range qi.partitions {
						item.index++
					}
					qi.tableSaved = false
					qi.partitions = append([]*QcowImagePartition{np}, qi.partitions...)
					return newPart, nil
				} else {
					if endSec == diskSects {
						endSec--
						size = (endSec - startSec + 1) * blkSize
					}
					if endSec*blkSize > qi.size {
						return nil, errors.New("Cannot create partition as it will be larger than the extents")
					}
					np, err := NewQcowImagePartition(name, startSec, endSec, size, blkSize, qi.isGpt, partitionType, bootable)
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
					np, err := NewQcowImagePartition(name, startSec, endSec, size, blkSize, qi.isGpt, partitionType, bootable)
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
		np, err := NewQcowImagePartition(name, startSec, endSec, size, blkSize, qi.isGpt, partitionType, bootable)
		if err != nil {
			return nil, err
		}
		newPart = np
		newPart.setImage(0, qi)
		qi.tableSaved = false
		qi.partitions = []*QcowImagePartition{np}
		return newPart, nil
	}
	return newPart, errors.New("Error creating new partition")
}

func (qi *QcowImage) CreatePartitionAt(name string, start uint64, end uint64, partitionType *QcowImagePartitionType, bootable bool) (*QcowImagePartition, error) {
	//need to figure out where in the array we place the new part...
	//additionally we need to see if there is an overlap..
	if end > qi.size {
		return nil, errors.New(fmt.Sprintf("Cannot create a new partition as it ends past the extents of the disk"))
	}
	indexWatermark := 0
	var newPart *QcowImagePartition = nil
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
					np, err := NewQcowImagePartition(name, startSec, endSec, size, blkSize, qi.isGpt, partitionType, bootable)
					if err != nil {
						return nil, err
					}
					newPart = np
					newPart.setImage(0, qi)
					for _, item := range qi.partitions {
						item.index++
					}
					qi.tableSaved = false
					qi.partitions = append([]*QcowImagePartition{np}, qi.partitions...)
					return newPart, nil
				} else {
					np, err := NewQcowImagePartition(name, startSec, endSec, size, blkSize, qi.isGpt, partitionType, bootable)
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
					np, err := NewQcowImagePartition(name, startSec, endSec, size, blkSize, qi.isGpt, partitionType, bootable)
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
		np, err := NewQcowImagePartition(name, startSec, endSec, size, blkSize, qi.isGpt, partitionType, bootable)
		if err != nil {
			return nil, err
		}
		newPart = np
		newPart.setImage(0, qi)
		qi.tableSaved = false
		qi.partitions = []*QcowImagePartition{np}
		return newPart, nil
	}
	return newPart, errors.New("Error creating new partition")
}

func (qi *QcowImage) Destroy() error {
	return os.Remove(qi.path)
}

type QcowImagePartitionType struct {
	GPT gpt.Type
	MBR mbr.Type
}

var UNSUPPORTED_MBR_PARTITION_TYPE mbr.Type = 0xff
var UNSUPPORTED_GPT_PARTITION_TYPE gpt.Type = "UNSUPPORTED"

var (
	Unused                   = &QcowImagePartitionType{gpt.Unused, mbr.Empty}
	MbrBoot                  = &QcowImagePartitionType{gpt.MbrBoot, mbr.Empty}
	EFISystemPartition       = &QcowImagePartitionType{gpt.EFISystemPartition, mbr.EFISystem}
	BiosBoot                 = &QcowImagePartitionType{gpt.BiosBoot, mbr.Empty}
	MicrosoftReserved        = &QcowImagePartitionType{gpt.MicrosoftReserved, mbr.Empty}
	MicrosoftBasicData       = &QcowImagePartitionType{gpt.MicrosoftBasicData, mbr.Empty}
	MicrosoftLDMMetadata     = &QcowImagePartitionType{gpt.MicrosoftLDMMetadata, mbr.Empty}
	MicrosoftLDMData         = &QcowImagePartitionType{gpt.MicrosoftLDMData, mbr.Empty}
	MicrosoftWindowsRecovery = &QcowImagePartitionType{gpt.MicrosoftWindowsRecovery, mbr.Empty}
	LinuxFilesystem          = &QcowImagePartitionType{gpt.LinuxFilesystem, mbr.Linux}
	LinuxRaid                = &QcowImagePartitionType{gpt.LinuxRaid, mbr.Linux}
	LinuxRootX86             = &QcowImagePartitionType{gpt.LinuxRootX86, mbr.Linux}
	LinuxRootX86_64          = &QcowImagePartitionType{gpt.LinuxRootX86_64, mbr.Linux}
	LinuxRootArm32           = &QcowImagePartitionType{gpt.LinuxRootArm32, mbr.Linux}
	LinuxRootArm64           = &QcowImagePartitionType{gpt.LinuxRootArm64, mbr.Linux}
	LinuxSwap                = &QcowImagePartitionType{gpt.LinuxSwap, mbr.Linux}
	LinuxLVM                 = &QcowImagePartitionType{gpt.LinuxLVM, mbr.LinuxLVM}
	LinuxDMCrypt             = &QcowImagePartitionType{gpt.LinuxDMCrypt, mbr.Linux}
	LinuxLUKS                = &QcowImagePartitionType{gpt.LinuxLUKS, mbr.Linux}
	VMWareFilesystem         = &QcowImagePartitionType{gpt.VMWareFilesystem, mbr.VMWareFS}
	Fat12                    = &QcowImagePartitionType{UNSUPPORTED_GPT_PARTITION_TYPE, mbr.Fat12}
	XenixRoot                = &QcowImagePartitionType{UNSUPPORTED_GPT_PARTITION_TYPE, mbr.XenixRoot}
	XenixUsr                 = &QcowImagePartitionType{UNSUPPORTED_GPT_PARTITION_TYPE, mbr.XenixUsr}
	Fat16                    = &QcowImagePartitionType{UNSUPPORTED_GPT_PARTITION_TYPE, mbr.Fat16}
	ExtendedCHS              = &QcowImagePartitionType{UNSUPPORTED_GPT_PARTITION_TYPE, mbr.ExtendedCHS}
	Fat16b                   = &QcowImagePartitionType{UNSUPPORTED_GPT_PARTITION_TYPE, mbr.Fat16b}
	NTFS                     = &QcowImagePartitionType{UNSUPPORTED_GPT_PARTITION_TYPE, mbr.NTFS}
	CommodoreFAT             = &QcowImagePartitionType{UNSUPPORTED_GPT_PARTITION_TYPE, mbr.CommodoreFAT}
	Fat32CHS                 = &QcowImagePartitionType{UNSUPPORTED_GPT_PARTITION_TYPE, mbr.Fat32CHS}
	Fat32LBA                 = &QcowImagePartitionType{UNSUPPORTED_GPT_PARTITION_TYPE, mbr.Fat32LBA}
	Fat16bLBA                = &QcowImagePartitionType{UNSUPPORTED_GPT_PARTITION_TYPE, mbr.Fat16bLBA}
	ExtendedLBA              = &QcowImagePartitionType{UNSUPPORTED_GPT_PARTITION_TYPE, mbr.ExtendedLBA}
	Linux                    = &QcowImagePartitionType{gpt.LinuxFilesystem, mbr.Linux}
	LinuxExtended            = &QcowImagePartitionType{gpt.LinuxFilesystem, mbr.LinuxExtended}
	Iso9660                  = &QcowImagePartitionType{UNSUPPORTED_GPT_PARTITION_TYPE, mbr.Iso9660}
	MacOSXUFS                = &QcowImagePartitionType{UNSUPPORTED_GPT_PARTITION_TYPE, mbr.MacOSXUFS}
	MacOSXBoot               = &QcowImagePartitionType{UNSUPPORTED_GPT_PARTITION_TYPE, mbr.MacOSXBoot}
	HFS                      = &QcowImagePartitionType{UNSUPPORTED_GPT_PARTITION_TYPE, mbr.HFS}
	Solaris8Boot             = &QcowImagePartitionType{UNSUPPORTED_GPT_PARTITION_TYPE, mbr.Solaris8Boot}
	GPTProtective            = &QcowImagePartitionType{gpt.EFISystemPartition, mbr.EFISystem}
	EFISystem                = &QcowImagePartitionType{gpt.EFISystemPartition, mbr.EFISystem}
	VMWareFS                 = &QcowImagePartitionType{gpt.VMWareFilesystem, mbr.VMWareFS}
	VMWareSwap               = &QcowImagePartitionType{UNSUPPORTED_GPT_PARTITION_TYPE, mbr.VMWareSwap}
)

func getMbrPartitionTypeCheck(inType mbr.Type) *QcowImagePartitionType {
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

func getGptPartitionTypeCheck(inType gpt.Type) *QcowImagePartitionType {
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

func NewQcowImagePartition(name string, start uint64, end uint64, size uint64, blkSize uint64, isGpt bool, partType *QcowImagePartitionType, bootable bool) (*QcowImagePartition, error) {
	prt := &QcowImagePartition{
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

func NewQcowImagePartitionMbr(sectorSize int, part *mbr.Partition) *QcowImagePartition {
	mbrType := getMbrPartitionTypeCheck(part.Type)
	if mbrType == nil {
		return nil
	}
	blkSize := uint64(sectorSize)
	prt := &QcowImagePartition{
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

func NewQcowImagePartitionGpt(sectorSize int, part *gpt.Partition) *QcowImagePartition {
	gptType := getGptPartitionTypeCheck(part.Type)
	if gptType == nil {
		return nil
	}
	blkSize := uint64(sectorSize)
	prt := &QcowImagePartition{
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

type QcowPartitionFsType string

const (
	FS_EXFAT QcowPartitionFsType = "exfat"
	FS_EXT4  QcowPartitionFsType = "ext4"
	FS_FAT32 QcowPartitionFsType = "fat32"
	FS_NTFS  QcowPartitionFsType = "ntfs"
	FS_ZFS   QcowPartitionFsType = "zfs"
)

type QcowImagePartition struct {
	index      int
	start      uint64
	startByte  uint64
	size       uint64
	end        uint64
	endByte    uint64
	partType   *QcowImagePartitionType
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
	img        *QcowImage
	dev        string
}

func (qp *QcowImagePartition) setImage(index int, img *QcowImage) {
	qp.img = img
	qp.index = index
	qp.dev = fmt.Sprintf("%sp%d", qp.img.connectedDevice, qp.index+1)
	if !vutils.Files.PathExists(qp.dev) && qp.img.tableSaved {
		//run part probe..
		runPartProbe(qp.img.connectedDevice)
	}
	qp.GetFilesystem()
}

func (qp *QcowImagePartition) SetName(name string) {
	if !qp.isGpt || qp.partType.GPT == gpt.EFISystemPartition {
		return
	}
	qp.name = name
	qp.gptPart.Name = name
}

func (qp *QcowImagePartition) SetBootable(isBootable bool) {
	if qp.isGpt {
		return
	}
	qp.bootable = isBootable
	qp.mbrPart.Bootable = true
}

func (qp *QcowImagePartition) GetFilesystem() {
	if !qp.img.tableSaved {
		return
	}
	typ, err := getFilesystemType(qp.dev)
	if err != nil {
		return
	}
	fmt.Println("FS TYPE: " + typ)
}

func (qp *QcowImagePartition) MakeFilesystem(fsType QcowPartitionFsType) error {
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

func (qp *QcowImagePartition) Mount() (string, error) {
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

func (qp *QcowImagePartition) MountAt(path string) error {
	if v, ok := qp.mountMap[path]; !ok || (ok && !v) {
		if !vutils.Files.PathExists(path) {
			return errors.New("Mount point does not exist")
		}
		cmd := vutils.Exec.CreateAsyncCommand("mount", false, qp.dev, path).Sudo()
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

func (qp *QcowImagePartition) Unmount() error {
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

func (qp *QcowImagePartition) UnmountAt(path string) error {
	if v, ok := qp.mountMap[path]; ok && v {
		cmd := vutils.Exec.CreateAsyncCommand("umount", false, path).Sudo()
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

package vmm

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/768bit/firecracker-go-sdk"
	models "github.com/768bit/firecracker-go-sdk/client/models"
	"github.com/768bit/firecracker-go-sdk/client/operations"
	"github.com/768bit/vutils"
	"github.com/cloudius-systems/capstan/core"
	"github.com/cloudius-systems/capstan/util"
	log "github.com/sirupsen/logrus"
)

const UNKOWN_STATUS = "UNKNOWN_STATUS"

func NewFireCrackerProcess(id string, name string, cmd string, entryPoint string, cpus int64, memory int64, imageSize int64, autoStart bool) (*FireCrackerProcess, error) {
	fcp := &FireCrackerProcess{
		osvImageSize:      imageSize,
		id:                id,
		isOsv:             true,
		name:              name,
		cmd:               cmd,
		entryPoint:        entryPoint,
		cpus:              cpus,
		memory:            memory,
		autoStart:         autoStart,
		networkInterfaces: []string{},
	}
	return fcp.init()
}

func NewFireCrackerProcessImg(id string, name string, boot string, cpus int64, memory int64, kernelPath string, driveImages []string, networkInterfaces []string, autoStart bool) (*FireCrackerProcess, error) {
	if networkInterfaces == nil {
		networkInterfaces = []string{}
	}
	fcp := &FireCrackerProcess{
		id:                id,
		isOsv:             false,
		name:              name,
		kernelPath:        kernelPath,
		cpus:              cpus,
		memory:            memory,
		imageList:         driveImages,
		cmd:               boot,
		autoStart:         autoStart,
		networkInterfaces: networkInterfaces,
	}
	return fcp.init()
}

type FireCrackerProcess struct {
	name     string
	id       string
	logger   *log.Entry
	machine  *firecracker.Machine
	fcConfig firecracker.Config

	isOsv        bool
	autoStart    bool
	osvImageSize int64

	memory int64
	cpus   int64

	imageList         []string
	networkInterfaces []string

	jailerProc            *vutils.ExecAsyncCommand
	jailerBinaryPath      string
	firecrackerBinaryPath string
	kernelPath            string
	chrootPath            string
	socketPath            string
	imagePath             string
	cmd                   string
	entryPoint            string
	statusResp            *operations.DescribeInstanceOK
	template              *core.Template
	image                 *core.Image
	repo                  *util.Repo
	conn                  *firecracker.Client
	Status                string
	cloudInit             []byte

	isPolling bool
	exitChan  chan error
	stateChan chan string

	isRestarting   bool
	isShuttingDown bool

	procExitWaitChan chan error

	err error
	ctx context.Context

	cancelFunc context.CancelFunc
}

func (fcp *FireCrackerProcess) init() (*FireCrackerProcess, error) {
	firecrackerPath, jailerPath, err := getFirecrackerBinary()
	if err != nil {
		return fcp, err
	}
	fcp.firecrackerBinaryPath = firecrackerPath
	fcp.jailerBinaryPath = jailerPath

	fcp.chrootPath = filepath.Join(ROOT_PATH, "firecracker", fcp.id, "root")
	os.RemoveAll(fcp.chrootPath)
	fcp.socketPath = filepath.Join(fcp.chrootPath, "api.socket")
	fcp.procExitWaitChan = make(chan error)
	fcp.exitChan = make(chan error)
	fcp.stateChan = make(chan string)
	fcp.Status = UNKOWN_STATUS
	if fcp.jailerProc == nil {
		err = fcp.startFirecrackerProcess()
		if err != nil {
			return nil, err
		}
	}
	if fcp.autoStart {
		if err := fcp.Start(); err != nil {
			return fcp, err
		}
	}
	return fcp, nil
}

func (fcp *FireCrackerProcess) startFirecrackerProcess() error {
	fcp.jailerProc = vutils.Exec.CreateAsyncCommand(fcp.jailerBinaryPath, false,
		"--id", fcp.id,
		"--node", "0",
		"--exec-file", fcp.firecrackerBinaryPath,
		"--uid", strconv.Itoa(os.Getuid()),
		"--gid", strconv.Itoa(os.Getgid()),
		"--chroot-base-dir", ROOT_PATH).Sudo() //.BindToStdoutAndStdErr() //.CaptureStdoutAndStdErr(false, false)
	e := fcp.jailerProc.Start()
	go func() {
		fcp.procExitWaitChan <- fcp.jailerProc.Wait()
	}()
	return e
}

func (fcp *FireCrackerProcess) checkImage() {

}

func (fcp *FireCrackerProcess) GetStatus() string {
	return fcp.Status
}

func (fcp *FireCrackerProcess) Send(input string) error {
	return fcp.jailerProc.Write([]byte(input))
}

func (fcp *FireCrackerProcess) SetCloudInit(cloudConfig []byte) {
	fcp.cloudInit = cloudConfig
	fcp.cmd = fcp.cmd + " ds=nocloud-net"
}

func (fcp *FireCrackerProcess) Build() error {
	if img, repo, imgPath, err := fcp.runBuild(); err != nil {
		return err
	} else {
		fcp.repo = repo
		fcp.image = img
		fcp.imagePath = imgPath
	}
	return nil
}

func (fcp *FireCrackerProcess) Console() (io.ReadCloser, io.ReadCloser, io.WriteCloser, error) {
	if fcp.jailerProc == nil || fcp.Status != "Running" {
		return nil, nil, nil, errors.New("Cannot connect to console of non running VM")
	} else {
		outP, errP, inP := fcp.jailerProc.GetPipes()
		return outP, errP, inP, nil
	}
}

func (fcp *FireCrackerProcess) runBuild() (*core.Image, *util.Repo, string, error) {
	return BuildBaseCapstanImage(fcp.name, fcp.cmd, fcp.entryPoint, fcp.osvImageSize)
}

func (fcp *FireCrackerProcess) beginPollingLoop() {
	err := fcp.pollStatus()
	if err != nil {
		//fcp.logger.Warnf("Error when doing intial polling: %s", err.Error())
		fcp.exitChan <- err
		return
	}
	fcp.isPolling = true
	go func() {
		for {
			time.Sleep(time.Second * 2)
			if !fcp.isPolling {
				return
			}
			//perform the polling..
			err := fcp.pollStatus()
			if err != nil {
				fcp.exitChan <- err
				return
			}
		}
	}()
	go func() {
		//use our select switcher to check for changes in state...
		for {
			select {
			case err := <-fcp.exitChan:
				if err != nil {
					//there was an error in execution - lets break out of the loops and clean everything up..
					//the error will also imply a change of state
					fcp.err = err
					fcp.logger.Debugf("Error when doing polling: %s", err.Error())
					fcp.Status = "ERROR"
				}
				fcp.isPolling = false
				if fcp.jailerProc != nil && fcp.jailerProc.Proc != nil && fcp.jailerProc.Proc.ProcessState != nil && fcp.jailerProc.Proc.ProcessState.Exited() {
					//it has exited.. lets remake the process..
					err := fcp.startFirecrackerProcess()
					if err != nil {
						fcp.logger.Debugf("Error when restarting firecracker: %s", err.Error())
					} else {
						if fcp.autoStart {
							err = fcp.Start()
							if err != nil {
								fcp.logger.Debugf("Error when restarting vmm in autostart mode: %s", err.Error())
							}
						} else {
							fcp.cancelFunc()
						}
					}
				}
				return
				//break
			case newStatus := <-fcp.stateChan:
				if fcp.Status != newStatus {
					//state has changed - process this...
					fcp.logger.Warnf("Status Has changed %s -> %s", fcp.Status, newStatus)
				}
				//break
			}
		}
	}()
}

func (fcp *FireCrackerProcess) cleanUp() {
	//clean up firecracker and the jailer - lets tear everything down...

	os.Remove(fcp.fcConfig.SocketPath)
	os.RemoveAll(filepath.Join(fcp.chrootPath, "dev"))

}

func (fcp *FireCrackerProcess) pollStatus() error {
	res, err := fcp.conn.GetInstanceDescription()
	if err != nil {
		return err
	}
	fcp.statusResp = res
	status := *(fcp.statusResp.Payload.State)
	//fcp.logger.Warnf("Status %s", status)
	if fcp.Status == UNKOWN_STATUS {
		fcp.Status = status
	} else {
		fcp.stateChan <- status
	}
	return nil
}

func (fcp *FireCrackerProcess) Start() error {

	if fcp.jailerProc != nil && fcp.jailerProc.Proc != nil && fcp.jailerProc.Proc.ProcessState != nil && fcp.jailerProc.Proc.ProcessState.Exited() {
		fcp.cleanUp()
		err := fcp.startFirecrackerProcess()
		if err != nil {
			return err
		}
	} else if fcp.jailerProc == nil {
		fcp.cleanUp()
		err := fcp.startFirecrackerProcess()
		if err != nil {
			return err
		}
	}
	if fcp.isOsv {
		return fcp.startOsv()
	}

	log.Println("Firecracker started for", fcp.id, "KERNEL:", fcp.kernelPath, "IMAGE:", fcp.imagePath, "BOOT:", fcp.cmd)
	//now we need to connect to the relevant socket...
	fcp.ctx, fcp.cancelFunc = context.WithCancel(context.Background())

	driveList := make([]models.Drive, len(fcp.imageList))

	for index, img := range fcp.imageList {
		driveName := fmt.Sprintf("drive%d", index)
		if index == 0 {
			driveName = "rootfs"
		}
		destPath := filepath.Join(fcp.chrootPath, driveName+".img")
		if !vutils.Files.PathExists(destPath) {
			err := qemuConvertImgRaw(img, destPath)
			if err != nil {
				return err
			}
		}
		driveList[index] = models.Drive{
			DriveID:      firecracker.String(driveName),
			PathOnHost:   firecracker.String("/" + driveName + ".img"),
			IsRootDevice: firecracker.Bool(false),
			IsReadOnly:   firecracker.Bool(false),
		}
		if index == 0 {
			driveList[index].IsRootDevice = firecracker.Bool(true)
		}
	}

	if fcp.cloudInit != nil && len(fcp.cloudInit) > 0 {
		destPath := filepath.Join(fcp.chrootPath, "cloud-init.img")
		if !vutils.Files.PathExists(destPath) {
			err := ioutil.WriteFile(destPath, fcp.cloudInit, 0660)
			if err != nil {
				return err
			}
		}
		driveList = append(driveList, models.Drive{
			DriveID:      firecracker.String("cloud_init"),
			PathOnHost:   firecracker.String("/cloud-init.img"),
			IsRootDevice: firecracker.Bool(false),
			IsReadOnly:   firecracker.Bool(false),
		})
	}

	ifaceList := make([]firecracker.NetworkInterface, len(fcp.networkInterfaces))

	for index, iface := range fcp.networkInterfaces {
		ifaceList[index] = firecracker.NetworkInterface{
			MacAddress:  fmt.Sprintf("AA:FC:00:00:00:0%d", index),
			HostDevName: iface,
		}
	}

	logger := log.New()
	//logger.SetLevel(log.DebugLevel)
	fcp.logger = log.NewEntry(logger)

	fcp.conn = firecracker.NewClient(fcp.socketPath, fcp.logger, true)

	fcp.fcConfig = firecracker.Config{
		SocketPath:        fcp.socketPath,
		KernelImagePath:   "/kernel.elf",
		KernelArgs:        fcp.cmd,
		Drives:            driveList,
		NetworkInterfaces: ifaceList,
		//LogLevel:          "Debug",
		//LogFifo: filepath.Join(fcp.chrootPath, "log"),
		//MetricsFifo:filepath.Join(fcp.chrootPath, "metrics"),
		MachineCfg: models.MachineConfiguration{
			VcpuCount:  firecracker.Int64(fcp.cpus),
			MemSizeMib: firecracker.Int64(fcp.memory),
			HtEnabled:  firecracker.Bool(false),
		},
		//Debug: true,
	}

	//err := ioutil.WriteFile(filepath.Join(fcp.chrootPath, "log"), []byte{}, 0666)
	//if err != nil {
	//  return err
	//}
	//err = ioutil.WriteFile(filepath.Join(fcp.chrootPath, "metrics"), []byte{}, 0666)
	//if err != nil {
	//  return err
	//}
	log.Println("Creating machine")
	m, err := firecracker.NewMachine(fcp.ctx, fcp.fcConfig, firecracker.WithLogger(fcp.logger), firecracker.WithClient(fcp.conn), firecracker.WithProcessRunner(fcp.jailerProc.Proc))
	if err != nil {
		return err
	}

	m.Handlers.FcInit = FCHandlerList
	m.Handlers.Validation = m.Handlers.Validation.Clear()
	kpath := filepath.Join(fcp.chrootPath, "kernel.elf")
	//create hard links for resources...
	if !vutils.Files.PathExists(kpath) {
		err = vutils.Files.Copy(fcp.kernelPath, kpath)
		if err != nil {
			return err
		}
	}

	log.Println("Starting machine")

	fcp.machine = m

	err = chown(fcp.chrootPath)
	if err != nil {
		return err
	}

	if err := m.Start(fcp.ctx); err != nil {
		return err
	}

	fcp.beginPollingLoop() //while we wait we will begin polling for status...

	return nil

}

func (fcp *FireCrackerProcess) startOsv() error {

	osvRelease, err := getCapstanDevPath()
	if err != nil {
		return err
	}
	fcp.kernelPath = filepath.Join(osvRelease, "loader-stripped.elf")
	if !vutils.Files.CheckPathExists(fcp.kernelPath) {
		return errors.New("Unable to fins OSV ELF Kernel")
	}
	log.Println("Firecracker started for", fcp.id, "KERNEL", fcp.kernelPath, "IMAGE", fcp.imagePath)
	//now we need to connect to the relevant socket...
	fcp.ctx, fcp.cancelFunc = context.WithCancel(context.Background())
	db := models.Drive{
		DriveID:      firecracker.String("rootfs"),
		PathOnHost:   firecracker.String("/usr.img"),
		IsRootDevice: firecracker.Bool(false),
		IsReadOnly:   firecracker.Bool(false),
	}

	fcp.logger = log.NewEntry(log.New())

	fcp.conn = firecracker.NewClient(fcp.socketPath, fcp.logger, false)

	fcp.fcConfig = firecracker.Config{
		SocketPath:      fcp.socketPath,
		KernelImagePath: "/kernel.elf",
		KernelArgs:      "--power-off-on-abort --nopci --verbose " + fcp.cmd,
		Drives:          []models.Drive{db},
		MachineCfg: models.MachineConfiguration{
			VcpuCount:  firecracker.Int64(fcp.cpus),
			MemSizeMib: firecracker.Int64(fcp.memory),
			HtEnabled:  firecracker.Bool(false),
		},
	}
	log.Println("Creating machine")
	m, err := firecracker.NewMachine(fcp.ctx, fcp.fcConfig, firecracker.WithLogger(fcp.logger), firecracker.WithClient(fcp.conn), firecracker.WithProcessRunner(fcp.jailerProc.Proc))
	if err != nil {
		return err
	}

	m.Handlers.FcInit = FCHandlerList
	m.Handlers.Validation = m.Handlers.Validation.Clear()
	kpath := filepath.Join(fcp.chrootPath, "kernel.elf")
	//create hard links for resources...
	if !vutils.Files.PathExists(kpath) {
		err = vutils.Files.Copy(fcp.kernelPath, kpath)
		if err != nil {
			return err
		}
	}
	ipath := filepath.Join(fcp.chrootPath, "usr.img")
	if !vutils.Files.PathExists(ipath) {
		err = qemuConvertImgRaw(fcp.imagePath, ipath)
		if err != nil {
			return err
		}
	}

	log.Println("Starting machine")

	fcp.machine = m

	if err := m.Start(fcp.ctx); err != nil {
		return err
	}

	fcp.beginPollingLoop() //while we wait we will begin polling for status...

	return nil

}

func (fcp *FireCrackerProcess) Wait() error {
	for {
		select {
		case waitErr := <-fcp.procExitWaitChan:
			if !fcp.isShuttingDown && !fcp.isRestarting {
				return waitErr
			}
		}
	}
	//if fcp.autoStart {
	return nil
	//}
	//return fcp.jailerProc.Wait()
	//return fcp.machine.Wait(fcp.ctx)

}

func (fcp *FireCrackerProcess) terminateInitialise() error {

	//basically forcefully tear down the entire stack and reinitialise it - building etc. will not be necessary...

	fcp.cleanUp()

	return nil

}

func (fcp *FireCrackerProcess) Stop() error {
	fcp.isShuttingDown = true
	if fcp.jailerProc != nil && fcp.jailerProc.Proc != nil && fcp.jailerProc.Proc.ProcessState != nil && !fcp.jailerProc.Proc.ProcessState.Exited() {
		fcp.jailerProc.Proc.Process.Kill()
	}
	fcp.cleanUp()
	fcp.isShuttingDown = false
	// a forced termination
	return nil

}

func (fcp *FireCrackerProcess) Shutdown() error {
	fcp.isShuttingDown = true

	// a graceful shutdown..
	e := fcp.machine.Shutdown(fcp.ctx)
	fcp.isShuttingDown = false
	return e
}

func (fcp *FireCrackerProcess) ShutdownTimeout(timeout time.Duration) error {
	fcp.isShuttingDown = true
	// a graceful shutdown..
	ctx, cancel := context.WithTimeout(fcp.ctx, timeout)
	defer cancel()

	e := fcp.machine.Shutdown(ctx)
	if e != nil {
		return e
	}
	select {
	case <-time.After(1 * time.Second):
		fmt.Println("overslept")
	case <-ctx.Done():
		fmt.Println(ctx.Err()) // prints "context deadline exceeded"
	}
	fcp.isShuttingDown = false
	return fcp.Wait()

}

func (fcp *FireCrackerProcess) Restart() error {

	fcp.isRestarting = true

	// a graceful shutdown..
	e := fcp.machine.Shutdown(fcp.ctx)
	if e != nil {
		return e
	}
	go func() {
		e := fcp.jailerProc.Wait()
		if e != nil {
			return
		}
		e = fcp.machine.Start(fcp.ctx)
		fcp.isRestarting = false
	}()
	return e

}

func (fcp *FireCrackerProcess) RestartTimeout() error {

	// a graceful restart
	return nil

}

func (fcp *FireCrackerProcess) Reset() error {

	// forced reset
	return nil

}

func (fcp *FireCrackerProcess) Terminate() error {

	// forcfully remove all items associated with the firecracker process (a full cleanup)
	return nil

}

func (fcp *FireCrackerProcess) Destroy() error {

	// forcfully remove all items associated with the firecracker process (a full cleanup)
	return nil

}

func qemuConvertImgRaw(src string, dest string) error {
	return vutils.Exec.CreateAsyncCommand("qemu-img", false, "convert", "-O", "raw", src, dest).Sudo().BindToStdoutAndStdErr().StartAndWait()
}

func chown(path string) error {
	return vutils.Exec.CreateAsyncCommand("chown", false, "-R", fmt.Sprintf("%d:%d", os.Getuid(), os.Getgid()), path).Sudo().BindToStdoutAndStdErr().StartAndWait()
}

var FCHandlerList = firecracker.HandlerList{}.Append(
	//StartVMMHandler, //we handle the jailer process - this is to make it usable across a number of scenarios - docker/lxc and qemu can be supported
	firecracker.CreateLogFilesHandler,
	firecracker.BootstrapLoggingHandler,
	firecracker.CreateMachineHandler,
	firecracker.CreateBootSourceHandler,
	firecracker.AttachDrivesHandler,
	firecracker.CreateNetworkInterfacesHandler,
	firecracker.AddVsocksHandler,
)

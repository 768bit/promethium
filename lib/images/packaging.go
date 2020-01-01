package images

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/768bit/promethium/lib/common"
	"github.com/768bit/promethium/lib/config"
	"github.com/768bit/vutils"
)

//packaging takes a source and builds it into a nice packaged image

type ImageBuildSpec struct {
	Name         string
	Version      string
	BuildScript  string
	SourceURI    string
	Type         config.VmmType
	Architecture common.ImageArchitecture
	Source       common.ImageSourceType
	Args         []string
	RootFs       common.ImageFsType
	Size         uint64
	workspace    string
	imgPath      string
	mountPoint   string
	img          *QemuImage
	part         *QemuImagePartition
}

func (ib *ImageBuildSpec) Prepare(workspace string) error {
	ib.workspace = filepath.Join(workspace, ib.Name, ib.Version)
	os.RemoveAll(ib.workspace)
	if err := vutils.Files.CreateDirIfNotExist(ib.workspace); err != nil {
		return err
	} else {
		//create the meta file with all the details needed...
		err := ioutil.WriteFile(filepath.Join(ib.workspace, "type"), []byte(ib.Type), 0644)
		if err != nil {
			return err
		}
		ib.imgPath = filepath.Join(ib.workspace, "root.qcow2")
		img, err := CreateNewQemuImage(ib.imgPath, ib.Size)
		if err != nil {
			return err
		}
		err = img.Connect()
		if err != nil {
			return err
		}
		err = img.MakeFilesystem("root", ib.RootFs)
		if err != nil {
			return err
		}
		mp, err := img.Mount()
		if err != nil {
			return err
		}
		ib.mountPoint = mp
		ib.img = img
		return nil
	}
}

func (ib *ImageBuildSpec) RunBuild() error {
	buildCmd := vutils.Exec.CreateAsyncCommand(ib.BuildScript, false, ib.workspace, ib.Version, ib.mountPoint)
	err := buildCmd.BindToStdoutAndStdErr().StartAndWait()
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}

func (ib *ImageBuildSpec) cleanup() error {
	if err := ib.img.Unmount(); err != nil {
		return err
	} else {
		return ib.img.Disconnect()
	}
}

func (ib *ImageBuildSpec) Package(output string) error {
	//kernelPath := filepath.Join(ib.workspace, "kernel.elf")
	imagePath := filepath.Join(ib.workspace, "root.qcow2")

	_, err := NewImageFromQcow(ib.Name, ib.Version, ib.Type, ib.Size, ib.Source, ib.SourceURI, imagePath)
	println(err)
	return ib.cleanup()
}

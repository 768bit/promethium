package common

import (
	"io"
	"os"
)

type ImageArchitecture string

const (
	X86_64  ImageArchitecture = "x86_64"
	AARCH64 ImageArchitecture = "aarch64"
)

var ImageArchitectures = []ImageArchitecture{
	X86_64,
	AARCH64,
}

type ImageFsType string

const (
	Zfs   ImageFsType = "zfs"
	Ext4  ImageFsType = "ext4"
	ExFat ImageFsType = "exfat"
	Ntfs  ImageFsType = "ntfs"
	Fat32 ImageFsType = "fat32"
)

type ImageSourceType string

const (
	PromethiumImage ImageSourceType = "promethium"
	DockerImage     ImageSourceType = "docker"
	TarImage        ImageSourceType = "tar"
	RawImage        ImageSourceType = "raw"
	Qcow2Image      ImageSourceType = "qcow2"
	CapstanImage    ImageSourceType = "capstan"
)

type ImageBackendType uint8

const (
	QcowImageBackend    ImageBackendType = 0x00
	BlockImageBackend   ImageBackendType = 0x01
	UnknownImageBackend ImageBackendType = 0xff
)

//func NewImage() (*Image, error) {
////
////}

type ImageContainsBits uint8

const (
	ImageHasNothing ImageContainsBits = 1 << iota
	ImageHasDisk
	ImageHasKernel
	ImageHasCloudInit
	ImageHasAdditionalDisk
)

type Image interface {
	GetID() string
	GetType() string
	GetBootParams() string
	ExtractKernelToPath(id string, outPath string) (string, string, error)
	ExtractRootDiskToPath(id string, outPath string, size uint64) (string, string, error)
	AddDisk(diskPath string) error
	AddKernel(kernelPath string) error
	AddCloudInitUserData(cloudInitPath string) error
	HasRootDisk() bool
	HasAdditionalDisks() bool
	HasKernel() bool
	HasCloudInit() bool
	CreatePackage(workspacePath string, outputRoot string) (string, error)
	GetRootDiskReader() (*os.File, io.Reader, error)
	GetKernelReader() (*os.File, io.Reader, error)
	GetCloudInitReader() (*os.File, io.Reader, error)
	GetAdditionalDiskReader(index int) (*os.File, io.Reader, error)
}

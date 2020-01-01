package common

type KernelArchitecture string

const (
	KernelAmd64 KernelArchitecture = "amd64"
)

type VmmKernel struct {
	id            string
	path          string
	storageDriver StorageDriver
}

func NewKernel(id string, path string, driver StorageDriver) *VmmKernel {
	return &VmmKernel{
		id:            id,
		path:          path,
		storageDriver: driver,
	}
}

func (vmk *VmmKernel) GetURI() string {
	return vmk.storageDriver.GetURI() + "/kernels/" + vmk.id + ".elf"
}

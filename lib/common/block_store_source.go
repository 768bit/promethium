package common

type BlockStoreSourceType string

const (
	FileStore   BlockStoreSourceType = "file"
	BlockDevice BlockStoreSourceType = "block"
)

type BlockStoreType string

const (
	KernelImage BlockStoreType = "kernel"
	DiskImage   BlockStoreType = "image"
)

type BlockStoreSource struct {
	SourceType  BlockStoreSourceType
	Type        BlockStoreType
	path        string //the path to the back end type..
	storage     StorageDriver
	nodeUUID    string
	clusterUUID string
	isImage     bool
}

func (bss *BlockStoreSource) GetPath() string {
	//return the poath to this store on this device...
	return ""
}

func (bss *BlockStoreSource) GetURI() string {
	//return the poath to this store on this device...
	return ""
}

func (bss *BlockStoreSource) GetClusterURI() string {
	//return the poath to this store on this device...
	return ""
}

func (bss *BlockStoreSource) IsShared(clusterUUID string) (bool, error) {
	//check if this target is shared for the cluster specified
	return false, nil
}

func (bss *BlockStoreSource) IsKernel() bool {
	//check if this target is shared for the cluster specified
	return false
}

func (bss *BlockStoreSource) IsImage() bool {
	//check if this target is shared for the cluster specified
	return false
}

func (bss *BlockStoreSource) IsDisk() bool {
	//check if this target is shared for the cluster specified
	return false
}

func (bss *BlockStoreSource) IsBlockDevice() (bool, error) {
	//check if this target is a block device (natively) or not
	return false, nil
}

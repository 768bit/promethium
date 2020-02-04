package storage

import (
	"errors"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/768bit/promethium/lib/common"
	"github.com/768bit/promethium/lib/images"
	"github.com/768bit/vutils"
)

//this represents a local data storage driver - it is installed as the default storage driver when the node is first deployed
//and uses the location /opt/promethium/storage/default-local

type LocalFileStorage struct {
	id            string
	sm            *StorageManager
	rootFolder    string
	imagesFolder  string
	imagesEnabled bool
	disksFolder   string
	disksEnabled  bool
	kernelsFolder string
	imagesCache   map[string]string
}

func getRootFolderFromConfig(config map[string]interface{}) (string, error) {
	if config == nil {
		return "", errors.New("Unable to get root folder from config")
	} else if v, ok := config["rootFolder"]; !ok || v == nil {
		return "", errors.New("Unable to get root folder from config")
	} else if vs, ok := config["rootFolder"].(string); !ok || vs == "" {
		return "", errors.New("Unable to get root folder from config")
	} else {
		return vs, nil
	}
}

func LoadLocalFileStorage(sm *StorageManager, id string, config map[string]interface{}) (*LocalFileStorage, error) {
	rootFolder, err := getRootFolderFromConfig(config)
	if err != nil {
		return nil, err
	}
	if !vutils.Files.CheckPathExists(rootFolder) {
		return nil, errors.New("The specified local file storage folder: " + rootFolder + " doesn't exist")
	}
	imagesFolder := filepath.Join(rootFolder, "images")
	disksFolder := filepath.Join(rootFolder, "disks")
	kernelsFolder := filepath.Join(rootFolder, "kernels")
	//create the instance...
	lfs := &LocalFileStorage{
		id:            id,
		sm:            sm,
		rootFolder:    rootFolder,
		disksFolder:   disksFolder,
		imagesFolder:  imagesFolder,
		imagesCache:   map[string]string{},
		kernelsFolder: kernelsFolder,
	}
	if !vutils.Files.CheckPathExists(disksFolder) {
		lfs.disksEnabled = false
	} else {
		lfs.disksEnabled = true
	}
	if !vutils.Files.CheckPathExists(imagesFolder) {
		lfs.imagesEnabled = false
	} else {
		lfs.imagesEnabled = true
	}
	return lfs, err
}

func InitLocalFileStorage(id string, config map[string]interface{}) error {
	rootFolder, err := getRootFolderFromConfig(config)
	if err != nil {
		return err
	}
	vutils.Files.CreateDirIfNotExist(rootFolder)
	imagesPath := filepath.Join(rootFolder, "images")
	vutils.Files.CreateDirIfNotExist(imagesPath)
	disksFolder := filepath.Join(rootFolder, "disks")
	vutils.Files.CreateDirIfNotExist(disksFolder)
	kernelsFolder := filepath.Join(rootFolder, "kernels")
	vutils.Files.CreateDirIfNotExist(kernelsFolder)
	return nil
}

func (lfs *LocalFileStorage) GetImages() ([]common.Image, error) {
	//get images needs to get a list of files from the directory
	lfs.imagesCache = map[string]string{}
	if !lfs.imagesEnabled {
		return nil, errors.New("Unable to get images from this storage medium as it doesnt support images")
	}
	files := vutils.Files.GetFilesInDirWithExtension(lfs.imagesFolder, ".prk")
	imagesList := []common.Image{}
	for _, file := range files {
		fullPath := filepath.Join(lfs.imagesFolder, file)
		img, err := images.LoadImageFromPrk(fullPath, lfs.sm.imagesCache)
		if err != nil {
			return nil, err
		}
		lfs.imagesCache[img.ID] = file[:len(file)-4]
		imagesList = append(imagesList, img)
	}
	return imagesList, nil
}

func (lfs *LocalFileStorage) GetURI() string {
	return "local-file://" + lfs.id
}

func (lfs *LocalFileStorage) GetImage(name string) (common.Image, error) {
	imageFilePath := filepath.Join(lfs.imagesFolder, name+".prk")
	if !lfs.imagesEnabled {
		return nil, errors.New("Unable to get images from this storage medium as it doesnt support images")
	} else if !vutils.Files.CheckPathExists(imageFilePath) {
		return nil, errors.New("Unable to find image with name " + name)
	} else {
		img, err := images.LoadImageFromPrk(imageFilePath, lfs.sm.imagesCache)
		if err != nil {
			return nil, err
		}
		return img, nil
	}
}
func (lfs *LocalFileStorage) ImportImageFromRdr(stream io.ReadCloser) error {
	//store reader in images area
	tuuid, _ := vutils.UUID.MakeUUIDString()
	fpath := filepath.Join(lfs.imagesFolder, tuuid+".prk.tmp")
	defer stream.Close()
	of, err := os.Create(fpath)
	if err != nil {
		return err
	}
	_, err = io.Copy(of, stream)
	if err != nil {
		of.Close()
		os.RemoveAll(fpath)
		return err
	}
	//on success we move it now so it can be used..
	//lets peek into the image - perhaps it exists already...
	img, err := images.GetMetaFromPrk(fpath)
	if err != nil {
		of.Close()
		os.RemoveAll(fpath)
		return err
	} else {
		fimg, err := lfs.sm.GetImageByID(img.GetID())
		if err == nil && fimg != nil {
			of.Close()
			return errors.New("Image with that ID already exists")
		}
	}
	of.Close()
	destPath := filepath.Join(lfs.imagesFolder, img.Name+"-"+img.Version+".prk")
	if vutils.Files.CheckPathExists(destPath) {
		os.RemoveAll(fpath)
		return errors.New("Image with that Name and Version already exists")
	}
	err = os.Rename(fpath, destPath)
	if err != nil {
		//of.Close()
		os.RemoveAll(fpath)
		return err
	}
	_, err = images.LoadImageFromPrk(destPath, lfs.sm.imagesCache)
	if err != nil {
		//of.Close()
		//os.RemoveAll(destPath)
		return err
	}
	return nil

}

func (lfs *LocalFileStorage) GetImageById(id string) (common.Image, error) {

	if !lfs.imagesEnabled {
		return nil, errors.New("Unable to get images from this storage medium as it doesnt support images")
	} else if v, ok := lfs.imagesCache[id]; !ok || v == "" {
		_, err := lfs.GetImages()
		if err != nil {
			return nil, err
		}
		if v, ok := lfs.imagesCache[id]; !ok || v == "" {
			return nil, errors.New("Unable to find image with id " + id)
		} else {
			return lfs.GetImage(lfs.imagesCache[id])
		}
	} else {
		imageFilePath := filepath.Join(lfs.imagesFolder, lfs.imagesCache[id]+".prk")
		img, err := images.LoadImageFromPrk(imageFilePath, lfs.sm.imagesCache)
		if err != nil {
			return nil, err
		}
		return img, nil
	}
}

func (lfs *LocalFileStorage) CreateDiskFromImage(id string, img common.Image, size uint64) (*common.VmmStorageDisk, *common.VmmKernel, error) {
	//first lets setup the image in a templ location and resize it...
	tdir, err := ioutil.TempDir("", "prmvm")
	if err != nil {
		return nil, nil, err
	} else {
		defer os.RemoveAll(tdir)
		imgPath, kernPath, err := img.ExtractRootDiskToPath(id, tdir, size)
		if err != nil {
			return nil, nil, err
		}

		//ok we have the built images - we need to put them where they need to be... for local this is a simple copy
		newImagePath := filepath.Join(lfs.disksFolder, id, "root.qcow2")
		newKernelPath := filepath.Join(lfs.kernelsFolder, id+".elf")
		if err := vutils.Files.Copy(imgPath, newImagePath); err != nil {
			return nil, nil, err
		} else if err := vutils.Files.Copy(kernPath, newKernelPath); err != nil {
			return nil, nil, err
		}

		//everything copied ok.. lets instantiate the Disk...

		dsk, err := common.NewStorageDisk(id, "root", newImagePath, lfs)
		if err != nil {
			return nil, nil, err
		}

		kern := common.NewKernel(id, newKernelPath, lfs)
		return dsk, kern, nil
	}

}

func (lfs *LocalFileStorage) LookupPath(path string) (string, bool, error) {
	//path lookups are all about finding different classes of stuff
	if strings.HasPrefix(path, "/") {
		path = path[1:]
	}
	splitPath := strings.Split(path, "/")
	storageClass := splitPath[0]
	switch storageClass {
	case "kernels":
		if !lfs.disksEnabled {
			return "", false, errors.New("This storage target is not enabled for disk/kernel storage")
		}
		if len(splitPath) != 2 {
			return "", false, errors.New("The supplied path is invalid")
		}
		return filepath.Join(lfs.kernelsFolder, splitPath[1]), false, nil
	case "disks":
		if !lfs.disksEnabled {
			return "", false, errors.New("This storage target is not enabled for disk/kernel storage")
		}
		if len(splitPath) != 3 {
			return "", false, errors.New("The supplied path is invalid")
		}
		return filepath.Join(lfs.disksFolder, splitPath[1], splitPath[2]), false, nil
	}
	return "", false, nil
}

func (lfs *LocalFileStorage) WriteKernel(id string, source io.Reader) (string, error) {
	//pump the source to the output file with ID...
	newKernelPath := filepath.Join(lfs.kernelsFolder, id+".elf")
	f, err := os.Create(newKernelPath)
	if err != nil {
		return newKernelPath, err
	}
	defer f.Close()
	_, err = io.Copy(f, source)
	return newKernelPath, err
}
func (lfs *LocalFileStorage) WriteRootDisk(id string, source io.Reader, newSize int64, sourceIsRaw bool, growPart bool) (string, error) {
	//output the qcow2 image somewhere (it came from an image), unless its raw, in which case we just pump it to dest...
	newDiskPath := filepath.Join(lfs.disksFolder, id, "root.img")
	vutils.Files.CreateDirIfNotExist(filepath.Join(lfs.disksFolder, id))
	if sourceIsRaw {
		//create the dest file by pumping to dest

		f, err := os.Create(newDiskPath)
		if err != nil {
			return newDiskPath, err
		}
		defer f.Close()
		_, err = io.Copy(f, source)
		if err != nil {
			return newDiskPath, err
		}
		//
		err = resizeRawImage(newDiskPath, newSize, growPart)
		return newDiskPath, err
	} else {
		err := writeAndConvertQcow2(newDiskPath, source, newSize, false, growPart)
		return newDiskPath, err
	}
	return "", nil
}
func (lfs *LocalFileStorage) WriteAdditionalDisk(id string, index int, source io.Reader, newSize int64, sourceIsRaw bool, growPart bool) (string, error) {
	return "", nil
}
func (lfs *LocalFileStorage) WriteCloudInit(id string, source io.Reader) (string, error) {
	return "", nil
}

func resizeRawImage(path string, newSize int64, growPart bool) error {
	qcimg, err := images.LoadQemuImage(path)
	if err != nil {
		return err
	}
	nsu := uint64(newSize)
	if nsu < qcimg.VirtualSize() {
		//err
	} else if nsu > qcimg.VirtualSize() {
		//resize/expand
		err := qcimg.Resize(nsu)
		if err != nil {
			return err
		}
		if growPart {
			err = qcimg.GrowFullPart()
			if err != nil {
				println(err.Error())
			}
		}
	}
	return nil
}

func writeAndConvertQcow2(outpath string, source io.Reader, newSize int64, destIsDevice bool, growPart bool) error {
	//create a temporary file path...
	tdir, err := ioutil.TempDir("", "prmextract")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tdir)
	//lets pup to a temp file..
	tempDiskPath := filepath.Join(tdir, "disk.qcow2")
	f, err := os.Create(tempDiskPath)
	if err != nil {
		return err
	}
	_, err = io.Copy(f, source)
	if err != nil {
		err = f.Close()
		return err
	}
	err = f.Close()
	if err != nil {
		return err
	}
	//now we have outputted we need to open the qcow image so we can interrogate it and resize if necessary...
	qcimg, err := images.LoadQemuImage(tempDiskPath)
	if err != nil {
		return err
	}

	nsu := uint64(newSize)
	if nsu < qcimg.VirtualSize() {
		return errors.New("Size is smaller")
	} else if nsu > qcimg.VirtualSize() {
		//resize/expand
		err := qcimg.Resize(nsu)
		if err != nil {
			return err
		}
		if growPart {
			err = qcimg.Connect()
			if err != nil {
				return err
			}
			err = qcimg.GrowFullPart()
			if err != nil {
				println("Error growing part:" + err.Error())
			}
		}
	}

	//now we can use qemu-convert to target what we need to...
	if destIsDevice {
		err := qcimg.ConvertImgRawDevice(outpath)
		if err != nil {
			return err
		}
	} else {
		err := qcimg.ConvertImgRaw(outpath)
		if err != nil {
			return err
		}
	}

	return nil
}

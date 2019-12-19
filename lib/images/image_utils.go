package images

import (
  "fmt"
  "github.com/768bit/promethium/lib/cloudconfig"
  "github.com/768bit/vutils"
  "io/ioutil"
  "os"
  "path/filepath"
)

//capstan was managin images before.. we will continue to use capstan for managing these images but we create isntances of these images as required...

//these utils allow for the management of these image instances


type PImage struct {

}


func (pi *PImage) loadFromDisk() {



}

func MakeCloudInitImageBuilt(hostname string, networkConfig *cloudconfig.MetaDataNetworkConfig, userData *cloudconfig.UserData) ([]byte, error) {
  tmp, _ := ioutil.TempDir("", "promethium")
  // Once this function is finished, remove temporary file.
  defer os.RemoveAll(tmp)

  fmt.Printf("Creating Image In: %s\n", tmp)

  md := cloudconfig.NewMetaDataWithNetworking(hostname, networkConfig)
  err := md.WriteMetaData(filepath.Join(tmp, "meta_data"))
  if err != nil {
    fmt.Println(err)
    return nil, err
  }

  err = userData.WriteUserData(filepath.Join(tmp, "user_data"))
  if err != nil {
    fmt.Println(err)
    return nil, err
  }
  ddCmd := vutils.Exec.CreateAsyncCommand("cloud-localds", false, filepath.Join(tmp, "cidata.img"), filepath.Join(tmp, "user_data"), filepath.Join(tmp, "meta_data"))
  err = ddCmd.BindToStdoutAndStdErr().StartAndWait()
  if err != nil {
    fmt.Println(err)
    return nil, err
  }

  img, err := ioutil.ReadFile(filepath.Join(tmp, "cidata.img"))
  if err != nil {
    fmt.Println(err)
    return nil, err
  }

  return img, nil

}

func MakeCloudInitImage(hostname string, networkConfig *cloudconfig.MetaDataNetworkConfig, userData *cloudconfig.UserData) ([]byte, error) {
  tmp, _ := ioutil.TempDir("", "promethium")
  // Once this function is finished, remove temporary file.
  defer os.RemoveAll(tmp)

  fmt.Printf("Creating Image In: %s\n", tmp)

  mntPoint := filepath.Join(tmp, "mnt")
  imagePath := filepath.Join(tmp, "img")

  err := vutils.Files.CreateDirIfNotExist(mntPoint)
  if err != nil {
    fmt.Println(err)
    return nil, err
  }
  //dd if=/dev/zero of=/rootfs.ext4 bs=1M count=50
  ddCmd := vutils.Exec.CreateAsyncCommand("dd", false, "if=/dev/zero", fmt.Sprintf("of=%s", imagePath), "bs=1M", "count=16")
  err = ddCmd.BindToStdoutAndStdErr().StartAndWait()
  if err != nil {
    fmt.Println(err)
    return nil, err
  }

  fmt.Printf("Image Created\n")
  //make the exfat file system...

  mkfsCmd := vutils.Exec.CreateAsyncCommand("mkfs.exfat", false, "-n", "config-2", imagePath)
  err = mkfsCmd.BindToStdoutAndStdErr().StartAndWait()
  if err != nil {
    fmt.Println(err)
    return nil, err
  }

  fmt.Printf("File System Initialised\n")

  //mount the image...

  mountCmd := vutils.Exec.CreateAsyncCommand("mount", false, "-o", "loop", "-t", "exfat", imagePath, mntPoint).Sudo()
  err = mountCmd.BindToStdoutAndStdErr().StartAndWait()
  if err != nil {
    fmt.Println(err)
    return nil, err
  }


  fmt.Printf("Image Mounted: %s\n", mntPoint)
  //create the structure

  ciPath := filepath.Join(mntPoint, "openstack", "latest")

  err = vutils.Files.CreateDirIfNotExist(ciPath)
  if err != nil {
    fmt.Println(err)
    return nil, err
  }

  md := cloudconfig.NewMetaDataWithNetworking(hostname, networkConfig)
  err = md.WriteMetaDataJSON(filepath.Join(ciPath, "meta_data.json"))
  if err != nil {
    fmt.Println(err)
    return nil, err
  }

  err = userData.WriteUserData(filepath.Join(ciPath, "user_data"))
  if err != nil {
    fmt.Println(err)
    return nil, err
  }

  fmt.Printf("Assets Copied to: %s\n", ciPath)

  //unmount

  umountCmd := vutils.Exec.CreateAsyncCommand("umount", false, mntPoint).Sudo()
  err = umountCmd.BindToStdoutAndStdErr().StartAndWait()
  if err != nil {
   fmt.Println(err)
   return nil, err
  }

  fmt.Printf("Image Unmounted: %s\n", mntPoint)

  //read in the image so that the build can be deleted..
  img, err := ioutil.ReadFile(imagePath)
  if err != nil {
    fmt.Println(err)
    return nil, err
  }

  return img, nil

}

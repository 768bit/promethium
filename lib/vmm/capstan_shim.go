package vmm

import (
  "errors"
  "fmt"
  "github.com/768bit/vutils"
  "github.com/cloudius-systems/capstan/cmd"
  "github.com/cloudius-systems/capstan/core"
  "github.com/cloudius-systems/capstan/util"
  "os"
  "path/filepath"
  "runtime"
)

func BuildBaseCapstanImage(name string, cmdPath string, entryPoint string, imageSize int64) (*core.Image, *util.Repo, string, error) {
  //template := makeCapstanTemplate("cloudius/osv", cmdPath, files)
  repo := util.NewRepo(util.DefaultRepositoryUrl)
  image := &core.Image{
    Name:       name,
    Hypervisor: "qemu",
  }

  bootOpts := cmd.BootOptions{
    Cmd:        cmdPath,
    Boot:       []string{},
    EnvList:    []string{},
    PackageDir: entryPoint,
  }

  err := cmd.ComposePackage(repo, imageSize, true, true, true, entryPoint, name, &bootOpts, "zfs")
  //err := cmd.Compose(repo, "", 64000000, entryPoint, name, cmdPath, true)
  //err := cmd.Build(repo, image, template, true, "512M")
  imgPath := repo.ImagePath(image.Hypervisor, image.Name)
  return image, repo, imgPath, err
}

func GetCapstanImagePath(name string) (*core.Image, *util.Repo, string) {
  repo := util.NewRepo(util.DefaultRepositoryUrl)
  image := &core.Image{
    Name:       name,
    Hypervisor: "qemu",
  }
  imgPath := repo.ImagePath(image.Hypervisor, image.Name)
  return image, repo, imgPath
}

func capstanPkgCompose(repo *util.Repo, imageSize int64, updatePackage, verbose, pullMissing bool,
    packageDir, appName string, bootOpts *cmd.BootOptions, filesystem string) error {

    // Package content should be collected in a subdirectory called mpm-pkg.
    //targetPath := filepath.Join(packageDir, "mpm-pkg")
    //vutils.Files.CreateDirIfNotExist(targetPath)
    // Remove collected directory afterwards.
    //defer os.RemoveAll(targetPath)

    // Construct final bootcmd for the image.
    commandLine, err := bootOpts.GetCmd()
    if err != nil {
      return err
    }

    // First, collect the contents of the package.
    //if err := cmd.CollectPackage(repo, packageDir, pullMissing, false, verbose); err != nil {
    //  return err
    //}

    // If all is well, we have to start preparing the files for upload.
    paths, err := cmd.CollectDirectoryContents(packageDir)
    if err != nil {
      return err
    }

    // Get the path of imported image.
    imagePath := repo.ImagePath("qemu", appName)
    // Check whether the image already exists.
    imageExists := false
    if _, err = os.Stat(imagePath); !os.IsNotExist(err) {
      imageExists = true
    }

    if filesystem == "zfs" {
      imageCachePath := repo.ImageCachePath("qemu", appName)
      var imageCache core.HashCache

      // If the user requested new image or requested to update a non-existent image,
      // initialize it first.
      if !updatePackage || !imageExists {
      // Initialize an empty image based on the provided loader image. imageSize is used to
      // determine the size of the user partition. Use default loader image.
      if err := repo.InitializeZfsImage("", appName, imageSize); err != nil {
        return fmt.Errorf("Failed to initialize empty image named %s.\nError was: %s", appName, err)
      }
    } else {
      // We are updating an existing image so try to parse the cache
      // config file. Note that we are not interested in any errors as
      // no-cache or invalid cache means that all files will be uploaded.
      imageCache, _ = core.ParseHashCache(imageCachePath)
    }

      // Upload the specified path onto virtual image.
      imageCache, err = cmd.UploadPackageContents(repo, imagePath, paths, imageCache, verbose)
      if err != nil {
      return err
    }

    // Save the new image cache
    imageCache.WriteToFile(imageCachePath)
  }

  if err = util.SetCmdLine(imagePath, commandLine); err != nil {
    return err
  }
  fmt.Printf("Command line set to: '%s'\n", commandLine)

  return nil
}

func makeCapstanTemplate(base string, cmd string, files map[string]string) *core.Template {
  return &core.Template{
    Base:    base,
    Cmdline: cmd,
    Files:   files,
    Rootfs:  "ROOTFS",
  }
}


func getCapstanDevPath() (string, error) {
  _, callerFile, _, _ := runtime.Caller(0)
  executablePath := filepath.Dir(callerFile)
  executablePath = filepath.Join(executablePath, "..", "workspace", "osv", "build", "release")
  if !vutils.Files.CheckPathExists(executablePath) {
    return executablePath, errors.New("OSV Dev Path missing")
  }
  return executablePath, nil
}

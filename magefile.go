// +build mage

package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/768bit/promethium/lib/images"
	"github.com/768bit/promethium/lib/vmm"
	"github.com/768bit/vutils"
	"github.com/magefile/mage/mg"
	"gitlab.768bit.com/pub/vpkg"
)

var CWD, _ = os.Getwd()
var WORKSPACE = filepath.Join(CWD, "workspace")
var SCRIPTS = filepath.Join(CWD, "scripts")
var BASE_IMAGES = filepath.Join(CWD, "base-images")
var ASSET_OUT_DIR = filepath.Join(CWD, "assets", "images")
var VDATA *vpkg.VersionData
var LD_FLAGS_FMT_STR = "-w -s -X main.Version=%s -X main.Build=%s -X \"main.BuildDate=%s\" -X main.BuildUUID=%s -X main.GitCommit=%s"

func InitialiseVersionData() error {

	vd, err := vpkg.LoadVersionData(CWD)
	if err != nil {
		fmt.Println(err)
		vd = vpkg.NewVersionData()
		err = vd.Save(CWD)
		if err != nil {
			return err
		}
		fmt.Println("Created a new version.json file. Please re-run command to use it.")
		return nil
	}
	VDATA = vd
	return nil

}

func CopyAssets() error {
	fc, jailer, err := vmm.GetBuiltFirecracker()
	if err != nil {
		return err
	}
	//copy the binaries to logical folder..
	assetsPath := filepath.Join(CWD, "assets")
	fcAssetsPath := filepath.Join(assetsPath, "firecracker")
	osvAssetsPath := filepath.Join(assetsPath, "osv")
	err = vutils.Files.Copy(fc, filepath.Join(fcAssetsPath, "firecracker"))
	if err != nil {
		return err
	}
	err = vutils.Files.Copy(jailer, filepath.Join(fcAssetsPath, "jailer"))
	if err != nil {
		return err
	}

	osvWorkspaceRelease := filepath.Join(CWD, "workspace", "osv", "build", "release")
	kernelImg := filepath.Join(osvWorkspaceRelease, "loader-stripped.elf")

	return vutils.Files.Copy(kernelImg, filepath.Join(osvAssetsPath, "kernel.elf"))

}

func Build() error {
	mg.Deps(InitialiseVersionData)
	if err := packAssets(); err != nil {
		return err
	}

	defer cleanPackedAssets()

	//build the binary...

	buildDir := filepath.Join(CWD, "build")
	vutils.Files.CreateDirIfNotExist(buildDir)
	promethiumBuildOut := filepath.Join(buildDir, "promethium")
	ldflags := fmt.Sprintf(LD_FLAGS_FMT_STR, VDATA.FullVersionString(), VDATA.ShortID, VDATA.DateString, VDATA.UUID, VDATA.GitCommit)
	fmt.Printf("Building with LDFLAGS: %s\n", ldflags)
	cmd := vutils.Exec.CreateAsyncCommand("go", false, "build", "-v", "-ldflags", ldflags, "-o", promethiumBuildOut)
	err := cmd.BindToStdoutAndStdErr().SetWorkingDir(filepath.Join(CWD, "cmd")).CopyEnv().StartAndWait()
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil

}

func packAssets() error {
	fmt.Println("Packing Assets...")

	pkrBuildCmd := vutils.Exec.CreateAsyncCommand("packr", false, "-z")
	err := pkrBuildCmd.BindToStdoutAndStdErr().SetWorkingDir(filepath.Join(CWD, "lib", "assets")).StartAndWait()
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}

func cleanPackedAssets() error {
	fmt.Println("Cleaning Packed Assets...")
	packCleanCmd := vutils.Exec.CreateAsyncCommand("packr", false, "clean")
	err := packCleanCmd.BindToStdoutAndStdErr().SetWorkingDir(filepath.Join(CWD, "lib", "assets")).StartAndWait()
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}

var ImagesMap = []images.ImageBuildSpec{
	{
		Name:         "ubuntu",
		Version:      "bionic",
		RootFs:       "ext4",
		SourceURI:    "ubuntu:bionic",
		Type:         images.StandardImage,
		BuildScript:  filepath.Join(SCRIPTS, "build-fc-ubuntu.sh"),
		Size:         1 * 1024 * 1024 * 1024,
		Source:       images.DockerImage,
		Architecture: images.X86_64,
	},
	{
		Name:         "ubuntu",
		Version:      "bionic-cloud",
		RootFs:       "ext4",
		SourceURI:    "ubuntu:bionic",
		Type:         images.StandardImage,
		BuildScript:  filepath.Join(SCRIPTS, "build-fc-ubuntu-cloud.sh"),
		Size:         1 * 1024 * 1024 * 1024,
		Source:       images.DockerImage,
		Architecture: images.X86_64,
	},
	{
		Name:         "alpine",
		Version:      "3.10",
		RootFs:       "ext4",
		SourceURI:    "alpine:3.10",
		Type:         images.StandardImage,
		BuildScript:  filepath.Join(SCRIPTS, "build-fc-alpine.sh"),
		Size:         1 * 1024 * 1024 * 1024,
		Source:       images.DockerImage,
		Architecture: images.X86_64,
	},
	{
		Name:         "alpine",
		Version:      "3.10-cloud",
		RootFs:       "ext4",
		SourceURI:    "alpine:3.10",
		Type:         images.StandardImage,
		BuildScript:  filepath.Join(SCRIPTS, "build-fc-alpine-cloud.sh"),
		Size:         1 * 1024 * 1024 * 1024,
		Source:       images.DockerImage,
		Architecture: images.X86_64,
	},
}

func BuildBaseImages() error {
	//build the base images based on the configuration map
	for _, imgSpec := range ImagesMap {
		err := imgSpec.Prepare(WORKSPACE)
		if err != nil {
			return err
		}
		err = imgSpec.RunBuild()
		if err != nil {
			return err
		}
		err = imgSpec.Package(ASSET_OUT_DIR)
		if err != nil {
			return err
		}
	}
	return nil
}

func BuildImages() error {
	//build the base images based on the configuration map

	//using the directory structure - build the images...

	dirs := map[string][]string{
		"ubuntu": {
			"bionic-cloud",
			"bionic",
		},
		"alpine": {
			"3.10-cloud",
			"3.10",
		},
	}

	for os, items := range dirs {
		for _, version := range items {
			imageRootPath := filepath.Join(BASE_IMAGES, os, version)
			println(imageRootPath)
			img, err := images.BuildPackageFrom(imageRootPath, ASSET_OUT_DIR)
			if err != nil {
				println(err)
				return err
			}
			println(img.ID + " :: " + img.Name + " " + img.Version)
		}
	}
	return nil
}

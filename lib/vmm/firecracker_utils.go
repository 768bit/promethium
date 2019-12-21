package vmm

import (
	"errors"
	"github.com/768bit/vutils"
	"github.com/kardianos/osext"
	"log"
	"path/filepath"
	"runtime"
)

func getFirecrackerBinary() (string, string, error) {

	//go get the firecracker binary - it is either in the bin directory, embedded or within the workspace..
	// if it is embedded it needs to be extracted for execution

	firecrackerPath, jailerPath, err := lookupExecFolder()
	if err != nil {
		//check the embedded asset
		firecrackerPath, jailerPath, err = lookupEmbeddedBinary()
		if err != nil {
			firecrackerPath, jailerPath, err = lookupDevPath()
		}
	}

	return firecrackerPath, jailerPath, err

}

func lookupEmbeddedBinary() (string, string, error) {
	return checkBinariesExist(filepath.Join(ROOT_PATH, "bin"))
}

func lookupExecFolder() (string, string, error) {
	folderPath, err := osext.ExecutableFolder()
	if err != nil {
		return "", "", err
	}
	log.Printf("Looking up Firecracker and Jailer Binaries in %s\n", folderPath)
	return checkBinariesExist(folderPath)
}

func GetBuiltFirecracker() (string, string, error) {
	return lookupDevPath()
}

func lookupDevPath() (string, string, error) {
	_, callerFile, _, _ := runtime.Caller(0)
	executablePath := filepath.Dir(callerFile)
	executablePath = filepath.Join(executablePath, "..", "workspace", "firecracker", "build", "release-musl")
	log.Printf("Looking up Firecracker and Jailer Binaries in Workspace %s\n", executablePath)
	return checkBinariesExist(executablePath)
}

func getRootPath() string {
	_, callerFile, _, _ := runtime.Caller(0)
	executablePath := filepath.Dir(callerFile)
	executablePath = filepath.Join(executablePath, "..")
	return executablePath
}

func checkBinariesExist(root string) (string, string, error) {

	firecrackerPath := filepath.Join(root, "firecracker")
	jailerPath := filepath.Join(root, "jailer")

	if !vutils.Files.CheckPathExists(firecrackerPath) {
		return firecrackerPath, jailerPath, errors.New("Unable to find firecracker binary")
	} else if !vutils.Files.CheckPathExists(jailerPath) {
		return firecrackerPath, jailerPath, errors.New("Unable to find jailer binary")
	}

	return firecrackerPath, jailerPath, nil

}

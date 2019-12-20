package config

import (
	"os"
	"path/filepath"

	"github.com/768bit/promethium/lib/networking"
	"github.com/768bit/promethium/lib/storage"
	"github.com/768bit/vutils"
)

type PromethiumDaemonConfig struct {
	NodeID    string                      `json:"nodeID"`
	Clusters  []*ClusterConfig            `json:"clusters"`
	Storage   []*storage.StorageConfig    `json:"storage"`
	Networks  []*networking.NetworkConfig `json:"networks"`
	AppRoot   string                      `json:"appRoot"`
	JailUser  string                      `json:"jailUser"`
	JailGroup string                      `json:"jailGroup"`
	API       *APIConfig                  `json:"api"`
}

type ClusterConfig struct {
	ID    string        `json:"id"`
	Nodes []interface{} `json:"nodes"`
}

type APIConfig struct {
	BindAddress string `json:"bindAddress"`
	Port        uint   `json:"port"`
}

const PROMETHIUM_CONFIG_DIR = "/etc/promethium"
const PROMETHIUM_DAEMON_CONFIG_LOCATION = PROMETHIUM_CONFIG_DIR + "/daemon.json"

var PROMETHIUM_DAEMON_CONFIG_LOAD_LIST = []string{PROMETHIUM_DAEMON_CONFIG_LOCATION}
var CWD, _ = os.Getwd()

func LoadPromethiumDaemonConfig() (*PromethiumDaemonConfig, error) {
	if !vutils.Files.CheckPathExists(PROMETHIUM_DAEMON_CONFIG_LOCATION) {
		if err := vutils.Files.CreateDirIfNotExist(PROMETHIUM_CONFIG_DIR); err != nil {
			return nil, err
		} else if err := generateNewPromethiumDaemonConfig(); err != nil {
			return nil, err
		}
	}
	oconfig := &PromethiumDaemonConfig{}
	err := vutils.Config.GetConfigFromDefaultList("promethium.daemon", CWD, PROMETHIUM_DAEMON_CONFIG_LOAD_LIST, oconfig)
	if err != nil {
		return nil, err
	}
	return oconfig, nil
}

func generateNewPromethiumDaemonConfig() error {
	defaultPromDir := "/opt/promethium"
	storageDir := "/opt/promethium/storage/default-local"
	storageName := "default-local"
	newUUID, _ := vutils.UUID.MakeUUIDString()
	oconfig := &PromethiumDaemonConfig{
		NodeID:   newUUID,
		Clusters: []*ClusterConfig{},
		Storage: []*storage.StorageConfig{
			{
				ID:     storageName,
				Driver: "local-file",
				Config: map[string]interface{}{
					"rootFolder": storageDir,
				},
			},
		},
		Networks:  []*networking.NetworkConfig{},
		AppRoot:   defaultPromDir,
		JailUser:  "promethium_jail",
		JailGroup: "promethium_jail",
		API: &APIConfig{
			BindAddress: "0.0.0.0",
			Port:        8921,
		},
	}
	err, _ := vutils.Config.TrySaveConfig(CWD, PROMETHIUM_DAEMON_CONFIG_LOAD_LIST, oconfig)
	if err != nil {
		//return err
		return err
	}
	//now create the default folder structure...
	vutils.Files.CreateDirIfNotExist(defaultPromDir)
	//we also need to create others
	firecrackerDir := filepath.Join(defaultPromDir, "firecracker")
	storageDirPath := filepath.Join(defaultPromDir, "storage")
	instancesDir := filepath.Join(defaultPromDir, "instances")
	vutils.Files.CreateDirIfNotExist(firecrackerDir)
	vutils.Files.CreateDirIfNotExist(storageDirPath)
	vutils.Files.CreateDirIfNotExist(instancesDir)

	_, err = storage.InitLocalFileStorage(storageName, oconfig.Storage[0].Config)

	return err
}

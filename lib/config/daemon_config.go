package config

import (
	"os"

	"github.com/768bit/promethium/lib/networking"
	"github.com/768bit/vutils"
)

type PromethiumDaemonConfig struct {
	NodeID    string                      `json:"nodeID"`
	Clusters  []*ClusterConfig            `json:"clusters"`
	Storage   []*StorageConfig            `json:"storage"`
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

type StorageConfig struct {
	ID     string                 `json:"id"`
	Driver string                 `json:"driver"`
	Config map[string]interface{} `json:"config"`
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
	newUUID, _ := vutils.UUID.MakeUUIDString()
	oconfig := &PromethiumDaemonConfig{
		NodeID:   newUUID,
		Clusters: []*ClusterConfig{},
		Storage: []*StorageConfig{
			{
				ID:     "default",
				Driver: "local-file",
				Config: map[string]interface{}{
					"rootFolder": "/opt/promethium/storage/default-local",
				},
			},
		},
		Networks:  []*networking.NetworkConfig{},
		AppRoot:   "/opt/promethium",
		JailUser:  "promethium_jail",
		JailGroup: "promethium_jail",
		API: &APIConfig{
			BindAddress: "0.0.0.0",
			Port:        8921,
		},
	}
	err, _ := vutils.Config.TrySaveConfig(CWD, PROMETHIUM_DAEMON_CONFIG_LOAD_LIST, oconfig)
	return err
}

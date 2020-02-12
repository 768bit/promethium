package config

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"syscall"
	"time"

	"text/template"

	"github.com/768bit/promethium/lib/networking"
	"github.com/768bit/vutils"
	"github.com/fsnotify/fsnotify"
	"github.com/gofrs/flock"
)

var IS_NEW_CONFIG bool = false

var RootCommandList = []string{
	"brctl",
	"qemu-nbd",
	"partprobe",
	"udevadm|settle",
}

func init() {
	olist := []string{}
	for _, item := range RootCommandList {
		spl := strings.Split(item, "|")
		if p, exists := getCommandPath(spl[0]); exists && p != "" {
			cmdPath := strings.TrimSpace(p)
			if len(spl) > 1 {
				for i := 1; i < len(spl); i++ {
					olist = append(olist, cmdPath+" "+spl[i])
				}
			} else {
				olist = append(olist, cmdPath)
			}

		}
	}
	RootCommandList = olist
}

func getRootCommandListForPath(installBinPath string) []string {
	olist := append([]string{}, RootCommandList...)
	//now lets enable sudo on the binaries we need...
	olist = append(olist, filepath.Join(installBinPath, "jailer"))
	return olist
}

type PromethiumDaemonConfigUpdateCallbackArea string

const (
	NetworkingUpdate   PromethiumDaemonConfigUpdateCallbackArea = "networking"
	StorageUpdate      PromethiumDaemonConfigUpdateCallbackArea = "storage"
	ClusterNodesUpdate PromethiumDaemonConfigUpdateCallbackArea = "cluster-node"
)

type PromethiumDaemonConfigUpdateCallback func(area PromethiumDaemonConfigUpdateCallbackArea, scope string, add []string, update []string, remove []string)

type PromethiumDaemonConfig struct {
	NodeID           string                      `json:"nodeID"`
	Clusters         []*ClusterConfig            `json:"clusters"`
	Storage          []*StorageConfig            `json:"storage"`
	Networks         []*networking.NetworkConfig `json:"networks"`
	AppRoot          string                      `json:"appRoot"`
	User             string                      `json:"user"`
	Group            string                      `json:"group"`
	JailUser         string                      `json:"jailUser"`
	JailGroup        string                      `json:"jailGroup"`
	Http             *HttpAPIConfig              `json:"http"`
	Https            *HttpsAPIConfig             `json:"https"`
	Unix             *UnixAPIConfig              `json:"unix"`
	isNew            bool
	linuxBridgeAvail bool
	ovsBridgeAvail   bool
	kvmDevAccess     bool
	configPath       string
	watcher          *fsnotify.Watcher
	fd               *os.File
	lock             *flock.Flock
	locked           bool
	updateCallback   PromethiumDaemonConfigUpdateCallback
}

type SudoersTemplateData struct {
	Group           string
	RootCommandList string
}

var PROMETHIUM_SUDOERS_TEMPLATE = `%{{ .Group }} ALL=(ALL) !ALL
%{{ .Group }} ALL=(root) NOPASSWD: {{ .RootCommandList }}
`

type SystemdTemplateData struct {
	User       string
	Group      string
	BinaryPath string
	ConfigPath string
}

var PROMETHIUM_DAEMON_SYSTEMD_UNIT_TEMPLATE = `[Unit]
Description=Promethium Daemon Service

[Service]
Type=notify
ExecStart={{ .BinaryPath }} daemon --service.uid={{ .User }} --service.gid={{ .Group }}
Restart=always
RestartSec=30

[Install]
WantedBy=multi-user.target
`

func InstallServiceUnit(destDir string, name string, binPath string, confPath string, user string, group string, enable bool, start bool) error {

	t, err := template.New("sudoers").Parse(PROMETHIUM_DAEMON_SYSTEMD_UNIT_TEMPLATE)
	if err != nil {
		return err
	}
	fullDestPath := filepath.Join(destDir, name)
	of, err := os.Create(fullDestPath)
	if err != nil {
		return err
	}
	defer of.Close()
	err = t.Execute(of, &SystemdTemplateData{
		BinaryPath: binPath,
		ConfigPath: confPath,
		User:       user,
		Group:      group,
	})
	if err != nil {
		return err
	}
	//in any case run systemctl daemon-reload
	cmd := vutils.Exec.CreateAsyncCommand("systemctl", false, "daemon-reload")
	err = cmd.Sudo().BindToStdoutAndStdErr().StartAndWait()
	if err != nil {
		return err
	}
	if enable {
		cmd = vutils.Exec.CreateAsyncCommand("systemctl", false, "enable", name)
		err = cmd.Sudo().BindToStdoutAndStdErr().StartAndWait()
		if err != nil {
			return err
		}
	}
	if start {
		cmd = vutils.Exec.CreateAsyncCommand("systemctl", false, "start", name)
		err = cmd.Sudo().BindToStdoutAndStdErr().StartAndWait()
		if err != nil {
			return err
		}
	}
	return nil

}

func (pdc *PromethiumDaemonConfig) validate(doCreate bool) error {
	err := pdc.validateUsersAndGroups(doCreate)
	if err != nil {
		return err
	}
	uid, err := GetUserId(pdc.User)
	if err != nil {
		return err
	}
	gid, err := GetGroupId(pdc.Group)
	if err != nil {
		return err
	}
	firecrackerDir := filepath.Join(pdc.AppRoot, "firecracker")
	storageDirPath := filepath.Join(pdc.AppRoot, "storage")
	instancesDir := filepath.Join(pdc.AppRoot, "instances")
	binPath := filepath.Join(pdc.AppRoot, "bin")
	if !vutils.Files.CheckPathExists(pdc.AppRoot) {
		if doCreate {
			vutils.Files.CreateDirIfNotExist(pdc.AppRoot)
		} else {
			return os.ErrNotExist
		}
	}

	if !vutils.Files.CheckPathExists(binPath) {
		if doCreate {
			vutils.Files.CreateDirIfNotExist(binPath)
			DoChmod(binPath, 0700, true)
			err = DoChown(binPath, uid, gid, true)
			if err != nil {
				println(err.Error())
			}
		} else {
			return os.ErrNotExist
		}
	}

	if !vutils.Files.CheckPathExists(firecrackerDir) {
		if doCreate {
			vutils.Files.CreateDirIfNotExist(firecrackerDir)
			DoChmod(firecrackerDir, 0700, false)
			err = DoChown(firecrackerDir, uid, gid, false)
			if err != nil {
				println(err.Error())
			}
		} else {
			return os.ErrNotExist
		}
	}

	if !vutils.Files.CheckPathExists(storageDirPath) {
		if doCreate {
			vutils.Files.CreateDirIfNotExist(storageDirPath)
			DoChmod(storageDirPath, 0700, true)
			err = DoChown(storageDirPath, uid, gid, true)
			if err != nil {
				println(err.Error())
			}
		} else {
			return os.ErrNotExist
		}
	}

	if !vutils.Files.CheckPathExists(instancesDir) {
		if doCreate {
			vutils.Files.CreateDirIfNotExist(instancesDir)
			DoChmod(instancesDir, 0700, true)
			err = DoChown(instancesDir, uid, gid, true)
			if err != nil {
				println(err.Error())
			}
		} else {
			return os.ErrNotExist
		}
	}

	if pdc.Storage != nil && len(pdc.Storage) > 0 {
		for _, store := range pdc.Storage {
			if store != nil && store.Driver == "local-file" && store.Config != nil {
				v, ok := store.Config["rootFolder"]
				if !ok {
					fmt.Printf("Error loading local-file store as the root folder is not specified\n")
					continue
				}
				vstr, ok := v.(string)
				if !ok {
					fmt.Printf("Error loading local-file store as the root folder is not a valid string\n")
					continue
				}
				if !vutils.Files.CheckPathExists(vstr) {
					if doCreate {
						vutils.Files.CreateDirIfNotExist(vstr)
						vutils.Files.CreateDirIfNotExist(filepath.Join(vstr, "images"))
						vutils.Files.CreateDirIfNotExist(filepath.Join(vstr, "kernels"))
						vutils.Files.CreateDirIfNotExist(filepath.Join(vstr, "disks"))
						DoChmod(vstr, 0700, true)
						if err != nil {
							fmt.Printf("Error setting up local-file store: %s -> %s\n", vstr, err.Error())
							continue
						}
						err = DoChown(vstr, uid, gid, true)
						if err != nil {
							fmt.Printf("Error setting up local-file store: %s -> %s\n", vstr, err.Error())
							continue
						}
					} else {
						fmt.Printf("Error setting up local-file store: %s -> The target directory doesn't exist.\n", vstr)
					}
				}
			}
		}
	}
	err = DoChown(pdc.AppRoot, uid, gid, false)
	if err != nil {
		println(err.Error())
	}

	if pdc.Https != nil && pdc.Https.Enable {
		return pdc.Https.checkPaths()
	}

	return pdc.verifyOwnershipAndAccess()

}

func (pdc *PromethiumDaemonConfig) validateUsersAndGroups(doCreate bool) error {
	if _, err := user.Lookup(pdc.User); err != nil {
		if doCreate {
			if err := createUser(pdc.User, "Promethium Daemon User Account"); err != nil {
				return err
			}
		} else {
			return err
		}
	}
	if _, err := user.LookupGroup(pdc.Group); err != nil {
		if doCreate {
			if err := createGroup(pdc.User, pdc.Group); err != nil {
				return err
			}
		} else {
			return err
		}
	}
	if _, err := user.Lookup(pdc.JailUser); err != nil {
		if doCreate {
			if err := createUser(pdc.JailUser, "Promethium Daemon Jail User Account"); err != nil {
				return err
			}
		} else {
			return err
		}
	}
	if _, err := user.LookupGroup(pdc.JailGroup); err != nil {
		if doCreate {
			if err := createGroup(pdc.JailUser, pdc.JailGroup); err != nil {
				return err
			}
		} else {
			return err
		}
	}
	if err := installSudoers(pdc.Group, pdc.AppRoot); err != nil {
		return err
	}
	return nil
}

func createUser(username string, gecos string) error {
	cmd := vutils.Exec.CreateAsyncCommand("adduser", false, "--system", "--shell", "/dev/null", "--no-create-home", "--gecos", gecos, "--disabled-password", "--disabled-login", username)
	return cmd.Sudo().BindToStdoutAndStdErr().StartAndWait()
}

func createGroup(username string, groupname string) error {
	cmd := vutils.Exec.CreateAsyncCommand("addgroup", false, "--system", groupname)
	err := cmd.Sudo().BindToStdoutAndStdErr().StartAndWait()
	if err != nil {
		return err
	}
	cmd = vutils.Exec.CreateAsyncCommand("usermod", false, "-aG", groupname, username)
	return cmd.Sudo().BindToStdoutAndStdErr().StartAndWait()
}

func installSudoers(groupname string, rootPath string) error {
	//if vutils.Files.CheckPathExists("/etc/sudoers.d/promethium") {
	//	return nil
	//}
	binPath := filepath.Join(rootPath, "bin")
	t, err := template.New("sudoers").Parse(PROMETHIUM_SUDOERS_TEMPLATE)
	if err != nil {
		return err
	}
	var tpl bytes.Buffer
	t.Execute(&tpl, &SudoersTemplateData{
		Group:           groupname,
		RootCommandList: strings.Join(getRootCommandListForPath(binPath), ", "),
	})
	println(tpl.String())
	cmd := vutils.Exec.CreateAsyncCommand("/bin/bash", false, "-c", "echo \""+tpl.String()+"\" > /etc/sudoers.d/promethium")
	return cmd.Sudo().BindToStdoutAndStdErr().StartAndWait()
}

func checkPathPerms(path string, uid int, gid int, filePerms os.FileMode, dirPerms os.FileMode, recursive bool) error {

	fi, err := os.Stat(path)
	if err != nil {
		return err
	} else if fi.IsDir() {
		//check dir perms
		dperms := fi.Mode().Perm()
		duid, dgid, err := getUidAndGidForFileInfo(fi)
		if err != nil {
			return err
		} else if uid != duid {
			return fmt.Errorf("Dir Path %s not owned by correct UID: %d. Current UID: %d", path, uid, duid)
		} else if gid != dgid {
			return fmt.Errorf("Dir Path %s not owned by correct GID: %d. Current GID: %d", path, gid, dgid)
		} else if dperms != dirPerms {
			return fmt.Errorf("Dir Path %s permissions arent set correctly. Required: %#o. Current: %#o", path, dirPerms, dperms)
		} else if recursive {
			//with all children perform the check...
			return filepath.Walk(path, func(sub_path string, _ os.FileInfo, err error) error {
				if path == sub_path {
					return nil
				}
				if err != nil {
					return err
				}
				return checkPathPerms(sub_path, uid, gid, filePerms, dirPerms, recursive)
			})
		}
	} else {
		//check file perms
		fperms := fi.Mode().Perm()
		fuid, fgid, err := getUidAndGidForFileInfo(fi)
		if err != nil {
			return err
		} else if uid != fuid {
			return fmt.Errorf("File Path %s not owned by correct UID: %d. Current UID: %d", path, uid, fuid)
		} else if gid != fgid {
			return fmt.Errorf("File Path %s not owned by correct GID: %d. Current GID: %d", path, gid, fgid)
		} else if fperms != filePerms {
			return fmt.Errorf("File Path %s permissions arent set correctly. Required: %#o. Current: %#o", path, dirPerms, fperms)
		}
	}
	return nil
}

func getUidAndGidForFileInfo(fi os.FileInfo) (UID int, GID int, err error) {
	if stat, ok := fi.Sys().(*syscall.Stat_t); ok {
		UID = int(stat.Uid)
		GID = int(stat.Gid)
	} else {
		err = errors.New("Unable to perform syscall to determing file ownership")
	}
	return
}

func getCommandPath(name string) (string, bool) {
	cmd := exec.Command("/bin/sh", "-c", "command -v "+name)
	if out, err := cmd.Output(); err != nil {
		return "", false
	} else {
		return string(out), true
	}
}

func isCommandAvailable(name string) bool {
	cmd := exec.Command("/bin/sh", "-c", "command -v "+name)
	if err := cmd.Run(); err != nil {
		return false
	}
	return true
}

func checkCommandVersion(name string, flags []string, expected string, allowSuffix bool) bool {
	if !isCommandAvailable(name) {
		return false
	} else {
		cmd := exec.Command(name, flags...)
		if out, err := cmd.Output(); err != nil {
			return false
		} else {
			outStr := strings.ToLower(strings.TrimSpace(string(out)))
			//so based ont he output string lets see if they are equal...
			lowerExp := strings.ToLower(expected)
			if outStr == lowerExp {
				return true
			} else if allowSuffix && strings.HasSuffix(outStr, " "+expected) {
				return true
			}
			return false
		}
	}
}

func (pdc *PromethiumDaemonConfig) verifyOwnershipAndAccess() error {

	uid, err := GetUserId(pdc.User)
	if err != nil {
		return err
	}
	gid, err := GetGroupId(pdc.Group)
	if err != nil {
		return err
	}
	// jail_uid, err := getUserId(pdc.JailUser)
	// if err != nil {
	// 	return err
	// }
	// jail_gid, err := getGroupId(pdc.JailGroup)
	// if err != nil {
	// 	return err
	// }

	//with each file or directory in the traget verify correct ownership...

	//1. config should be owned by promethium user and group...
	binPath := filepath.Join(pdc.AppRoot, "bin")
	err = checkPathPerms(binPath, uid, gid, 0700, 0700, true)
	if err != nil {
		return err
	}

	//firecrackerDir := filepath.Join(pdc.AppRoot, "firecracker")

	//checkPathPerms(firecrackerDir)

	storageDirPath := filepath.Join(pdc.AppRoot, "storage")
	err = checkPathPerms(storageDirPath, uid, gid, 0600, 0700, true)
	if err != nil {
		return err
	}
	instancesDir := filepath.Join(pdc.AppRoot, "instances")
	err = checkPathPerms(instancesDir, uid, gid, 0600, 0700, true)
	if err != nil {
		return err
	}

	return pdc.verifySudo()

}

var SUDO_DEFAULTS_RX = regexp.MustCompile(`\s*((([^=\s,\-+]+)([-+]?=[^,]+)?))(,\s*)?\s*`)
var SUDO_USER_RIGHTS_RX = regexp.MustCompile(`^\s*\((\S+)\)\s+(PASSWD|NOPASSWD|ALL|!ALL)\s*[:]?\s*(.+)?\s*$`)

func parseSudoListOut(name string) bool {
	cmd := exec.Command("sudo", "-l", "-U", name)
	if out, err := cmd.Output(); err != nil {
		fmt.Println(err)
		return false
	} else {
		lines := strings.Split(string(out), "\n")
		for _, line := range lines {
			//defMatches := SUDO_DEFAULTS_RX.FindStringSubmatch(line)
			ruleMatches := SUDO_USER_RIGHTS_RX.FindStringSubmatch(line)
			//println(strings.Join(defMatches, " :: "))
			if len(ruleMatches) == 4 {
				user := ruleMatches[1]
				inst := ruleMatches[2]
				bins := ruleMatches[3]
				println(user, inst, bins)
			}

		}
	}
	return true
}

func (pdc *PromethiumDaemonConfig) verifySudo() error {
	//there are a number of things the daemon user needs to do...
	//we need to verify that is possible...
	//we also need to make sure the sudo that is setup is too permissive!
	parseSudoListOut(pdc.User)
	parseSudoListOut(pdc.JailUser)
	return nil
}

type ClusterConfig struct {
	ID    string   `json:"id"`
	Nodes []string `json:"nodes"`
}

type UnixAPIConfig struct {
	Enable bool   `json:"enable"`
	Path   string `json:"path"`
}

type HttpAPIConfig struct {
	Enable      bool   `json:"enable"`
	BindAddress string `json:"bindAddress"`
	Port        uint   `json:"port"`
}

type ClientCertPolicy string

const (
	DisableClientCerts          ClientCertPolicy = "disable"
	RequestClientCerts          ClientCertPolicy = "request"
	RequireAnyClientCerts       ClientCertPolicy = "require-any"
	VerifyClientCertsIfGiven    ClientCertPolicy = "verify-if-provided"
	RequireAndVerifyClientCerts ClientCertPolicy = "require-and-verify"
)

type HttpsAPIConfig struct {
	*HttpAPIConfig
	PrivateKey        string           `json:"privateKey"`
	Certificate       string           `json:"certificate"`
	CACertificate     string           `json:"caCertificate"`
	ClientCertPolicy  ClientCertPolicy `json:"enableClientCert"`
	ClientCertCACerts []string         `json:"clientCertCACerts"`
}

func (httpsConf *HttpsAPIConfig) checkPaths() error {
	//check the paths of all defined config items
	if !vutils.Files.CheckPathExists(httpsConf.PrivateKey) {
		return os.ErrNotExist
	} else if !vutils.Files.CheckPathExists(httpsConf.Certificate) {
		return os.ErrNotExist
	} else {
		//check format now...
		_, err := tls.LoadX509KeyPair(httpsConf.PrivateKey, httpsConf.Certificate)
		if err != nil {
			return err
		}
		caCertPool := x509.NewCertPool()
		if httpsConf.CACertificate != "" && !vutils.Files.CheckPathExists(httpsConf.CACertificate) {
			return os.ErrNotExist
		} else if httpsConf.CACertificate != "" {
			caCert, caCertErr := ioutil.ReadFile(httpsConf.CACertificate)
			if caCertErr != nil {
				return caCertErr
			}

			ok := caCertPool.AppendCertsFromPEM(caCert)
			if !ok {
				return fmt.Errorf("cannot parse CA certificate")
			}
		}

		if httpsConf.ClientCertCACerts != nil && len(httpsConf.ClientCertCACerts) > 0 {
			for _, certPath := range httpsConf.ClientCertCACerts {
				caCert, caCertErr := ioutil.ReadFile(certPath)
				if caCertErr != nil {
					return fmt.Errorf("cannot parse client CA certificate at path %s: %s", certPath, caCertErr.Error())
				}

				ok := caCertPool.AppendCertsFromPEM(caCert)
				if !ok {
					return fmt.Errorf("cannot parse client CA certificate at path %s", certPath)
				}
			}
		}
	}
	return nil
}

type StorageConfig struct {
	ID     string                 `json:"id"`
	Driver string                 `json:"driver"`
	Config map[string]interface{} `json:"config"`
}

const PROMETHIUM_DEFAULT_USER = "promethium"
const PROMETHIUM_DEFAULT_GROUP = "promethium"

const PROMETHIUM_DEFAULT_JAIL_USER = "promethium_jail"
const PROMETHIUM_DEFAULT_JAIL_GROUP = "promethium_jail"

const PROMETHIUM_DEFAULT_HTTP_PORT = uint(8921)
const PROMETHIUM_DEFAULT_HTTP_BIND_ADDRESS = "127.0.0.1"

const PROMETHIUM_DEFAULT_HTTPS_PORT = uint(8922)
const PROMETHIUM_DEFAULT_HTTPS_BIND_ADDRESS = "127.0.0.1"

const PROMETHIUM_SOCKET_PATH = "/tmp/promethium.sock"

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
	oconfig.configPath = PROMETHIUM_DAEMON_CONFIG_LOCATION
	// err = oconfig.Lock()
	// if err != nil {
	// 	return nil, err
	// }
	err = oconfig.initiateWatcher()
	if err != nil {
		return nil, err
	}
	return oconfig, nil
}

func LoadPromethiumDaemonConfigAtPath(configPath string) (*PromethiumDaemonConfig, error) {
	configOutPath := configPath
	if !strings.HasSuffix(configPath, string(filepath.Separator)+"daemon.json") {
		configOutPath = filepath.Join(configPath, "daemon.json")
	}
	if !vutils.Files.CheckPathExists(configOutPath) {
		return nil, errors.New("Promethium daemon config not loadable at config location")
	}
	oconfig := &PromethiumDaemonConfig{}
	err := vutils.Config.GetConfigFromDefaultList("promethium.daemon", CWD, []string{configOutPath}, oconfig)
	if err != nil {
		return nil, err
	}
	oconfig.configPath = configOutPath
	// if err := oconfig.validate(true); err != nil {
	// 	return nil, err
	// }
	// err = oconfig.Lock()
	// if err != nil {
	// 	return nil, err
	// }
	err = oconfig.initiateWatcher()
	if err != nil {
		return nil, err
	}
	return oconfig, nil
}

func (pdc *PromethiumDaemonConfig) SetUpdateCallback(callback PromethiumDaemonConfigUpdateCallback) {
	pdc.updateCallback = callback
}

func (pdc *PromethiumDaemonConfig) IsNewConfig() bool {
	return IS_NEW_CONFIG
}

func (pdc *PromethiumDaemonConfig) EnableHttp(address string, port uint) {
	pdc.Http = &HttpAPIConfig{
		Enable:      true,
		BindAddress: address,
		Port:        port,
	}
}

func (pdc *PromethiumDaemonConfig) DisableHttp() {
	if pdc.Http != nil {
		pdc.Http.Enable = false
	}
}

func (pdc *PromethiumDaemonConfig) EnableHttps(address string, port uint, privateKey string, cert string, caCert string, clientCertPolicy ClientCertPolicy, clientCertCACerts []string) {
	pdc.Https = &HttpsAPIConfig{
		HttpAPIConfig: &HttpAPIConfig{
			Enable:      true,
			BindAddress: address,
			Port:        port,
		},
		PrivateKey:        privateKey,
		Certificate:       cert,
		CACertificate:     caCert,
		ClientCertCACerts: []string{},
	}
	if caCert != "" {
		pdc.Https.ClientCertCACerts = append(pdc.Https.ClientCertCACerts, caCert)
	}
	if clientCertCACerts != nil && len(clientCertCACerts) > 0 {
		pdc.Https.ClientCertCACerts = append(pdc.Https.ClientCertCACerts, clientCertCACerts...)
	}
}

func (pdc *PromethiumDaemonConfig) DisableHttps() {
	if pdc.Https != nil {
		pdc.Https.Enable = false
	}
}

func (pdc *PromethiumDaemonConfig) callCallback(area PromethiumDaemonConfigUpdateCallbackArea, scope string, add []string, update []string, remove []string) {
	if pdc.updateCallback != nil {
		pdc.updateCallback(area, scope, add, update, remove)
	}
}

func (pdc *PromethiumDaemonConfig) tryLoadConf() []error {
	pdc.locked = true
	log.Println("Trying to load modified config")
	errList := []error{}
	oconfig := &PromethiumDaemonConfig{}
	err := vutils.Config.GetConfigFromDefaultList("promethium.daemon", CWD, []string{pdc.configPath}, oconfig)
	if err != nil {
		pdc.locked = false
		return append(errList, err)
	}
	//now we have a loaded config lets see what needs to be inserted
	if oconfig.NodeID != pdc.NodeID {
		errList = append(errList, errors.New("Unable to change NodeID at runtime"))
	}
	if oconfig.AppRoot != pdc.AppRoot {
		errList = append(errList, errors.New("Unable to change AppRoot at runtime"))
	}
	if oconfig.User != pdc.User {
		errList = append(errList, errors.New("Unable to change User at runtime"))
	}
	if oconfig.Group != pdc.Group {
		errList = append(errList, errors.New("Unable to change Group at runtime"))
	}
	if oconfig.JailUser != pdc.JailUser {
		errList = append(errList, errors.New("Unable to change JailUser at runtime"))
	}
	if oconfig.JailGroup != pdc.JailGroup {
		errList = append(errList, errors.New("Unable to change JailGroup at runtime"))
	}
	if err := pdc.validateUnixConfig(oconfig.Unix); err != nil {
		errList = append(errList, err)
	}
	if err := pdc.validateHttpConfig(oconfig.Http); err != nil {
		errList = append(errList, err)
	}
	if err := pdc.validateHttpsConfig(oconfig.Https); err != nil {
		errList = append(errList, err)
	}
	//now lets check the rest of the config.. cluster stuff cannot be changed..
	if err := pdc.validateClusterConfig(oconfig.Clusters); err != nil {
		errList = append(errList, err)
	}
	if err := pdc.validateStorageConfig(oconfig.Storage); err != nil {
		errList = append(errList, err)
	}
	if err := pdc.validateNetworkConfig(oconfig.Networks); err != nil {
		errList = append(errList, err)
	}
	if err := pdc.SaveConfig(); err != nil {
		errList = append(errList, err)
	}
	time.Sleep(2 * time.Second)
	pdc.locked = false
	log.Println("Loaded modified config")
	return errList
}
func contains(s []string, searchterm string) (bool, int) {
	i := sort.SearchStrings(s, searchterm)
	return i < len(s) && s[i] == searchterm, i
}
func storageConfContains(s []*StorageConfig, ID string) (bool, int) {
	for ind, item := range s {
		if item.ID == ID {
			return true, ind
		}
	}
	return false, 0
}
func networkConfContains(s []*networking.NetworkConfig, ID string) (bool, int) {
	for ind, item := range s {
		if item.ID == ID {
			return true, ind
		}
	}
	return false, 0
}
func removeAtindex(s []string, index int) []string {
	return append(s[:index], s[index+1:]...)
}
func removeStorageConfAtindex(s []*StorageConfig, index int) []*StorageConfig {
	return append(s[:index], s[index+1:]...)
}
func removeNetworkConfAtindex(s []*networking.NetworkConfig, index int) []*networking.NetworkConfig {
	return append(s[:index], s[index+1:]...)
}
func (pdc *PromethiumDaemonConfig) GetClusterConf(clusterID string) *ClusterConfig {
	if pdc.Clusters == nil || len(pdc.Clusters) == 0 {
		return nil
	} else {
		for _, clusterConf := range pdc.Clusters {
			if clusterConf.ID == clusterID {
				return clusterConf
			}
		}
	}
	return nil
}
func (pdc *PromethiumDaemonConfig) GetStorageConf(id string) (*StorageConfig, int) {
	if pdc.Storage == nil || len(pdc.Storage) == 0 {
		return nil, 0
	} else {
		for ind, storageConf := range pdc.Storage {
			if storageConf.ID == id {
				return storageConf, ind
			}
		}
	}
	return nil, 0
}
func (pdc *PromethiumDaemonConfig) GetNetworkConf(id string) (*networking.NetworkConfig, int) {
	if pdc.Networks == nil || len(pdc.Networks) == 0 {
		return nil, 0
	} else {
		for ind, networkConf := range pdc.Networks {
			if networkConf.ID == id {
				return networkConf, ind
			}
		}
	}
	return nil, 0
}
func (pdc *PromethiumDaemonConfig) validateClusterConfig(newConf []*ClusterConfig) error {
	if newConf != nil {
		if pdc.Clusters == nil {
			return errors.New("Cannot join node to cluster by changing config. Use promethium cluster join or promethium cluster create.")
		}
		//iterate the configs...
		for _, clusterConf := range newConf {
			if clusterConf == nil {
				continue
			}
			existClusConf := pdc.GetClusterConf(clusterConf.ID)
			if existClusConf == nil {
				//doestnt exist
				return errors.New("Cannot join node to cluster by changing config. Use promethium cluster join.")
			} else if clusterConf.Nodes != nil && len(clusterConf.Nodes) > 0 {
				add := []string{}
				remove := []string{}
				nodes := []string{}
				for _, node := range clusterConf.Nodes {
					if xist, _ := contains(existClusConf.Nodes, node); !xist {
						add = append(add, node)
					}
					nodes = append(nodes, node)
				}

				for _, node := range existClusConf.Nodes {
					if xist, _ := contains(clusterConf.Nodes, node); !xist {
						remove = append(remove, node)
					} else {
						nodes = append(nodes, node)
					}
				}
				existClusConf.Nodes = nodes
				if len(add) > 0 || len(remove) > 0 {
					pdc.callCallback(ClusterNodesUpdate, clusterConf.ID, add, nil, remove)
				}
			}
		}
	}
	return nil
}
func (pdc *PromethiumDaemonConfig) validateStorageConfig(newConf []*StorageConfig) error {
	if newConf != nil {
		if pdc.Storage == nil {
			return errors.New("Unable to set new Storage Config at runtime.")
		}
		//iterate the configs...
		add := []string{}
		update := []string{}
		remove := []string{}
		removeInds := []int{}
		for _, storeConf := range newConf {
			if storeConf == nil {
				continue
			}
			existStoreConf, _ := pdc.GetStorageConf(storeConf.ID)
			if existStoreConf == nil {
				//doestnt exist
				add = append(add, storeConf.ID)
				pdc.Storage = append(pdc.Storage, storeConf)
			} else {
				//it exists.. is it to be updated?
				if existStoreConf.Driver != storeConf.Driver {
					return errors.New("Unable to change Storage driver at runtime.")
				} else {
					existStoreConf.Config = storeConf.Config
					update = append(update, existStoreConf.ID)
				}
			}
		}
		for _, storeConf := range pdc.Storage {
			if storeConf == nil {
				continue
			}
			if xists, index := storageConfContains(newConf, storeConf.ID); !xists {
				remove = append(remove, storeConf.ID)
				removeInds = append(removeInds, index)
			}
		}
		if len(removeInds) > 0 {
			for _, ind := range removeInds {
				removeStorageConfAtindex(pdc.Storage, ind)
			}
		}
		if len(add) > 0 || len(update) > 0 || len(remove) > 0 {
			pdc.callCallback(StorageUpdate, "", add, update, remove)
		}
	}
	return nil
}
func (pdc *PromethiumDaemonConfig) validateNetworkConfig(newConf []*networking.NetworkConfig) error {
	if newConf != nil {
		if pdc.Networks == nil {
			return errors.New("Unable to set new Network Config at runtime.")
		}
		//iterate the configs...
		add := []string{}
		update := []string{}
		remove := []string{}
		removeInds := []int{}
		for _, netConf := range newConf {
			if netConf == nil {
				continue
			}
			existNetConf, _ := pdc.GetNetworkConf(netConf.ID)
			if existNetConf == nil {
				//doestnt exist
				add = append(add, netConf.ID)
				pdc.Networks = append(pdc.Networks, netConf)
			} else {
				//it exists.. is it to be updated?
				if existNetConf.Type != netConf.Type {
					return errors.New("Unable to change Network driver/type at runtime.")
				} else {
					if existNetConf.Enabled != netConf.Enabled {
						existNetConf.Enabled = netConf.Enabled
					}
					if existNetConf.Name != netConf.Name {
						existNetConf.Name = netConf.Name
					}
					if existNetConf.MasterInterface != netConf.MasterInterface {
						existNetConf.MasterInterface = netConf.MasterInterface
					}
					existNetConf.IPV4 = netConf.IPV4
					existNetConf.IPV6 = netConf.IPV6
					update = append(update, netConf.ID)
				}
			}
		}
		for _, netConf := range pdc.Networks {
			if netConf == nil {
				continue
			}
			if xists, index := networkConfContains(newConf, netConf.ID); !xists {
				remove = append(remove, netConf.ID)
				removeInds = append(removeInds, index)
			}
		}
		if len(removeInds) > 0 {
			for _, ind := range removeInds {
				removeNetworkConfAtindex(pdc.Networks, ind)
			}
		}
		if len(add) > 0 || len(update) > 0 || len(remove) > 0 {
			pdc.callCallback(NetworkingUpdate, "", add, update, remove)
		}
	}
	return nil
}
func (pdc *PromethiumDaemonConfig) validateHttpConfig(newConf *HttpAPIConfig) error {
	if newConf != nil {
		currConf := pdc.Http
		if currConf.Enable && !newConf.Enable {
			return errors.New("Unable to disable HTTP endpoint at runtime")
		} else if !currConf.Enable && newConf.Enable {
			return errors.New("Unable to enable HTTP endpoint at runtime")
		} else if currConf.BindAddress != newConf.BindAddress {
			return errors.New("Unable to change HTTP endpoint BindAddress at runtime")
		} else if currConf.Port != newConf.Port {
			return errors.New("Unable to change HTTP endpoint Port at runtime")
		}
	}
	return nil
}
func (pdc *PromethiumDaemonConfig) validateHttpsConfig(newConf *HttpsAPIConfig) error {
	if newConf != nil {
		currConf := pdc.Https
		if currConf.Enable && !newConf.Enable {
			return errors.New("Unable to disable HTTPS endpoint at runtime")
		} else if !currConf.Enable && newConf.Enable {
			return errors.New("Unable to enable HTTPS endpoint at runtime")
		} else if currConf.BindAddress != newConf.BindAddress {
			return errors.New("Unable to change HTTPS endpoint BindAddress at runtime")
		} else if currConf.Port != newConf.Port {
			return errors.New("Unable to change HTTPS endpoint Port at runtime")
		} else if currConf.PrivateKey != newConf.PrivateKey {
			return errors.New("Unable to change HTTPS endpoint PrivateKey at runtime")
		} else if currConf.Certificate != newConf.Certificate {
			return errors.New("Unable to change HTTPS endpoint Certificate at runtime")
		} else if currConf.ClientCertPolicy != newConf.ClientCertPolicy {
			return errors.New("Unable to change HTTPS endpoint ClientCertPolicy at runtime")
		} else if newConf.ClientCertCACerts != nil && len(newConf.ClientCertCACerts) > 0 && currConf.ClientCertCACerts != nil && len(currConf.ClientCertCACerts) > 0 {
			if newConf.ClientCertCACerts[0] != currConf.ClientCertCACerts[0] {
				return errors.New("Unable to change HTTPS endpoint ClientCertCACerts at runtime")
			}
		}
	}
	return nil
}
func (pdc *PromethiumDaemonConfig) validateUnixConfig(newConf *UnixAPIConfig) error {
	if newConf != nil {
		currConf := pdc.Unix
		if currConf.Enable && !newConf.Enable {
			return errors.New("Unable to disable UNIX endpoint at runtime")
		} else if !currConf.Enable && newConf.Enable {
			return errors.New("Unable to enable UNIX endpoint at runtime")
		}
	}
	return nil
}

func (pdc *PromethiumDaemonConfig) initiateWatcher() error {
	var err error = nil
	pdc.watcher, err = fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	go func() {
		for {
			select {
			case event, ok := <-pdc.watcher.Events:
				if !ok {
					return
				}
				log.Println("event:", event)
				if !pdc.locked && event.Op&fsnotify.Write == fsnotify.Write {
					log.Println("modified file:", event.Name)
					elist := pdc.tryLoadConf()
					if len(elist) > 0 {
						for _, e := range elist {
							log.Println(e.Error())
						}
					}
				}
			case err, ok := <-pdc.watcher.Errors:
				if !ok {
					return
				}
				log.Println("error:", err)
			}
		}
	}()
	return pdc.watcher.Add(pdc.configPath)

}

var ConfigAlreadyLockedErr = errors.New("The config file is already locked")
var ConfigNotLockedErr = errors.New("The config file is not currently")
var ConfigUnableToLockErr = errors.New("The config file cannot be locked")

func (pdc *PromethiumDaemonConfig) Lock() error {
	if pdc.locked {
		return ConfigAlreadyLockedErr
	}
	var err error = nil
	// pdc.fd, err = os.Open(pdc.configPath)
	// if err != nil {
	// 	return err
	// }
	pdc.lock = flock.New(pdc.configPath)
	err = pdc.lock.Lock()
	if err != nil {
		//pdc.fd.Close()
		return err
	}
	println("obtained config lock", pdc.configPath)
	pdc.locked = true
	return nil
}

func (pdc *PromethiumDaemonConfig) Unlock() error {
	if !pdc.locked || pdc.lock == nil {
		pdc.locked = false
		return ConfigNotLockedErr
	}
	pdc.lock.Unlock()
	//pdc.fd.Close()
	pdc.locked = false
	println("released config lock", pdc.configPath)
	return nil
}

func (pdc *PromethiumDaemonConfig) SaveConfig() error {
	// if pdc.locked {
	// 	pdc.Unlock()
	// }
	err, _ := vutils.Config.TrySaveConfig(CWD, []string{pdc.configPath}, pdc)
	if err != nil {
		err = pdc.Lock()
		if err != nil {
			println(err.Error())
		}
		return err
	}
	//return pdc.Lock()
	return nil
}

func generateNewPromethiumDaemonConfig() error {
	println("Generating new default config")
	defaultPromDir := "/opt/promethium"

	return NewPromethiumDaemonConfig(PROMETHIUM_DAEMON_CONFIG_LOCATION, defaultPromDir, PROMETHIUM_SOCKET_PATH, PROMETHIUM_DEFAULT_USER, PROMETHIUM_DEFAULT_GROUP, PROMETHIUM_DEFAULT_JAIL_USER, PROMETHIUM_DEFAULT_JAIL_GROUP, true, PROMETHIUM_DEFAULT_HTTP_PORT, PROMETHIUM_DEFAULT_HTTP_BIND_ADDRESS, false, 0, "", "", "", "")

}

func NewPromethiumDaemonConfig(configPath string, dataPath string, socketPath string, user string, group string, jailUser string, jailGroup string, enableHttp bool, httpPort uint, httpBindAddr string, enableHttps bool, httpsPort uint, httpsBindAddress string, privateKey string, cert string, caCert string) error {
	println("Creating new config")
	defaultPromDir := dataPath
	storageName := "default-local"
	configOutPath := configPath
	if !strings.HasSuffix(configPath, string(filepath.Separator)+"daemon.json") {
		configOutPath = filepath.Join(configPath, "daemon.json")
	}
	storageDir := filepath.Join(defaultPromDir, "storage", storageName)
	newUUID, _ := vutils.UUID.MakeUUIDString()
	oconfig := &PromethiumDaemonConfig{
		NodeID:   newUUID,
		Clusters: []*ClusterConfig{},
		Storage: []*StorageConfig{
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
		User:      user,
		Group:     group,
		JailUser:  jailUser,
		JailGroup: jailGroup,
		Unix: &UnixAPIConfig{
			Enable: true,
			Path:   PROMETHIUM_SOCKET_PATH,
		},
	}

	if enableHttp && httpPort > 0 && httpPort < 65535 && httpBindAddr != "" {
		oconfig.EnableHttp(httpBindAddr, httpPort)
	}

	if enableHttps && httpsPort > 0 && httpsPort < 65535 && httpsBindAddress != "" {
		oconfig.EnableHttps(httpsBindAddress, httpsPort, privateKey, cert, caCert, DisableClientCerts, nil)
	}

	if err := oconfig.validate(true); err != nil {
		return err
	}

	oconfig.configPath = configOutPath

	err := oconfig.SaveConfig()
	if err != nil {
		//return err
		return err
	}
	IS_NEW_CONFIG = true
	//now create the default folder structure...

	//
	return oconfig.initiateWatcher()

	//return err
}

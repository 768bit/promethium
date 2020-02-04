package config

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"

	"text/template"

	"github.com/768bit/promethium/lib/networking"
	"github.com/768bit/vutils"
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
	ID    string        `json:"id"`
	Nodes []interface{} `json:"nodes"`
}

type UnixAPIConfig struct {
	Enable bool   `json:"enable"`
	Path   string `json:"path"`
	User   string `json:"user"`
	Group  string `json:"group"`
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
	// if err := oconfig.validate(true); err != nil {
	// 	return nil, err
	// }
	return oconfig, nil
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

	err, _ := vutils.Config.TrySaveConfig(CWD, []string{configOutPath}, oconfig)
	if err != nil {
		//return err
		return err
	}
	IS_NEW_CONFIG = true
	//now create the default folder structure...

	//

	return err
}

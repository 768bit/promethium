package cloudconfig

import (
  "encoding/json"
  "gopkg.in/yaml.v2"
  "io/ioutil"
)

type UserData struct {
  Groups []string `yaml:"groups,omitempty,flow"`
  Users  []*UserDataUserConfig `yaml:"users,omitempty,flow"`
  PackageUpdate bool `yaml:"package_update,omitempty"`
  PackageUpgrade bool `yaml:"package_upgrade,omitempty"`
  Packages []string `yaml:"packages,omitempty,flow"`
  RunCmd []string `yaml:"runcmd,omitempty,flow"`
  BootCmd []string `yaml:"bootcmd,omitempty,flow"`
  Locale string `yaml:"locale,omitempty"`
  SshAuthorisedKeys []string `yaml:"ssh_authorized_keys,omitempty,flow"`
  DisableRoot bool `yaml:"disable_root,omitempty"`

}

type UserDataUserConfig struct {
  Name string `yaml:"name,omitempty"`
  Gecos string `yaml:"gecos,omitempty"`
  PrimaryGroup string `yaml:"primary_group,omitempty"`
  Groups []string `yaml:"groups,omitempty,flow"`
  Sudo string `yaml:"sudo,omitempty"`
  ExpireDate string `yaml:"expiredate,omitempty"`
  SshAuthorisedKeys []string `yaml:"ssh_authorized_keys,omitempty,flow"`
  Inactive bool `yaml:"inactive,omitempty"`
  System bool `yaml:"system,omitempty"`
  LockPassword bool `yaml:"lock_passwd,omitempty"`
  Password string `yaml:"passwd,omitempty"`
  Shell string `yaml:"shell,omitempty"`
}

func (md *UserData) WriteUserData(dest string) error {
  x, err := yaml.Marshal(md)
  if err != nil {
    return err
  }
  x = append([]byte("#cloud-config\n"), x...)
  return ioutil.WriteFile(dest, x, 0660)
}

func (md *UserData) WriteUserDataJSON(dest string) error {
  x, err := json.Marshal(md)
  if err != nil {
    return err
  }
  return ioutil.WriteFile(dest, x, 0660)
}

package config

import (
	"io/ioutil"
	"os"
	"testing"
)

func TestMakeCloudInitImage(t *testing.T) {
	tdir, err := ioutil.TempDir("", "prmtest")
	defer os.RemoveAll(tdir)
	tdir2, err := ioutil.TempDir("", "prmtest2")
	defer os.RemoveAll(tdir2)
	testUser := "prom_test"
	testJailUser := "prom_jail_test"
	err = NewPromethiumDaemonConfig(tdir, tdir2, "/etc/prom", testUser, testUser, testJailUser, testJailUser, false, 0, "", false, 0, "", "", "", "")
	if err != nil {
		t.Error(err)
	}
}

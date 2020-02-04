package config

import (
	"fmt"
	"os"
	"os/user"
	"strconv"

	"github.com/768bit/vutils"
)

func DoChown(path string, user int, group int, recursive bool) error {
	cmdParams := []string{}
	if recursive {
		cmdParams = append(cmdParams, "-R")
	}
	cmdParams = append(cmdParams, fmt.Sprintf("%d:%d", user, group), path)
	cmd := vutils.Exec.CreateAsyncCommand("chown", false, cmdParams...)
	return cmd.Sudo().BindToStdoutAndStdErr().StartAndWait()
}

func DoChmod(path string, mode os.FileMode, recursive bool) error {
	cmdParams := []string{}
	if recursive {
		cmdParams = append(cmdParams, "-R")
	}
	cmdParams = append(cmdParams, fmt.Sprintf("%#o", mode), path)
	cmd := vutils.Exec.CreateAsyncCommand("chmod", false, cmdParams...)
	return cmd.Sudo().BindToStdoutAndStdErr().StartAndWait()
}

func GetGroupId(groupname string) (int, error) {
	if u, err := user.LookupGroup(groupname); err != nil {
		return -1, err
	} else {
		return strconv.Atoi(u.Gid)
	}
}

func GetUserId(username string) (int, error) {
	if u, err := user.Lookup(username); err != nil {
		return -1, err
	} else {
		return strconv.Atoi(u.Uid)
	}
}

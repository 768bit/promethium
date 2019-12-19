package images

import (
	"github.com/768bit/vutils"
)

func MakeExt4(name string, target string) error {
	cmd := vutils.Exec.CreateAsyncCommand("mkfs.ext4", false, "-L", name, target).Sudo().BindToStdoutAndStdErr()
	return cmd.StartAndWait()
}

func MakeExfat(name string, target string) error {
	cmd := vutils.Exec.CreateAsyncCommand("mkfs.exfat", false, "-n", name, target).Sudo().BindToStdoutAndStdErr()
	return cmd.StartAndWait()
}

func MakeFat32(name string, target string) error {
	cmd := vutils.Exec.CreateAsyncCommand("mkfs.fat", false, "-F", "32", "-n", name, target).Sudo().BindToStdoutAndStdErr()
	return cmd.StartAndWait()
}

func MakeNtfs(name string, target string) error {
	cmd := vutils.Exec.CreateAsyncCommand("mkfs.ntfs", false, "-f", "-L", name, target).Sudo().BindToStdoutAndStdErr()
	return cmd.StartAndWait()
}

//go:build !windows

package runtime

import (
	"os/exec"
	"syscall"
)

func configureChroot(cmd *exec.Cmd, rootfs string) error {
	cmd.SysProcAttr = &syscall.SysProcAttr{Chroot: rootfs}
	return nil
}
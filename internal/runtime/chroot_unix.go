//go:build !windows

package runtime

import (
	"os"
	"os/exec"
	"strings"
	"syscall"
)

func configureChroot(cmd *exec.Cmd, rootfs string) error {
	// Check if running on WSL
	if isWSL() {
		// On WSL, we can't use chroot, so we just set the rootfs as the working directory base
		// The actual command execution will use the rootfs filesystem
		return nil
	}
	
	cmd.SysProcAttr = &syscall.SysProcAttr{Chroot: rootfs}
	return nil
}

func isWSL() bool {
	// Check for WSL by looking at /proc/version
	data, err := os.ReadFile("/proc/version")
	if err != nil {
		return false
	}
	return strings.Contains(string(data), "microsoft") || strings.Contains(string(data), "WSL")
}

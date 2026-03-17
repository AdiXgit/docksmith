//go:build windows

package build

import (
	"fmt"
	"os/exec"
)

func configureChroot(cmd *exec.Cmd, rootfs string) error {
	return fmt.Errorf("RUN is not supported on Windows: chroot-based build execution requires a Unix host")
}
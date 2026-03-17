//go:build windows

package runtime

import (
	"fmt"
	"os/exec"
)

func configureChroot(cmd *exec.Cmd, rootfs string) error {
	return fmt.Errorf("run is not supported on Windows: chroot-based image execution requires a Unix host")
}
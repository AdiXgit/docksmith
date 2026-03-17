package build

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func ensureWorkdir(rootfs, workdir string) error {
	if workdir == "" {
		return nil
	}
	return os.MkdirAll(filepath.Join(rootfs, workdir), 0755)
}

func runInRoot(rootfs, workdir string, env map[string]string, command string) error {
	if err := ensureWorkdir(rootfs, workdir); err != nil {
		return err
	}

	cmd := exec.Command("/bin/sh", "-c", command)
	cmd.Dir = "/"

	cmd.Env = []string{}
	for k, v := range env {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := configureChroot(cmd, rootfs); err != nil {
		return err
	}

	if workdir != "" {
		cmd.Dir = workdir
	}
	return cmd.Run()
}

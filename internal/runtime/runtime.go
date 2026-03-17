package runtime

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"docksmith/internal/image"
	"docksmith/internal/util"
)

type RunOptions struct {
	NameTag string
	Cmd     []string
	Env     map[string]string
}

func RunImage(opts RunOptions) error {
	if err := image.EnsureStore(); err != nil {
		return err
	}
	m, err := image.LoadManifest(opts.NameTag)
	if err != nil {
		return err
	}

	rootfs, err := os.MkdirTemp("", "docksmith-run-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(rootfs)

	for _, layer := range m.Layers {
		if err := util.UntarInto(image.LayerPath(layer.Digest), rootfs); err != nil {
			return err
		}
	}

	cmd := opts.Cmd
	if len(cmd) == 0 {
		cmd = m.Config.Cmd
	}
	if len(cmd) == 0 {
		return fmt.Errorf("no CMD defined in image and no runtime command provided")
	}

	envMap := map[string]string{}
	for _, kv := range m.Config.Env {
		var k, v string
		for i := 0; i < len(kv); i++ {
			if kv[i] == '=' {
				k = kv[:i]
				v = kv[i+1:]
				break
			}
		}
		if k != "" {
			envMap[k] = v
		}
	}
	for k, v := range opts.Env {
		envMap[k] = v
	}

	command := exec.Command(cmd[0], cmd[1:]...)
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	command.Stdin = os.Stdin
	if err := configureChroot(command, rootfs); err != nil {
		return err
	}

	command.Dir = "/"
	if m.Config.WorkingDir != "" {
		if err := os.MkdirAll(filepath.Join(rootfs, m.Config.WorkingDir), 0755); err != nil {
			return err
		}
		command.Dir = m.Config.WorkingDir
	}

	for k, v := range envMap {
		command.Env = append(command.Env, fmt.Sprintf("%s=%s", k, v))
	}

	err = command.Run()
	if exitErr, ok := err.(*exec.ExitError); ok {
		fmt.Printf("Container exited with code %d\n", exitErr.ExitCode())
		return nil
	}
	if err == nil {
		fmt.Println("Container exited with code 0")
	}
	return err
}

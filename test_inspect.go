package main

import (
	"fmt"
	"docksmith/internal/image"
)

func main() {
	m, _ := image.LoadManifest("test:1.0")
	fmt.Printf("WorkingDir: %s\n", m.Config.WorkingDir)
	fmt.Printf("Cmd: %v\n", m.Config.Cmd)
}

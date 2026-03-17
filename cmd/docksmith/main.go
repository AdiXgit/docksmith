package main

import (
	"fmt"
	"os"

	"docksmith/internal/build"
	"docksmith/internal/cli"
	"docksmith/internal/image"
	"docksmith/internal/runtime"
)

func main() {
	args, err := cli.Parse(os.Args[1:])
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}

	switch args.Command {
	case "build":
		if err := build.Build(build.BuildOptions{
			NameTag:  args.NameTag,
			Context:  args.Context,
			NoCache:  args.NoCache,
			Override: nil,
		}); err != nil {
			fmt.Fprintln(os.Stderr, "Build failed:", err)
			os.Exit(1)
		}
	case "images":
		if err := image.ListImages(); err != nil {
			fmt.Fprintln(os.Stderr, "images failed:", err)
			os.Exit(1)
		}
	case "run":
		if err := runtime.RunImage(runtime.RunOptions{
			NameTag: args.NameTag,
			Cmd:     args.Cmd,
			Env:     args.Env,
		}); err != nil {
			fmt.Fprintln(os.Stderr, "run failed:", err)
			os.Exit(1)
		}
	case "rmi":
		if err := image.RemoveImage(args.NameTag); err != nil {
			fmt.Fprintln(os.Stderr, "rmi failed:", err)
			os.Exit(1)
		}
	default:
		fmt.Fprintln(os.Stderr, "unknown command")
		os.Exit(1)
	}
}

package cli

import (
	"errors"
	"fmt"
	"strings"
)

type ParsedArgs struct {
	Command string
	NameTag string
	Context string
	NoCache bool
	Cmd     []string
	Env     map[string]string
}

func Parse(argv []string) (ParsedArgs, error) {
	if len(argv) == 0 {
		return ParsedArgs{}, errors.New("usage: docksmith <build|images|run|rmi> ...")
	}

	switch argv[0] {
	case "build":
		return parseBuild(argv[1:])
	case "images":
		return ParsedArgs{Command: "images"}, nil
	case "run":
		return parseRun(argv[1:])
	case "rmi":
		if len(argv) != 2 {
			return ParsedArgs{}, errors.New("usage: docksmith rmi <name:tag>")
		}
		return ParsedArgs{
			Command: "rmi",
			NameTag: argv[1],
		}, nil
	default:
		return ParsedArgs{}, fmt.Errorf("unknown command %q", argv[0])
	}
}

func parseBuild(argv []string) (ParsedArgs, error) {
	out := ParsedArgs{Command: "build"}
	i := 0
	for i < len(argv) {
		switch argv[i] {
		case "-t":
			if i+1 >= len(argv) {
				return out, errors.New("missing value after -t")
			}
			out.NameTag = argv[i+1]
			i += 2
		case "--no-cache":
			out.NoCache = true
			i++
		default:
			if strings.HasPrefix(argv[i], "-") {
				return out, fmt.Errorf("unknown flag %s", argv[i])
			}
			out.Context = argv[i]
			i++
		}
	}
	if out.NameTag == "" || out.Context == "" {
		return out, errors.New("usage: docksmith build -t <name:tag> <context> [--no-cache]")
	}
	return out, nil
}

func parseRun(argv []string) (ParsedArgs, error) {
	out := ParsedArgs{
		Command: "run",
		Env:     map[string]string{},
	}
	i := 0
	for i < len(argv) {
		if argv[i] == "-e" {
			if i+1 >= len(argv) {
				return out, errors.New("missing KEY=VALUE after -e")
			}
			kv := argv[i+1]
			parts := strings.SplitN(kv, "=", 2)
			if len(parts) != 2 {
				return out, fmt.Errorf("invalid env override %q", kv)
			}
			out.Env[parts[0]] = parts[1]
			i += 2
			continue
		}
		if out.NameTag == "" {
			out.NameTag = argv[i]
			i++
			continue
		}
		out.Cmd = append(out.Cmd, argv[i:]...)
		break
	}
	if out.NameTag == "" {
		return out, errors.New("usage: docksmith run [-e KEY=VALUE] <name:tag> [cmd ...]")
	}
	return out, nil
}

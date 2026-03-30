package build

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type Instruction struct {
	Op      string
	Raw     string
	Args    string
	LineNum int
}

func ParseDocksmithfile(path string) ([]Instruction, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var out []Instruction
	scanner := bufio.NewScanner(f)
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, " ", 2)
		op := strings.ToUpper(parts[0])

		switch op {
		case "FROM", "COPY", "RUN", "WORKDIR", "ENV", "CMD":
		default:
			return nil, fmt.Errorf("unrecognized instruction at line %d: %s", lineNum, op)
		}

		args := ""
		if len(parts) > 1 {
			args = parts[1]
		}

		out = append(out, Instruction{
			Op:      op,
			Raw:     line,
			Args:    args,
			LineNum: lineNum,
		})
	}

	return out, scanner.Err()
}

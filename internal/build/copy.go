package build

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"docksmith/internal/util"
)

func parseCopyArgs(args string) (string, string, error) {
	parts := strings.Fields(args)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("COPY requires exactly 2 args")
	}
	return parts[0], parts[1], nil
}

func hashFile(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func resolveCopySources(contextDir, src string) ([]string, error) {
	var matches []string

	if strings.Contains(src, "*") {
		m, err := filepath.Glob(filepath.Join(contextDir, src))
		if err != nil {
			return nil, err
		}
		for _, abs := range m {
			rel, err := filepath.Rel(contextDir, abs)
			if err != nil {
				return nil, err
			}
			matches = append(matches, rel)
		}
	} else {
		matches = append(matches, src)
	}

	sort.Strings(matches)
	return matches, nil
}

func copyIntoRoot(contextDir string, sources []string, dest string, rootfs string) ([]string, []string, error) {
	var changed []string
	var sourceHashes []string

	for _, src := range sources {
		absSrc := filepath.Join(contextDir, src)
		info, err := os.Stat(absSrc)
		if err != nil {
			return nil, nil, err
		}

		if info.IsDir() {
			err := filepath.Walk(absSrc, func(path string, fi os.FileInfo, err error) error {
				if err != nil {
					return err
				}
				relToContext, err := filepath.Rel(contextDir, path)
				if err != nil {
					return err
				}
				targetPath := filepath.Join(rootfs, dest, filepath.Base(src))
				if src == "." || src == "./" {
					targetPath = filepath.Join(rootfs, dest)
				}
				relInsideSrc, _ := filepath.Rel(absSrc, path)
				finalTarget := filepath.Join(targetPath, relInsideSrc)

				if fi.IsDir() {
					if err := os.MkdirAll(finalTarget, 0755); err != nil {
						return err
					}
					changed = append(changed, strings.TrimPrefix(strings.TrimPrefix(finalTarget, rootfs), string(filepath.Separator)))
					return nil
				}
				if err := util.CopyFile(path, finalTarget, fi.Mode()); err != nil {
					return err
				}
				h, err := hashFile(path)
				if err != nil {
					return err
				}
				sourceHashes = append(sourceHashes, relToContext+":"+h)
				changed = append(changed, strings.TrimPrefix(strings.TrimPrefix(finalTarget, rootfs), string(filepath.Separator)))
				return nil
			})
			if err != nil {
				return nil, nil, err
			}
			continue
		}

		baseDest := filepath.Join(rootfs, dest)
		var target string
		if strings.HasSuffix(dest, "/") || pathLooksLikeDir(baseDest) {
			target = filepath.Join(baseDest, filepath.Base(src))
		} else {
			target = baseDest
		}
		if err := util.CopyFile(absSrc, target, info.Mode()); err != nil {
			return nil, nil, err
		}
		h, err := hashFile(absSrc)
		if err != nil {
			return nil, nil, err
		}
		sourceHashes = append(sourceHashes, src+":"+h)
		changed = append(changed, strings.TrimPrefix(strings.TrimPrefix(target, rootfs), string(filepath.Separator)))
	}

	sort.Strings(changed)
	sort.Strings(sourceHashes)
	return changed, sourceHashes, nil
}

func pathLooksLikeDir(p string) bool {
	info, err := os.Stat(p)
	return err == nil && info.IsDir()
}

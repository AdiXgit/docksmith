package build

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"docksmith/internal/image"
	"docksmith/internal/util"
)

type BuildOptions struct {
	NameTag  string
	Context  string
	NoCache  bool
	Override map[string]string
}

func Build(opts BuildOptions) error {
	if err := image.EnsureStore(); err != nil {
		return err
	}

	dfPath := filepath.Join(opts.Context, "Docksmithfile")
	instructions, err := ParseDocksmithfile(dfPath)
	if err != nil {
		return err
	}
	if len(instructions) == 0 || instructions[0].Op != "FROM" {
		return fmt.Errorf("first instruction must be FROM")
	}

	name, tag, err := image.ParseNameTag(opts.NameTag)
	if err != nil {
		return err
	}

	tmpRoot, err := os.MkdirTemp("", "docksmith-build-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpRoot)

	var manifest image.Manifest
	manifest.Name = name
	manifest.Tag = tag
	manifest.Created = util.UTCNowISO()

	envMap := map[string]string{}
	workdir := ""
	var lastLayerDigest string
	baseNameTag := strings.TrimSpace(instructions[0].Args)

	fmt.Printf("Step 1/%d : %s\n", len(instructions), instructions[0].Raw)
	base, err := image.BaseManifest(baseNameTag)
	if err != nil {
		return err
	}
	for _, l := range base.Layers {
		if err := util.UntarInto(image.LayerPath(l.Digest), tmpRoot); err != nil {
			return err
		}
	}
	manifest.Layers = append(manifest.Layers, base.Layers...)
	manifest.Config = base.Config
	workdir = manifest.Config.WorkingDir
	for _, kv := range manifest.Config.Env {
		parts := strings.SplitN(kv, "=", 2)
		if len(parts) == 2 {
			envMap[parts[0]] = parts[1]
		}
	}
	if len(manifest.Layers) > 0 {
		lastLayerDigest = manifest.Layers[len(manifest.Layers)-1].Digest
	} else {
		lastLayerDigest = base.Digest
	}

	cascadeMiss := false

	for i, instr := range instructions[1:] {
		stepNo := i + 2
		start := time.Now()

		switch instr.Op {
		case "WORKDIR":
			workdir = strings.TrimSpace(instr.Args)
			manifest.Config.WorkingDir = workdir
			fmt.Printf("Step %d/%d : %s\n", stepNo, len(instructions), instr.Raw)

		case "ENV":
			fmt.Printf("Step %d/%d : %s\n", stepNo, len(instructions), instr.Raw)
			parts := strings.SplitN(instr.Args, "=", 2)
			if len(parts) != 2 {
				return fmt.Errorf("invalid ENV at line %d", instr.LineNum)
			}
			envMap[parts[0]] = parts[1]
			manifest.Config.Env = sortedEnv(envMap)

		case "CMD":
			fmt.Printf("Step %d/%d : %s\n", stepNo, len(instructions), instr.Raw)
			var cmd []string
			if err := json.Unmarshal([]byte(instr.Args), &cmd); err != nil {
				return fmt.Errorf("invalid CMD JSON at line %d: %w", instr.LineNum, err)
			}
			manifest.Config.Cmd = cmd

		case "COPY":
			src, dest, err := parseCopyArgs(instr.Args)
			if err != nil {
				return fmt.Errorf("line %d: %w", instr.LineNum, err)
			}
			sources, err := resolveCopySources(opts.Context, src)
			if err != nil {
				return err
			}

			var srcHashes []string
			var changed []string
			cacheKey := ""

			if !opts.NoCache && !cascadeMiss {
				for _, s := range sources {
					abs := filepath.Join(opts.Context, s)
					info, err := os.Stat(abs)
					if err != nil {
						return err
					}
					if info.IsDir() {
						err = filepath.Walk(abs, func(path string, fi os.FileInfo, err error) error {
							if err != nil || fi.IsDir() {
								return err
							}
							rel, _ := filepath.Rel(opts.Context, path)
							h, err := hashFile(path)
							if err != nil {
								return err
							}
							srcHashes = append(srcHashes, rel+":"+h)
							return nil
						})
						if err != nil {
							return err
						}
					} else {
						h, err := hashFile(abs)
						if err != nil {
							return err
						}
						srcHashes = append(srcHashes, s+":"+h)
					}
				}
				cacheKey = BuildCacheKey(lastLayerDigest, instr, workdir, envMap, srcHashes)
				if ce, ok := LoadCache(cacheKey); ok {
					fmt.Printf("Step %d/%d : %s [CACHE HIT] %.2fs\n", stepNo, len(instructions), instr.Raw, time.Since(start).Seconds())
					if err := util.UntarInto(image.LayerPath(ce.LayerDigest), tmpRoot); err != nil {
						return err
					}
					info, err := os.Stat(image.LayerPath(ce.LayerDigest))
					if err != nil {
						return err
					}
					manifest.Layers = append(manifest.Layers, image.LayerEntry{
						Digest:    ce.LayerDigest,
						Size:      info.Size(),
						CreatedBy: instr.Raw,
					})
					lastLayerDigest = ce.LayerDigest
					continue
				}
			}

			fmt.Printf("Step %d/%d : %s [CACHE MISS] ", stepNo, len(instructions), instr.Raw)
			changed, srcHashes, err = copyIntoRoot(opts.Context, sources, dest, tmpRoot)
			if err != nil {
				return err
			}
			layerDigest, size, err := writeLayerFromChanged(tmpRoot, changed)
			if err != nil {
				return err
			}
			if !opts.NoCache && !cascadeMiss {
				if cacheKey == "" {
					cacheKey = BuildCacheKey(lastLayerDigest, instr, workdir, envMap, srcHashes)
				}
				if err := SaveCache(cacheKey, layerDigest); err != nil {
					return err
				}
			}
			manifest.Layers = append(manifest.Layers, image.LayerEntry{
				Digest:    layerDigest,
				Size:      size,
				CreatedBy: instr.Raw,
			})
			lastLayerDigest = layerDigest
			cascadeMiss = true
			fmt.Printf("%.2fs\n", time.Since(start).Seconds())

		case "RUN":
			cacheKey := ""
			if !opts.NoCache && !cascadeMiss {
				cacheKey = BuildCacheKey(lastLayerDigest, instr, workdir, envMap, nil)
				if ce, ok := LoadCache(cacheKey); ok {
					fmt.Printf("Step %d/%d : %s [CACHE HIT] %.2fs\n", stepNo, len(instructions), instr.Raw, time.Since(start).Seconds())
					if err := util.UntarInto(image.LayerPath(ce.LayerDigest), tmpRoot); err != nil {
						return err
					}
					info, err := os.Stat(image.LayerPath(ce.LayerDigest))
					if err != nil {
						return err
					}
					manifest.Layers = append(manifest.Layers, image.LayerEntry{
						Digest:    ce.LayerDigest,
						Size:      info.Size(),
						CreatedBy: instr.Raw,
					})
					lastLayerDigest = ce.LayerDigest
					continue
				}
			}

			fmt.Printf("Step %d/%d : %s [CACHE MISS] ", stepNo, len(instructions), instr.Raw)
			beforeSnap, err := snapshot(tmpRoot)
			if err != nil {
				return err
			}
			if err := runInRoot(tmpRoot, workdir, envMap, instr.Args); err != nil {
				return err
			}
			changed, err := diffSnapshot(tmpRoot, beforeSnap)
			if err != nil {
				return err
			}
			layerDigest, size, err := writeLayerFromChanged(tmpRoot, changed)
			if err != nil {
				return err
			}
			if !opts.NoCache && !cascadeMiss {
				if err := SaveCache(cacheKey, layerDigest); err != nil {
					return err
				}
			}
			manifest.Layers = append(manifest.Layers, image.LayerEntry{
				Digest:    layerDigest,
				Size:      size,
				CreatedBy: instr.Raw,
			})
			lastLayerDigest = layerDigest
			cascadeMiss = true
			fmt.Printf("%.2fs\n", time.Since(start).Seconds())
		}
	}

	if err := image.SaveManifest(manifest); err != nil {
		return err
	}
	fmt.Printf("Successfully built %s %s:%s\n", manifest.Digest, manifest.Name, manifest.Tag)
	return nil
}

func sortedEnv(envMap map[string]string) []string {
	var out []string
	for k, v := range envMap {
		out = append(out, fmt.Sprintf("%s=%s", k, v))
	}
	// stable enough if you want strict lexical sorting, sort.Strings(out)
	return out
}

func writeLayerFromChanged(rootfs string, changed []string) (string, int64, error) {
	tmp, err := os.CreateTemp("", "docksmith-layer-*.tar")
	if err != nil {
		return "", 0, err
	}
	tmp.Close()
	defer os.Remove(tmp.Name())

	if err := util.TarPaths(rootfs, changed, tmp.Name()); err != nil {
		return "", 0, err
	}
	data, err := os.ReadFile(tmp.Name())
	if err != nil {
		return "", 0, err
	}
	digest := util.HashBytes(data)
	layerPath := image.LayerPath(digest)
	if _, err := os.Stat(layerPath); os.IsNotExist(err) {
		if err := os.WriteFile(layerPath, data, 0644); err != nil {
			return "", 0, err
		}
	}
	info, err := os.Stat(layerPath)
	if err != nil {
		return "", 0, err
	}
	return digest, info.Size(), nil
}

func snapshot(root string) (map[string]string, error) {
	out := map[string]string{}
	files, err := util.SortedFiles(root)
	if err != nil {
		return nil, err
	}
	for _, f := range files {
		info, err := os.Stat(f)
		if err != nil {
			return nil, err
		}
		rel, _ := filepath.Rel(root, f)
		if info.IsDir() {
			out[rel] = "DIR"
			continue
		}
		h, err := hashFile(f)
		if err != nil {
			return nil, err
		}
		out[rel] = h
	}
	return out, nil
}

func diffSnapshot(root string, before map[string]string) ([]string, error) {
	after, err := snapshot(root)
	if err != nil {
		return nil, err
	}
	var changed []string
	for path, hash := range after {
		if old, ok := before[path]; !ok || old != hash {
			changed = append(changed, path)
		}
	}
	return changed, nil
}

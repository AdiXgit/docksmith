package build

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"docksmith/internal/image"
	"docksmith/internal/util"
)

type CacheEntry struct {
	LayerDigest string `json:"layerDigest"`
}

func cachePath(key string) string {
	name := strings.TrimPrefix(key, "sha256:")
	return filepath.Join(image.DocksmithHome(), "cache", name+".json")
}

func LoadCache(key string) (CacheEntry, bool) {
	data, err := os.ReadFile(cachePath(key))
	if err != nil {
		return CacheEntry{}, false
	}
	var e CacheEntry
	if err := json.Unmarshal(data, &e); err != nil {
		return CacheEntry{}, false
	}
	if _, err := os.Stat(image.LayerPath(e.LayerDigest)); err != nil {
		return CacheEntry{}, false
	}
	return e, true
}

func SaveCache(key, layerDigest string) error {
	b, err := json.Marshal(CacheEntry{LayerDigest: layerDigest})
	if err != nil {
		return err
	}
	return os.WriteFile(cachePath(key), b, 0644)
}

func BuildCacheKey(prevDigest string, instr Instruction, workdir string, env map[string]string, copySourceHashes []string) string {
	var envKeys []string
	for k := range env {
		envKeys = append(envKeys, k)
	}
	sort.Strings(envKeys)

	var envParts []string
	for _, k := range envKeys {
		envParts = append(envParts, k+"="+env[k])
	}

	payload := struct {
		PrevDigest       string
		InstructionText  string
		Workdir          string
		Env              []string
		CopySourceHashes []string
	}{
		PrevDigest:       prevDigest,
		InstructionText:  instr.Raw,
		Workdir:          workdir,
		Env:              envParts,
		CopySourceHashes: copySourceHashes,
	}

	b, _ := json.Marshal(payload)
	return util.HashBytes(b)
}

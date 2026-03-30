package image

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"docksmith/internal/util"
)

func DocksmithHome() string {
	home, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	return filepath.Join(home, ".docksmith")
}

func EnsureStore() error {
	dirs := []string{
		filepath.Join(DocksmithHome(), "images"),
		filepath.Join(DocksmithHome(), "layers"),
		filepath.Join(DocksmithHome(), "cache"),
	}
	for _, d := range dirs {
		if err := os.MkdirAll(d, 0755); err != nil {
			return err
		}
	}
	return nil
}

func ParseNameTag(nameTag string) (string, string, error) {
	parts := strings.Split(nameTag, ":")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("invalid name:tag %q", nameTag)
	}
	return parts[0], parts[1], nil
}

func ImagePath(name, tag string) string {
	return filepath.Join(DocksmithHome(), "images", fmt.Sprintf("%s_%s.json", name, tag))
}

func LayerPath(digest string) string {
	name := strings.TrimPrefix(digest, "sha256:")
	return filepath.Join(DocksmithHome(), "layers", name+".tar")
}

func SaveManifest(m Manifest) error {
	tmp := m
	tmp.Digest = ""
	canonical, err := json.Marshal(tmp)
	if err != nil {
		return err
	}
	m.Digest = util.HashBytes(canonical)
	finalBytes, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(ImagePath(m.Name, m.Tag), finalBytes, 0644)
}

func LoadManifest(nameTag string) (Manifest, error) {
	name, tag, err := ParseNameTag(nameTag)
	if err != nil {
		return Manifest{}, err
	}
	data, err := os.ReadFile(ImagePath(name, tag))
	if err != nil {
		return Manifest{}, err
	}
	var m Manifest
	if err := json.Unmarshal(data, &m); err != nil {
		return Manifest{}, err
	}
	return m, nil
}

func ListImages() error {
	if err := EnsureStore(); err != nil {
		return err
	}
	dir := filepath.Join(DocksmithHome(), "images")
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	type row struct {
		Name, Tag, ID, Created string
	}
	var rows []row
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			return err
		}
		var m Manifest
		if err := json.Unmarshal(data, &m); err != nil {
			return err
		}
		id := strings.TrimPrefix(m.Digest, "sha256:")
		if len(id) > 12 {
			id = id[:12]
		}
		rows = append(rows, row{m.Name, m.Tag, id, m.Created})
	}
	sort.Slice(rows, func(i, j int) bool {
		if rows[i].Name == rows[j].Name {
			return rows[i].Tag < rows[j].Tag
		}
		return rows[i].Name < rows[j].Name
	})
	fmt.Printf("%-20s %-12s %-14s %-25s\n", "NAME", "TAG", "ID", "CREATED")
	for _, r := range rows {
		fmt.Printf("%-20s %-12s %-14s %-25s\n", r.Name, r.Tag, r.ID, r.Created)
	}
	return nil
}

func RemoveImage(nameTag string) error {
	m, err := LoadManifest(nameTag)
	if err != nil {
		return err
	}
	if err := os.Remove(ImagePath(m.Name, m.Tag)); err != nil {
		return err
	}
	for _, layer := range m.Layers {
		_ = os.Remove(LayerPath(layer.Digest))
	}
	return nil
}

func BaseManifest(nameTag string) (Manifest, error) {
	m, err := LoadManifest(nameTag)
	if err != nil {
		return Manifest{}, errors.New("base image not found in local store")
	}
	return m, nil
}

package util

import (
	"archive/tar"
	"io"
	"os"
	"path/filepath"
	"sort"
	"time"
)

func TarPaths(baseDir string, relPaths []string, tarPath string) error {
	sort.Strings(relPaths)

	f, err := os.Create(tarPath)
	if err != nil {
		return err
	}
	defer f.Close()

	tw := tar.NewWriter(f)
	defer tw.Close()

	zeroTime := time.Unix(0, 0)

	for _, rel := range relPaths {
		full := filepath.Join(baseDir, rel)
		info, err := os.Lstat(full)
		if err != nil {
			return err
		}

		hdr, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return err
		}
		hdr.Name = rel
		hdr.ModTime = zeroTime
		hdr.AccessTime = zeroTime
		hdr.ChangeTime = zeroTime

		if err := tw.WriteHeader(hdr); err != nil {
			return err
		}

		if info.Mode().IsRegular() {
			in, err := os.Open(full)
			if err != nil {
				return err
			}
			if _, err := io.Copy(tw, in); err != nil {
				in.Close()
				return err
			}
			in.Close()
		}
	}
	return nil
}

func UntarInto(tarPath, dest string) error {
	f, err := os.Open(tarPath)
	if err != nil {
		return err
	}
	defer f.Close()

	tr := tar.NewReader(f)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		target := filepath.Join(dest, hdr.Name)
		switch hdr.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0755); err != nil {
				return err
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return err
			}
			out, err := os.OpenFile(target, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, os.FileMode(hdr.Mode))
			if err != nil {
				return err
			}
			if _, err := io.Copy(out, tr); err != nil {
				out.Close()
				return err
			}
			out.Close()
		}
	}
}


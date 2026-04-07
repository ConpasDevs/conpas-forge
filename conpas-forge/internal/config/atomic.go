package config

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func AtomicWrite(path string, data []byte, perm os.FileMode) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("create parent directory: %w", err)
	}

	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, perm); err != nil {
		os.Remove(tmp)
		return fmt.Errorf("write temp file: %w", err)
	}

	if err := os.Rename(tmp, path); err != nil {
		if err2 := copyFile(tmp, path, perm); err2 != nil {
			os.Remove(tmp)
			return fmt.Errorf("rename failed (%w) and copy fallback failed: %w", err, err2)
		}
		os.Remove(tmp)
	}

	return nil
}

func copyFile(src, dst string, perm os.FileMode) error {
	srcF, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcF.Close()

	dstF, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, perm)
	if err != nil {
		return err
	}
	defer dstF.Close()

	_, err = io.Copy(dstF, srcF)
	return err
}

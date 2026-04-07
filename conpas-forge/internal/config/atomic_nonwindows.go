//go:build !windows

package config

import "os"

func replaceFileAtomic(tmp, path string) error {
	return os.Rename(tmp, path)
}

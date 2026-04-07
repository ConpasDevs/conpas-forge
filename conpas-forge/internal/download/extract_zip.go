package download

import (
	"archive/zip"
	"fmt"
	"io"
	"path/filepath"
	"strings"
)

func ExtractFileFromZip(archivePath, targetName string) ([]byte, error) {
	r, err := zip.OpenReader(archivePath)
	if err != nil {
		return nil, fmt.Errorf("open zip archive: %w", err)
	}
	defer r.Close()

	for _, f := range r.File {
		base := filepath.Base(f.Name)
		if strings.Contains(f.Name, "..") || base != targetName {
			continue
		}

		rc, err := f.Open()
		if err != nil {
			return nil, fmt.Errorf("open zip entry %q: %w", f.Name, err)
		}
		defer rc.Close()

		data, err := io.ReadAll(rc)
		if err != nil {
			return nil, fmt.Errorf("read zip entry %q: %w", f.Name, err)
		}
		return data, nil
	}

	return nil, fmt.Errorf("file %q not found in zip archive", targetName)
}

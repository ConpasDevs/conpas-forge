package download

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func ExtractFileFromTarGz(archivePath, targetName string) ([]byte, error) {
	f, err := os.Open(archivePath)
	if err != nil {
		return nil, fmt.Errorf("open archive: %w", err)
	}
	defer f.Close()

	gz, err := gzip.NewReader(f)
	if err != nil {
		return nil, fmt.Errorf("create gzip reader: %w", err)
	}
	defer gz.Close()

	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("read tar header: %w", err)
		}

		base := filepath.Base(hdr.Name)
		if strings.Contains(hdr.Name, "..") || base != targetName {
			continue
		}

		data, err := io.ReadAll(tr)
		if err != nil {
			return nil, fmt.Errorf("read tar entry %q: %w", hdr.Name, err)
		}
		return data, nil
	}

	return nil, fmt.Errorf("file %q not found in tar.gz archive", targetName)
}

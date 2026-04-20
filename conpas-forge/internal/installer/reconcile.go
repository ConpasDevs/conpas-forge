package installer

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/conpasDEVS/conpas-forge/internal/config"
)

// ReconcileOutputStyles ensures that dir contains exactly one .md file — the active
// output-style file for the configured persona. All other .md files are removed.
// Non-.md files are left untouched. Returns the list of removed basenames.
func ReconcileOutputStyles(dir, filename string, content []byte) (removed []string, err error) {
	if filename == "" {
		return nil, fmt.Errorf("output-style filename must not be empty")
	}

	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("create output-styles dir: %w", err)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("read output-styles dir: %w", err)
	}

	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if !strings.HasSuffix(name, ".md") {
			continue
		}
		if name == filename {
			continue
		}
		if err := os.Remove(filepath.Join(dir, name)); err != nil {
			return nil, fmt.Errorf("remove orphan output-style %q: %w", name, err)
		}
		removed = append(removed, name)
	}

	dest := filepath.Join(dir, filename)
	if err := config.AtomicWrite(dest, content, 0644); err != nil {
		return removed, fmt.Errorf("write output-style %q: %w", filename, err)
	}

	return removed, nil
}

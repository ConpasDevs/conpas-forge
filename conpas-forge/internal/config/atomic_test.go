package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestAtomicWriteReplacesExistingFile(t *testing.T) {
	tests := []struct {
		name string
		old  string
		new  string
	}{
		{name: "replace existing file contents", old: "before", new: "after"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "config.yaml")
			if err := os.WriteFile(path, []byte(tt.old), 0644); err != nil {
				t.Fatalf("seed file: %v", err)
			}

			if err := AtomicWrite(path, []byte(tt.new), 0644); err != nil {
				t.Fatalf("AtomicWrite() error = %v", err)
			}

			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("read file: %v", err)
			}
			if string(data) != tt.new {
				t.Fatalf("contents = %q, want %q", string(data), tt.new)
			}
		})
	}
}

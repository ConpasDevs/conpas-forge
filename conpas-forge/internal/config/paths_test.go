package config

import (
	"path/filepath"
	"testing"
)

func TestClaudeMCPPaths(t *testing.T) {
	tests := []struct {
		name    string
		fn      func() string
		wantSfx string // suffix relative to home
	}{
		{
			name:    "ClaudeMCPDir",
			fn:      ClaudeMCPDir,
			wantSfx: filepath.Join(".claude", "mcp"),
		},
		{
			name:    "EngramMCPFile",
			fn:      EngramMCPFile,
			wantSfx: filepath.Join(".claude", "mcp", "engram.json"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			home := t.TempDir()
			old := HomeDir()
			OverrideHomeDir(home)
			defer OverrideHomeDir(old)

			want := filepath.Join(home, tt.wantSfx)
			if got := tt.fn(); got != want {
				t.Fatalf("%s() = %q, want %q", tt.name, got, want)
			}
		})
	}
}

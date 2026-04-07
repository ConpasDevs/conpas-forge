package installer

import (
	"os"
	"strings"
	"testing"

	"github.com/conpasDEVS/conpas-forge/internal/config"
)

func TestMergeRefreshesBackupForLatestRestorePoint(t *testing.T) {
	homeDir := t.TempDir()
	oldHomeDir := config.HomeDir()
	config.OverrideHomeDir(homeDir)
	defer config.OverrideHomeDir(oldHomeDir)

	tests := []struct {
		name           string
		initial        string
		manuallyEdited string
		wantBackup     string
	}{
		{
			name:           "backup tracks latest pre-merge state",
			initial:        "{\n  \"theme\": \"dark\"\n}\n",
			manuallyEdited: "{\n  \"theme\": \"light\"\n}\n",
			wantBackup:     "\"theme\": \"light\"",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := config.AtomicWrite(config.SettingsJSON(), []byte(tt.initial), 0644); err != nil {
				t.Fatalf("seed settings: %v", err)
			}
			if err := Merge(map[string]any{"mcpServers": map[string]any{"engram": map[string]any{"type": "stdio"}}}); err != nil {
				t.Fatalf("first Merge(): %v", err)
			}

			if err := config.AtomicWrite(config.SettingsJSON(), []byte(tt.manuallyEdited), 0644); err != nil {
				t.Fatalf("manual edit settings: %v", err)
			}
			if err := Merge(map[string]any{"telemetry": map[string]any{"enabled": true}}); err != nil {
				t.Fatalf("second Merge(): %v", err)
			}

			backup, err := os.ReadFile(config.SettingsJSONBak())
			if err != nil {
				t.Fatalf("read backup: %v", err)
			}
			if !strings.Contains(string(backup), tt.wantBackup) {
				t.Fatalf("backup = %q, want substring %q", string(backup), tt.wantBackup)
			}

			if err := config.AtomicWrite(config.SettingsJSON(), []byte("{\n  \"theme\": \"broken\"\n}\n"), 0644); err != nil {
				t.Fatalf("overwrite settings before restore: %v", err)
			}
			if err := Restore(); err != nil {
				t.Fatalf("Restore() error = %v", err)
			}

			restored, err := os.ReadFile(config.SettingsJSON())
			if err != nil {
				t.Fatalf("read restored settings: %v", err)
			}
			if !strings.Contains(string(restored), tt.wantBackup) {
				t.Fatalf("restored settings = %q, want substring %q", string(restored), tt.wantBackup)
			}
		})
	}
}

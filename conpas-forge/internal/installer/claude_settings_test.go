package installer

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/conpasDEVS/conpas-forge/internal/config"
)

func TestClaudeSettingsInstallerWritesBypassMode(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "bypass mode written to settings.json"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			homeDir := t.TempDir()
			oldHomeDir := config.HomeDir()
			config.OverrideHomeDir(homeDir)
			defer config.OverrideHomeDir(oldHomeDir)

			cfg := config.DefaultConfig()
			opts := &InstallOptions{Config: &cfg}

			inst := NewClaudeSettingsInstaller()
			result := inst.Install(context.Background(), opts, nil)

			if !result.Success {
				t.Fatalf("expected success, got err: %v", result.Err)
			}
			if result.ModuleName != "Claude Code Settings" {
				t.Fatalf("module name = %q, want %q", result.ModuleName, "Claude Code Settings")
			}

			data, err := os.ReadFile(config.SettingsJSON())
			if err != nil {
				t.Fatalf("read settings.json: %v", err)
			}
			var parsed map[string]any
			if err := json.Unmarshal(data, &parsed); err != nil {
				t.Fatalf("parse settings.json: %v", err)
			}
			val, ok := parsed["bypassPermissionsModeAccepted"]
			if !ok {
				t.Fatal("bypassPermissionsModeAccepted not found in settings.json")
			}
			if val != true {
				t.Fatalf("bypassPermissionsModeAccepted = %v, want true", val)
			}

			perms, ok := parsed["permissions"].(map[string]any)
			if !ok {
				t.Fatal("permissions block not found in settings.json")
			}
			if perms["defaultMode"] != "bypassPermissions" {
				t.Fatalf("permissions.defaultMode = %v, want bypassPermissions", perms["defaultMode"])
			}
		})
	}
}

func TestBuildModulesAlwaysIncludesClaudeSettings(t *testing.T) {
	cfg := config.DefaultConfig()
	modules := BuildModules([]string{}, &cfg)

	found := false
	for _, m := range modules {
		if m.Name() == "Claude Code Settings" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("ClaudeSettingsInstaller must always be present in BuildModules")
	}
}

package installer

import (
	"context"
	"testing"

	"github.com/conpasDEVS/conpas-forge/internal/config"
)

func TestClaudeSettingsInstallerSucceeds(t *testing.T) {
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

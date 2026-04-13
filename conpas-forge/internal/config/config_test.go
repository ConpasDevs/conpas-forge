package config

import (
	"os"
	"strings"
	"testing"
)

func TestSaveRefreshesBackupBeforeWrite(t *testing.T) {
	homeDir := t.TempDir()
	oldHomeDir := HomeDir()
	OverrideHomeDir(homeDir)
	defer OverrideHomeDir(oldHomeDir)

	if err := os.MkdirAll(ForgeDir(), 0755); err != nil {
		t.Fatalf("mkdir forge dir: %v", err)
	}

	// Seed an initial config.
	initial := DefaultConfig()
	initial.Persona = "asturiano"
	if err := Save(&initial); err != nil {
		t.Fatalf("first Save(): %v", err)
	}

	// Manually overwrite config so the backup will differ from the next save.
	manualContent := "version: 1\npersona: yoda\n"
	if err := AtomicWrite(ConfigPath(), []byte(manualContent), 0644); err != nil {
		t.Fatalf("manual overwrite: %v", err)
	}

	// Second save — backup must capture the manually-written state.
	updated := DefaultConfig()
	updated.Persona = "jedi"
	if err := Save(&updated); err != nil {
		t.Fatalf("second Save(): %v", err)
	}

	backup, err := os.ReadFile(ConfigBak())
	if err != nil {
		t.Fatalf("read backup: %v", err)
	}
	if !strings.Contains(string(backup), "yoda") {
		t.Fatalf("backup = %q, want substring %q", string(backup), "yoda")
	}
}

func TestLoadReturnsDescriptiveErrorOnCorruptYAML(t *testing.T) {
	homeDir := t.TempDir()
	oldHomeDir := HomeDir()
	OverrideHomeDir(homeDir)
	defer OverrideHomeDir(oldHomeDir)

	if err := os.MkdirAll(ForgeDir(), 0755); err != nil {
		t.Fatalf("mkdir forge dir: %v", err)
	}

	if err := AtomicWrite(ConfigPath(), []byte(":\tinvalid: yaml: [\n"), 0644); err != nil {
		t.Fatalf("write corrupt config: %v", err)
	}

	_, err := Load()
	if err == nil {
		t.Fatal("Load() expected error on corrupt YAML, got nil")
	}
	if !strings.Contains(err.Error(), "corrupted") {
		t.Fatalf("error = %q, want substring %q", err.Error(), "corrupted")
	}
	if !strings.Contains(err.Error(), ConfigPath()) {
		t.Fatalf("error = %q, want config path in message", err.Error())
	}
}

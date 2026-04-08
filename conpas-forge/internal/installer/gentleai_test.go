package installer

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/conpasDEVS/conpas-forge/internal/config"
)

func TestGentleAISkillCount(t *testing.T) {
	tests := []struct {
		name string
		want int
	}{
		{name: "matches embedded skill list", want: 20},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GentleAISkillCount(); got != tt.want {
				t.Fatalf("GentleAISkillCount() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestCountGentleAISkillsDeployed(t *testing.T) {
	homeDir := t.TempDir()
	oldHomeDir := config.HomeDir()
	config.OverrideHomeDir(homeDir)
	defer config.OverrideHomeDir(oldHomeDir)

	tests := []struct {
		name  string
		paths []string
		want  int
	}{
		{
			name: "counts only deployed skill files",
			paths: []string{
				config.ClaudeMD(),
				filepath.Join(config.SkillDir("sdd-init"), "SKILL.md"),
				filepath.Join(config.SkillDir("engram-memory"), "SKILL.md"),
				filepath.Join(config.SharedSkillsDir(), "engram-convention.md"),
			},
			want: 2,
		},
		{
			name: "ignores nested non-skill paths",
			paths: []string{
				filepath.Join(config.SkillsDir(), "custom", "nested", "SKILL.md"),
				filepath.Join(config.SkillDir("go-testing"), "notes.txt"),
			},
			want: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CountGentleAISkillsDeployed(tt.paths); got != tt.want {
				t.Fatalf("CountGentleAISkillsDeployed() = %d, want %d; paths=%v", got, tt.want, tt.paths)
			}
		})
	}
}

func TestCountGentleAISkillsDeployedMatchesSourceList(t *testing.T) {
	homeDir := t.TempDir()
	oldHomeDir := config.HomeDir()
	config.OverrideHomeDir(homeDir)
	defer config.OverrideHomeDir(oldHomeDir)

	paths := make([]string, 0, GentleAISkillCount())
	for _, skill := range sddSkills {
		paths = append(paths, filepath.Join(config.SkillDir(skill), "SKILL.md"))
	}

	if got := CountGentleAISkillsDeployed(paths); got != GentleAISkillCount() {
		t.Fatalf("deployed skill count = %d, want %d", got, GentleAISkillCount())
	}

	if got := fmt.Sprintf("%d skills", CountGentleAISkillsDeployed(paths)); got != "20 skills" {
		t.Fatalf("formatted deployed skill count = %q, want %q", got, "20 skills")
	}
}

func TestGentleAIInstallerFailsOnPartialDeployment(t *testing.T) {
	homeDir := t.TempDir()
	oldHomeDir := config.HomeDir()
	config.OverrideHomeDir(homeDir)
	defer config.OverrideHomeDir(oldHomeDir)

	tests := []struct {
		name string
	}{
		{name: "write failures return installer error"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := os.MkdirAll(config.ClaudeDir(), 0755); err != nil {
				t.Fatalf("mkdir claude dir: %v", err)
			}
			if err := os.WriteFile(config.SkillsDir(), []byte("blocked"), 0644); err != nil {
				t.Fatalf("block skills dir: %v", err)
			}

			cfg := config.DefaultConfig()
			result := NewGentleAIInstaller().Install(context.Background(), &InstallOptions{Config: &cfg}, nil)
			if result.Err == nil {
				t.Fatal("expected installer error, got nil")
			}
			if result.Success {
				t.Fatal("expected installer success=false on partial deployment")
			}
		})
	}
}

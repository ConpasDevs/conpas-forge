package tui

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/conpasDEVS/conpas-forge/internal/config"
	"github.com/conpasDEVS/conpas-forge/internal/installer"
)

func TestNewModulesModelUsesRealGentleAISkillCount(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{name: "gentle ai description reflects source skill count", want: "19 skills + CLAUDE.md + output styles"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := NewModulesModel(&config.Config{})
			if got := model.choices[1].Description; got != tt.want {
				t.Fatalf("description = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestCtrlCCancelsInstallPipeline(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "ctrl+c during install marks model cancelled and cancels context"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := config.DefaultConfig()
			ctx, cancel := context.WithCancel(context.Background())
			model := New(&cfg, installer.Platform{}, t.TempDir())
			model.screen = ScreenInstall
			model.installing = true
			model.installCtx = ctx
			model.cancelFn = cancel

			updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
			got := updated.(Model)

			if !got.Cancelled() {
				t.Fatal("expected model to be cancelled")
			}
			if ctx.Err() == nil {
				t.Fatal("expected install context to be cancelled")
			}
		})
	}
}

func TestInstallDoneSurfacesConfigSaveFailure(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "config save error becomes failed result"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			oldHomeDir := config.HomeDir()
			config.OverrideHomeDir(tempDir)
			defer config.OverrideHomeDir(oldHomeDir)

			blockingPath := filepath.Join(tempDir, ".conpas-forge")
			if err := config.AtomicWrite(blockingPath, []byte("not a directory"), 0644); err != nil {
				t.Fatalf("setup blocking path: %v", err)
			}

			cfg := config.DefaultConfig()
			model := New(&cfg, installer.Platform{}, tempDir)

			writtenSkills := []string{
				config.ClaudeMD(),
				filepath.Join(config.SkillDir("sdd-init"), "SKILL.md"),
				filepath.Join(config.SkillDir("engram-memory"), "SKILL.md"),
			}

			updated, _ := model.Update(InstallDoneMsg{Results: []installer.Result{{
				ModuleName:   "Gentle AI",
				Success:      true,
				PathsWritten: writtenSkills,
			}}})

			got := updated.(Model)
			if len(got.results) != 2 {
				t.Fatalf("results length = %d, want 2", len(got.results))
			}

			configResult := got.results[1]
			if configResult.ModuleName != "Config" {
				t.Fatalf("config result module = %q, want %q", configResult.ModuleName, "Config")
			}
			if configResult.Err == nil {
				t.Fatal("expected config save error, got nil")
			}
			if !strings.Contains(configResult.Err.Error(), "save config") {
				t.Fatalf("config save error = %q, want to contain %q", configResult.Err.Error(), "save config")
			}
			if got.cfg.Modules.GentleAI.SkillsDeployed != 2 {
				t.Fatalf("skills deployed = %d, want 2", got.cfg.Modules.GentleAI.SkillsDeployed)
			}
			if !got.cfg.Modules.GentleAI.Installed {
				t.Fatal("expected Gentle AI to remain marked installed from module result")
			}
			if !installer.HasErrors(got.results) {
				t.Fatal("expected install results to be marked as failed")
			}
		})
	}
}

package installer

import (
	"context"
	"fmt"

	"github.com/conpasDEVS/conpas-forge/internal/config"
)

// ClaudeSettingsInstaller writes Claude Code global settings to settings.json.
// Currently handles: bypassPermissionsModeAccepted.
type ClaudeSettingsInstaller struct{}

func NewClaudeSettingsInstaller() *ClaudeSettingsInstaller { return &ClaudeSettingsInstaller{} }

func (c *ClaudeSettingsInstaller) Name() string { return "Claude Code Settings" }

func (c *ClaudeSettingsInstaller) Install(ctx context.Context, opts *InstallOptions, progress func(ProgressEvent)) Result {
	result := Result{ModuleName: "Claude Code Settings"}
	emit := func(status, detail string) {
		if progress != nil {
			progress(ProgressEvent{Module: "Claude Code Settings", Status: status, Detail: detail, Percent: -1})
		}
	}

	emit("writing", "Enabling bypass permissions mode...")
	if err := Merge(map[string]any{
		"bypassPermissionsModeAccepted": true,
		"permissions": map[string]any{
			"defaultMode": "bypassPermissions",
		},
	}); err != nil {
		result.Err = fmt.Errorf("enable bypass mode: %w", err)
		return result
	}
	result.PathsWritten = append(result.PathsWritten, config.SettingsJSON())
	emit("done", "bypass permissions mode enabled")
	result.Success = true
	return result
}

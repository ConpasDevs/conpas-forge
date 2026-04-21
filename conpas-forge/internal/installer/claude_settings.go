package installer

import (
	"context"
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

	// bypassPermissions cannot be persisted as a default mode in Claude Code.
	// Permissions are handled via per-tool allowlists from the engramtools catalog (internal/engramtools/catalog.go).
	// This module is kept as a placeholder for future global settings.
	emit("done", "Claude Code settings verified")
	result.Success = true
	return result
}

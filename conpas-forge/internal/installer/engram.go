package installer

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/conpasDEVS/conpas-forge/internal/config"
	"github.com/conpasDEVS/conpas-forge/internal/download"
)

// Engram GitHub repository — confirmed from exploration
const engramOwner = "Gentleman-Programming"
const engramRepo = "engram"

type EngramInstaller struct {
	httpClient *http.Client
}

func NewEngramInstaller() *EngramInstaller {
	return &EngramInstaller{
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

func (e *EngramInstaller) Name() string { return "Engram" }

func (e *EngramInstaller) Install(ctx context.Context, opts *InstallOptions, progress func(ProgressEvent)) Result {
	result := Result{ModuleName: "Engram"}
	emit := func(status, detail string, pct int) {
		if progress != nil {
			progress(ProgressEvent{Module: "Engram", Status: status, Detail: detail, Percent: pct})
		}
	}

	// Step 1: Query GitHub releases API
	emit("downloading", "Querying GitHub releases...", -1)
	release, err := download.FetchLatestRelease(ctx, e.httpClient, engramOwner, engramRepo)
	if err != nil {
		result.Err = fmt.Errorf("fetch engram release: %w", err)
		return result
	}

	// Step 2: Select asset for current platform
	archiveAsset, checksumAsset, err := download.SelectAsset(release, opts.Platform.OS, opts.Platform.Arch)
	if err != nil {
		result.Err = fmt.Errorf("select engram asset: %w", err)
		return result
	}

	// Step 3: Download archive with progress
	emit("downloading", fmt.Sprintf("Downloading %s...", archiveAsset.Name), 0)
	tmpPath, err := download.DownloadToTemp(ctx, e.httpClient, archiveAsset.BrowserDownloadURL, func(read, total int64) {
		if total > 0 {
			emit("downloading", fmt.Sprintf("Downloading %s...", archiveAsset.Name), int(read*100/total))
		}
	})
	if err != nil {
		result.Err = fmt.Errorf("download engram: %w", err)
		return result
	}
	defer os.Remove(tmpPath)

	// Steps 4-5: Checksum verification
	if checksumAsset != nil {
		emit("verifying", "Verifying checksum...", -1)
		expectedHex, err := download.FetchChecksumHex(ctx, e.httpClient, checksumAsset, archiveAsset.Name)
		if err != nil {
			result.Err = fmt.Errorf("fetch checksum for %s: %w", archiveAsset.Name, err)
			return result
		} else {
			if err := download.VerifyChecksum(tmpPath, expectedHex); err != nil {
				result.Err = fmt.Errorf("checksum verification: %w", err)
				return result
			}
		}
	} else {
		result.Err = fmt.Errorf("no checksum asset found for %s", archiveAsset.Name)
		return result
	}

	// Step 6: Extract binary
	emit("extracting", "Extracting binary...", -1)
	binName := download.BinaryName(opts.Platform.OS)
	var binaryBytes []byte
	if opts.Platform.OS == "windows" {
		binaryBytes, err = download.ExtractFileFromZip(tmpPath, binName)
	} else {
		binaryBytes, err = download.ExtractFileFromTarGz(tmpPath, binName)
	}
	if err != nil {
		result.Err = fmt.Errorf("extract engram binary: %w", err)
		return result
	}

	// Step 7: Place binary
	emit("writing", "Installing engram binary...", -1)
	destPath := config.EngramBinary()
	if err := config.AtomicWrite(destPath, binaryBytes, 0755); err != nil {
		result.Err = fmt.Errorf("write engram binary: %w", err)
		return result
	}
	result.PathsWritten = append(result.PathsWritten, destPath)

	// Step 8: Verify binary
	stat, err := os.Stat(destPath)
	if err != nil || stat.Size() == 0 {
		result.Err = fmt.Errorf("extracted binary is empty or missing at %s", destPath)
		return result
	}

	// Step 9: PATH check
	if !CheckPathContains(config.BinDir()) {
		result.Warnings = append(result.Warnings, PathWarning(config.BinDir(), runtime.GOOS))
	}

	// Step 10: Register MCP server via `claude mcp add --scope user`
	emit("writing", "Registering Engram MCP server...", -1)
	if err := registerEngramMCP(ctx, destPath); err != nil {
		result.Err = fmt.Errorf("register engram MCP: %w", err)
		return result
	}
	result.PathsWritten = append(result.PathsWritten, config.ClaudeJSON())

	// Step 11: Allowlist Engram MCP tools (no marketplace — uses local binary via mcp file)
	emit("writing", "Allowlisting Engram tools...", -1)
	allowEntry := map[string]any{
		"permissions": map[string]any{
			"allow": engramMCPTools,
		},
	}
	if err := Merge(allowEntry); err != nil {
		result.Warnings = append(result.Warnings, fmt.Sprintf("permissions.allow update failed: %v", err))
	} else {
		result.PathsWritten = append(result.PathsWritten, config.SettingsJSON())
	}

	emit("done", fmt.Sprintf("Engram %s installed", release.TagName), 100)
	result.Success = true
	return result
}

// registerEngramMCP registers the Engram binary as a user-scoped MCP server
// via `claude mcp add --scope user engram -- <binary> mcp --tools=agent`.
// If engram is already registered, it removes the old entry first (idempotent re-install).
func registerEngramMCP(ctx context.Context, binaryPath string) error {
	claudePath, err := exec.LookPath("claude")
	if err != nil {
		return fmt.Errorf("'claude' not found in PATH — install Claude Code first: %w", err)
	}

	// Remove existing entry if present (idempotent re-install)
	removeCmd := exec.CommandContext(ctx, claudePath, "mcp", "remove", "engram", "--scope", "user")
	_ = removeCmd.Run() // ignore error — entry may not exist

	addCmd := exec.CommandContext(ctx, claudePath,
		"mcp", "add",
		"--scope", "user",
		"engram",
		"--",
		binaryPath, "mcp", "--tools=agent",
	)
	out, err := addCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("claude mcp add failed: %w\noutput: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}

// engramMCPTools lists the Engram tool names that must be allowlisted in
// ~/.claude/settings.json so Claude Code does not prompt for permission on each call.
// Engram is registered with --tools=agent which enables plugin mode in Claude Code.
// In plugin mode the tool prefix is mcp__plugin_<server>_<server>__ (not mcp__<server>__).
var engramMCPTools = []any{
	"mcp__plugin_engram_engram__mem_capture_passive",
	"mcp__plugin_engram_engram__mem_context",
	"mcp__plugin_engram_engram__mem_delete",
	"mcp__plugin_engram_engram__mem_get_observation",
	"mcp__plugin_engram_engram__mem_merge_projects",
	"mcp__plugin_engram_engram__mem_save",
	"mcp__plugin_engram_engram__mem_save_prompt",
	"mcp__plugin_engram_engram__mem_search",
	"mcp__plugin_engram_engram__mem_session_end",
	"mcp__plugin_engram_engram__mem_session_start",
	"mcp__plugin_engram_engram__mem_session_summary",
	"mcp__plugin_engram_engram__mem_stats",
	"mcp__plugin_engram_engram__mem_suggest_topic_key",
	"mcp__plugin_engram_engram__mem_timeline",
	"mcp__plugin_engram_engram__mem_update",
}

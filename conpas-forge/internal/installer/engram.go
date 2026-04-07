package installer

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"runtime"
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
			result.Warnings = append(result.Warnings, fmt.Sprintf("checksum fetch failed (skipping): %v", err))
		} else {
			if err := download.VerifyChecksum(tmpPath, expectedHex); err != nil {
				result.Err = fmt.Errorf("checksum verification: %w", err)
				return result
			}
		}
	} else {
		result.Warnings = append(result.Warnings, "no checksum asset found — skipping verification")
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

	// Step 10: Merge settings.json
	emit("writing", "Updating settings.json...", -1)
	mcpEntry := map[string]any{
		"mcpServers": map[string]any{
			"engram": map[string]any{
				"command": destPath,
				"type":    "stdio",
			},
		},
	}
	if err := Merge(mcpEntry); err != nil {
		result.Warnings = append(result.Warnings, fmt.Sprintf("settings.json update failed: %v", err))
	} else {
		result.PathsWritten = append(result.PathsWritten, config.SettingsJSON())
	}

	emit("done", fmt.Sprintf("Engram %s installed", release.TagName), 100)
	result.Success = true
	return result
}

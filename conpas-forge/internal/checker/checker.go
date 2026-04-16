package checker

import (
	"context"
	"net/http"

	"github.com/conpasDEVS/conpas-forge/internal/config"
	"github.com/conpasDEVS/conpas-forge/internal/download"
	"github.com/conpasDEVS/conpas-forge/internal/version"
	"golang.org/x/mod/semver"
)

const (
	StatusUpToDate     = "up-to-date"
	StatusOutdated     = "outdated"
	StatusNotInstalled = "not-installed"
	StatusUnknown      = "unknown"
)

// ModuleCheck holds the version comparison result for a single module.
type ModuleCheck struct {
	Module           string `json:"name"`
	InstalledVersion string `json:"installedVersion"`
	LatestVersion    string `json:"latestVersion"`
	Status           string `json:"status"`
	DownloadURL      string `json:"downloadUrl"`
}

// CheckVersions fetches the latest release tags from GitHub and compares them
// against the locally-installed versions stored in cfg. It never returns an error
// for API failures — those are reflected as StatusUnknown in the result.
func CheckVersions(ctx context.Context, client *http.Client, cfg *config.Config) ([]ModuleCheck, error) {
	checks := []ModuleCheck{
		checkEngram(ctx, client, cfg),
		checkConpasForge(ctx, client, cfg),
	}
	return checks, nil
}

func checkEngram(ctx context.Context, client *http.Client, cfg *config.Config) ModuleCheck {
	const owner, repo = "Gentleman-Programming", "engram"
	const downloadURL = "https://github.com/Gentleman-Programming/engram/releases/latest"

	installed := cfg.Modules.Engram.Version
	if !cfg.Modules.Engram.Installed || installed == "" {
		return ModuleCheck{
			Module:           "Engram",
			InstalledVersion: installed,
			LatestVersion:    "",
			Status:           StatusNotInstalled,
			DownloadURL:      downloadURL,
		}
	}

	latest, err := download.FetchLatestTag(ctx, client, owner, repo)
	if err != nil {
		return ModuleCheck{
			Module:           "Engram",
			InstalledVersion: installed,
			LatestVersion:    "unknown",
			Status:           StatusUnknown,
			DownloadURL:      downloadURL,
		}
	}

	return ModuleCheck{
		Module:           "Engram",
		InstalledVersion: installed,
		LatestVersion:    latest,
		Status:           compareVersions(installed, latest),
		DownloadURL:      downloadURL,
	}
}

func checkConpasForge(ctx context.Context, client *http.Client, cfg *config.Config) ModuleCheck {
	const owner, repo = "conpas-ai", "conpas-forge"
	const downloadURL = "https://github.com/conpas-ai/conpas-forge/releases/latest"

	// conpas-forge is always installed — it IS the running binary.
	// Use version.Version (set via ldflags in releases); fall back to config if set.
	installed := version.Version
	if installed == "" {
		installed = cfg.Modules.GentleAI.Version
	}

	latest, err := download.FetchLatestTag(ctx, client, owner, repo)
	if err != nil {
		return ModuleCheck{
			Module:           "conpas-forge",
			InstalledVersion: installed,
			LatestVersion:    "unknown",
			Status:           StatusUnknown,
			DownloadURL:      downloadURL,
		}
	}

	return ModuleCheck{
		Module:           "conpas-forge",
		InstalledVersion: installed,
		LatestVersion:    latest,
		Status:           compareVersions(installed, latest),
		DownloadURL:      downloadURL,
	}
}

// compareVersions uses semver to determine status. Non-semver versions are treated as unknown.
func compareVersions(installed, latest string) string {
	if !semver.IsValid(installed) || !semver.IsValid(latest) {
		return StatusUnknown
	}
	cmp := semver.Compare(installed, latest)
	switch {
	case cmp < 0:
		return StatusOutdated
	default:
		return StatusUpToDate
	}
}

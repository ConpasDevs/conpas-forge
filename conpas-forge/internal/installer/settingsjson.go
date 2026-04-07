package installer

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/conpasDEVS/conpas-forge/internal/config"
)

// Merge safely merges newEntries into ~/.claude/settings.json using 4-layer protection.
func Merge(newEntries map[string]any) error {
	settingsPath := config.SettingsJSON()
	bakPath := config.SettingsJSONBak()

	// Layer 1: immutable backup (create once, never overwrite)
	if _, err := os.Stat(settingsPath); err == nil {
		if _, err := os.Stat(bakPath); os.IsNotExist(err) {
			data, err := os.ReadFile(settingsPath)
			if err != nil {
				return fmt.Errorf("read settings for backup: %w", err)
			}
			if err := config.AtomicWrite(bakPath, data, 0644); err != nil {
				return fmt.Errorf("create settings backup: %w", err)
			}
		}
	}

	// Layer 2: read + unmarshal existing
	existing := make(map[string]any)
	if data, err := os.ReadFile(settingsPath); err == nil {
		if err := json.Unmarshal(data, &existing); err != nil {
			return fmt.Errorf("settings.json is not valid JSON: %w — restore with: conpas-forge config restore", err)
		}
	}

	merged := deepMerge(existing, newEntries)

	// Layer 3: re-marshal + round-trip validate
	jsonBytes, err := json.MarshalIndent(merged, "", "  ")
	if err != nil {
		return fmt.Errorf("merge produced invalid JSON: %w — aborting write", err)
	}
	jsonBytes = append(jsonBytes, '\n')

	var roundTrip map[string]any
	if err := json.Unmarshal(jsonBytes, &roundTrip); err != nil {
		return fmt.Errorf("merge validation failed: %w — aborting write", err)
	}

	// Layer 4: atomic write
	return config.AtomicWrite(settingsPath, jsonBytes, 0644)
}

// Restore replaces settings.json with the backup.
func Restore() error {
	bakPath := config.SettingsJSONBak()
	settingsPath := config.SettingsJSON()

	data, err := os.ReadFile(bakPath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("no backup found at %s", bakPath)
		}
		return fmt.Errorf("read backup: %w", err)
	}

	var check map[string]any
	if err := json.Unmarshal(data, &check); err != nil {
		return fmt.Errorf("backup file contains invalid JSON — restore aborted: %w", err)
	}

	return config.AtomicWrite(settingsPath, data, 0644)
}

func deepMerge(dst, src map[string]any) map[string]any {
	for key, srcVal := range src {
		dstVal, exists := dst[key]
		if !exists {
			dst[key] = srcVal
			continue
		}
		srcMap, srcIsMap := srcVal.(map[string]any)
		dstMap, dstIsMap := dstVal.(map[string]any)
		if srcIsMap && dstIsMap {
			dst[key] = deepMerge(dstMap, srcMap)
		} else {
			dst[key] = srcVal
		}
	}
	return dst
}

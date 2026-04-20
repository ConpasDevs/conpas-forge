package installer

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/conpasDEVS/conpas-forge/internal/config"
)

// ForgeManifest tracks the set of skills and output-style files deployed by conpas-forge.
type ForgeManifest struct {
	Skills       []string `json:"skills"`
	OutputStyles []string `json:"output_styles,omitempty"`
}

// ReadManifest reads the manifest at the given path.
// If the file does not exist, it returns an empty manifest (no error).
// If the file contains invalid JSON, it returns an error so the caller can skip cleanup.
func ReadManifest(path string) (*ForgeManifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return &ForgeManifest{}, nil
		}
		return nil, fmt.Errorf("read manifest: %w", err)
	}
	var m ForgeManifest
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("parse manifest: %w", err)
	}
	return &m, nil
}

// WriteManifest atomically writes the given skill list to the manifest file at path.
// Kept for backward compatibility — does not write the OutputStyles field.
func WriteManifest(path string, skills []string) error {
	return WriteManifestFull(path, skills, nil)
}

// WriteManifestFull atomically writes skills and outputStyles to the manifest file at path.
func WriteManifestFull(path string, skills []string, outputStyles []string) error {
	data, err := json.Marshal(ForgeManifest{Skills: skills, OutputStyles: outputStyles})
	if err != nil {
		return fmt.Errorf("marshal manifest: %w", err)
	}
	if err := config.AtomicWrite(path, data, 0644); err != nil {
		return fmt.Errorf("write manifest: %w", err)
	}
	return nil
}

// CalculateStale returns the skills in the manifest that are not in expected.
// These are candidates for removal.
func CalculateStale(manifest *ForgeManifest, expected []string) []string {
	expectedSet := make(map[string]struct{}, len(expected))
	for _, s := range expected {
		expectedSet[s] = struct{}{}
	}
	var stale []string
	for _, s := range manifest.Skills {
		if _, ok := expectedSet[s]; !ok {
			stale = append(stale, s)
		}
	}
	return stale
}

// CleanupStale removes each stale skill directory from skillsDir.
// Permission or other errors are logged and skipped — cleanup never fails the deploy.
func CleanupStale(skillsDir string, stale []string) error {
	return cleanupStaleWith(skillsDir, stale, os.RemoveAll)
}

// cleanupStaleWith is the testable core of CleanupStale.
// It accepts a removeAll func so tests can inject failures without OS tricks.
func cleanupStaleWith(skillsDir string, stale []string, removeAll func(string) error) error {
	for _, name := range stale {
		target := filepath.Join(skillsDir, name)
		if err := removeAll(target); err != nil {
			log.Printf("conpas-forge: cleanup stale skill %q: %v (skipped)", name, err)
		}
	}
	return nil
}

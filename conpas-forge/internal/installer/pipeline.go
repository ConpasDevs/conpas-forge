package installer

import (
	"context"
	"log"

	"github.com/conpasDEVS/conpas-forge/internal/config"
)

// SkillProvider is an optional interface that installers may implement
// to report which skill names they will deploy.
type SkillProvider interface {
	ExpectedSkills() []string
}

func RunPipeline(ctx context.Context, modules []Module, opts *InstallOptions, progress func(ProgressEvent)) []Result {
	// Collect expected skills from all SkillProvider modules.
	var expected []string
	for _, m := range modules {
		if sp, ok := m.(SkillProvider); ok {
			expected = append(expected, sp.ExpectedSkills()...)
		}
	}

	// If any module provides skills, run manifest-based cleanup before deploying.
	if len(expected) > 0 {
		manifestPath := config.ForgeManifest()
		manifest, err := ReadManifest(manifestPath)
		if err != nil {
			// Corrupted manifest: log warning and skip cleanup to protect user files.
			log.Printf("conpas-forge: skipping cleanup — manifest unreadable: %v", err)
		} else {
			stale := CalculateStale(manifest, expected)
			if len(stale) > 0 {
				if err := CleanupStale(config.SkillsDir(), stale); err != nil {
					// CleanupStale only logs per-skill errors and always returns nil.
					log.Printf("conpas-forge: cleanup error: %v", err)
				}
			}
		}
	}

	// Run all installers.
	results := make([]Result, 0, len(modules))
	for _, m := range modules {
		if ctx.Err() != nil {
			results = append(results, Result{
				ModuleName: m.Name(),
				Success:    false,
				Err:        ctx.Err(),
			})
			continue
		}
		r := m.Install(ctx, opts, progress)
		results = append(results, r)
	}

	// Write manifest only if ALL installers succeeded.
	if len(expected) > 0 && !HasErrors(results) {
		if err := WriteManifest(config.ForgeManifest(), expected); err != nil {
			log.Printf("conpas-forge: failed to write manifest: %v", err)
		}
	}

	return results
}

func HasErrors(results []Result) bool {
	for _, r := range results {
		if r.Err != nil {
			return true
		}
	}
	return false
}

func AllPaths(results []Result) []string {
	var paths []string
	for _, r := range results {
		paths = append(paths, r.PathsWritten...)
	}
	return paths
}

func AllWarnings(results []Result) []string {
	var warnings []string
	for _, r := range results {
		warnings = append(warnings, r.Warnings...)
	}
	return warnings
}

// collectExpected is a helper exposed for tests.
func collectExpected(modules []Module) []string {
	var expected []string
	for _, m := range modules {
		if sp, ok := m.(SkillProvider); ok {
			expected = append(expected, sp.ExpectedSkills()...)
		}
	}
	return expected
}

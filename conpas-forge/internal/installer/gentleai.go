package installer

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/conpasDEVS/conpas-forge/internal/assets"
	"github.com/conpasDEVS/conpas-forge/internal/config"
	"github.com/conpasDEVS/conpas-forge/internal/persona"
	"github.com/conpasDEVS/conpas-forge/internal/version"
)

// sddSkills lists the SDD skill names from gentle-ai (excludes zoho-deluge).
var sddSkills = []string{
	"sdd-init", "sdd-explore", "sdd-propose", "sdd-spec",
	"sdd-design", "sdd-tasks", "sdd-apply", "sdd-verify",
	"sdd-archive", "sdd-onboard", "engram-memory",
	"branch-pr", "issue-creation", "judgment-day",
	"go-testing",
	"skill-creator", "skill-registry",
}

func GentleAISkillCount() int {
	return len(sddSkills)
}

func CountGentleAISkillsDeployed(paths []string) int {
	skillsDir := config.SkillsDir()
	sharedDir := config.SharedSkillsDir()
	count := 0
	for _, path := range paths {
		if filepath.Base(path) != "SKILL.md" {
			continue
		}
		if filepath.Dir(path) == sharedDir {
			continue
		}
		if filepath.Dir(filepath.Dir(path)) != skillsDir {
			continue
		}
		count++
	}
	return count
}

type GentleAIInstaller struct{}

func NewGentleAIInstaller() *GentleAIInstaller { return &GentleAIInstaller{} }

func (g *GentleAIInstaller) Name() string { return "Gentle AI" }

func (g *GentleAIInstaller) Install(ctx context.Context, opts *InstallOptions, progress func(ProgressEvent)) Result {
	result := Result{ModuleName: "Gentle AI"}
	emit := func(status, detail string) {
		if progress != nil {
			progress(ProgressEvent{Module: "Gentle AI", Status: status, Detail: detail, Percent: -1})
		}
	}

	// Step 1: Generate and write CLAUDE.md
	emit("writing", "Generating CLAUDE.md...")
	if err := persona.WriteCLAUDEMD(opts.Config, version.Version); err != nil {
		result.Err = fmt.Errorf("write CLAUDE.md: %w", err)
		return result
	}
	result.PathsWritten = append(result.PathsWritten, config.ClaudeMD())

	// Step 2: Deploy SDD skills
	emit("writing", fmt.Sprintf("Deploying %d SDD skills...", len(sddSkills)))
	successCount := 0
	for _, name := range sddSkills {
		data, err := assets.FS.ReadFile("skills/" + name + "/SKILL.md")
		if err != nil {
			result.Warnings = append(result.Warnings, fmt.Sprintf("read skill %s: %v", name, err))
			continue
		}
		dest := filepath.Join(config.SkillDir(name), "SKILL.md")
		if err := config.AtomicWrite(dest, data, 0644); err != nil {
			result.Warnings = append(result.Warnings, fmt.Sprintf("write skill %s: %v", name, err))
			continue
		}
		result.PathsWritten = append(result.PathsWritten, dest)
		successCount++
	}
	if successCount < len(sddSkills) {
		result.Warnings = append(result.Warnings, fmt.Sprintf("%d/%d skills deployed", successCount, len(sddSkills)))
	}

	// Step 3: Deploy _shared assets
	emit("writing", "Deploying shared assets...")
	if entries, err := assets.FS.ReadDir("skills/_shared"); err == nil {
		for _, e := range entries {
			if e.IsDir() || strings.HasPrefix(e.Name(), ".") {
				continue
			}
			data, err := assets.FS.ReadFile("skills/_shared/" + e.Name())
			if err != nil {
				result.Warnings = append(result.Warnings, fmt.Sprintf("read _shared/%s: %v", e.Name(), err))
				continue
			}
			dest := filepath.Join(config.SharedSkillsDir(), e.Name())
			if err := config.AtomicWrite(dest, data, 0644); err != nil {
				result.Warnings = append(result.Warnings, fmt.Sprintf("write _shared/%s: %v", e.Name(), err))
				continue
			}
			result.PathsWritten = append(result.PathsWritten, dest)
		}
	}

	result.Success = true
	return result
}

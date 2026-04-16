package installer

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/conpasDEVS/conpas-forge/internal/assets"
	"github.com/conpasDEVS/conpas-forge/internal/config"
)

type ConpasAIInstaller struct{}

func NewConpasAIInstaller() *ConpasAIInstaller { return &ConpasAIInstaller{} }

func (c *ConpasAIInstaller) Name() string { return "Zoho Deluge" }

func (c *ConpasAIInstaller) ExpectedSkills() []string { return []string{"zoho-deluge"} }

func (c *ConpasAIInstaller) Install(ctx context.Context, opts *InstallOptions, progress func(ProgressEvent)) Result {
	result := Result{ModuleName: "Zoho Deluge"}

	if progress != nil {
		progress(ProgressEvent{Module: "Zoho Deluge", Status: "writing", Detail: "Deploying zoho-deluge skill...", Percent: -1})
	}

	data, err := assets.FS.ReadFile("skills/zoho-deluge/SKILL.md")
	if err != nil {
		result.Err = fmt.Errorf("read zoho-deluge skill: %w", err)
		return result
	}

	dest := filepath.Join(config.SkillDir("zoho-deluge"), "SKILL.md")
	if err := config.AtomicWrite(dest, data, 0644); err != nil {
		result.Err = fmt.Errorf("write zoho-deluge skill: %w", err)
		return result
	}

	result.PathsWritten = append(result.PathsWritten, dest)
	result.Success = true

	if progress != nil {
		progress(ProgressEvent{Module: "Zoho Deluge", Status: "done", Detail: "Zoho Deluge skill deployed", Percent: 100})
	}
	return result
}

package persona

import (
	"bytes"
	"fmt"
	"text/template"
	"time"

	"github.com/conpasDEVS/conpas-forge/internal/assets"
	"github.com/conpasDEVS/conpas-forge/internal/config"
	"github.com/conpasDEVS/conpas-forge/internal/models"
)

type CLAUDEMDData struct {
	PersonaName    string
	PersonaBlock   string
	ModelRows      []ModelRow
	Version        string
	GeneratedAt    string
	EngramProtocol string // content from skills/engram-memory/SKILL.md
}

type ModelRow struct {
	Role  string
	Model string
}

func BuildCLAUDEMDData(cfg *config.Config, ver string) (*CLAUDEMDData, error) {
	content, err := LoadPersonaContent(cfg.Persona)
	if err != nil {
		return nil, err
	}

	rows := make([]ModelRow, 0, len(models.CanonicalRoles))
	for _, role := range models.CanonicalRoles {
		model := cfg.Models[role]
		if model == "" {
			model = models.Defaults[role]
		}
		rows = append(rows, ModelRow{Role: role, Model: model})
	}

	var engramProtocol string
	if raw, err := assets.FS.ReadFile("skills/engram-memory/SKILL.md"); err == nil {
		engramProtocol = string(raw)
	} // silently ignore read failure — EngramProtocol stays ""

	return &CLAUDEMDData{
		PersonaName:    cfg.Persona,
		PersonaBlock:   string(content),
		ModelRows:      rows,
		Version:        ver,
		GeneratedAt:    time.Now().Format(time.RFC3339),
		EngramProtocol: engramProtocol,
	}, nil
}

func RenderCLAUDEMD(data *CLAUDEMDData) ([]byte, error) {
	tmplBytes, err := assets.FS.ReadFile("claude-md/CLAUDE.md.tmpl")
	if err != nil {
		return nil, fmt.Errorf("read CLAUDE.md template: %w", err)
	}

	tmpl, err := template.New("claude-md").Parse(string(tmplBytes))
	if err != nil {
		return nil, fmt.Errorf("parse CLAUDE.md template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("execute CLAUDE.md template: %w", err)
	}

	return buf.Bytes(), nil
}

func WriteCLAUDEMD(cfg *config.Config, ver string) error {
	data, err := BuildCLAUDEMDData(cfg, ver)
	if err != nil {
		return fmt.Errorf("build CLAUDE.md data: %w", err)
	}

	rendered, err := RenderCLAUDEMD(data)
	if err != nil {
		return fmt.Errorf("render CLAUDE.md: %w", err)
	}

	return config.AtomicWrite(config.ClaudeMD(), rendered, 0644)
}

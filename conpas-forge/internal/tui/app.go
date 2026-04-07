package tui

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/conpasDEVS/conpas-forge/internal/config"
	"github.com/conpasDEVS/conpas-forge/internal/installer"
)

type Screen int

const (
	ScreenModules Screen = iota
	ScreenPersona
	ScreenModels
	ScreenInstall
	ScreenSummary
)

type Model struct {
	screen    Screen
	cfg       *config.Config
	platform  installer.Platform
	homeDir   string
	modules   ModulesModel
	persona   PersonaModel
	models    ModelsModel
	summary   SummaryModel
	progress  []installer.ProgressEvent
	results   []installer.Result
	width     int
	height    int
	cancelled bool
	program   *tea.Program
}

func New(cfg *config.Config, platform installer.Platform, homeDir string) Model {
	return Model{
		screen:   ScreenModules,
		cfg:      cfg,
		platform: platform,
		homeDir:  homeDir,
		modules:  NewModulesModel(cfg),
		persona:  NewPersonaModel(cfg),
		models:   NewModelsModel(cfg),
	}
}

func (m *Model) SetProgram(p *tea.Program) {
	m.program = p
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		if msg.String() == "ctrl+c" && m.screen < ScreenInstall {
			m.cancelled = true
			return m, tea.Quit
		}

	case SetProgramMsg:
		m.program = msg.P
		return m, nil

	case AdvanceMsg:
		m.screen++
		return m, nil

	case BackMsg:
		if m.screen > ScreenModules {
			m.screen--
		}
		return m, nil

	case ConfirmInstallMsg:
		m.screen = ScreenInstall
		// Update config with TUI selections before installing
		m.cfg.Persona = m.persona.Selected()
		m.cfg.Models = m.models.Assignments()
		return m, m.runPipelineCmd()

	case ProgressMsg:
		m.progress = append(m.progress, msg.Event)
		return m, nil

	case InstallDoneMsg:
		m.results = msg.Results
		// Update module status in config
		for _, r := range msg.Results {
			switch r.ModuleName {
			case "Engram":
				m.cfg.Modules.Engram.Installed = r.Success
			case "Gentle AI":
				m.cfg.Modules.GentleAI.Installed = r.Success
				m.cfg.Modules.GentleAI.SkillsDeployed = 16
			case "Zoho Deluge":
				m.cfg.Modules.ZohoDeluge.Installed = r.Success
			}
		}
		_ = config.Save(m.cfg)
		m.summary = NewSummaryModel(m.results)
		m.screen = ScreenSummary
		return m, nil
	}

	// Route messages to active sub-model
	var cmd tea.Cmd
	switch m.screen {
	case ScreenModules:
		m.modules, cmd = m.modules.Update(msg)
	case ScreenPersona:
		m.persona, cmd = m.persona.Update(msg)
	case ScreenModels:
		m.models, cmd = m.models.Update(msg)
	case ScreenSummary:
		m.summary, cmd = m.summary.Update(msg)
	}
	return m, cmd
}

func (m Model) View() string {
	switch m.screen {
	case ScreenModules:
		return m.modules.View()
	case ScreenPersona:
		return m.persona.View()
	case ScreenModels:
		return m.models.View()
	case ScreenInstall:
		return m.progressView()
	case ScreenSummary:
		return m.summary.View()
	}
	return ""
}

func (m Model) progressView() string {
	s := titleStyle.Render("Installing...") + "\n\n"
	for _, evt := range m.progress {
		icon := "  "
		switch evt.Status {
		case "done":
			icon = checkStyle.Render("✓ ")
		case "error":
			icon = errorStyle.Render("✗ ")
		default:
			icon = "  "
		}
		s += icon + "[" + evt.Module + "] " + evt.Detail + "\n"
	}
	return s
}

func (m Model) Cancelled() bool            { return m.cancelled }
func (m Model) Results() []installer.Result { return m.results }
func (m Model) SelectedModules() []string   { return m.modules.Selected() }

func (m Model) Selections() *installer.InstallOptions {
	return &installer.InstallOptions{
		Config:   m.cfg,
		Persona:  m.persona.Selected(),
		Models:   m.models.Assignments(),
		Platform: m.platform,
		HomeDir:  m.homeDir,
	}
}

func (m Model) runPipelineCmd() tea.Cmd {
	opts := m.Selections()
	selectedIDs := m.SelectedModules()
	p := m.program

	return func() tea.Msg {
		modules := installer.BuildModules(selectedIDs)
		progress := func(evt installer.ProgressEvent) {
			if p != nil {
				p.Send(ProgressMsg{Event: evt})
			}
		}
		results := installer.RunPipeline(context.Background(), modules, opts, progress)
		return InstallDoneMsg{Results: results}
	}
}


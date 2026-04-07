package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/conpasDEVS/conpas-forge/internal/config"
	"github.com/conpasDEVS/conpas-forge/internal/installer"
)

var (
	checkStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
	cursorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("6")).Bold(true)
	errorStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("1"))
	titleStyle  = lipgloss.NewStyle().Bold(true).Underline(true)
)

type ModuleChoice struct {
	ID          string
	Label       string
	Description string
	Checked     bool
}

type ModulesModel struct {
	choices []ModuleChoice
	cursor  int
	errMsg  string
}

func NewModulesModel(cfg *config.Config) ModulesModel {
	return ModulesModel{
		choices: []ModuleChoice{
			{ID: "engram", Label: "Engram", Description: "Persistent memory MCP server", Checked: cfg.Modules.Engram.Installed},
			{ID: "gentle-ai", Label: "Gentle AI Skills", Description: fmt.Sprintf("%d skills + CLAUDE.md + output styles", installer.GentleAISkillCount()), Checked: cfg.Modules.GentleAI.Installed},
			{ID: "zoho-deluge", Label: "Zoho Deluge Skill", Description: "Conpas AI coding standard for Zoho Deluge", Checked: cfg.Modules.ZohoDeluge.Installed},
			{ID: "all", Label: "All modules", Description: "Select all three", Checked: false},
		},
	}
}

func (m ModulesModel) Init() tea.Cmd { return nil }

func (m ModulesModel) Update(msg tea.Msg) (ModulesModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.choices)-1 {
				m.cursor++
			}
		case " ":
			m.errMsg = ""
			if m.choices[m.cursor].ID == "all" {
				// Toggle all individual modules
				allChecked := true
				for _, c := range m.choices[:3] {
					if !c.Checked {
						allChecked = false
						break
					}
				}
				for i := 0; i < 3; i++ {
					m.choices[i].Checked = !allChecked
				}
			} else {
				m.choices[m.cursor].Checked = !m.choices[m.cursor].Checked
			}
		case "enter":
			if err := m.Validate(); err != nil {
				m.errMsg = err.Error()
			} else {
				return m, func() tea.Msg { return AdvanceMsg{} }
			}
		case "q", "ctrl+c":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m ModulesModel) View() string {
	var sb strings.Builder
	sb.WriteString(titleStyle.Render("Select modules to install") + "\n\n")
	for i, c := range m.choices {
		cursor := "  "
		if i == m.cursor {
			cursor = cursorStyle.Render("▶ ")
		}
		check := "[ ]"
		if c.Checked {
			check = checkStyle.Render("[✓]")
		}
		sb.WriteString(fmt.Sprintf("%s%s %s — %s\n", cursor, check, c.Label, c.Description))
	}
	if m.errMsg != "" {
		sb.WriteString("\n" + errorStyle.Render(m.errMsg) + "\n")
	}
	sb.WriteString("\n" + lipgloss.NewStyle().Faint(true).Render("Space: toggle • Enter: confirm • q: quit"))
	return sb.String()
}

func (m ModulesModel) Selected() []string {
	var ids []string
	for _, c := range m.choices {
		if c.Checked && c.ID != "all" {
			ids = append(ids, c.ID)
		}
	}
	return ids
}

func (m ModulesModel) Validate() error {
	if len(m.Selected()) == 0 {
		return fmt.Errorf("Select at least one module")
	}
	return nil
}

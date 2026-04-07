package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/conpasDEVS/conpas-forge/internal/config"
	"github.com/conpasDEVS/conpas-forge/internal/models"
)

var (
	editingStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("3"))
	headerStyle  = lipgloss.NewStyle().Bold(true)
)

type RoleEntry struct {
	Role  string
	Model string
}

type ModelsModel struct {
	roles     []RoleEntry
	cursor    int
	editing   bool
	textInput textinput.Model
	oldValue  string
	errMsg    string
}

func NewModelsModel(cfg *config.Config) ModelsModel {
	rows := make([]RoleEntry, 0, len(models.CanonicalRoles))
	for _, role := range models.CanonicalRoles {
		model := cfg.Models[role]
		if model == "" {
			model = models.Defaults[role]
		}
		rows = append(rows, RoleEntry{Role: role, Model: model})
	}

	ti := textinput.New()
	ti.CharLimit = 128

	return ModelsModel{roles: rows, textInput: ti}
}

func (m ModelsModel) Init() tea.Cmd { return nil }

func (m ModelsModel) Update(msg tea.Msg) (ModelsModel, tea.Cmd) {
	if m.editing {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "enter":
				val := strings.TrimSpace(m.textInput.Value())
				if val == "" {
					m.errMsg = fmt.Sprintf("Role '%s' requires a model", m.roles[m.cursor].Role)
					return m, nil
				}
				m.roles[m.cursor].Model = val
				m.editing = false
				m.errMsg = ""
				return m, nil
			case "esc":
				m.roles[m.cursor].Model = m.oldValue
				m.editing = false
				m.errMsg = ""
				return m, nil
			}
		}
		var cmd tea.Cmd
		m.textInput, cmd = m.textInput.Update(msg)
		return m, cmd
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.roles)-1 {
				m.cursor++
			}
		case "enter":
			if err := m.Validate(); err != nil {
				m.errMsg = err.Error()
				return m, nil
			}
			return m, func() tea.Msg { return ConfirmInstallMsg{} }
		case " ", "e":
			// Start editing current row
			m.oldValue = m.roles[m.cursor].Model
			m.textInput.SetValue(m.roles[m.cursor].Model)
			m.textInput.Focus()
			m.editing = true
			m.errMsg = ""
		case "esc", "left", "h":
			return m, func() tea.Msg { return BackMsg{} }
		}
	}
	return m, nil
}

func (m ModelsModel) View() string {
	var sb strings.Builder
	sb.WriteString(titleStyle.Render("Assign models to SDD roles") + "\n\n")
	sb.WriteString(headerStyle.Render(fmt.Sprintf("%-20s %s", "Role", "Model")) + "\n")
	sb.WriteString(strings.Repeat("─", 60) + "\n")

	for i, r := range m.roles {
		cursor := "  "
		if i == m.cursor {
			cursor = cursorStyle.Render("▶ ")
		}
		modelVal := r.Model
		if i == m.cursor && m.editing {
			modelVal = editingStyle.Render(m.textInput.View())
		}
		sb.WriteString(fmt.Sprintf("%s%-20s %s\n", cursor, r.Role, modelVal))
	}

	if m.errMsg != "" {
		sb.WriteString("\n" + errorStyle.Render(m.errMsg) + "\n")
	}
	sb.WriteString("\n" + lipgloss.NewStyle().Faint(true).Render("↑/↓: navigate • Space/E: edit • Enter: confirm • Esc: back"))
	return sb.String()
}

func (m ModelsModel) Assignments() map[string]string {
	result := make(map[string]string, len(m.roles))
	for _, r := range m.roles {
		result[r.Role] = r.Model
	}
	return result
}

func (m ModelsModel) Validate() error {
	for _, r := range m.roles {
		if r.Model == "" {
			return fmt.Errorf("Role '%s' requires a model", r.Role)
		}
	}
	return nil
}

package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/conpasDEVS/conpas-forge/internal/config"
	"github.com/conpasDEVS/conpas-forge/internal/persona"
)

var selectedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("5")).Bold(true)

var personaDescriptions = map[string]string{
	"asturiano":   "Direct Asturian colleague — warm, pragmatic, regional wit",
	"galleguinho": "Galician veteran with retranca — ironic, wise, entrañable",
	"tony-stark":  "Eccentric genius — charismatic, fast, engineering metaphors",
	"yoda":        "Cryptic Jedi master — OSV syntax, minimal, wise",
	"sargento":    "Iron sergeant — cold, authoritarian, zero social, hyper-technical",
	"argentino":   "Rioplatense mentor — passionate, voseo, constructively challenging",
}

type PersonaEntry struct {
	Name        string
	Description string
}

type PersonaModel struct {
	personas []PersonaEntry
	cursor   int
	selected int
}

func NewPersonaModel(cfg *config.Config) PersonaModel {
	names := persona.ValidNames()
	entries := make([]PersonaEntry, 0, len(names))
	for _, name := range names {
		desc := personaDescriptions[name]
		if desc == "" {
			desc = name
		}
		entries = append(entries, PersonaEntry{Name: name, Description: desc})
	}

	selected := 0
	for i, e := range entries {
		if e.Name == cfg.Persona {
			selected = i
			break
		}
	}

	return PersonaModel{personas: entries, cursor: selected, selected: selected}
}

func (m PersonaModel) Init() tea.Cmd { return nil }

func (m PersonaModel) Update(msg tea.Msg) (PersonaModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.personas)-1 {
				m.cursor++
			}
		case "enter":
			m.selected = m.cursor
			return m, func() tea.Msg { return AdvanceMsg{} }
		case "esc", "left", "h":
			return m, func() tea.Msg { return BackMsg{} }
		}
	}
	return m, nil
}

func (m PersonaModel) View() string {
	var sb strings.Builder
	sb.WriteString(titleStyle.Render("Select persona") + "\n\n")
	for i, p := range m.personas {
		cursor := "  "
		if i == m.cursor {
			cursor = cursorStyle.Render("▶ ")
		}
		line := fmt.Sprintf("%s%s — %s", cursor, p.Name, p.Description)
		if i == m.selected {
			line = selectedStyle.Render(line)
		}
		sb.WriteString(line + "\n")
	}
	sb.WriteString("\n" + lipgloss.NewStyle().Faint(true).Render("↑/↓: navigate • Enter: select • Esc: back"))
	return sb.String()
}

func (m PersonaModel) Selected() string {
	if m.selected >= 0 && m.selected < len(m.personas) {
		return m.personas[m.selected].Name
	}
	return "asturiano"
}

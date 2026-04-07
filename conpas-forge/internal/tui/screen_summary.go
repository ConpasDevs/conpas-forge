package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/conpasDEVS/conpas-forge/internal/installer"
)

var (
	successStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("2")).Bold(true)
	warningStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("3"))
	moduleStyle  = lipgloss.NewStyle().Bold(true)
)

type SummaryModel struct {
	results []installer.Result
}

func NewSummaryModel(results []installer.Result) SummaryModel {
	return SummaryModel{results: results}
}

func (m SummaryModel) Init() tea.Cmd { return nil }

func (m SummaryModel) Update(msg tea.Msg) (SummaryModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "enter", " ", "esc", "ctrl+c":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m SummaryModel) View() string {
	var sb strings.Builder
	sb.WriteString(titleStyle.Render("Installation Summary") + "\n\n")

	hasErrors := installer.HasErrors(m.results)

	for _, r := range m.results {
		sb.WriteString(moduleStyle.Render("▸ "+r.ModuleName) + "\n")
		if r.Err != nil {
			sb.WriteString("  " + errorStyle.Render(fmt.Sprintf("ERROR: %v", r.Err)) + "\n")
		} else {
			for _, p := range r.PathsWritten {
				sb.WriteString("  " + checkStyle.Render("✓") + " " + p + "\n")
			}
		}
		for _, w := range r.Warnings {
			sb.WriteString("  " + warningStyle.Render("⚠ "+w) + "\n")
		}
		sb.WriteString("\n")
	}

	if hasErrors {
		sb.WriteString(errorStyle.Render("Installation completed with errors.") + "\n")
	} else {
		sb.WriteString(successStyle.Render("Installation complete!") + "\n")
	}

	sb.WriteString("\n" + lipgloss.NewStyle().Faint(true).Render("Press any key to exit"))
	return sb.String()
}

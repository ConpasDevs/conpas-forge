package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/conpasDEVS/conpas-forge/internal/installer"
)

type AdvanceMsg struct{}
type BackMsg struct{}
type ValidationErrorMsg struct{ Message string }
type ConfirmInstallMsg struct{}
type ProgressMsg struct{ Event installer.ProgressEvent }
type InstallDoneMsg struct{ Results []installer.Result }
type SetProgramMsg struct{ P *tea.Program }

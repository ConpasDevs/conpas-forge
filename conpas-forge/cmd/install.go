package cmd

import (
	"fmt"
	"io"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/conpasDEVS/conpas-forge/internal/checker"
	"github.com/conpasDEVS/conpas-forge/internal/config"
	"github.com/conpasDEVS/conpas-forge/internal/installer"
	"github.com/conpasDEVS/conpas-forge/internal/tui"
)

var runHealthForInstall = checker.RunHealth

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Interactive TUI installer for Claude Code environment",
	RunE:  runInstall,
}

func init() {
	rootCmd.AddCommand(installCmd)
}

func runInstall(_ *cobra.Command, _ []string) error {
	if !term.IsTerminal(int(os.Stdin.Fd())) {
		fmt.Fprintln(os.Stderr, "install requires an interactive terminal")
		os.Exit(1)
	}

	cfg, err := config.Load()
	if err != nil {
		cfg2 := config.DefaultConfig()
		cfg = &cfg2
	}

	platform := installer.DetectPlatform()
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("cannot determine home directory: %w", err)
	}

	m := tui.New(cfg, platform, homeDir)
	p := tea.NewProgram(m)

	// Send program reference into the model on first tick
	go func() {
		p.Send(tui.SetProgramMsg{P: p})
	}()

	finalModel, err := p.Run()
	if err != nil {
		return fmt.Errorf("TUI error: %w", err)
	}

	fm, ok := finalModel.(tui.Model)
	if !ok {
		return fmt.Errorf("unexpected model type after TUI")
	}

	if fm.Cancelled() {
		fmt.Fprintln(os.Stderr, "Install cancelled.")
		return nil
	}

	if installer.HasErrors(fm.Results()) {
		os.Exit(1)
	}

	postInstallHealthSummary(os.Stdout, homeDir, fm.Results(), fm.Cancelled())

	return nil
}

func postInstallHealthSummary(out io.Writer, homeDir string, results []installer.Result, cancelled bool) {
	if cancelled || installer.HasErrors(results) {
		return
	}

	report, err := runHealthForInstall(checker.HealthOptions{HomeDir: homeDir})
	if err != nil {
		fmt.Fprintf(out, "\nPost-install health summary unavailable: %v\n", err) //nolint:errcheck
		return
	}

	fmt.Fprintln(out)                                           //nolint:errcheck
	fmt.Fprint(out, checker.RenderConciseHealthSummary(report)) //nolint:errcheck
}

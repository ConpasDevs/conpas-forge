package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/conpasDEVS/conpas-forge/internal/assets"
	"github.com/conpasDEVS/conpas-forge/internal/config"
	"github.com/conpasDEVS/conpas-forge/internal/installer"
	"github.com/conpasDEVS/conpas-forge/internal/models"
	"github.com/conpasDEVS/conpas-forge/internal/persona"
	"github.com/conpasDEVS/conpas-forge/internal/version"
)

var (
	personaFlag string
	modelFlags  []string
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage conpas-forge configuration",
	RunE:  runConfig,
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "Show current configuration",
	RunE:  runConfigList,
}

var restoreCmd = &cobra.Command{
	Use:   "restore",
	Short: "Restore settings.json from backup",
	RunE:  runConfigRestore,
}

func init() {
	configCmd.Flags().StringVar(&personaFlag, "persona", "", "Set active persona (e.g. yoda, asturiano)")
	configCmd.Flags().StringArrayVar(&modelFlags, "model", nil, "Set model for a role (e.g. orchestrator=claude-opus-4-6)")
	configCmd.AddCommand(listCmd)
	configCmd.AddCommand(restoreCmd)
	rootCmd.AddCommand(configCmd)
}

func runConfig(cmd *cobra.Command, _ []string) error {
	if personaFlag == "" && len(modelFlags) == 0 {
		return cmd.Help()
	}

	if !config.Exists() {
		fmt.Fprintln(os.Stderr, "config not found: run 'conpas-forge install' first")
		os.Exit(1)
	}

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	// Validate all inputs first (atomic: fail before any mutation)
	if personaFlag != "" && !models.IsValidPersona(personaFlag) {
		fmt.Fprintf(os.Stderr, "unknown persona %q. Valid personas: %s\n",
			personaFlag, strings.Join(models.ValidPersonas, ", "))
		os.Exit(1)
	}

	type rolePair struct{ role, model string }
	var pairs []rolePair
	for _, flag := range modelFlags {
		parts := strings.SplitN(flag, "=", 2)
		if len(parts) != 2 {
			fmt.Fprintf(os.Stderr, "invalid format: expected role=model, got %q\n", flag)
			os.Exit(1)
		}
		role, model := parts[0], parts[1]
		if !models.IsValidRole(role) {
			fmt.Fprintf(os.Stderr, "unknown role %q. Valid roles: %s\n",
				role, strings.Join(models.CanonicalRoles, ", "))
			os.Exit(1)
		}
		if model == "" {
			fmt.Fprintf(os.Stderr, "model for role %q cannot be empty\n", role)
			os.Exit(1)
		}
		pairs = append(pairs, rolePair{role, model})
	}

	// Apply all changes in memory
	if personaFlag != "" {
		cfg.Persona = personaFlag
	}
	for _, p := range pairs {
		cfg.Models[p.role] = p.model
		fmt.Printf("  %s → %s\n", p.role, p.model)
	}

	// Save once
	if err := config.Save(cfg); err != nil {
		return fmt.Errorf("save config: %w", err)
	}

	// Regenerate CLAUDE.md once
	if err := persona.WriteCLAUDEMD(cfg, version.Version); err != nil {
		return fmt.Errorf("regenerate CLAUDE.md: %w", err)
	}

	// Reconcile output styles when persona changed.
	if personaFlag != "" {
		styleFile := persona.OutputStyleFor(personaFlag)
		if styleFile == "" {
			return fmt.Errorf("output-style mapping missing for persona %q", personaFlag)
		}
		styleData, err := assets.FS.ReadFile("output-styles/" + styleFile)
		if err != nil {
			return fmt.Errorf("read output-style asset %q: %w", styleFile, err)
		}
		removed, err := installer.ReconcileOutputStyles(config.OutputStylesDir(), styleFile, styleData)
		if err != nil {
			return fmt.Errorf("reconcile output styles: %w", err)
		}
		for _, name := range removed {
			fmt.Printf("  removed orphan output-style: %s\n", name)
		}
	}

	if personaFlag != "" {
		fmt.Printf("Persona updated to %s. CLAUDE.md regenerated.\n", personaFlag)
	} else {
		fmt.Println("CLAUDE.md regenerated.")
	}

	return nil
}

func runConfigList(_ *cobra.Command, _ []string) error {
	if !config.Exists() {
		fmt.Fprintln(os.Stderr, "config not found: run 'conpas-forge install' first")
		os.Exit(1)
	}

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	fmt.Printf("Persona: %s\n\n", cfg.Persona)
	fmt.Println("Models:")
	for _, role := range models.CanonicalRoles {
		fmt.Printf("  %-20s %s\n", role, cfg.Models[role])
	}
	fmt.Println("\nModules:")
	printModule := func(name string, s config.ModuleStatus) {
		if s.Installed {
			fmt.Printf("  %-20s installed", name)
			if s.Version != "" {
				fmt.Printf(" (v%s)", s.Version)
			}
			if s.SkillsDeployed > 0 {
				fmt.Printf(" (%d skills)", s.SkillsDeployed)
			}
			fmt.Println()
		} else {
			fmt.Printf("  %-20s (not installed)\n", name)
		}
	}
	printModule("engram", cfg.Modules.Engram)
	printModule("gentle-ai", cfg.Modules.GentleAI)
	printModule("zoho-deluge", cfg.Modules.ZohoDeluge)

	return nil
}

func runConfigRestore(_ *cobra.Command, _ []string) error {
	if err := installer.Restore(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println("settings.json restored from backup.")
	return nil
}

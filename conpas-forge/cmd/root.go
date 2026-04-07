package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "conpas-forge",
	Short: "conpas-forge — Claude Code environment installer",
	Long:  "conpas-forge installs and configures a complete Claude Code environment with Engram memory, SDD skills, and custom personas.",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

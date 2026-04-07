package cmd

import (
	"fmt"
	"runtime"

	"github.com/conpasDEVS/conpas-forge/internal/version"
	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("conpas-forge %s\n", version.Version)
		fmt.Printf("Go runtime: %s\n", runtime.Version())
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}

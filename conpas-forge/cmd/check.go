package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"text/tabwriter"
	"time"

	"github.com/conpasDEVS/conpas-forge/internal/checker"
	"github.com/conpasDEVS/conpas-forge/internal/config"
	"github.com/spf13/cobra"
)

var checkJSONFlag bool

var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Check installed vs latest versions of all modules",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("load config: %w", err)
		}

		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		client := &http.Client{Timeout: 10 * time.Second}
		results, _ := checker.CheckVersions(ctx, client, cfg)

		if checkJSONFlag {
			enc := json.NewEncoder(cmd.OutOrStdout())
			enc.SetIndent("", "  ")
			return enc.Encode(struct {
				Modules []checker.ModuleCheck `json:"modules"`
			}{Modules: results})
		}

		w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 3, ' ', 0)
		fmt.Fprintln(w, "MODULE\tINSTALLED\tLATEST\tSTATUS")
		fmt.Fprintln(w, "------\t---------\t------\t------")
		for _, r := range results {
			installed := r.InstalledVersion
			if installed == "" {
				installed = "—"
			}
			latest := r.LatestVersion
			if latest == "" {
				latest = "—"
			}
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", r.Module, installed, latest, r.Status)
		}
		return w.Flush()
	},
}

func init() {
	checkCmd.Flags().BoolVar(&checkJSONFlag, "json", false, "Output results as JSON")
	rootCmd.AddCommand(checkCmd)
}

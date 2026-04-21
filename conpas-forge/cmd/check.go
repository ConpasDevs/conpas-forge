package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"text/tabwriter"
	"time"

	"github.com/conpasDEVS/conpas-forge/internal/checker"
	"github.com/conpasDEVS/conpas-forge/internal/config"
	"github.com/conpasDEVS/conpas-forge/internal/installer"
	"github.com/spf13/cobra"
)

var checkJSONFlag bool
var checkHealthFlag bool
var checkRepairFlag bool

var (
	loadConfigForCheck = config.Load
	runVersionCheck    = checker.CheckVersions
	runHealthCheck     = checker.RunHealth
)

var errHealthCheckFailed = errors.New("health check reported failures")

var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Check installed vs latest versions of all modules",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Flag validation: --repair requires --health
		if checkRepairFlag && !checkHealthFlag {
			return errors.New("--repair requires --health")
		}

		cfg, err := loadConfigForCheck()
		if err != nil {
			return fmt.Errorf("load config: %w", err)
		}

		if checkHealthFlag {
			homeDir, err := os.UserHomeDir()
			if err != nil {
				return fmt.Errorf("resolve home directory: %w", err)
			}

			report, err := runHealthCheck(checker.HealthOptions{HomeDir: homeDir})
			if err != nil {
				return fmt.Errorf("run health checks: %w", err)
			}

			// Run repair if requested
			var repairReport *installer.RepairReport
			var repairErr error
			if checkRepairFlag {
				ctx, cancel := context.WithTimeout(cmd.Context(), 30*time.Second)
				defer cancel()
				r, rErr := installer.RepairEngram(ctx)
				repairReport = &r
				repairErr = rErr
			}

			if checkJSONFlag {
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				if repairReport != nil {
					// Combined JSON output: {"health": ..., "repair": ...}
					combined := struct {
						Health checker.HealthReport    `json:"health"`
						Repair *installer.RepairReport `json:"repair,omitempty"`
					}{Health: report, Repair: repairReport}
					if err := enc.Encode(combined); err != nil {
						return fmt.Errorf("encode combined JSON: %w", err)
					}
				} else {
					if err := enc.Encode(report); err != nil {
						return fmt.Errorf("encode health JSON: %w", err)
					}
				}
			} else {
				if _, err := cmd.OutOrStdout().Write([]byte(checker.RenderDetailedHealth(report))); err != nil {
					return fmt.Errorf("write health output: %w", err)
				}
				if repairReport != nil {
					fmt.Fprint(cmd.OutOrStdout(), installer.RenderRepairSummary(*repairReport))
				}
			}

			if repairErr != nil {
				return repairErr
			}

			if report.Summary.Fail > 0 {
				return errHealthCheckFailed
			}
			return nil
		}

		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		client := &http.Client{Timeout: 10 * time.Second}
		results, _ := runVersionCheck(ctx, client, cfg)

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
	checkCmd.Flags().BoolVar(&checkHealthFlag, "health", false, "Run Claude Code installation health checks")
	checkCmd.Flags().BoolVar(&checkRepairFlag, "repair", false, "Config-only repair of Engram MCP registration and allowlist (requires --health)")
	rootCmd.AddCommand(checkCmd)
}

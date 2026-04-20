package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"testing"

	"github.com/conpasDEVS/conpas-forge/internal/checker"
	"github.com/conpasDEVS/conpas-forge/internal/config"
)

func TestCheckCommand_HasJSONFlag(t *testing.T) {
	flag := checkCmd.Flags().Lookup("json")
	if flag == nil {
		t.Fatal("check command missing --json flag")
	}
}

func TestCheckCommand_HasHealthFlag(t *testing.T) {
	flag := checkCmd.Flags().Lookup("health")
	if flag == nil {
		t.Fatal("check command missing --health flag")
	}
}

func TestCheckCommand_DefaultRoutingUnchanged(t *testing.T) {
	withCheckDeps(t)

	checkJSONFlag = false
	checkHealthFlag = false

	loadConfigForCheck = func() (*config.Config, error) {
		return &config.Config{}, nil
	}

	calledVersion := false
	calledHealth := false
	runVersionCheck = func(_ context.Context, _ *http.Client, _ *config.Config) ([]checker.ModuleCheck, error) {
		calledVersion = true
		return []checker.ModuleCheck{{Module: "Engram", Status: checker.StatusUpToDate}}, nil
	}
	runHealthCheck = func(_ checker.HealthOptions) (checker.HealthReport, error) {
		calledHealth = true
		return checker.HealthReport{}, nil
	}

	buf := &bytes.Buffer{}
	checkCmd.SetOut(buf)
	defer checkCmd.SetOut(nil)

	if err := checkCmd.RunE(checkCmd, []string{}); err != nil {
		t.Fatalf("RunE returned error: %v", err)
	}

	if !calledVersion {
		t.Fatal("expected default check to call version checker")
	}
	if calledHealth {
		t.Fatal("expected default check to skip health checker")
	}

	output := buf.String()
	if !strings.Contains(output, "MODULE") || !strings.Contains(output, "STATUS") {
		t.Fatalf("table output missing expected headers, got:\n%q", output)
	}
}

func TestCheckCommand_HealthDetailedOutput(t *testing.T) {
	withCheckDeps(t)

	checkJSONFlag = false
	checkHealthFlag = true

	loadConfigForCheck = func() (*config.Config, error) {
		return &config.Config{}, nil
	}

	report := checker.HealthReport{
		Scope: "claude-code",
		Summary: checker.HealthSummary{
			OK: 1,
		},
		Checks: []checker.HealthCheck{{
			ID:       "core.claude_dir",
			Category: "core",
			Status:   checker.HealthOK,
			Message:  "ok",
		}},
	}

	calledHealth := false
	runHealthCheck = func(_ checker.HealthOptions) (checker.HealthReport, error) {
		calledHealth = true
		return report, nil
	}

	buf := &bytes.Buffer{}
	checkCmd.SetOut(buf)
	defer checkCmd.SetOut(nil)

	if err := checkCmd.RunE(checkCmd, []string{}); err != nil {
		t.Fatalf("RunE returned error: %v", err)
	}
	if !calledHealth {
		t.Fatal("expected health checker to be called")
	}

	output := buf.String()
	if !strings.Contains(output, "Health report") || !strings.Contains(output, "core.claude_dir") {
		t.Fatalf("detailed health output missing expected content, got:\n%q", output)
	}
}

func TestCheckCommand_HealthJSONOutput(t *testing.T) {
	withCheckDeps(t)

	checkJSONFlag = true
	checkHealthFlag = true

	loadConfigForCheck = func() (*config.Config, error) {
		return &config.Config{}, nil
	}

	report := checker.HealthReport{
		Scope: "claude-code",
		Summary: checker.HealthSummary{
			OK:   1,
			Warn: 1,
		},
		Checks: []checker.HealthCheck{
			{ID: "core.settings_json", Category: "core", Status: checker.HealthOK, Message: "ok"},
			{ID: "optional.output_styles", Category: "optional", Status: checker.HealthWarn, Message: "missing"},
		},
	}

	runHealthCheck = func(_ checker.HealthOptions) (checker.HealthReport, error) {
		return report, nil
	}

	buf := &bytes.Buffer{}
	checkCmd.SetOut(buf)
	defer checkCmd.SetOut(nil)

	if err := checkCmd.RunE(checkCmd, []string{}); err != nil {
		t.Fatalf("RunE returned error: %v", err)
	}

	var decoded checker.HealthReport
	if err := json.Unmarshal(buf.Bytes(), &decoded); err != nil {
		t.Fatalf("failed to decode JSON output: %v\noutput:\n%s", err, buf.String())
	}
	if decoded.Scope != "claude-code" {
		t.Fatalf("scope = %q, want claude-code", decoded.Scope)
	}
	if len(decoded.Checks) != 2 {
		t.Fatalf("checks length = %d, want 2", len(decoded.Checks))
	}
	if decoded.Summary.Warn != 1 {
		t.Fatalf("warn count = %d, want 1", decoded.Summary.Warn)
	}
}

func TestCheckCommand_HealthFailReturnsError(t *testing.T) {
	withCheckDeps(t)

	checkJSONFlag = false
	checkHealthFlag = true

	loadConfigForCheck = func() (*config.Config, error) {
		return &config.Config{}, nil
	}
	runHealthCheck = func(_ checker.HealthOptions) (checker.HealthReport, error) {
		return checker.HealthReport{
			Scope:   "claude-code",
			Summary: checker.HealthSummary{Fail: 1},
		}, nil
	}

	err := checkCmd.RunE(checkCmd, []string{})
	if !errors.Is(err, errHealthCheckFailed) {
		t.Fatalf("expected errHealthCheckFailed, got %v", err)
	}
}

func withCheckDeps(t *testing.T) {
	t.Helper()
	originalLoad := loadConfigForCheck
	originalVersion := runVersionCheck
	originalHealth := runHealthCheck
	originalJSON := checkJSONFlag
	originalHealthFlag := checkHealthFlag
	t.Cleanup(func() {
		loadConfigForCheck = originalLoad
		runVersionCheck = originalVersion
		runHealthCheck = originalHealth
		checkJSONFlag = originalJSON
		checkHealthFlag = originalHealthFlag
	})
}

package cmd

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/conpasDEVS/conpas-forge/internal/checker"
	"github.com/conpasDEVS/conpas-forge/internal/installer"
)

func TestPostInstallHealthSummary_SuccessPrintsSummary(t *testing.T) {
	originalRun := runHealthForInstall
	t.Cleanup(func() { runHealthForInstall = originalRun })

	runCalled := false
	runHealthForInstall = func(_ checker.HealthOptions) (checker.HealthReport, error) {
		runCalled = true
		return checker.HealthReport{
			Scope: "claude-code",
			Summary: checker.HealthSummary{
				OK:   1,
				Warn: 1,
			},
			Checks: []checker.HealthCheck{{
				ID:          "optional.output_styles",
				Category:    "optional",
				Status:      checker.HealthWarn,
				Message:     "missing",
				Remediation: "Optional fix",
			}},
		}, nil
	}

	buf := &bytes.Buffer{}
	results := []installer.Result{{ModuleName: "Engram", Success: true}}
	postInstallHealthSummary(buf, "C:/tmp", results, false)

	if !runCalled {
		t.Fatal("expected health engine to be called")
	}
	output := buf.String()
	if !strings.Contains(output, "Post-install health summary") || !strings.Contains(output, "Action items") {
		t.Fatalf("summary output missing expected content:\n%s", output)
	}
}

func TestPostInstallHealthSummary_SkipsWhenCancelledOrError(t *testing.T) {
	originalRun := runHealthForInstall
	t.Cleanup(func() { runHealthForInstall = originalRun })

	runCount := 0
	runHealthForInstall = func(_ checker.HealthOptions) (checker.HealthReport, error) {
		runCount++
		return checker.HealthReport{}, nil
	}

	buf := &bytes.Buffer{}
	postInstallHealthSummary(buf, "C:/tmp", []installer.Result{{Err: errors.New("boom")}}, false)
	postInstallHealthSummary(buf, "C:/tmp", []installer.Result{{Success: true}}, true)

	if runCount != 0 {
		t.Fatalf("health engine called %d times, want 0", runCount)
	}
	if buf.Len() != 0 {
		t.Fatalf("expected no output when skipped, got:\n%s", buf.String())
	}
}

func TestPostInstallHealthSummary_BestEffortErrorNote(t *testing.T) {
	originalRun := runHealthForInstall
	t.Cleanup(func() { runHealthForInstall = originalRun })

	runHealthForInstall = func(_ checker.HealthOptions) (checker.HealthReport, error) {
		return checker.HealthReport{}, errors.New("health unavailable")
	}

	buf := &bytes.Buffer{}
	postInstallHealthSummary(buf, "C:/tmp", []installer.Result{{Success: true}}, false)

	if !strings.Contains(buf.String(), "health summary unavailable") {
		t.Fatalf("expected best-effort diagnostic note, got:\n%s", buf.String())
	}
}

func TestPostInstallHealthSummary_SuccessWithFailuresPrintsConciseFailureSummary(t *testing.T) {
	originalRun := runHealthForInstall
	t.Cleanup(func() { runHealthForInstall = originalRun })

	runHealthForInstall = func(_ checker.HealthOptions) (checker.HealthReport, error) {
		return checker.HealthReport{
			Scope: "claude-code",
			Summary: checker.HealthSummary{
				OK:   1,
				Warn: 1,
				Fail: 2,
			},
			Checks: []checker.HealthCheck{
				{ID: "core.claude_dir", Category: "core", Status: checker.HealthOK, Message: "ok"},
				{ID: "optional.output_styles", Category: "optional", Status: checker.HealthWarn, Message: "missing", Remediation: "Optional fix"},
				{ID: "engram.permissions_allow", Category: "engram", Status: checker.HealthFail, Message: "missing tools", Remediation: "Run install"},
				{ID: "skills.manifest", Category: "skills", Status: checker.HealthFail, Message: "manifest invalid", Remediation: "Reinstall skills"},
			},
		}, nil
	}

	buf := &bytes.Buffer{}
	postInstallHealthSummary(buf, "C:/tmp", []installer.Result{{ModuleName: "Engram", Success: true}}, false)

	output := buf.String()
	for _, want := range []string{
		"Post-install health summary (claude-code)",
		"Totals: ok=1 warn=1 fail=2 skip=0",
		"WARN optional.output_styles",
		"FAIL engram.permissions_allow",
		"FAIL skills.manifest",
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("summary output missing %q:\n%s", want, output)
		}
	}

	if strings.Contains(output, "core.claude_dir") {
		t.Fatalf("concise summary should omit ok checks:\n%s", output)
	}
}

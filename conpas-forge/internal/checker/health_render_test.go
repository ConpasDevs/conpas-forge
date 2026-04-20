package checker

import (
	"strings"
	"testing"
)

func TestRenderDetailedHealth_IncludesSectionsAndRemediation(t *testing.T) {
	report := HealthReport{
		Scope: "claude-code",
		Summary: HealthSummary{
			OK:   1,
			Warn: 1,
			Fail: 1,
		},
		Checks: []HealthCheck{
			{ID: "core.claude_dir", Category: "core", Status: HealthOK, Message: "ok", Path: "/tmp/.claude"},
			{ID: "optional.output_styles", Category: "optional", Status: HealthWarn, Message: "missing", Remediation: "Optional: reinstall output styles."},
			{ID: "engram.binary", Category: "engram", Status: HealthFail, Message: "missing", Remediation: "Run install."},
		},
	}

	output := RenderDetailedHealth(report)

	for _, want := range []string{"[core]", "[engram]", "[optional]", "remediation: Run install.", "Summary: ok=1 warn=1 fail=1 skip=0"} {
		if !strings.Contains(output, want) {
			t.Fatalf("output missing %q\n%s", want, output)
		}
	}
}

func TestRenderConciseHealthSummary_OnlyWarnFailHighlights(t *testing.T) {
	report := HealthReport{
		Scope: "claude-code",
		Summary: HealthSummary{
			OK:   2,
			Warn: 1,
			Fail: 1,
		},
		Checks: []HealthCheck{
			{ID: "core.claude_dir", Category: "core", Status: HealthOK, Message: "ok"},
			{ID: "optional.output_styles", Category: "optional", Status: HealthWarn, Message: "missing", Remediation: "Optional fix"},
			{ID: "engram.permissions_allow", Category: "engram", Status: HealthFail, Message: "missing tools", Remediation: "Run install"},
		},
	}

	output := RenderConciseHealthSummary(report)

	if !strings.Contains(output, "Action items:") {
		t.Fatalf("concise output missing action header:\n%s", output)
	}
	if !strings.Contains(output, "WARN optional.output_styles") || !strings.Contains(output, "FAIL engram.permissions_allow") {
		t.Fatalf("concise output missing warn/fail entries:\n%s", output)
	}
	if strings.Contains(output, "core.claude_dir") {
		t.Fatalf("concise output should not include ok checks:\n%s", output)
	}
}

package checker

import (
	"fmt"
	"strings"
)

type ConciseSummary struct {
	Summary    HealthSummary
	Highlights []HealthCheck
}

func BuildConciseSummary(report HealthReport) ConciseSummary {
	highlights := make([]HealthCheck, 0)
	for _, check := range report.Checks {
		if check.Status == HealthWarn || check.Status == HealthFail {
			highlights = append(highlights, check)
		}
	}
	return ConciseSummary{Summary: report.Summary, Highlights: highlights}
}

func RenderDetailedHealth(report HealthReport) string {
	var b strings.Builder
	b.WriteString("Health report (claude-code)\n")
	b.WriteString(fmt.Sprintf("Summary: ok=%d warn=%d fail=%d skip=%d\n\n", report.Summary.OK, report.Summary.Warn, report.Summary.Fail, report.Summary.Skip))

	categories := []string{"core", "skills", "engram", "optional"}
	for _, category := range categories {
		checks := checksByCategory(report.Checks, category)
		if len(checks) == 0 {
			continue
		}
		b.WriteString(fmt.Sprintf("[%s]\n", category))
		for _, check := range checks {
			b.WriteString(fmt.Sprintf("- [%s] %s: %s\n", strings.ToUpper(string(check.Status)), check.ID, check.Message))
			if check.Path != "" {
				b.WriteString(fmt.Sprintf("  path: %s\n", check.Path))
			}
			if check.Expected != "" {
				b.WriteString(fmt.Sprintf("  expected: %s\n", check.Expected))
			}
			if check.Actual != "" {
				b.WriteString(fmt.Sprintf("  actual: %s\n", check.Actual))
			}
			if (check.Status == HealthWarn || check.Status == HealthFail) && check.Remediation != "" {
				b.WriteString(fmt.Sprintf("  remediation: %s\n", check.Remediation))
			}
		}
		b.WriteString("\n")
	}

	return strings.TrimRight(b.String(), "\n") + "\n"
}

func RenderConciseHealthSummary(report HealthReport) string {
	concise := BuildConciseSummary(report)

	var b strings.Builder
	b.WriteString("Post-install health summary (claude-code)\n")
	b.WriteString(fmt.Sprintf("Totals: ok=%d warn=%d fail=%d skip=%d\n", concise.Summary.OK, concise.Summary.Warn, concise.Summary.Fail, concise.Summary.Skip))

	if len(concise.Highlights) == 0 {
		b.WriteString("No warnings or failures detected.\n")
		return b.String()
	}

	b.WriteString("Action items:\n")
	for _, check := range concise.Highlights {
		line := fmt.Sprintf("- %s %s", strings.ToUpper(string(check.Status)), check.ID)
		if check.Remediation != "" {
			line = fmt.Sprintf("%s — %s", line, check.Remediation)
		} else {
			line = fmt.Sprintf("%s — %s", line, check.Message)
		}
		b.WriteString(line + "\n")
	}

	return b.String()
}

func checksByCategory(checks []HealthCheck, category string) []HealthCheck {
	out := make([]HealthCheck, 0)
	for _, check := range checks {
		if check.Category == category {
			out = append(out, check)
		}
	}
	return out
}

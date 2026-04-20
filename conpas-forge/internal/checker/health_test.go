package checker

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestRunHealth_TableDriven(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(t *testing.T, home string)
		assertion func(t *testing.T, report HealthReport)
	}{
		{
			name: "healthy install passes all required checks",
			setup: func(t *testing.T, home string) {
				writeHealthyFixture(t, home, true)
			},
			assertion: func(t *testing.T, report HealthReport) {
				if report.Summary.Fail != 0 {
					t.Fatalf("fail count = %d, want 0", report.Summary.Fail)
				}
				if report.Summary.Warn != 0 {
					t.Fatalf("warn count = %d, want 0", report.Summary.Warn)
				}
				mustStatus(t, report, "core.settings_json", HealthOK)
				mustStatus(t, report, "skills.manifest_artifacts", HealthOK)
				mustStatus(t, report, "engram.permissions_allow", HealthOK)
			},
		},
		{
			name: "missing settings is fail",
			setup: func(t *testing.T, home string) {
				writeHealthyFixture(t, home, true)
				mustRemove(t, filepath.Join(home, ".claude", "settings.json"))
			},
			assertion: func(t *testing.T, report HealthReport) {
				mustStatus(t, report, "core.settings_json", HealthFail)
				mustStatus(t, report, "engram.permissions_allow", HealthSkip)
				if report.Summary.Fail == 0 {
					t.Fatal("expected at least one fail")
				}
			},
		},
		{
			name: "invalid settings json is fail",
			setup: func(t *testing.T, home string) {
				writeHealthyFixture(t, home, true)
				mustWriteFile(t, filepath.Join(home, ".claude", "settings.json"), []byte("{"))
			},
			assertion: func(t *testing.T, report HealthReport) {
				mustStatus(t, report, "core.settings_json", HealthFail)
				mustStatus(t, report, "engram.permissions_allow", HealthSkip)
			},
		},
		{
			name: "partial allowlist is fail",
			setup: func(t *testing.T, home string) {
				writeHealthyFixture(t, home, true)
				settingsPath := filepath.Join(home, ".claude", "settings.json")
				settings := map[string]any{
					"permissions": map[string]any{
						"allow": []any{"mcp__engram__mem_save"},
					},
				}
				mustWriteJSON(t, settingsPath, settings)
			},
			assertion: func(t *testing.T, report HealthReport) {
				mustStatus(t, report, "engram.permissions_allow", HealthFail)
			},
		},
		{
			name: "zero-byte engram binary is fail",
			setup: func(t *testing.T, home string) {
				writeHealthyFixture(t, home, true)
				mustWriteFile(t, filepath.Join(home, ".conpas-forge", "bin", engramBinaryName()), []byte{})
			},
			assertion: func(t *testing.T, report HealthReport) {
				mustStatus(t, report, "engram.binary", HealthFail)
			},
		},
		{
			name: "invalid manifest json is fail",
			setup: func(t *testing.T, home string) {
				writeHealthyFixture(t, home, true)
				mustWriteFile(t, filepath.Join(home, ".claude", "skills", ".forge-manifest.json"), []byte("{"))
			},
			assertion: func(t *testing.T, report HealthReport) {
				mustStatus(t, report, "skills.manifest", HealthFail)
				mustStatus(t, report, "skills.manifest_artifacts", HealthSkip)
			},
		},
		{
			name: "manifest declared artifact missing is fail",
			setup: func(t *testing.T, home string) {
				writeHealthyFixture(t, home, true)
				mustRemove(t, filepath.Join(home, ".claude", "skills", "sdd-apply", "SKILL.md"))
			},
			assertion: func(t *testing.T, report HealthReport) {
				mustStatus(t, report, "skills.manifest_artifacts", HealthFail)
			},
		},
		{
			name: "optional output-styles missing is warn",
			setup: func(t *testing.T, home string) {
				writeHealthyFixture(t, home, false)
			},
			assertion: func(t *testing.T, report HealthReport) {
				mustStatus(t, report, "optional.output_styles", HealthWarn)
				if report.Summary.Fail != 0 {
					t.Fatalf("fail count = %d, want 0", report.Summary.Fail)
				}
			},
		},
		{
			name: "missing .claude yields fail + skips",
			setup: func(t *testing.T, home string) {
				bin := filepath.Join(home, ".conpas-forge", "bin")
				if err := os.MkdirAll(bin, 0o755); err != nil {
					t.Fatalf("mkdir bin: %v", err)
				}
				mustWriteFile(t, filepath.Join(bin, engramBinaryName()), []byte("binary"))
			},
			assertion: func(t *testing.T, report HealthReport) {
				mustStatus(t, report, "core.claude_dir", HealthFail)
				mustStatus(t, report, "core.claude_md_non_empty", HealthSkip)
				mustStatus(t, report, "core.settings_json", HealthSkip)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			home := t.TempDir()
			tt.setup(t, home)

			report, err := RunHealth(HealthOptions{HomeDir: home})
			if err != nil {
				t.Fatalf("RunHealth() error = %v", err)
			}

			if report.Scope != "claude-code" {
				t.Fatalf("report.Scope = %q, want claude-code", report.Scope)
			}

			tt.assertion(t, report)
		})
	}
}

func TestEngramBinaryNameMatchesOS(t *testing.T) {
	name := engramBinaryName()
	if runtime.GOOS == "windows" && name != "engram.exe" {
		t.Fatalf("windows binary name = %q, want engram.exe", name)
	}
	if runtime.GOOS != "windows" && name != "engram" {
		t.Fatalf("non-windows binary name = %q, want engram", name)
	}
}

func TestRunHealth_DoesNotProbeClaudeMCPList(t *testing.T) {
	home := t.TempDir()
	writeHealthyFixture(t, home, true)

	binDir := filepath.Join(t.TempDir(), "bin")
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		t.Fatalf("mkdir fake bin: %v", err)
	}

	probeMarker := filepath.Join(home, "claude-probe-invoked.txt")
	createFakeClaudeCLI(t, binDir, probeMarker)
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))

	if _, err := RunHealth(HealthOptions{HomeDir: home}); err != nil {
		t.Fatalf("RunHealth() error = %v", err)
	}

	if _, err := os.Stat(probeMarker); err == nil {
		t.Fatalf("expected no claude probe invocation, marker exists: %s", probeMarker)
	} else if !os.IsNotExist(err) {
		t.Fatalf("stat marker: %v", err)
	}
}

func writeHealthyFixture(t *testing.T, home string, withOutputStyles bool) {
	t.Helper()

	claudeDir := filepath.Join(home, ".claude")
	skillsDir := filepath.Join(claudeDir, "skills")
	sharedDir := filepath.Join(skillsDir, "_shared")
	binDir := filepath.Join(home, ".conpas-forge", "bin")

	for _, dir := range []string{claudeDir, skillsDir, sharedDir, binDir} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", dir, err)
		}
	}

	mustWriteFile(t, filepath.Join(claudeDir, "CLAUDE.md"), []byte("# CLAUDE\ncontent\n"))
	mustWriteFile(t, filepath.Join(binDir, engramBinaryName()), []byte("binary"))

	settings := map[string]any{
		"permissions": map[string]any{
			"allow": asAny(requiredEngramMCPTools),
		},
	}
	mustWriteJSON(t, filepath.Join(claudeDir, "settings.json"), settings)

	skills := []string{"sdd-apply", "sdd-spec", "go-testing"}
	manifest := map[string]any{"skills": skills}
	mustWriteJSON(t, filepath.Join(skillsDir, ".forge-manifest.json"), manifest)

	for _, skill := range skills {
		path := filepath.Join(skillsDir, skill, "SKILL.md")
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("mkdir skill dir: %v", err)
		}
		mustWriteFile(t, path, []byte("# skill"))
	}

	if withOutputStyles {
		outputStyles := filepath.Join(claudeDir, "output-styles")
		if err := os.MkdirAll(outputStyles, 0o755); err != nil {
			t.Fatalf("mkdir output-styles: %v", err)
		}
		mustWriteFile(t, filepath.Join(outputStyles, "default.md"), []byte("style"))
	}
}

func asAny(items []string) []any {
	out := make([]any, 0, len(items))
	for _, item := range items {
		out = append(out, item)
	}
	return out
}

func mustWriteJSON(t *testing.T, path string, v any) {
	t.Helper()
	data, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("marshal json: %v", err)
	}
	mustWriteFile(t, path, data)
}

func mustWriteFile(t *testing.T, path string, data []byte) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir parent for %s: %v", path, err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func mustRemove(t *testing.T, path string) {
	t.Helper()
	if err := os.Remove(path); err != nil {
		t.Fatalf("remove %s: %v", path, err)
	}
}

func mustStatus(t *testing.T, report HealthReport, id string, want HealthStatus) {
	t.Helper()
	for _, check := range report.Checks {
		if check.ID == id {
			if check.Status != want {
				t.Fatalf("check %s status = %s, want %s", id, check.Status, want)
			}
			return
		}
	}
	t.Fatalf("check %s not found", id)
}

func createFakeClaudeCLI(t *testing.T, binDir, marker string) {
	t.Helper()

	if runtime.GOOS == "windows" {
		script := "@echo off\r\necho invoked>\"" + marker + "\"\r\n"
		mustWriteFile(t, filepath.Join(binDir, "claude.bat"), []byte(script))
		return
	}

	scriptPath := filepath.Join(binDir, "claude")
	mustWriteFile(t, scriptPath, []byte("#!/bin/sh\necho invoked > \""+marker+"\"\n"))
	if err := os.Chmod(scriptPath, 0o755); err != nil {
		t.Fatalf("chmod fake claude script: %v", err)
	}
}

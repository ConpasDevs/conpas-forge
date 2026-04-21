package checker

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/conpasDEVS/conpas-forge/internal/engramtools"
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
					t.Fatalf("fail count = %d, want 0\nchecks: %s", report.Summary.Fail, dumpChecks(report))
				}
				if report.Summary.Warn != 0 {
					t.Fatalf("warn count = %d, want 0\nchecks: %s", report.Summary.Warn, dumpChecks(report))
				}
				mustStatus(t, report, "core.settings_json", HealthOK)
				mustStatus(t, report, "skills.manifest_artifacts", HealthOK)
				mustStatus(t, report, "engram.permissions_allow", HealthOK)
				mustStatus(t, report, "engram.mcp_registration", HealthOK)
				mustStatus(t, report, "engram.tool_name_mapping", HealthOK)
				mustStatus(t, report, "engram.settings_consistency", HealthOK)
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
				mustStatus(t, report, "engram.settings_consistency", HealthSkip)
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
				// T2.10: settings_json_prereq_skips_consistency
				mustStatus(t, report, "engram.settings_consistency", HealthSkip)
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
					t.Fatalf("fail count = %d, want 0\nchecks: %s", report.Summary.Fail, dumpChecks(report))
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
		// T2.10 — New health test table entries (16 entries)
		{
			name: "mcp_registration_missing_claude_json",
			setup: func(t *testing.T, home string) {
				writeHealthyFixture(t, home, true)
				mustRemove(t, filepath.Join(home, ".claude.json"))
			},
			assertion: func(t *testing.T, report HealthReport) {
				mustStatus(t, report, "engram.mcp_registration", HealthFail)
				check := findCheck(t, report, "engram.mcp_registration")
				if !strings.Contains(check.Message, "missing") {
					t.Errorf("message %q should mention missing", check.Message)
				}
			},
		},
		{
			name: "mcp_registration_malformed_json",
			setup: func(t *testing.T, home string) {
				writeHealthyFixture(t, home, true)
				mustWriteFile(t, filepath.Join(home, ".claude.json"), []byte("{"))
			},
			assertion: func(t *testing.T, report HealthReport) {
				mustStatus(t, report, "engram.mcp_registration", HealthFail)
				check := findCheck(t, report, "engram.mcp_registration")
				if !strings.Contains(check.Message, "invalid JSON") {
					t.Errorf("message %q should mention invalid JSON", check.Message)
				}
			},
		},
		{
			name: "mcp_registration_no_mcp_servers_key",
			setup: func(t *testing.T, home string) {
				writeHealthyFixture(t, home, true)
				writeClaudeJSON(t, home, map[string]any{})
			},
			assertion: func(t *testing.T, report HealthReport) {
				mustStatus(t, report, "engram.mcp_registration", HealthFail)
				check := findCheck(t, report, "engram.mcp_registration")
				if !strings.Contains(check.Message, "missing") {
					t.Errorf("message %q should mention missing", check.Message)
				}
			},
		},
		{
			name: "mcp_registration_no_engram_entry",
			setup: func(t *testing.T, home string) {
				writeHealthyFixture(t, home, true)
				writeClaudeJSON(t, home, map[string]any{"mcpServers": map[string]any{}})
			},
			assertion: func(t *testing.T, report HealthReport) {
				mustStatus(t, report, "engram.mcp_registration", HealthFail)
			},
		},
		{
			name: "mcp_registration_engram_not_object",
			setup: func(t *testing.T, home string) {
				writeHealthyFixture(t, home, true)
				writeClaudeJSON(t, home, map[string]any{
					"mcpServers": map[string]any{"engram": "bad-value"},
				})
			},
			assertion: func(t *testing.T, report HealthReport) {
				mustStatus(t, report, "engram.mcp_registration", HealthFail)
				check := findCheck(t, report, "engram.mcp_registration")
				if !strings.Contains(check.Message, "unexpected shape") {
					t.Errorf("message %q should mention unexpected shape", check.Message)
				}
			},
		},
		{
			name: "mcp_registration_missing_command",
			setup: func(t *testing.T, home string) {
				writeHealthyFixture(t, home, true)
				writeClaudeJSON(t, home, map[string]any{
					"mcpServers": map[string]any{
						"engram": map[string]any{
							"args": []any{"mcp", "--tools=agent"},
						},
					},
				})
			},
			assertion: func(t *testing.T, report HealthReport) {
				mustStatus(t, report, "engram.mcp_registration", HealthFail)
				check := findCheck(t, report, "engram.mcp_registration")
				if !strings.Contains(check.Message, "command") {
					t.Errorf("message %q should mention command", check.Message)
				}
			},
		},
		{
			name: "mcp_registration_missing_args",
			setup: func(t *testing.T, home string) {
				writeHealthyFixture(t, home, true)
				writeClaudeJSON(t, home, map[string]any{
					"mcpServers": map[string]any{
						"engram": map[string]any{
							"command": "/path/to/engram",
						},
					},
				})
			},
			assertion: func(t *testing.T, report HealthReport) {
				mustStatus(t, report, "engram.mcp_registration", HealthFail)
				check := findCheck(t, report, "engram.mcp_registration")
				if !strings.Contains(check.Message, "args") {
					t.Errorf("message %q should mention args", check.Message)
				}
			},
		},
		{
			name: "mcp_registration_well_formed",
			setup: func(t *testing.T, home string) {
				writeHealthyFixture(t, home, true)
				// Already written by writeHealthyFixture — just verify it passes
			},
			assertion: func(t *testing.T, report HealthReport) {
				mustStatus(t, report, "engram.mcp_registration", HealthOK)
			},
		},
		{
			name: "permissions_allow_extra_user_entries",
			setup: func(t *testing.T, home string) {
				writeHealthyFixture(t, home, true)
				// Add extra user entries on top of the 15 required tools
				allow := append(engramtools.RequiredAllowlistAsAny(), "Bash", "mcp__custom__my_tool")
				mustWriteJSON(t, filepath.Join(home, ".claude", "settings.json"), map[string]any{
					"permissions": map[string]any{"allow": allow},
				})
			},
			assertion: func(t *testing.T, report HealthReport) {
				mustStatus(t, report, "engram.permissions_allow", HealthOK)
			},
		},
		{
			name: "tool_mapping_skill_missing",
			setup: func(t *testing.T, home string) {
				writeHealthyFixture(t, home, true)
				mustRemove(t, filepath.Join(home, ".claude", "skills", "engram-memory", "SKILL.md"))
			},
			assertion: func(t *testing.T, report HealthReport) {
				mustStatus(t, report, "engram.tool_name_mapping", HealthFail)
				check := findCheck(t, report, "engram.tool_name_mapping")
				if !strings.Contains(check.Message, "missing") {
					t.Errorf("message %q should mention missing", check.Message)
				}
			},
		},
		{
			name: "tool_mapping_extra_tool_in_asset",
			setup: func(t *testing.T, home string) {
				writeHealthyFixture(t, home, true)
				// Write SKILL.md with all 15 tools + one extra
				extra := append(engramtools.RequiredAliases(), "engram_mem_nonexistent")
				mustWriteFile(t, filepath.Join(home, ".claude", "skills", "engram-memory", "SKILL.md"),
					buildEngramSkillMDFromList(extra))
			},
			assertion: func(t *testing.T, report HealthReport) {
				mustStatus(t, report, "engram.tool_name_mapping", HealthFail)
				check := findCheck(t, report, "engram.tool_name_mapping")
				if !strings.Contains(check.Message, "not in canonical catalog") {
					t.Errorf("message %q should mention catalog", check.Message)
				}
			},
		},
		{
			name: "tool_mapping_catalog_tool_missing_from_asset",
			setup: func(t *testing.T, home string) {
				writeHealthyFixture(t, home, true)
				// Write SKILL.md with one tool removed
				aliases := engramtools.RequiredAliases()
				reduced := make([]string, 0, len(aliases)-1)
				for _, a := range aliases {
					if a != "engram_mem_timeline" {
						reduced = append(reduced, a)
					}
				}
				mustWriteFile(t, filepath.Join(home, ".claude", "skills", "engram-memory", "SKILL.md"),
					buildEngramSkillMDFromList(reduced))
			},
			assertion: func(t *testing.T, report HealthReport) {
				mustStatus(t, report, "engram.tool_name_mapping", HealthFail)
				check := findCheck(t, report, "engram.tool_name_mapping")
				if !strings.Contains(check.Message, "missing catalog tools") {
					t.Errorf("message %q should mention missing catalog tools", check.Message)
				}
			},
		},
		{
			name: "tool_mapping_no_declarations_parseable",
			setup: func(t *testing.T, home string) {
				writeHealthyFixture(t, home, true)
				// Write an empty SKILL.md
				mustWriteFile(t, filepath.Join(home, ".claude", "skills", "engram-memory", "SKILL.md"), []byte("# Empty skill\n"))
			},
			assertion: func(t *testing.T, report HealthReport) {
				mustStatus(t, report, "engram.tool_name_mapping", HealthWarn)
			},
		},
		{
			name: "settings_consistency_legacy_mcpservers",
			setup: func(t *testing.T, home string) {
				writeHealthyFixture(t, home, true)
				// Write settings with legacy mcpServers.engram
				allow := engramtools.RequiredAllowlistAsAny()
				mustWriteJSON(t, filepath.Join(home, ".claude", "settings.json"), map[string]any{
					"permissions": map[string]any{"allow": allow},
					"mcpServers": map[string]any{
						"engram": map[string]any{
							"command": "/old/path/engram",
							"args":    []any{"mcp"},
						},
					},
				})
			},
			assertion: func(t *testing.T, report HealthReport) {
				mustStatus(t, report, "engram.settings_consistency", HealthWarn)
			},
		},
		// Skip chain validations
		{
			// settings_json_prereq_skips_consistency
			name: "settings_json_prereq_skips_consistency",
			setup: func(t *testing.T, home string) {
				writeHealthyFixture(t, home, true)
				mustWriteFile(t, filepath.Join(home, ".claude", "settings.json"), []byte("{"))
			},
			assertion: func(t *testing.T, report HealthReport) {
				mustStatus(t, report, "core.settings_json", HealthFail)
				mustStatus(t, report, "engram.settings_consistency", HealthSkip)
			},
		},
		{
			// skills_dir_prereq_skips_mapping
			name: "skills_dir_prereq_skips_mapping",
			setup: func(t *testing.T, home string) {
				writeHealthyFixture(t, home, true)
				// Remove the entire skills directory
				if err := os.RemoveAll(filepath.Join(home, ".claude", "skills")); err != nil {
					t.Fatalf("remove skills dir: %v", err)
				}
			},
			assertion: func(t *testing.T, report HealthReport) {
				mustStatus(t, report, "skills.dir", HealthFail)
				mustStatus(t, report, "engram.tool_name_mapping", HealthSkip)
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

	// Snapshot ~/.claude.json mtime before RunHealth
	claudeJSONPath := filepath.Join(home, ".claude.json")
	beforeStat, err := os.Stat(claudeJSONPath)
	if err != nil {
		t.Fatalf("stat ~/.claude.json before RunHealth: %v", err)
	}
	beforeMtime := beforeStat.ModTime()

	if _, err := RunHealth(HealthOptions{HomeDir: home}); err != nil {
		t.Fatalf("RunHealth() error = %v", err)
	}

	// Assert no claude probe invocation
	if _, err := os.Stat(probeMarker); err == nil {
		t.Fatalf("expected no claude probe invocation, marker exists: %s", probeMarker)
	} else if !os.IsNotExist(err) {
		t.Fatalf("stat marker: %v", err)
	}

	// Assert ~/.claude.json was NOT written during health
	afterStat, err := os.Stat(claudeJSONPath)
	if err != nil {
		t.Fatalf("stat ~/.claude.json after RunHealth: %v", err)
	}
	if afterStat.ModTime() != beforeMtime {
		t.Fatalf("~/.claude.json was modified during RunHealth — health must be read-only")
	}
}

// T2.9: writeHealthyFixture extended to write ~/.claude.json and SKILL.md.
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

	binaryPath := filepath.Join(binDir, engramBinaryName())
	mustWriteFile(t, binaryPath, []byte("binary"))

	settings := map[string]any{
		"permissions": map[string]any{
			"allow": engramtools.RequiredAllowlistAsAny(),
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

	// Write engram-memory SKILL.md with all catalog aliases
	engramSkillDir := filepath.Join(skillsDir, "engram-memory")
	if err := os.MkdirAll(engramSkillDir, 0o755); err != nil {
		t.Fatalf("mkdir engram skill dir: %v", err)
	}
	mustWriteFile(t, filepath.Join(engramSkillDir, "SKILL.md"), buildEngramSkillMD())

	// Write ~/.claude.json with valid mcpServers.engram
	writeClaudeJSON(t, home, map[string]any{
		"mcpServers": map[string]any{
			"engram": map[string]any{
				"command": binaryPath,
				"args":    []any{"mcp", "--tools=agent"},
			},
		},
	})

	if withOutputStyles {
		outputStyles := filepath.Join(claudeDir, "output-styles")
		if err := os.MkdirAll(outputStyles, 0o755); err != nil {
			t.Fatalf("mkdir output-styles: %v", err)
		}
		mustWriteFile(t, filepath.Join(outputStyles, "default.md"), []byte("style"))
	}
}

// T2.9: writeClaudeJSON marshals content to ~/.claude.json.
func writeClaudeJSON(t *testing.T, home string, content any) {
	t.Helper()
	data, err := json.Marshal(content)
	if err != nil {
		t.Fatalf("marshal claude.json: %v", err)
	}
	mustWriteFile(t, filepath.Join(home, ".claude.json"), data)
}

// T2.9: buildEngramSkillMD generates SKILL.md with all 15 catalog aliases.
func buildEngramSkillMD() []byte {
	return buildEngramSkillMDFromList(engramtools.RequiredAliases())
}

// buildEngramSkillMDFromList generates SKILL.md bullet section from a given list.
func buildEngramSkillMDFromList(aliases []string) []byte {
	var sb strings.Builder
	sb.WriteString("## Engram Tools\n\n")
	for _, alias := range aliases {
		sb.WriteString(fmt.Sprintf("- **%s** — tool description\n", alias))
	}
	return []byte(sb.String())
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
				t.Fatalf("check %s status = %s, want %s\nchecks:\n%s", id, check.Status, want, dumpChecks(report))
			}
			return
		}
	}
	t.Fatalf("check %s not found\nchecks:\n%s", id, dumpChecks(report))
}

func findCheck(t *testing.T, report HealthReport, id string) HealthCheck {
	t.Helper()
	for _, check := range report.Checks {
		if check.ID == id {
			return check
		}
	}
	t.Fatalf("check %s not found", id)
	return HealthCheck{}
}

func dumpChecks(report HealthReport) string {
	var sb strings.Builder
	for _, c := range report.Checks {
		sb.WriteString(fmt.Sprintf("  [%s] %s: %s\n", c.Status, c.ID, c.Message))
	}
	return sb.String()
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

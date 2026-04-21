// Tests mutate config.HomeDir global; do not t.Parallel().
package installer

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/conpasDEVS/conpas-forge/internal/config"
	"github.com/conpasDEVS/conpas-forge/internal/engramtools"
)

// setupRepairHome creates a temp home, overrides the HomeDir global,
// and returns a cleanup function that restores the original home.
func setupRepairHome(t *testing.T) string {
	t.Helper()
	home := t.TempDir()
	oldHome := config.HomeDir()
	config.OverrideHomeDir(home)
	t.Cleanup(func() { config.OverrideHomeDir(oldHome) })
	return home
}

// writeEngramBinary creates a non-empty fake engram binary in the expected location.
func writeEngramBinary(t *testing.T, home string) string {
	t.Helper()
	binDir := filepath.Join(home, ".conpas-forge", "bin")
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		t.Fatalf("mkdir bin: %v", err)
	}
	binaryName := "engram"
	if runtime.GOOS == "windows" {
		binaryName = "engram.exe"
	}
	binaryPath := filepath.Join(binDir, binaryName)
	if err := os.WriteFile(binaryPath, []byte("binary"), 0o755); err != nil {
		t.Fatalf("write fake engram binary: %v", err)
	}
	return binaryPath
}

// createFakeClaudeForRepair writes a fake claude script that:
//   - Exits 0 for any invocation (simulates successful mcp add/remove).
//   - Writes a marker file capturing the full argv for assertions.
//   - On "mcp add": writes a minimal ~/.claude.json with mcpServers.engram.
func createFakeClaudeForRepair(t *testing.T, binDir, home, markerPath string) {
	t.Helper()
	claudeJSONPath := filepath.Join(home, ".claude.json")

	if runtime.GOOS == "windows" {
		// On Windows, use a .bat that writes a marker and creates claude.json on "mcp add".
		// Uses PowerShell inline to avoid findstr dependency issues.
		script := "@echo off\r\n" +
			"echo %* >> \"" + markerPath + "\"\r\n" +
			"if \"%2\"==\"add\" (\r\n" +
			"  echo {\"mcpServers\":{\"engram\":{\"command\":\"fake\",\"args\":[\"mcp\",\"--tools=agent\"]}}} > \"" + claudeJSONPath + "\"\r\n" +
			")\r\n" +
			"exit /b 0\r\n"
		if err := os.WriteFile(filepath.Join(binDir, "claude.bat"), []byte(script), 0o644); err != nil {
			t.Fatalf("write fake claude.bat: %v", err)
		}
		return
	}

	// Unix shell script
	script := "#!/bin/sh\n" +
		"echo \"$@\" >> \"" + markerPath + "\"\n" +
		"if echo \"$*\" | grep -q 'add'; then\n" +
		"  printf '{\"mcpServers\":{\"engram\":{\"command\":\"fake\",\"args\":[\"mcp\",\"--tools=agent\"]}}}' > \"" + claudeJSONPath + "\"\n" +
		"fi\n" +
		"exit 0\n"
	scriptPath := filepath.Join(binDir, "claude")
	if err := os.WriteFile(scriptPath, []byte(script), 0o755); err != nil {
		t.Fatalf("write fake claude script: %v", err)
	}
}

// writeSettingsJSON writes a settings.json with the given allow list.
func writeSettingsJSON(t *testing.T, home string, allow []any) {
	t.Helper()
	claudeDir := filepath.Join(home, ".claude")
	if err := os.MkdirAll(claudeDir, 0o755); err != nil {
		t.Fatalf("mkdir .claude: %v", err)
	}
	settings := map[string]any{
		"permissions": map[string]any{
			"allow": allow,
		},
	}
	data, err := json.Marshal(settings)
	if err != nil {
		t.Fatalf("marshal settings: %v", err)
	}
	if err := os.WriteFile(filepath.Join(claudeDir, "settings.json"), data, 0o644); err != nil {
		t.Fatalf("write settings.json: %v", err)
	}
}

func TestRepairEngram_MissingMCPEntry(t *testing.T) {
	home := setupRepairHome(t)

	// Setup: binary present, no ~/.claude.json
	writeEngramBinary(t, home)

	// Fake claude binary
	binDir := filepath.Join(t.TempDir(), "bin")
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		t.Fatalf("mkdir fake bin: %v", err)
	}
	markerPath := filepath.Join(home, "claude-argv.txt")
	createFakeClaudeForRepair(t, binDir, home, markerPath)
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))

	rep, err := RepairEngram(context.Background())
	if err != nil {
		t.Fatalf("RepairEngram() error = %v", err)
	}

	if !rep.MCPRegistered {
		t.Errorf("MCPRegistered = false, want true")
	}

	// Verify fake claude was invoked (marker file should exist)
	if _, statErr := os.Stat(markerPath); statErr != nil {
		t.Errorf("fake claude marker not found — was claude not invoked? %v", statErr)
	}

	// Verify marker contains "add" (mcp add was called)
	markerData, readErr := os.ReadFile(markerPath)
	if readErr == nil && !strings.Contains(string(markerData), "add") {
		t.Errorf("fake claude marker %q does not contain 'add' — mcp add may not have been called", string(markerData))
	}
}

func TestRepairEngram_MissingBinary(t *testing.T) {
	_ = setupRepairHome(t)

	// No binary at all — RepairEngram should return an error
	rep, err := RepairEngram(context.Background())
	if err == nil {
		t.Fatalf("RepairEngram() expected error for missing binary, got nil")
	}
	if !strings.Contains(err.Error(), "Engram binary not found") {
		t.Errorf("error %q should mention 'Engram binary not found'", err.Error())
	}
	if rep.MCPRegistered {
		t.Errorf("MCPRegistered = true, want false when binary is missing")
	}
}

func TestRepairEngram_MissingAllowlistToolsUnion(t *testing.T) {
	home := setupRepairHome(t)

	// Setup: binary present, settings with partial allowlist + user entry
	writeEngramBinary(t, home)
	writeSettingsJSON(t, home, []any{"Bash", "mcp__engram__mem_save"})

	// Fake claude binary
	binDir := filepath.Join(t.TempDir(), "bin")
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		t.Fatalf("mkdir fake bin: %v", err)
	}
	markerPath := filepath.Join(home, "claude-argv.txt")
	createFakeClaudeForRepair(t, binDir, home, markerPath)
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))

	rep, err := RepairEngram(context.Background())
	if err != nil {
		t.Fatalf("RepairEngram() error = %v", err)
	}

	// Should have added 14 missing entries (1 was already there)
	if rep.AllowlistAdded != 14 {
		t.Errorf("AllowlistAdded = %d, want 14", rep.AllowlistAdded)
	}

	// Read back settings.json and verify union semantics
	data, err := os.ReadFile(filepath.Join(home, ".claude", "settings.json"))
	if err != nil {
		t.Fatalf("read settings.json after repair: %v", err)
	}
	var root map[string]any
	if err := json.Unmarshal(data, &root); err != nil {
		t.Fatalf("parse settings.json after repair: %v", err)
	}

	perms, _ := root["permissions"].(map[string]any)
	allow, _ := perms["allow"].([]any)

	// Build set for easy lookup
	allowSet := make(map[string]struct{})
	for _, v := range allow {
		if s, ok := v.(string); ok {
			allowSet[s] = struct{}{}
		}
	}

	// User entry "Bash" must be preserved
	if _, ok := allowSet["Bash"]; !ok {
		t.Errorf("user entry 'Bash' was removed — union semantics violated")
	}

	// All 15 catalog tools must be present
	for _, tool := range engramtools.RequiredAllowlist() {
		if _, ok := allowSet[tool]; !ok {
			t.Errorf("catalog tool %q missing after repair", tool)
		}
	}
}

func TestRepairEngram_AllToolsPresentNoop(t *testing.T) {
	home := setupRepairHome(t)

	// Setup: binary present, all 15 tools already in allowlist
	writeEngramBinary(t, home)
	writeSettingsJSON(t, home, engramtools.RequiredAllowlistAsAny())

	// Fake claude binary
	binDir := filepath.Join(t.TempDir(), "bin")
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		t.Fatalf("mkdir fake bin: %v", err)
	}
	markerPath := filepath.Join(home, "claude-argv.txt")
	createFakeClaudeForRepair(t, binDir, home, markerPath)
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))

	rep, err := RepairEngram(context.Background())
	if err != nil {
		t.Fatalf("RepairEngram() error = %v", err)
	}

	if rep.AllowlistAdded != 0 {
		t.Errorf("AllowlistAdded = %d, want 0 (no-op)", rep.AllowlistAdded)
	}

	summary := RenderRepairSummary(rep)
	if !strings.Contains(summary, "No allowlist changes needed") {
		t.Errorf("summary %q should mention 'No allowlist changes needed'", summary)
	}
}

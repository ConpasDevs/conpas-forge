package installer

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/conpasDEVS/conpas-forge/internal/config"
	"github.com/conpasDEVS/conpas-forge/internal/engramtools"
)

// RepairReport summarizes what the Engram repair did.
type RepairReport struct {
	MCPRegistered        bool     `json:"mcp_registered"`
	MCPAlreadyRegistered bool     `json:"mcp_already_registered"`
	AllowlistAdded       int      `json:"allowlist_added"`
	Errors               []string `json:"errors,omitempty"`
}

// RepairEngram performs config-only self-heal:
//  1. Verifies engram binary exists at config.EngramBinary(); aborts if not.
//  2. Re-registers MCP via registerEngramMCP (idempotent — remove-then-add).
//  3. Merges missing catalog allowlist entries via Merge() using union semantics.
//
// It does NOT download or replace the binary, and does NOT touch SKILL.md.
// Returns a RepairReport describing actions taken, plus an error if repair
// could not proceed (e.g. missing binary).
func RepairEngram(ctx context.Context) (RepairReport, error) {
	rep := RepairReport{}

	binaryPath := config.EngramBinary()
	stat, err := os.Stat(binaryPath)
	if err != nil || stat.IsDir() || stat.Size() == 0 {
		return rep, fmt.Errorf("Engram binary not found at %s — run 'conpas-forge install' first", binaryPath)
	}

	// (1) Re-register MCP — idempotent (remove-then-add).
	if err := registerEngramMCP(ctx, binaryPath); err != nil {
		return rep, fmt.Errorf("register engram MCP: %w", err)
	}
	rep.MCPRegistered = true

	// (2) Snapshot allowlist before Merge.
	before := currentAllowlistSet()

	entry := map[string]any{
		"permissions": map[string]any{
			"allow": engramtools.RequiredAllowlistAsAny(),
		},
	}
	if err := Merge(entry); err != nil {
		rep.Errors = append(rep.Errors, fmt.Sprintf("permissions.allow merge: %v", err))
		// Non-fatal: MCP repair already succeeded.
		return rep, nil
	}

	after := currentAllowlistSet()
	rep.AllowlistAdded = countNewEntries(before, after, engramtools.RequiredAllowlistSet())
	return rep, nil
}

// RenderRepairSummary returns a human-readable one-block summary for stdout.
func RenderRepairSummary(r RepairReport) string {
	var sb strings.Builder

	if r.MCPRegistered {
		sb.WriteString("Engram MCP server registered.\n")
	} else {
		sb.WriteString("Engram MCP server already registered.\n")
	}

	if r.AllowlistAdded > 0 {
		sb.WriteString(fmt.Sprintf("Added %d missing allowlist entries.\n", r.AllowlistAdded))
	} else {
		sb.WriteString("No allowlist changes needed.\n")
	}

	for _, e := range r.Errors {
		sb.WriteString(fmt.Sprintf("Warning: %s\n", e))
	}

	return sb.String()
}

// currentAllowlistSet reads the current permissions.allow from settings.json
// and returns it as a set. Returns an empty set on any read/parse error.
func currentAllowlistSet() map[string]struct{} {
	data, err := os.ReadFile(config.SettingsJSON())
	if err != nil {
		return make(map[string]struct{})
	}
	var root map[string]any
	if err := json.Unmarshal(data, &root); err != nil {
		return make(map[string]struct{})
	}
	out := make(map[string]struct{})
	perms, ok := root["permissions"].(map[string]any)
	if !ok {
		return out
	}
	allow, ok := perms["allow"].([]any)
	if !ok {
		return out
	}
	for _, raw := range allow {
		if s, ok := raw.(string); ok && s != "" {
			out[s] = struct{}{}
		}
	}
	return out
}

// countNewEntries returns the number of required entries that exist in after
// but did not exist in before.
func countNewEntries(before, after map[string]struct{}, required map[string]struct{}) int {
	count := 0
	for k := range required {
		_, wasBefore := before[k]
		_, isAfter := after[k]
		if !wasBefore && isAfter {
			count++
		}
	}
	return count
}

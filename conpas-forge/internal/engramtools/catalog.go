package engramtools

import (
	"regexp"
	"strings"
)

// Tool represents one Engram MCP tool with its allowlist-form and runtime-alias names.
type Tool struct {
	// Suffix is the common short name shared between allowlist and alias forms,
	// e.g. "mem_save", "mem_context".
	Suffix string
	// Allowlist is the name written to ~/.claude/settings.json permissions.allow.
	// Format: "mcp__engram__<suffix>".
	Allowlist string
	// Alias is the name the agent calls at runtime, as declared in SKILL.md.
	// Format: "engram_<suffix>".
	Alias string
}

// MCPAllowlistPrefix is the Claude Code MCP allowlist prefix for the engram server.
const MCPAllowlistPrefix = "mcp__engram__"

// AliasPrefix is the skill-facing runtime prefix for engram tools.
const AliasPrefix = "engram_"

// catalog is the canonical, ordered list of Engram MCP tools.
// Order matters for deterministic output (allowlist ordering, diagnostics).
var catalog = []Tool{
	{Suffix: "mem_capture_passive"},
	{Suffix: "mem_context"},
	{Suffix: "mem_delete"},
	{Suffix: "mem_get_observation"},
	{Suffix: "mem_merge_projects"},
	{Suffix: "mem_save"},
	{Suffix: "mem_save_prompt"},
	{Suffix: "mem_search"},
	{Suffix: "mem_session_end"},
	{Suffix: "mem_session_start"},
	{Suffix: "mem_session_summary"},
	{Suffix: "mem_stats"},
	{Suffix: "mem_suggest_topic_key"},
	{Suffix: "mem_timeline"},
	{Suffix: "mem_update"},
}

// init materializes Allowlist and Alias fields from Suffix — keeps the
// literal above minimal and eliminates duplication at the row level.
func init() {
	for i := range catalog {
		catalog[i].Allowlist = MCPAllowlistPrefix + catalog[i].Suffix
		catalog[i].Alias = AliasPrefix + catalog[i].Suffix
	}
}

// Catalog returns a copy of the canonical tool list. Callers must not mutate it.
func Catalog() []Tool {
	return append([]Tool(nil), catalog...)
}

// RequiredAllowlist returns the allowlist-form names ("mcp__engram__*") in canonical order.
func RequiredAllowlist() []string {
	out := make([]string, len(catalog))
	for i, t := range catalog {
		out[i] = t.Allowlist
	}
	return out
}

// RequiredAliases returns the runtime-alias names ("engram_*") in canonical order.
func RequiredAliases() []string {
	out := make([]string, len(catalog))
	for i, t := range catalog {
		out[i] = t.Alias
	}
	return out
}

// RequiredAllowlistAsAny returns the allowlist names as []any — installer shim
// to keep Merge() callsite unchanged (Merge expects map[string]any with []any values).
func RequiredAllowlistAsAny() []any {
	out := make([]any, len(catalog))
	for i, t := range catalog {
		out[i] = t.Allowlist
	}
	return out
}

// RequiredAllowlistSet returns the allowlist names as a set for subset checks.
func RequiredAllowlistSet() map[string]struct{} {
	out := make(map[string]struct{}, len(catalog))
	for _, t := range catalog {
		out[t.Allowlist] = struct{}{}
	}
	return out
}

// RequiredAliasSet returns the runtime aliases as a set for equality checks.
func RequiredAliasSet() map[string]struct{} {
	out := make(map[string]struct{}, len(catalog))
	for _, t := range catalog {
		out[t.Alias] = struct{}{}
	}
	return out
}

// AllowlistToAlias maps "mcp__engram__<suffix>" -> "engram_<suffix>".
// Returns ("", false) if input does not carry the allowlist prefix.
func AllowlistToAlias(allowlist string) (string, bool) {
	if !strings.HasPrefix(allowlist, MCPAllowlistPrefix) {
		return "", false
	}
	return AliasPrefix + strings.TrimPrefix(allowlist, MCPAllowlistPrefix), true
}

// AliasToAllowlist maps "engram_<suffix>" -> "mcp__engram__<suffix>".
// Returns ("", false) if input does not carry the alias prefix.
func AliasToAllowlist(alias string) (string, bool) {
	if !strings.HasPrefix(alias, AliasPrefix) {
		return "", false
	}
	return MCPAllowlistPrefix + strings.TrimPrefix(alias, AliasPrefix), true
}

// skillToolBulletRE matches lines of the form:
//
//	"- **engram_<suffix>**"  (optionally followed by spaces, em-dash, description)
//
// Anchored to line start (possibly preceded by whitespace), requires the
// bullet marker "- **", requires the name to start with "engram_", and
// requires the closing "**". Anything after the closing "**" is ignored.
// The (?m) flag makes ^ match start-of-line.
var skillToolBulletRE = regexp.MustCompile(`(?m)^\s*-\s+\*\*(engram_[A-Za-z0-9_]+)\*\*`)

// ParseSkillToolNames extracts tool aliases from SKILL.md content.
// Duplicates are deduplicated; order follows first occurrence.
// Returns an empty slice when no lines match the bold-bullet pattern.
func ParseSkillToolNames(content []byte) []string {
	matches := skillToolBulletRE.FindAllSubmatch(content, -1)
	seen := make(map[string]struct{})
	out := make([]string, 0, len(matches))
	for _, m := range matches {
		name := string(m[1])
		if _, ok := seen[name]; ok {
			continue
		}
		seen[name] = struct{}{}
		out = append(out, name)
	}
	return out
}

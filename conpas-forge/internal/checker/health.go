package checker

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"github.com/conpasDEVS/conpas-forge/internal/engramtools"
)

type HealthStatus string

const (
	HealthOK   HealthStatus = "ok"
	HealthWarn HealthStatus = "warn"
	HealthFail HealthStatus = "fail"
	HealthSkip HealthStatus = "skip"
)

type HealthCheck struct {
	ID          string       `json:"id"`
	Category    string       `json:"category"`
	Status      HealthStatus `json:"status"`
	Message     string       `json:"message"`
	Path        string       `json:"path,omitempty"`
	Expected    string       `json:"expected,omitempty"`
	Actual      string       `json:"actual,omitempty"`
	Remediation string       `json:"remediation,omitempty"`
}

type HealthSummary struct {
	OK   int `json:"ok"`
	Warn int `json:"warn"`
	Fail int `json:"fail"`
	Skip int `json:"skip"`
}

type HealthReport struct {
	Scope   string        `json:"scope"`
	Summary HealthSummary `json:"summary"`
	Checks  []HealthCheck `json:"checks"`
}

type HealthOptions struct {
	HomeDir string
}

type forgeManifest struct {
	Skills []string `json:"skills"`
}

// claudeJSONRoot is the minimal parsing struct for ~/.claude.json.
// Only the fields we validate are declared; all other keys are preserved
// implicitly by virtue of never being re-serialized (health is read-only).
type claudeJSONRoot struct {
	MCPServers map[string]json.RawMessage `json:"mcpServers"`
}

// claudeJSONMCPEntry is a decoded MCP server entry.
type claudeJSONMCPEntry struct {
	Command string    `json:"command"`
	// Args is a pointer to distinguish "key absent" from "key present but empty array"
	// nil => missing; non-nil (even if empty) => present.
	Args    *[]string `json:"args"`
}

// Remediation string constants — two canonical conventions:
// install-path fixes vs explicit-repair fixes.
const (
	remediationInstallEngram        = "Run 'conpas-forge install' to register the Engram MCP server."
	remediationInstallRestoreMCP    = "Run 'conpas-forge install' to restore the Engram MCP registration."
	remediationInstallRefreshAllow  = "Re-run 'conpas-forge install' to refresh permissions.allow."
	remediationInstallRedeploySkill = "Run 'conpas-forge install' to redeploy the Engram skill asset."
	remediationInstallMigrateMCP    = "Run 'conpas-forge install' to migrate to the correct registration location."
	remediationInstallFixClaudeJSON = "Fix ~/.claude.json JSON syntax or run 'conpas-forge install' to regenerate."
	remediationCheckPermsReinstall  = "Check file permissions and re-run 'conpas-forge install'."
	remediationRepairMCP            = "Run 'conpas-forge check --health --repair' to heal MCP registration."
	remediationRepairAllowlist      = "Run 'conpas-forge check --health --repair' to refresh permissions.allow."
)

func RunHealth(opts HealthOptions) (HealthReport, error) {
	homeDir := opts.HomeDir
	if homeDir == "" {
		resolved, err := os.UserHomeDir()
		if err != nil {
			return HealthReport{}, fmt.Errorf("resolve home directory: %w", err)
		}
		homeDir = resolved
	}

	claudeDir := filepath.Join(homeDir, ".claude")
	claudeMDPath := filepath.Join(claudeDir, "CLAUDE.md")
	settingsPath := filepath.Join(claudeDir, "settings.json")
	skillsDir := filepath.Join(claudeDir, "skills")
	sharedSkillsDir := filepath.Join(skillsDir, "_shared")
	manifestPath := filepath.Join(skillsDir, ".forge-manifest.json")
	outputStylesDir := filepath.Join(claudeDir, "output-styles")
	engramBinaryPath := filepath.Join(homeDir, ".conpas-forge", "bin", engramBinaryName())
	claudeJSONPath := filepath.Join(homeDir, ".claude.json")
	skillMDPath := filepath.Join(skillsDir, "engram-memory", "SKILL.md")

	checks := make([]HealthCheck, 0, 19)

	claudeExists, claudeIsDir, err := pathExistsAsDir(claudeDir)
	if err != nil {
		checks = append(checks, failCheck(
			"core.claude_dir",
			"core",
			"unable to read ~/.claude",
			claudeDir,
			"readable directory",
			err.Error(),
			"Verify directory permissions and re-run conpas-forge install.",
		))
	} else if !claudeExists {
		checks = append(checks, failCheck(
			"core.claude_dir",
			"core",
			"~/.claude is missing",
			claudeDir,
			"directory exists",
			"missing",
			"Run conpas-forge install to provision Claude Code artifacts.",
		))
	} else if !claudeIsDir {
		checks = append(checks, failCheck(
			"core.claude_dir",
			"core",
			"~/.claude exists but is not a directory",
			claudeDir,
			"directory",
			"file",
			"Replace ~/.claude with a directory and re-run conpas-forge install.",
		))
	} else {
		checks = append(checks, okCheck("core.claude_dir", "core", "~/.claude directory exists", claudeDir))
	}

	settingsCheck, settingsParsed, settingsAllowSet, settingsRoot := evaluateSettingsCheck(claudeExists && claudeIsDir, settingsPath)
	checks = append(checks, settingsCheck)

	if !(claudeExists && claudeIsDir) {
		checks = append(checks,
			skipCheck("core.claude_md_non_empty", "core", "skipped because ~/.claude prerequisite failed", claudeMDPath),
			skipCheck("skills.dir", "skills", "skipped because ~/.claude prerequisite failed", skillsDir),
			skipCheck("skills.shared_dir", "skills", "skipped because ~/.claude prerequisite failed", sharedSkillsDir),
			skipCheck("skills.manifest", "skills", "skipped because ~/.claude prerequisite failed", manifestPath),
			skipCheck("skills.manifest_artifacts", "skills", "skipped because manifest prerequisite failed", skillsDir),
			skipCheck("engram.tool_name_mapping", "engram", "skipped because ~/.claude prerequisite failed", skillMDPath),
			skipCheck("engram.permissions_allow", "engram", "skipped because settings.json prerequisite failed", settingsPath),
			skipCheck("engram.settings_consistency", "engram", "skipped because settings.json prerequisite failed", settingsPath),
			skipCheck("engram.mcp_registration", "engram", "skipped because ~/.claude prerequisite failed", claudeJSONPath),
			skipCheck("optional.output_styles", "optional", "skipped because ~/.claude prerequisite failed", outputStylesDir),
		)
		checks = append(checks, evaluateEngramBinaryCheck(engramBinaryPath))
		return buildReport(checks), nil
	}

	checks = append(checks, evaluateClaudeMDCheck(claudeMDPath))

	skillsExists, skillsIsDir, skillsErr := pathExistsAsDir(skillsDir)
	if skillsErr != nil {
		checks = append(checks, failCheck(
			"skills.dir",
			"skills",
			"unable to read ~/.claude/skills",
			skillsDir,
			"readable directory",
			skillsErr.Error(),
			"Verify directory permissions and re-run conpas-forge install.",
		))
	} else if !skillsExists {
		checks = append(checks, failCheck(
			"skills.dir",
			"skills",
			"skills directory is missing",
			skillsDir,
			"directory exists",
			"missing",
			"Run conpas-forge install to deploy skills.",
		))
	} else if !skillsIsDir {
		checks = append(checks, failCheck(
			"skills.dir",
			"skills",
			"skills path exists but is not a directory",
			skillsDir,
			"directory",
			"file",
			"Replace ~/.claude/skills with a directory and re-run install.",
		))
	} else {
		checks = append(checks, okCheck("skills.dir", "skills", "skills directory exists", skillsDir))
	}

	if skillsExists && skillsIsDir {
		checks = append(checks, evaluateSharedSkillsCheck(sharedSkillsDir))
		manifestCheck, manifestParsed, manifest := evaluateManifestCheck(manifestPath)
		checks = append(checks, manifestCheck)
		if manifestParsed {
			checks = append(checks, evaluateManifestArtifactsCheck(skillsDir, manifest))
		} else {
			checks = append(checks, skipCheck(
				"skills.manifest_artifacts",
				"skills",
				"skipped because manifest prerequisite failed",
				skillsDir,
			))
		}
		// T2.4: tool_name_mapping runs only when skills dir is present
		checks = append(checks, evaluateEngramToolNameMapping(skillMDPath))
	} else {
		checks = append(checks,
			skipCheck("skills.shared_dir", "skills", "skipped because skills directory prerequisite failed", sharedSkillsDir),
			skipCheck("skills.manifest", "skills", "skipped because skills directory prerequisite failed", manifestPath),
			skipCheck("skills.manifest_artifacts", "skills", "skipped because manifest prerequisite failed", skillsDir),
			skipCheck("engram.tool_name_mapping", "engram", "skipped because skills directory prerequisite failed", skillMDPath),
		)
	}

	checks = append(checks, evaluateEngramBinaryCheck(engramBinaryPath))

	// T2.3: MCP registration — no prerequisite on core.claude_dir (runs whenever homeDir resolved)
	checks = append(checks, evaluateEngramMCPRegistration(claudeJSONPath))

	if settingsParsed {
		checks = append(checks, evaluateEngramAllowlistCheck(settingsPath, settingsAllowSet))
		// T2.5: settings consistency — depends on parsed settings root
		checks = append(checks, evaluateEngramSettingsConsistency(settingsPath, settingsRoot))
	} else {
		checks = append(checks,
			skipCheck("engram.permissions_allow", "engram", "skipped because settings.json prerequisite failed", settingsPath),
			skipCheck("engram.settings_consistency", "engram", "skipped because settings.json prerequisite failed", settingsPath),
		)
	}

	checks = append(checks, evaluateOutputStylesCheck(outputStylesDir))

	return buildReport(checks), nil
}

func evaluateClaudeMDCheck(claudeMDPath string) HealthCheck {
	data, err := os.ReadFile(claudeMDPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return failCheck(
				"core.claude_md_non_empty",
				"core",
				"CLAUDE.md is missing",
				claudeMDPath,
				"non-empty file",
				"missing",
				"Re-run conpas-forge install to generate CLAUDE.md.",
			)
		}
		return failCheck(
			"core.claude_md_non_empty",
			"core",
			"unable to read CLAUDE.md",
			claudeMDPath,
			"readable non-empty file",
			err.Error(),
			"Check file permissions and re-run install if needed.",
		)
	}
	if strings.TrimSpace(string(data)) == "" {
		return failCheck(
			"core.claude_md_non_empty",
			"core",
			"CLAUDE.md is empty",
			claudeMDPath,
			"non-empty content",
			"empty",
			"Re-run conpas-forge install to regenerate CLAUDE.md.",
		)
	}
	return okCheck("core.claude_md_non_empty", "core", "CLAUDE.md exists and is non-empty", claudeMDPath)
}

// T2.1: Updated signature — returns fourth value: parsed settings root map[string]any.
func evaluateSettingsCheck(prereq bool, settingsPath string) (HealthCheck, bool, map[string]struct{}, map[string]any) {
	if !prereq {
		return skipCheck("core.settings_json", "core", "skipped because ~/.claude prerequisite failed", settingsPath), false, nil, nil
	}
	data, err := os.ReadFile(settingsPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return failCheck(
				"core.settings_json",
				"core",
				"settings.json is missing",
				settingsPath,
				"valid JSON file",
				"missing",
				"Run conpas-forge install to generate settings.json.",
			), false, nil, nil
		}
		return failCheck(
			"core.settings_json",
			"core",
			"unable to read settings.json",
			settingsPath,
			"readable valid JSON file",
			err.Error(),
			"Check file permissions and validate ~/.claude/settings.json.",
		), false, nil, nil
	}

	var root map[string]any
	if err := json.Unmarshal(data, &root); err != nil {
		return failCheck(
			"core.settings_json",
			"core",
			"settings.json contains invalid JSON",
			settingsPath,
			"valid JSON",
			err.Error(),
			"Fix JSON syntax or re-run conpas-forge install.",
		), false, nil, nil
	}

	allowSet := readAllowSet(root)
	return okCheck("core.settings_json", "core", "settings.json exists and parses as valid JSON", settingsPath), true, allowSet, root
}

func evaluateSharedSkillsCheck(sharedSkillsDir string) HealthCheck {
	exists, isDir, err := pathExistsAsDir(sharedSkillsDir)
	if err != nil {
		return failCheck(
			"skills.shared_dir",
			"skills",
			"unable to read skills/_shared",
			sharedSkillsDir,
			"readable directory",
			err.Error(),
			"Verify permissions and re-run install.",
		)
	}
	if !exists {
		return failCheck(
			"skills.shared_dir",
			"skills",
			"skills/_shared is missing",
			sharedSkillsDir,
			"directory exists",
			"missing",
			"Re-run conpas-forge install to deploy shared skill assets.",
		)
	}
	if !isDir {
		return failCheck(
			"skills.shared_dir",
			"skills",
			"skills/_shared exists but is not a directory",
			sharedSkillsDir,
			"directory",
			"file",
			"Replace skills/_shared with a directory and re-run install.",
		)
	}
	return okCheck("skills.shared_dir", "skills", "skills/_shared directory exists", sharedSkillsDir)
}

func evaluateManifestCheck(manifestPath string) (HealthCheck, bool, forgeManifest) {
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return failCheck(
				"skills.manifest",
				"skills",
				"skills manifest is missing",
				manifestPath,
				"valid .forge-manifest.json",
				"missing",
				"Re-run conpas-forge install to regenerate the skills manifest.",
			), false, forgeManifest{}
		}
		return failCheck(
			"skills.manifest",
			"skills",
			"unable to read skills manifest",
			manifestPath,
			"readable valid JSON",
			err.Error(),
			"Verify file permissions and re-run install.",
		), false, forgeManifest{}
	}

	var manifest forgeManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return failCheck(
			"skills.manifest",
			"skills",
			"skills manifest contains invalid JSON",
			manifestPath,
			"valid JSON with skills list",
			err.Error(),
			"Fix manifest JSON or re-run conpas-forge install.",
		), false, forgeManifest{}
	}

	return okCheck("skills.manifest", "skills", "skills manifest exists and parses", manifestPath), true, manifest
}

func evaluateManifestArtifactsCheck(skillsDir string, manifest forgeManifest) HealthCheck {
	if len(manifest.Skills) == 0 {
		return warnCheck(
			"skills.manifest_artifacts",
			"skills",
			"manifest has no declared skills",
			skillsDir,
			"non-empty skills list",
			"empty",
			"Re-run conpas-forge install to redeploy and refresh manifest entries.",
		)
	}

	missing := make([]string, 0)
	for _, skill := range manifest.Skills {
		skill = strings.TrimSpace(skill)
		if skill == "" {
			continue
		}
		target := filepath.Join(skillsDir, skill, "SKILL.md")
		if _, err := os.Stat(target); err != nil {
			missing = append(missing, target)
		}
	}

	if len(missing) > 0 {
		sort.Strings(missing)
		preview := missing[0]
		if len(missing) > 1 {
			preview = fmt.Sprintf("%s (+%d more)", preview, len(missing)-1)
		}
		return failCheck(
			"skills.manifest_artifacts",
			"skills",
			"manifest-declared artifacts missing on disk",
			skillsDir,
			"all manifest skills have SKILL.md",
			preview,
			"Re-run conpas-forge install to redeploy missing skill artifacts.",
		)
	}

	return okCheck("skills.manifest_artifacts", "skills", "all manifest-declared skill artifacts are present", skillsDir)
}

// T2.3: evaluateEngramMCPRegistration parses ~/.claude.json and validates the
// mcpServers.engram entry has the expected shape. No subprocess, no network.
func evaluateEngramMCPRegistration(claudeJSONPath string) HealthCheck {
	data, err := os.ReadFile(claudeJSONPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return failCheck(
				"engram.mcp_registration",
				"engram",
				"~/.claude.json is missing",
				claudeJSONPath,
				"file exists with mcpServers.engram entry",
				"missing",
				remediationInstallEngram,
			)
		}
		return failCheck(
			"engram.mcp_registration",
			"engram",
			"unable to read ~/.claude.json",
			claudeJSONPath,
			"readable file",
			err.Error(),
			remediationCheckPermsReinstall,
		)
	}

	var root claudeJSONRoot
	if err := json.Unmarshal(data, &root); err != nil {
		return failCheck(
			"engram.mcp_registration",
			"engram",
			"~/.claude.json contains invalid JSON",
			claudeJSONPath,
			"valid JSON with mcpServers.engram entry",
			err.Error(),
			remediationInstallFixClaudeJSON,
		)
	}

	if root.MCPServers == nil {
		return failCheck(
			"engram.mcp_registration",
			"engram",
			"~/.claude.json mcpServers.engram entry is missing",
			claudeJSONPath,
			"mcpServers.engram entry present",
			"missing",
			remediationInstallEngram,
		)
	}

	rawEntry, ok := root.MCPServers["engram"]
	if !ok {
		return failCheck(
			"engram.mcp_registration",
			"engram",
			"~/.claude.json mcpServers.engram entry is missing",
			claudeJSONPath,
			"mcpServers.engram entry present",
			"missing",
			remediationInstallEngram,
		)
	}

	// Detect non-object value before attempting struct unmarshal (A-REG-06)
	trimmed := strings.TrimSpace(string(rawEntry))
	if len(trimmed) == 0 || trimmed[0] != '{' {
		return failCheck(
			"engram.mcp_registration",
			"engram",
			"~/.claude.json mcpServers.engram has unexpected shape",
			claudeJSONPath,
			"object with command and args fields",
			"non-object value",
			remediationInstallRestoreMCP,
		)
	}

	var entry claudeJSONMCPEntry
	if err := json.Unmarshal(rawEntry, &entry); err != nil {
		return failCheck(
			"engram.mcp_registration",
			"engram",
			"~/.claude.json mcpServers.engram has unexpected shape",
			claudeJSONPath,
			"object with command and args fields",
			err.Error(),
			remediationInstallRestoreMCP,
		)
	}

	if strings.TrimSpace(entry.Command) == "" {
		return failCheck(
			"engram.mcp_registration",
			"engram",
			"~/.claude.json mcpServers.engram.command is missing or empty",
			claudeJSONPath,
			"non-empty command field",
			"missing or empty",
			remediationInstallRestoreMCP,
		)
	}

	if entry.Args == nil {
		return failCheck(
			"engram.mcp_registration",
			"engram",
			"~/.claude.json mcpServers.engram.args is missing or not an array",
			claudeJSONPath,
			"args array field",
			"missing",
			remediationInstallRestoreMCP,
		)
	}

	return okCheck("engram.mcp_registration", "engram", "~/.claude.json mcpServers.engram entry is well-formed", claudeJSONPath)
}

// T2.4: evaluateEngramToolNameMapping reads the engram-memory SKILL.md and validates
// that the declared tool aliases match the canonical catalog exactly.
func evaluateEngramToolNameMapping(skillMDPath string) HealthCheck {
	data, err := os.ReadFile(skillMDPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return failCheck(
				"engram.tool_name_mapping",
				"engram",
				"engram-memory skill asset is missing",
				skillMDPath,
				"SKILL.md with all canonical tool aliases",
				"missing",
				remediationInstallRedeploySkill,
			)
		}
		return failCheck(
			"engram.tool_name_mapping",
			"engram",
			"unable to read engram-memory skill asset",
			skillMDPath,
			"readable SKILL.md",
			err.Error(),
			remediationInstallRedeploySkill,
		)
	}

	names := engramtools.ParseSkillToolNames(data)
	if len(names) == 0 {
		return warnCheck(
			"engram.tool_name_mapping",
			"engram",
			"engram-memory skill asset declares no tools — unable to validate mapping",
			skillMDPath,
			"tool declarations matching canonical catalog",
			"no tool declarations found",
			remediationInstallRedeploySkill,
		)
	}

	// Build parsed set
	parsed := make(map[string]struct{}, len(names))
	for _, n := range names {
		parsed[n] = struct{}{}
	}

	required := engramtools.RequiredAliasSet()

	// Extras in asset but not in catalog
	extras := make([]string, 0)
	for n := range parsed {
		if _, ok := required[n]; !ok {
			extras = append(extras, n)
		}
	}

	// Missing in asset but in catalog
	missing := make([]string, 0)
	for n := range required {
		if _, ok := parsed[n]; !ok {
			missing = append(missing, n)
		}
	}

	if len(extras) > 0 {
		sort.Strings(extras)
		preview := extras[0]
		if len(extras) > 1 {
			preview = fmt.Sprintf("%s (+%d more)", preview, len(extras)-1)
		}
		return failCheck(
			"engram.tool_name_mapping",
			"engram",
			"engram-memory skill asset declares tools not in canonical catalog",
			skillMDPath,
			"tool aliases matching canonical catalog exactly",
			preview,
			remediationInstallRedeploySkill,
		)
	}

	if len(missing) > 0 {
		sort.Strings(missing)
		preview := missing[0]
		if len(missing) > 1 {
			preview = fmt.Sprintf("%s (+%d more)", preview, len(missing)-1)
		}
		return failCheck(
			"engram.tool_name_mapping",
			"engram",
			"engram-memory skill asset is missing catalog tools",
			skillMDPath,
			"tool aliases matching canonical catalog exactly",
			preview,
			remediationInstallRedeploySkill,
		)
	}

	return okCheck("engram.tool_name_mapping", "engram", "engram-memory skill asset tool aliases match canonical catalog", skillMDPath)
}

// T2.5: evaluateEngramSettingsConsistency inspects the already-parsed settings.json
// root for legacy Engram MCP keys that indicate a partial or pre-migration installation.
func evaluateEngramSettingsConsistency(settingsPath string, settingsRoot map[string]any) HealthCheck {
	if mcpServers, ok := settingsRoot["mcpServers"].(map[string]any); ok {
		if _, hasEngram := mcpServers["engram"]; hasEngram {
			return warnCheck(
				"engram.settings_consistency",
				"engram",
				"settings.json contains an mcpServers.engram entry — MCP registration should live in ~/.claude.json",
				settingsPath,
				"no mcpServers.engram in settings.json",
				"mcpServers.engram present",
				remediationInstallMigrateMCP,
			)
		}
	}
	return okCheck("engram.settings_consistency", "engram", "settings.json has no legacy Engram MCP keys", settingsPath)
}

func evaluateEngramBinaryCheck(binaryPath string) HealthCheck {
	info, err := os.Stat(binaryPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return failCheck(
				"engram.binary",
				"engram",
				"Engram binary is missing",
				binaryPath,
				"existing binary with size > 0",
				"missing",
				"Run conpas-forge install to install Engram.",
			)
		}
		return failCheck(
			"engram.binary",
			"engram",
			"unable to read Engram binary",
			binaryPath,
			"readable binary with size > 0",
			err.Error(),
			"Check file permissions or reinstall Engram via conpas-forge install.",
		)
	}

	if info.IsDir() {
		return failCheck(
			"engram.binary",
			"engram",
			"Engram binary path points to a directory",
			binaryPath,
			"binary file",
			"directory",
			"Re-run conpas-forge install to restore Engram binary.",
		)
	}

	if info.Size() <= 0 {
		return failCheck(
			"engram.binary",
			"engram",
			"Engram binary is empty",
			binaryPath,
			"size > 0",
			"size=0",
			"Re-run conpas-forge install to reinstall Engram binary.",
		)
	}

	return okCheck("engram.binary", "engram", "Engram binary exists and is non-empty", binaryPath)
}

// T2.6: evaluateEngramAllowlistCheck now consumes engramtools.RequiredAllowlist().
func evaluateEngramAllowlistCheck(settingsPath string, allowSet map[string]struct{}) HealthCheck {
	missing := make([]string, 0)
	for _, tool := range engramtools.RequiredAllowlist() {
		if _, ok := allowSet[tool]; !ok {
			missing = append(missing, tool)
		}
	}

	if len(missing) > 0 {
		sort.Strings(missing)
		preview := missing[0]
		if len(missing) > 1 {
			preview = fmt.Sprintf("%s (+%d more)", preview, len(missing)-1)
		}
		return failCheck(
			"engram.permissions_allow",
			"engram",
			"settings.json permissions.allow is missing required Engram tools",
			settingsPath,
			"contains all required Engram MCP tools",
			preview,
			remediationInstallRefreshAllow,
		)
	}

	return okCheck("engram.permissions_allow", "engram", "settings.json permissions.allow includes required Engram tools", settingsPath)
}

func evaluateOutputStylesCheck(outputStylesDir string) HealthCheck {
	exists, isDir, err := pathExistsAsDir(outputStylesDir)
	if err != nil {
		return warnCheck(
			"optional.output_styles",
			"optional",
			"unable to read output-styles directory",
			outputStylesDir,
			"directory available",
			err.Error(),
			"Re-run conpas-forge install if you rely on output styles.",
		)
	}
	if !exists {
		return warnCheck(
			"optional.output_styles",
			"optional",
			"output-styles directory is missing",
			outputStylesDir,
			"directory exists",
			"missing",
			"Optional: re-run conpas-forge install to deploy output styles.",
		)
	}
	if !isDir {
		return warnCheck(
			"optional.output_styles",
			"optional",
			"output-styles path exists but is not a directory",
			outputStylesDir,
			"directory",
			"file",
			"Optional: replace with a directory by re-running install.",
		)
	}
	return okCheck("optional.output_styles", "optional", "output-styles directory exists", outputStylesDir)
}

func buildReport(checks []HealthCheck) HealthReport {
	summary := HealthSummary{}
	for _, c := range checks {
		switch c.Status {
		case HealthOK:
			summary.OK++
		case HealthWarn:
			summary.Warn++
		case HealthFail:
			summary.Fail++
		case HealthSkip:
			summary.Skip++
		}
	}
	return HealthReport{
		Scope:   "claude-code",
		Summary: summary,
		Checks:  checks,
	}
}

func pathExistsAsDir(path string) (exists bool, isDir bool, err error) {
	info, err := os.Stat(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return false, false, nil
		}
		return false, false, err
	}
	return true, info.IsDir(), nil
}

func readAllowSet(root map[string]any) map[string]struct{} {
	out := make(map[string]struct{})
	permissions, ok := root["permissions"].(map[string]any)
	if !ok {
		return out
	}
	allow, ok := permissions["allow"].([]any)
	if !ok {
		return out
	}
	for _, raw := range allow {
		item, ok := raw.(string)
		if !ok {
			continue
		}
		item = strings.TrimSpace(item)
		if item != "" {
			out[item] = struct{}{}
		}
	}
	return out
}

func okCheck(id, category, message, path string) HealthCheck {
	return HealthCheck{ID: id, Category: category, Status: HealthOK, Message: message, Path: path}
}

func warnCheck(id, category, message, path, expected, actual, remediation string) HealthCheck {
	return HealthCheck{
		ID:          id,
		Category:    category,
		Status:      HealthWarn,
		Message:     message,
		Path:        path,
		Expected:    expected,
		Actual:      actual,
		Remediation: remediation,
	}
}

func failCheck(id, category, message, path, expected, actual, remediation string) HealthCheck {
	return HealthCheck{
		ID:          id,
		Category:    category,
		Status:      HealthFail,
		Message:     message,
		Path:        path,
		Expected:    expected,
		Actual:      actual,
		Remediation: remediation,
	}
}

func skipCheck(id, category, message, path string) HealthCheck {
	return HealthCheck{ID: id, Category: category, Status: HealthSkip, Message: message, Path: path}
}

func engramBinaryName() string {
	if runtime.GOOS == "windows" {
		return "engram.exe"
	}
	return "engram"
}

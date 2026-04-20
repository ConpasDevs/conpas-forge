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

// Keep in sync with installer/engram.go engramMCPTools.
var requiredEngramMCPTools = []string{
	"mcp__engram__mem_capture_passive",
	"mcp__engram__mem_context",
	"mcp__engram__mem_delete",
	"mcp__engram__mem_get_observation",
	"mcp__engram__mem_merge_projects",
	"mcp__engram__mem_save",
	"mcp__engram__mem_save_prompt",
	"mcp__engram__mem_search",
	"mcp__engram__mem_session_end",
	"mcp__engram__mem_session_start",
	"mcp__engram__mem_session_summary",
	"mcp__engram__mem_stats",
	"mcp__engram__mem_suggest_topic_key",
	"mcp__engram__mem_timeline",
	"mcp__engram__mem_update",
}

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

	checks := make([]HealthCheck, 0, 16)

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

	settingsCheck, settingsParsed, settingsAllowSet := evaluateSettingsCheck(claudeExists && claudeIsDir, settingsPath)
	checks = append(checks, settingsCheck)

	if !(claudeExists && claudeIsDir) {
		checks = append(checks,
			skipCheck("core.claude_md_non_empty", "core", "skipped because ~/.claude prerequisite failed", claudeMDPath),
			skipCheck("skills.dir", "skills", "skipped because ~/.claude prerequisite failed", skillsDir),
			skipCheck("skills.shared_dir", "skills", "skipped because ~/.claude prerequisite failed", sharedSkillsDir),
			skipCheck("skills.manifest", "skills", "skipped because ~/.claude prerequisite failed", manifestPath),
			skipCheck("skills.manifest_artifacts", "skills", "skipped because manifest prerequisite failed", skillsDir),
			skipCheck("engram.permissions_allow", "engram", "skipped because settings.json prerequisite failed", settingsPath),
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
	} else {
		checks = append(checks,
			skipCheck("skills.shared_dir", "skills", "skipped because skills directory prerequisite failed", sharedSkillsDir),
			skipCheck("skills.manifest", "skills", "skipped because skills directory prerequisite failed", manifestPath),
			skipCheck("skills.manifest_artifacts", "skills", "skipped because manifest prerequisite failed", skillsDir),
		)
	}

	checks = append(checks, evaluateEngramBinaryCheck(engramBinaryPath))

	if settingsParsed {
		checks = append(checks, evaluateEngramAllowlistCheck(settingsPath, settingsAllowSet))
	} else {
		checks = append(checks, skipCheck(
			"engram.permissions_allow",
			"engram",
			"skipped because settings.json prerequisite failed",
			settingsPath,
		))
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

func evaluateSettingsCheck(prereq bool, settingsPath string) (HealthCheck, bool, map[string]struct{}) {
	if !prereq {
		return skipCheck("core.settings_json", "core", "skipped because ~/.claude prerequisite failed", settingsPath), false, nil
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
			), false, nil
		}
		return failCheck(
			"core.settings_json",
			"core",
			"unable to read settings.json",
			settingsPath,
			"readable valid JSON file",
			err.Error(),
			"Check file permissions and validate ~/.claude/settings.json.",
		), false, nil
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
		), false, nil
	}

	allowSet := readAllowSet(root)
	return okCheck("core.settings_json", "core", "settings.json exists and parses as valid JSON", settingsPath), true, allowSet
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

func evaluateEngramAllowlistCheck(settingsPath string, allowSet map[string]struct{}) HealthCheck {
	missing := make([]string, 0)
	for _, tool := range requiredEngramMCPTools {
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
			"Re-run conpas-forge install to refresh permissions.allow.",
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

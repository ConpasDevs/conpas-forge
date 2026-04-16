package installer

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/conpasDEVS/conpas-forge/internal/config"
)

// ─── test doubles ─────────────────────────────────────────────────────────────

type fakeInstaller struct {
	name   string
	skills []string
	fail   bool
}

func (f *fakeInstaller) Name() string { return f.name }

func (f *fakeInstaller) ExpectedSkills() []string { return f.skills }

func (f *fakeInstaller) Install(_ context.Context, _ *InstallOptions, _ func(ProgressEvent)) Result {
	if f.fail {
		return Result{ModuleName: f.name, Success: false, Err: errors.New("simulated install failure")}
	}
	return Result{ModuleName: f.name, Success: true}
}

// ─── collectExpected ──────────────────────────────────────────────────────────

func TestCollectExpected(t *testing.T) {
	modules := []Module{
		&fakeInstaller{name: "a", skills: []string{"skill-a"}},
		&fakeInstaller{name: "b", skills: []string{"skill-b", "skill-c"}},
	}
	got := collectExpected(modules)
	if len(got) != 3 {
		t.Fatalf("collectExpected() = %v, want 3 skills", got)
	}
}

// ─── RunPipeline integration: stale skill removed ─────────────────────────────

func TestRunPipeline_RemovesStaleSkill(t *testing.T) {
	homeDir := t.TempDir()
	old := config.HomeDir()
	config.OverrideHomeDir(homeDir)
	defer config.OverrideHomeDir(old)

	// Pre-create a manifest that tracks "old-skill".
	manifestPath := config.ForgeManifest()
	if err := os.MkdirAll(filepath.Dir(manifestPath), 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	data, _ := json.Marshal(ForgeManifest{Skills: []string{"old-skill", "skill-a"}})
	if err := os.WriteFile(manifestPath, data, 0644); err != nil {
		t.Fatalf("write manifest: %v", err)
	}

	// Pre-create the stale skill directory.
	staleDir := config.SkillDir("old-skill")
	if err := os.MkdirAll(staleDir, 0755); err != nil {
		t.Fatalf("mkdir stale dir: %v", err)
	}

	// Run the pipeline with an installer that only knows "skill-a".
	modules := []Module{
		&fakeInstaller{name: "test", skills: []string{"skill-a"}},
	}
	cfg := config.DefaultConfig()
	results := RunPipeline(context.Background(), modules, &InstallOptions{Config: &cfg}, nil)

	if HasErrors(results) {
		t.Fatalf("RunPipeline() returned errors: %v", results[0].Err)
	}

	// Stale dir must be gone.
	if _, err := os.Stat(staleDir); !os.IsNotExist(err) {
		t.Fatal("expected stale skill dir to be removed after RunPipeline")
	}

	// Manifest must be updated to reflect only "skill-a".
	m, err := ReadManifest(manifestPath)
	if err != nil {
		t.Fatalf("ReadManifest after run: %v", err)
	}
	if fmt.Sprint(m.Skills) != fmt.Sprint([]string{"skill-a"}) {
		t.Fatalf("manifest skills = %v, want [skill-a]", m.Skills)
	}
}

// ─── RunPipeline integration: manifest NOT updated on failure ─────────────────

func TestRunPipeline_ManifestNotUpdatedOnError(t *testing.T) {
	homeDir := t.TempDir()
	old := config.HomeDir()
	config.OverrideHomeDir(homeDir)
	defer config.OverrideHomeDir(old)

	// Pre-create a manifest with known skills.
	manifestPath := config.ForgeManifest()
	if err := os.MkdirAll(filepath.Dir(manifestPath), 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	original := []string{"skill-x"}
	data, _ := json.Marshal(ForgeManifest{Skills: original})
	if err := os.WriteFile(manifestPath, data, 0644); err != nil {
		t.Fatalf("write manifest: %v", err)
	}

	// Run pipeline with a failing installer that reports different skills.
	modules := []Module{
		&fakeInstaller{name: "test", skills: []string{"skill-y"}, fail: true},
	}
	cfg := config.DefaultConfig()
	results := RunPipeline(context.Background(), modules, &InstallOptions{Config: &cfg}, nil)

	if !HasErrors(results) {
		t.Fatal("expected errors from failing installer")
	}

	// Manifest must remain unchanged.
	m, err := ReadManifest(manifestPath)
	if err != nil {
		t.Fatalf("ReadManifest after failed run: %v", err)
	}
	if fmt.Sprint(m.Skills) != fmt.Sprint(original) {
		t.Fatalf("manifest was mutated on failure: got %v, want %v", m.Skills, original)
	}
}

// ─── RunPipeline integration: no pre-existing manifest ────────────────────────

// R1.1 / R5.1: When NO manifest file exists before the pipeline runs, it must
// be created after a successful install so future runs can detect stale skills.
func TestRunPipeline_NoManifest_CreatesManifest(t *testing.T) {
	homeDir := t.TempDir()
	old := config.HomeDir()
	config.OverrideHomeDir(homeDir)
	defer config.OverrideHomeDir(old)

	// Intentionally do NOT create a manifest file.
	manifestPath := config.ForgeManifest()

	modules := []Module{
		&fakeInstaller{name: "test", skills: []string{"skill-a", "skill-b"}},
	}
	cfg := config.DefaultConfig()
	results := RunPipeline(context.Background(), modules, &InstallOptions{Config: &cfg}, nil)

	if HasErrors(results) {
		t.Fatalf("RunPipeline() unexpected error: %v", results[0].Err)
	}

	// Manifest must now exist and contain the expected skills.
	m, err := ReadManifest(manifestPath)
	if err != nil {
		t.Fatalf("ReadManifest after first install: %v", err)
	}
	if fmt.Sprint(m.Skills) != fmt.Sprint([]string{"skill-a", "skill-b"}) {
		t.Fatalf("manifest skills = %v, want [skill-a skill-b]", m.Skills)
	}
}

// ─── RunPipeline integration: corrupt manifest ────────────────────────────────

// R5.2: A corrupt manifest must not crash the pipeline. The run should succeed
// and write a fresh, valid manifest.
func TestRunPipeline_CorruptManifest_Succeeds(t *testing.T) {
	homeDir := t.TempDir()
	old := config.HomeDir()
	config.OverrideHomeDir(homeDir)
	defer config.OverrideHomeDir(old)

	// Write an intentionally broken manifest.
	manifestPath := config.ForgeManifest()
	if err := os.MkdirAll(filepath.Dir(manifestPath), 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(manifestPath, []byte(`{corrupt json!!!`), 0644); err != nil {
		t.Fatalf("write corrupt manifest: %v", err)
	}

	modules := []Module{
		&fakeInstaller{name: "test", skills: []string{"skill-a"}},
	}
	cfg := config.DefaultConfig()
	results := RunPipeline(context.Background(), modules, &InstallOptions{Config: &cfg}, nil)

	if HasErrors(results) {
		t.Fatalf("RunPipeline() with corrupt manifest returned error: %v", results[0].Err)
	}

	// After a successful run the manifest must be valid and up to date.
	m, err := ReadManifest(manifestPath)
	if err != nil {
		t.Fatalf("ReadManifest after run: %v", err)
	}
	if fmt.Sprint(m.Skills) != fmt.Sprint([]string{"skill-a"}) {
		t.Fatalf("manifest skills = %v, want [skill-a]", m.Skills)
	}
}

// ─── RunPipeline integration: user skill preservation ─────────────────────────

// R6.1: A skill directory created by the user (not tracked in any manifest)
// must survive the pipeline — it is never in the stale list.
func TestRunPipeline_UserSkillPreserved(t *testing.T) {
	homeDir := t.TempDir()
	old := config.HomeDir()
	config.OverrideHomeDir(homeDir)
	defer config.OverrideHomeDir(old)

	// Pre-create a manifest that only tracks "skill-a".
	manifestPath := config.ForgeManifest()
	if err := os.MkdirAll(filepath.Dir(manifestPath), 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	data, _ := json.Marshal(ForgeManifest{Skills: []string{"skill-a"}})
	if err := os.WriteFile(manifestPath, data, 0644); err != nil {
		t.Fatalf("write manifest: %v", err)
	}

	// Create a user-managed skill dir that is NOT in any manifest.
	userSkillDir := config.SkillDir("user-custom-skill")
	if err := os.MkdirAll(userSkillDir, 0755); err != nil {
		t.Fatalf("mkdir user skill: %v", err)
	}

	modules := []Module{
		&fakeInstaller{name: "test", skills: []string{"skill-a"}},
	}
	cfg := config.DefaultConfig()
	results := RunPipeline(context.Background(), modules, &InstallOptions{Config: &cfg}, nil)

	if HasErrors(results) {
		t.Fatalf("RunPipeline() unexpected error: %v", results[0].Err)
	}

	// User skill dir must still exist — it was never stale.
	if _, err := os.Stat(userSkillDir); err != nil {
		t.Fatalf("user-custom-skill dir was unexpectedly removed: %v", err)
	}
}

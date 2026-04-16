package installer

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

// ─── CalculateStale ────────────────────────────────────────────────────────────

func TestCalculateStale(t *testing.T) {
	tests := []struct {
		name     string
		manifest []string
		expected []string
		want     []string
	}{
		{
			name:     "empty manifest produces no stale",
			manifest: nil,
			expected: []string{"skill-a", "skill-b"},
			want:     nil,
		},
		{
			name:     "skill removed from expected is stale",
			manifest: []string{"skill-a", "skill-b"},
			expected: []string{"skill-a"},
			want:     []string{"skill-b"},
		},
		{
			name:     "all same skills produces no stale",
			manifest: []string{"skill-a", "skill-b"},
			expected: []string{"skill-a", "skill-b"},
			want:     nil,
		},
		{
			name:     "new skill added does not produce stale",
			manifest: []string{"skill-a"},
			expected: []string{"skill-a", "skill-b"},
			want:     nil,
		},
		{
			name:     "multiple stale skills detected",
			manifest: []string{"skill-a", "skill-b", "skill-c"},
			expected: []string{"skill-a"},
			want:     []string{"skill-b", "skill-c"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &ForgeManifest{Skills: tt.manifest}
			got := CalculateStale(m, tt.expected)
			if len(got) != len(tt.want) {
				t.Fatalf("CalculateStale() = %v, want %v", got, tt.want)
			}
			for i, s := range got {
				if s != tt.want[i] {
					t.Fatalf("CalculateStale()[%d] = %q, want %q", i, s, tt.want[i])
				}
			}
		})
	}
}

// ─── ReadManifest ─────────────────────────────────────────────────────────────

func TestReadManifest(t *testing.T) {
	t.Run("missing file returns empty manifest", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, ".forge-manifest.json")

		m, err := ReadManifest(path)
		if err != nil {
			t.Fatalf("ReadManifest() unexpected error: %v", err)
		}
		if len(m.Skills) != 0 {
			t.Fatalf("expected empty manifest, got %v", m.Skills)
		}
	})

	t.Run("corrupt JSON returns error", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, ".forge-manifest.json")
		if err := os.WriteFile(path, []byte(`{invalid json`), 0644); err != nil {
			t.Fatalf("setup: %v", err)
		}

		_, err := ReadManifest(path)
		if err == nil {
			t.Fatal("ReadManifest() expected error for corrupt JSON, got nil")
		}
	})

	t.Run("valid JSON returns manifest", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, ".forge-manifest.json")
		data, _ := json.Marshal(ForgeManifest{Skills: []string{"skill-a", "skill-b"}})
		if err := os.WriteFile(path, data, 0644); err != nil {
			t.Fatalf("setup: %v", err)
		}

		m, err := ReadManifest(path)
		if err != nil {
			t.Fatalf("ReadManifest() unexpected error: %v", err)
		}
		if len(m.Skills) != 2 || m.Skills[0] != "skill-a" || m.Skills[1] != "skill-b" {
			t.Fatalf("ReadManifest() = %v, want [skill-a skill-b]", m.Skills)
		}
	})
}

// ─── WriteManifest ────────────────────────────────────────────────────────────

func TestWriteManifest(t *testing.T) {
	t.Run("creates file with correct JSON content", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, ".forge-manifest.json")

		skills := []string{"sdd-init", "go-testing"}
		if err := WriteManifest(path, skills); err != nil {
			t.Fatalf("WriteManifest() error: %v", err)
		}

		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read manifest file: %v", err)
		}
		var got ForgeManifest
		if err := json.Unmarshal(data, &got); err != nil {
			t.Fatalf("parse manifest: %v", err)
		}
		if len(got.Skills) != 2 || got.Skills[0] != "sdd-init" || got.Skills[1] != "go-testing" {
			t.Fatalf("manifest skills = %v, want %v", got.Skills, skills)
		}
	})

	t.Run("overwrites existing file on second call", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, ".forge-manifest.json")

		if err := WriteManifest(path, []string{"skill-a"}); err != nil {
			t.Fatalf("first WriteManifest() error: %v", err)
		}
		if err := WriteManifest(path, []string{"skill-b", "skill-c"}); err != nil {
			t.Fatalf("second WriteManifest() error: %v", err)
		}

		data, _ := os.ReadFile(path)
		var got ForgeManifest
		if err := json.Unmarshal(data, &got); err != nil {
			t.Fatalf("parse manifest: %v", err)
		}
		if len(got.Skills) != 2 || got.Skills[0] != "skill-b" {
			t.Fatalf("manifest skills = %v, want [skill-b skill-c]", got.Skills)
		}
	})
}

// ─── CleanupStale ─────────────────────────────────────────────────────────────

func TestCleanupStale(t *testing.T) {
	t.Run("stale dir is removed", func(t *testing.T) {
		dir := t.TempDir()
		staleDir := filepath.Join(dir, "old-skill")
		if err := os.MkdirAll(staleDir, 0755); err != nil {
			t.Fatalf("setup: %v", err)
		}

		if err := CleanupStale(dir, []string{"old-skill"}); err != nil {
			t.Fatalf("CleanupStale() error: %v", err)
		}
		if _, err := os.Stat(staleDir); !os.IsNotExist(err) {
			t.Fatal("expected stale dir to be removed")
		}
	})

	t.Run("user-created dir not in stale list is untouched", func(t *testing.T) {
		dir := t.TempDir()
		userDir := filepath.Join(dir, "user-skill")
		if err := os.MkdirAll(userDir, 0755); err != nil {
			t.Fatalf("setup: %v", err)
		}

		if err := CleanupStale(dir, []string{"other-skill"}); err != nil {
			t.Fatalf("CleanupStale() error: %v", err)
		}
		if _, err := os.Stat(userDir); err != nil {
			t.Fatalf("expected user dir to remain, got: %v", err)
		}
	})

	t.Run("nonexistent stale dir does not cause error", func(t *testing.T) {
		dir := t.TempDir()
		// os.RemoveAll on nonexistent path returns nil, so this must not fail.
		if err := CleanupStale(dir, []string{"ghost-skill"}); err != nil {
			t.Fatalf("CleanupStale() error: %v", err)
		}
	})

	t.Run("removal failure is logged and skipped — other stale dirs still cleaned", func(t *testing.T) {
		dir := t.TempDir()
		// Create two stale dirs.
		for _, name := range []string{"fail-skill", "ok-skill"} {
			if err := os.MkdirAll(filepath.Join(dir, name), 0755); err != nil {
				t.Fatalf("setup: %v", err)
			}
		}
		// Inject a removeAll that fails for "fail-skill".
		removeFn := func(path string) error {
			if filepath.Base(path) == "fail-skill" {
				return fmt.Errorf("simulated permission error")
			}
			return os.RemoveAll(path)
		}

		err := cleanupStaleWith(dir, []string{"fail-skill", "ok-skill"}, removeFn)
		if err != nil {
			t.Fatalf("cleanupStaleWith() returned error: %v (expected nil — errors are skipped)", err)
		}
		// fail-skill should still exist because removal was blocked.
		if _, err := os.Stat(filepath.Join(dir, "fail-skill")); err != nil {
			t.Fatal("expected fail-skill to still exist after failed removal")
		}
		// ok-skill should have been removed despite the earlier failure.
		if _, err := os.Stat(filepath.Join(dir, "ok-skill")); !os.IsNotExist(err) {
			t.Fatal("expected ok-skill to be removed after CleanupStale")
		}
	})
}

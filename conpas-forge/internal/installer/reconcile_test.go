package installer

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// helper: write a file to dir with given content.
func writeFile(t *testing.T, dir, name, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0644); err != nil {
		t.Fatalf("writeFile(%s): %v", name, err)
	}
}

// helper: assert file exists in dir.
func assertExists(t *testing.T, dir, name string) {
	t.Helper()
	if _, err := os.Stat(filepath.Join(dir, name)); err != nil {
		t.Fatalf("expected %s to exist: %v", name, err)
	}
}

// helper: assert file does not exist in dir.
func assertAbsent(t *testing.T, dir, name string) {
	t.Helper()
	if _, err := os.Stat(filepath.Join(dir, name)); !os.IsNotExist(err) {
		t.Fatalf("expected %s to be absent, but it exists (err=%v)", name, err)
	}
}

// helper: assert slice contains element.
func assertContains(t *testing.T, slice []string, elem string) {
	t.Helper()
	for _, s := range slice {
		if s == elem {
			return
		}
	}
	t.Fatalf("expected %v to contain %q", slice, elem)
}

// Scenario 3.1 — Creates directory when absent
func TestReconcileCreatesDirWhenAbsent(t *testing.T) {
	base := t.TempDir()
	dir := filepath.Join(base, "output-styles")

	removed, err := ReconcileOutputStyles(dir, "neutra.md", []byte("# Neutra"))
	if err != nil {
		t.Fatalf("ReconcileOutputStyles() error = %v", err)
	}
	if len(removed) != 0 {
		t.Fatalf("expected no removed files, got %v", removed)
	}
	assertExists(t, dir, "neutra.md")
}

// Scenario 3.2 / 6.2 — Purges orphan .md files, leaves non-.md files intact
func TestReconcilePreservesNonMdFiles(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "old-persona.md", "# old")
	writeFile(t, dir, "readme.txt", "user file")
	writeFile(t, dir, ".hidden", "hidden")

	removed, err := ReconcileOutputStyles(dir, "sargento.md", []byte("# Sargento"))
	if err != nil {
		t.Fatalf("ReconcileOutputStyles() error = %v", err)
	}
	if len(removed) != 1 || removed[0] != "old-persona.md" {
		t.Fatalf("expected [old-persona.md] removed, got %v", removed)
	}
	assertAbsent(t, dir, "old-persona.md")
	assertExists(t, dir, "readme.txt")
	assertExists(t, dir, ".hidden")
	assertExists(t, dir, "sargento.md")
}

// Scenario 3.3 / 6.4 — Idempotency: calling twice with same arguments
func TestReconcileIdempotent(t *testing.T) {
	dir := t.TempDir()
	content := []byte("# Sargento")
	writeFile(t, dir, "sargento.md", "# Sargento")

	removed, err := ReconcileOutputStyles(dir, "sargento.md", content)
	if err != nil {
		t.Fatalf("second call error = %v", err)
	}
	if len(removed) != 0 {
		t.Fatalf("expected no removed files on second call, got %v", removed)
	}
	assertExists(t, dir, "sargento.md")
}

// Scenario 3.5 / 1.3 — Multiple orphans removed in single call
func TestReconcileMultipleOrphans(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "gentleman.md", "# old")
	writeFile(t, dir, "jedi-backend-erp.md", "# old")
	writeFile(t, dir, "tony-stark.md", "# target")

	removed, err := ReconcileOutputStyles(dir, "tony-stark.md", []byte("# Tony Stark"))
	if err != nil {
		t.Fatalf("ReconcileOutputStyles() error = %v", err)
	}
	if len(removed) != 2 {
		t.Fatalf("expected 2 removed files, got %v", removed)
	}
	assertContains(t, removed, "gentleman.md")
	assertContains(t, removed, "jedi-backend-erp.md")
	assertAbsent(t, dir, "gentleman.md")
	assertAbsent(t, dir, "jedi-backend-erp.md")
	assertExists(t, dir, "tony-stark.md")
}

// Boundary condition — empty filename returns error
func TestReconcileEmptyFilenameError(t *testing.T) {
	dir := t.TempDir()
	_, err := ReconcileOutputStyles(dir, "", []byte("content"))
	if err == nil {
		t.Fatal("expected error for empty filename, got nil")
	}
}

// Scenario 6.1 — Empty directory, writes target file
func TestReconcileEmptyDirWritesFile(t *testing.T) {
	dir := t.TempDir()

	removed, err := ReconcileOutputStyles(dir, "neutra.md", []byte("# Neutra"))
	if err != nil {
		t.Fatalf("ReconcileOutputStyles() error = %v", err)
	}
	if len(removed) != 0 {
		t.Fatalf("expected no removed files, got %v", removed)
	}
	assertExists(t, dir, "neutra.md")
}

// Scenario 6.3 — User-authored .md file is removed (forge owns directory)
func TestReconcileUserMdOrphanRemoved(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "my-custom-style.md", "# user custom")

	removed, err := ReconcileOutputStyles(dir, "tony-stark.md", []byte("# Tony Stark"))
	if err != nil {
		t.Fatalf("ReconcileOutputStyles() error = %v", err)
	}
	if len(removed) != 1 || removed[0] != "my-custom-style.md" {
		t.Fatalf("expected [my-custom-style.md] removed, got %v", removed)
	}
	assertAbsent(t, dir, "my-custom-style.md")
	assertExists(t, dir, "tony-stark.md")
}

// Scenario 6.6 — Unreadable directory returns error (skip on Windows)
func TestReconcileUnreadableDirError(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("unreadable-dir test not applicable on Windows")
	}

	base := t.TempDir()
	dir := filepath.Join(base, "output-styles")
	if err := os.MkdirAll(dir, 0000); err != nil {
		t.Fatalf("setup: %v", err)
	}
	t.Cleanup(func() { os.Chmod(dir, 0755) }) //nolint:errcheck

	_, err := ReconcileOutputStyles(dir, "neutra.md", []byte("# Neutra"))
	if err == nil {
		t.Fatal("expected error for unreadable directory, got nil")
	}
}

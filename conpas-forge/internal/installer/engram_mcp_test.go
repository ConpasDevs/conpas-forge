package installer

import (
	"context"
	"os"
	"os/exec"
	"testing"

	"github.com/conpasDEVS/conpas-forge/internal/config"
)

func TestRegisterEngramMCP(t *testing.T) {
	// Skip if claude is not installed — this test requires a real Claude Code installation.
	if _, err := exec.LookPath("claude"); err != nil {
		t.Skip("claude not in PATH — skipping MCP registration test")
	}

	t.Run("registers_and_is_idempotent", func(t *testing.T) {
		home := t.TempDir()
		old := config.HomeDir()
		config.OverrideHomeDir(home)
		defer func() {
			config.OverrideHomeDir(old)
			// Clean up the user-scoped engram entry added during the test.
			// Best-effort; ignore errors.
			if claudePath, err := exec.LookPath("claude"); err == nil {
				exec.Command(claudePath, "mcp", "remove", "engram", "--scope", "user").Run() //nolint:errcheck
			}
		}()

		// Create a fake binary that exists on disk (content irrelevant for registration).
		fakeBin := home + "/engram-fake"
		if err := os.WriteFile(fakeBin, []byte("#!/bin/sh\necho fake"), 0755); err != nil {
			t.Fatalf("create fake binary: %v", err)
		}

		// First call — must succeed.
		if err := registerEngramMCP(context.Background(), fakeBin); err != nil {
			t.Fatalf("registerEngramMCP() first call error = %v", err)
		}

		// Second call — must also succeed (idempotent re-install).
		if err := registerEngramMCP(context.Background(), fakeBin); err != nil {
			t.Fatalf("registerEngramMCP() second call (idempotent) error = %v", err)
		}
	})

	t.Run("fails_when_claude_not_in_path", func(t *testing.T) {
		// Temporarily shadow PATH so claude cannot be found.
		old := os.Getenv("PATH")
		os.Setenv("PATH", "")
		defer os.Setenv("PATH", old)

		err := registerEngramMCP(context.Background(), "/fake/engram")
		if err == nil {
			t.Fatal("expected error when claude not in PATH, got nil")
		}
	})
}

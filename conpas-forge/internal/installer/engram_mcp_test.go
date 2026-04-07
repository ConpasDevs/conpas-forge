package installer

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/conpasDEVS/conpas-forge/internal/config"
)

func TestMergeEngramMCPEntry(t *testing.T) {
	tests := []struct {
		name string
		run  func(t *testing.T)
	}{
		{
			name: "adds_mcpServers_engram_to_settings_json",
			run: func(t *testing.T) {
				entry := map[string]any{
					"mcpServers": map[string]any{
						"engram": map[string]any{
							"command": "/fake/engram",
							"args":    []any{"mcp", "--tools=agent"},
							"type":    "stdio",
						},
					},
				}
				if err := Merge(entry); err != nil {
					t.Fatalf("Merge() error = %v", err)
				}
				data, err := os.ReadFile(config.SettingsJSON())
				if err != nil {
					t.Fatalf("read settings.json: %v", err)
				}
				var parsed map[string]any
				if err := json.Unmarshal(data, &parsed); err != nil {
					t.Fatalf("json.Unmarshal() error = %v", err)
				}
				servers, ok := parsed["mcpServers"].(map[string]any)
				if !ok {
					t.Fatal("settings.json missing mcpServers object")
				}
				engram, ok := servers["engram"].(map[string]any)
				if !ok {
					t.Fatal("settings.json missing mcpServers.engram")
				}
				if engram["command"] != "/fake/engram" {
					t.Fatalf("command = %q, want %q", engram["command"], "/fake/engram")
				}
				if engram["type"] != "stdio" {
					t.Fatalf("type = %q, want %q", engram["type"], "stdio")
				}
			},
		},
		{
			name: "preserves_existing_mcpServers_entries",
			run: func(t *testing.T) {
				existing := map[string]any{
					"mcpServers": map[string]any{
						"other-tool": map[string]any{"command": "other"},
					},
				}
				if err := Merge(existing); err != nil {
					t.Fatalf("pre-seed Merge() error = %v", err)
				}
				entry := map[string]any{
					"mcpServers": map[string]any{
						"engram": map[string]any{
							"command": "/fake/engram",
							"args":    []any{"mcp", "--tools=agent"},
							"type":    "stdio",
						},
					},
				}
				if err := Merge(entry); err != nil {
					t.Fatalf("Merge() error = %v", err)
				}
				data, _ := os.ReadFile(config.SettingsJSON())
				var parsed map[string]any
				json.Unmarshal(data, &parsed)
				servers := parsed["mcpServers"].(map[string]any)
				if _, ok := servers["other-tool"]; !ok {
					t.Fatal("other-tool entry was lost after merge")
				}
				if _, ok := servers["engram"]; !ok {
					t.Fatal("engram entry missing after merge")
				}
			},
		},
		{
			name: "overwrites_stale_engram_entry",
			run: func(t *testing.T) {
				old := map[string]any{
					"mcpServers": map[string]any{
						"engram": map[string]any{"command": "/old/path", "type": "stdio"},
					},
				}
				if err := Merge(old); err != nil {
					t.Fatalf("pre-seed Merge() error = %v", err)
				}
				entry := map[string]any{
					"mcpServers": map[string]any{
						"engram": map[string]any{
							"command": "/new/path",
							"args":    []any{"mcp", "--tools=agent"},
							"type":    "stdio",
						},
					},
				}
				if err := Merge(entry); err != nil {
					t.Fatalf("Merge() error = %v", err)
				}
				data, _ := os.ReadFile(config.SettingsJSON())
				var parsed map[string]any
				json.Unmarshal(data, &parsed)
				servers := parsed["mcpServers"].(map[string]any)
				engram := servers["engram"].(map[string]any)
				if engram["command"] != "/new/path" {
					t.Fatalf("command = %q, want /new/path", engram["command"])
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			home := t.TempDir()
			old := config.HomeDir()
			config.OverrideHomeDir(home)
			defer config.OverrideHomeDir(old)

			tt.run(t)
		})
	}
}

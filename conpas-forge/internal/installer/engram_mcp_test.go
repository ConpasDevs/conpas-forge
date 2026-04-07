package installer

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/conpasDEVS/conpas-forge/internal/config"
)

func TestWriteEngramMCPFile(t *testing.T) {
	tests := []struct {
		name string
		run  func(t *testing.T, home string)
	}{
		{
			name: "creates_valid_JSON",
			run: func(t *testing.T, _ string) {
				if err := writeEngramMCPFile("/fake/engram"); err != nil {
					t.Fatalf("writeEngramMCPFile() error = %v", err)
				}
				data, err := os.ReadFile(config.EngramMCPFile())
				if err != nil {
					t.Fatalf("read MCP file: %v", err)
				}
				var parsed mcpServerConfig
				if err := json.Unmarshal(data, &parsed); err != nil {
					t.Fatalf("json.Unmarshal() error = %v", err)
				}
			},
		},
		{
			name: "command_field",
			run: func(t *testing.T, _ string) {
				if err := writeEngramMCPFile("/fake/engram"); err != nil {
					t.Fatalf("writeEngramMCPFile() error = %v", err)
				}
				data, _ := os.ReadFile(config.EngramMCPFile())
				var parsed mcpServerConfig
				json.Unmarshal(data, &parsed)
				if parsed.Command != "/fake/engram" {
					t.Fatalf("Command = %q, want %q", parsed.Command, "/fake/engram")
				}
			},
		},
		{
			name: "args_field",
			run: func(t *testing.T, _ string) {
				if err := writeEngramMCPFile("/fake/engram"); err != nil {
					t.Fatalf("writeEngramMCPFile() error = %v", err)
				}
				data, _ := os.ReadFile(config.EngramMCPFile())
				var parsed mcpServerConfig
				json.Unmarshal(data, &parsed)
				want := []string{"mcp", "--tools=agent"}
				if len(parsed.Args) != len(want) {
					t.Fatalf("Args = %v, want %v", parsed.Args, want)
				}
				for i := range want {
					if parsed.Args[i] != want[i] {
						t.Fatalf("Args[%d] = %q, want %q", i, parsed.Args[i], want[i])
					}
				}
			},
		},
		{
			name: "type_field",
			run: func(t *testing.T, _ string) {
				if err := writeEngramMCPFile("/fake/engram"); err != nil {
					t.Fatalf("writeEngramMCPFile() error = %v", err)
				}
				data, _ := os.ReadFile(config.EngramMCPFile())
				var parsed mcpServerConfig
				json.Unmarshal(data, &parsed)
				if parsed.Type != "stdio" {
					t.Fatalf("Type = %q, want %q", parsed.Type, "stdio")
				}
			},
		},
		{
			name: "creates_parent_directory",
			run: func(t *testing.T, home string) {
				if err := writeEngramMCPFile("/fake/engram"); err != nil {
					t.Fatalf("writeEngramMCPFile() error = %v", err)
				}
				info, err := os.Stat(config.ClaudeMCPDir())
				if err != nil {
					t.Fatalf("stat MCP dir: %v", err)
				}
				if !info.IsDir() {
					t.Fatalf("MCP dir is not a directory")
				}
			},
		},
		{
			name: "overwrites_existing_file",
			run: func(t *testing.T, _ string) {
				// Pre-create file with stale content
				if err := writeEngramMCPFile("/old/path"); err != nil {
					t.Fatalf("first writeEngramMCPFile() error = %v", err)
				}
				// Overwrite with new binary path
				if err := writeEngramMCPFile("/new/path"); err != nil {
					t.Fatalf("second writeEngramMCPFile() error = %v", err)
				}
				data, err := os.ReadFile(config.EngramMCPFile())
				if err != nil {
					t.Fatalf("read MCP file: %v", err)
				}
				var parsed mcpServerConfig
				if err := json.Unmarshal(data, &parsed); err != nil {
					t.Fatalf("json.Unmarshal() error = %v", err)
				}
				if parsed.Command != "/new/path" {
					t.Fatalf("Command = %q, want %q", parsed.Command, "/new/path")
				}
			},
		},
		{
			name: "settings_json_not_touched",
			run: func(t *testing.T, _ string) {
				if err := writeEngramMCPFile("/fake/engram"); err != nil {
					t.Fatalf("writeEngramMCPFile() error = %v", err)
				}
				data, err := os.ReadFile(config.SettingsJSON())
				if err != nil {
					// File does not exist — that's fine
					if os.IsNotExist(err) {
						return
					}
					t.Fatalf("read settings.json: %v", err)
				}
				if strings.Contains(string(data), "mcpServers") {
					t.Fatal("settings.json should not contain mcpServers after writeEngramMCPFile")
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

			tt.run(t, home)
		})
	}
}

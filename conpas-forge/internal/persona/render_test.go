package persona

import (
	"strings"
	"testing"

	"github.com/conpasDEVS/conpas-forge/internal/config"
)

func TestRenderCLAUDEMD(t *testing.T) {
	buildDefault := func(t *testing.T) *CLAUDEMDData {
		t.Helper()
		cfg := config.DefaultConfig()
		data, err := BuildCLAUDEMDData(&cfg, "v0.0.0-test")
		if err != nil {
			t.Fatalf("BuildCLAUDEMDData() error = %v", err)
		}
		return data
	}

	renderDefault := func(t *testing.T) string {
		t.Helper()
		data := buildDefault(t)
		out, err := RenderCLAUDEMD(data)
		if err != nil {
			t.Fatalf("RenderCLAUDEMD() error = %v", err)
		}
		return string(out)
	}

	tests := []struct {
		name string
		run  func(t *testing.T)
	}{
		{
			name: "contains_proactive_save_triggers",
			run: func(t *testing.T) {
				output := renderDefault(t)
				if !strings.Contains(output, "PROACTIVE SAVE TRIGGERS") {
					t.Fatal("rendered CLAUDE.md does not contain 'PROACTIVE SAVE TRIGGERS'")
				}
			},
		},
		{
			name: "contains_session_close_protocol",
			run: func(t *testing.T) {
				output := renderDefault(t)
				if !strings.Contains(output, "SESSION CLOSE PROTOCOL") {
					t.Fatal("rendered CLAUDE.md does not contain 'SESSION CLOSE PROTOCOL'")
				}
			},
		},
		{
			name: "protocol_appears_exactly_once",
			run: func(t *testing.T) {
				output := renderDefault(t)
				count := strings.Count(output, "PROACTIVE SAVE TRIGGERS")
				if count != 1 {
					t.Fatalf("'PROACTIVE SAVE TRIGGERS' appears %d times, want 1", count)
				}
			},
		},
		{
			name: "empty_protocol_renders_cleanly",
			run: func(t *testing.T) {
				data := &CLAUDEMDData{
					PersonaName:    "asturiano",
					PersonaBlock:   "# Test persona",
					ModelRows:      []ModelRow{{Role: "default", Model: "claude-opus-4-5"}},
					Version:        "v0.0.0-test",
					GeneratedAt:    "2026-01-01T00:00:00Z",
					EngramProtocol: "",
				}
				out, err := RenderCLAUDEMD(data)
				if err != nil {
					t.Fatalf("RenderCLAUDEMD() error = %v", err)
				}
				if strings.Contains(string(out), "PROACTIVE SAVE TRIGGERS") {
					t.Fatal("empty EngramProtocol should not render protocol content")
				}
			},
		},
		{
			name: "contains_core_block",
			run: func(t *testing.T) {
				output := renderDefault(t)
				if !strings.Contains(output, "## Core") {
					t.Fatal("rendered CLAUDE.md does not contain '## Core' section")
				}
				if !strings.Contains(output, "Base Personality") {
					t.Fatal("rendered CLAUDE.md does not contain 'Base Personality'")
				}
				if !strings.Contains(output, "Technical Expertise") {
					t.Fatal("rendered CLAUDE.md does not contain 'Technical Expertise'")
				}
			},
		},
		{
			name: "core_block_before_persona_block",
			run: func(t *testing.T) {
				output := renderDefault(t)
				coreIdx := strings.Index(output, "## Core")
				personaIdx := strings.Index(output, "## Persona")
				if coreIdx == -1 {
					t.Fatal("'## Core' not found in rendered CLAUDE.md")
				}
				if personaIdx == -1 {
					t.Fatal("'## Persona' not found in rendered CLAUDE.md")
				}
				if coreIdx > personaIdx {
					t.Fatal("'## Core' must appear before '## Persona'")
				}
			},
		},
		{
			name: "empty_core_block_renders_cleanly",
			run: func(t *testing.T) {
				data := &CLAUDEMDData{
					PersonaName:  "asturiano",
					PersonaBlock: "# Test persona",
					CoreBlock:    "",
					ModelRows:    []ModelRow{{Role: "default", Model: "claude-opus-4-5"}},
					Version:      "v0.0.0-test",
					GeneratedAt:  "2026-01-01T00:00:00Z",
				}
				out, err := RenderCLAUDEMD(data)
				if err != nil {
					t.Fatalf("RenderCLAUDEMD() error = %v", err)
				}
				if strings.Contains(string(out), "## Core") {
					t.Fatal("empty CoreBlock should not render '## Core' section")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.run(t)
		})
	}
}

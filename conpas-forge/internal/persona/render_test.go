package persona

import (
	"strings"
	"testing"

	"github.com/conpasDEVS/conpas-forge/internal/config"
)

func TestRenderCLAUDEMD(t *testing.T) {
	// buildDefault uses tony-stark so tests that check output-style content are deterministic.
	buildDefault := func(t *testing.T) *CLAUDEMDData {
		t.Helper()
		cfg := config.DefaultConfig()
		cfg.Persona = "tony-stark"
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
					PersonaName:     "asturiano",
					PersonaBlock:    "# Test persona",
					ModelRows:       []ModelRow{{Role: "default", Model: "claude-opus-4-5"}},
					Version:         "v0.0.0-test",
					GeneratedAt:     "2026-01-01T00:00:00Z",
					EngramProtocol:  "",
					OutputStyleFile: "asturiano.md",
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
					PersonaName:     "asturiano",
					PersonaBlock:    "# Test persona",
					CoreBlock:       "",
					ModelRows:       []ModelRow{{Role: "default", Model: "claude-opus-4-5"}},
					Version:         "v0.0.0-test",
					GeneratedAt:     "2026-01-01T00:00:00Z",
					OutputStyleFile: "asturiano.md",
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
		// Scenario 4.1 — Output Styles section names the active file
		{
			name: "output_style_file_named_in_section",
			run: func(t *testing.T) {
				output := renderDefault(t)
				if !strings.Contains(output, "~/.claude/output-styles/tony-stark.md") {
					t.Fatal("rendered CLAUDE.md does not contain the active output-style path")
				}
			},
		},
		// Scenario 4.1 — bound to persona line
		{
			name: "output_style_bound_to_persona",
			run: func(t *testing.T) {
				output := renderDefault(t)
				if !strings.Contains(output, "(bound to persona `tony-stark`)") {
					t.Fatal("rendered CLAUDE.md does not contain bound-to-persona line")
				}
			},
		},
		// Scenario 4.2 — old generic line absent
		{
			name: "old_generic_output_styles_line_absent",
			run: func(t *testing.T) {
				output := renderDefault(t)
				if strings.Contains(output, "Output style files are available in") {
					t.Fatal("rendered CLAUDE.md still contains the old generic Output Styles line")
				}
			},
		},
		// Scenario 4.2 — single-file instruction present
		{
			name: "output_style_single_file_instruction_present",
			run: func(t *testing.T) {
				output := renderDefault(t)
				if !strings.Contains(output, "only output-style file conpas-forge installs") {
					t.Fatal("rendered CLAUDE.md does not contain single-file instruction")
				}
			},
		},
		// Scenario 4.3 — BuildCLAUDEMDData populates OutputStyleFile for default persona
		{
			name: "build_data_output_style_file_populated",
			run: func(t *testing.T) {
				data := buildDefault(t)
				if data.OutputStyleFile == "" {
					t.Fatal("BuildCLAUDEMDData returned empty OutputStyleFile")
				}
				if data.OutputStyleFile != "tony-stark.md" {
					t.Fatalf("expected OutputStyleFile=tony-stark.md, got %q", data.OutputStyleFile)
				}
			},
		},
		// Scenario 4.4 — All seven personas return non-empty OutputStyleFile from BuildCLAUDEMDData
		{
			name: "all_personas_return_non_empty_output_style_file",
			run: func(t *testing.T) {
				cfg := config.DefaultConfig()
				for _, p := range []string{"argentino", "asturiano", "galleguinho", "neutra", "sargento", "tony-stark", "yoda"} {
					cfg.Persona = p
					data, err := BuildCLAUDEMDData(&cfg, "v0.0.0-test")
					if err != nil {
						t.Fatalf("BuildCLAUDEMDData(%s) error = %v", p, err)
					}
					if data.OutputStyleFile == "" {
						t.Fatalf("persona %q returned empty OutputStyleFile", p)
					}
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

// Scenario 2.1 — All ValidPersonas have a mapping entry
func TestOutputStyleForAllPersonas(t *testing.T) {
	seen := make(map[string]string)
	for _, p := range []string{"argentino", "asturiano", "galleguinho", "neutra", "sargento", "tony-stark", "yoda"} {
		f := OutputStyleFor(p)
		if f == "" {
			t.Errorf("OutputStyleFor(%q) returned empty string", p)
			continue
		}
		if !strings.HasSuffix(f, ".md") {
			t.Errorf("OutputStyleFor(%q) = %q, expected to end with .md", p, f)
		}
		if prev, dup := seen[f]; dup {
			t.Errorf("duplicate mapping: personas %q and %q both map to %q", prev, p, f)
		}
		seen[f] = p
	}
}

// Scenario 2.2 — Unknown persona returns empty string
func TestOutputStyleForUnknown(t *testing.T) {
	if f := OutputStyleFor("phantom"); f != "" {
		t.Fatalf("OutputStyleFor(phantom) = %q, want empty string", f)
	}
}

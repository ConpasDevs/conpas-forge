package engramtools

import (
	"testing"
)

func TestRequiredAllowlist_CountAndOrder(t *testing.T) {
	list := RequiredAllowlist()
	if len(list) != 15 {
		t.Fatalf("RequiredAllowlist() count = %d, want 15", len(list))
	}
	// Verify canonical first and last entries
	if list[0] != "mcp__engram__mem_capture_passive" {
		t.Errorf("list[0] = %q, want mcp__engram__mem_capture_passive", list[0])
	}
	if list[14] != "mcp__engram__mem_update" {
		t.Errorf("list[14] = %q, want mcp__engram__mem_update", list[14])
	}
	// All must have the allowlist prefix
	for _, s := range list {
		if len(s) <= len(MCPAllowlistPrefix) || s[:len(MCPAllowlistPrefix)] != MCPAllowlistPrefix {
			t.Errorf("entry %q missing prefix %q", s, MCPAllowlistPrefix)
		}
	}
}

func TestRequiredAllowlistAsAny_Count(t *testing.T) {
	list := RequiredAllowlistAsAny()
	if len(list) != 15 {
		t.Fatalf("RequiredAllowlistAsAny() count = %d, want 15", len(list))
	}
	for _, v := range list {
		if _, ok := v.(string); !ok {
			t.Errorf("entry %v is not a string", v)
		}
	}
}

func TestRequiredAliasSet_Membership(t *testing.T) {
	set := RequiredAliasSet()
	want := []string{
		"engram_mem_save",
		"engram_mem_search",
		"engram_mem_context",
		"engram_mem_update",
		"engram_mem_delete",
	}
	for _, w := range want {
		if _, ok := set[w]; !ok {
			t.Errorf("RequiredAliasSet missing %q", w)
		}
	}
	if len(set) != 15 {
		t.Fatalf("RequiredAliasSet() size = %d, want 15", len(set))
	}
}

func TestAllowlistToAlias(t *testing.T) {
	tests := []struct {
		input    string
		wantOut  string
		wantBool bool
	}{
		{"mcp__engram__mem_save", "engram_mem_save", true},
		{"mcp__engram__mem_context", "engram_mem_context", true},
		{"mcp__engram__mem_update", "engram_mem_update", true},
		{"bad_prefix__mem_save", "", false},
		{"", "", false},
		{"engram_mem_save", "", false}, // alias form, not allowlist form
	}
	for _, tt := range tests {
		got, ok := AllowlistToAlias(tt.input)
		if ok != tt.wantBool || got != tt.wantOut {
			t.Errorf("AllowlistToAlias(%q) = (%q, %v), want (%q, %v)",
				tt.input, got, ok, tt.wantOut, tt.wantBool)
		}
	}
}

func TestAliasToAllowlist(t *testing.T) {
	tests := []struct {
		input    string
		wantOut  string
		wantBool bool
	}{
		{"engram_mem_save", "mcp__engram__mem_save", true},
		{"engram_mem_context", "mcp__engram__mem_context", true},
		{"bad_mem_save", "", false},
		{"", "", false},
		{"mcp__engram__mem_save", "", false}, // allowlist form, not alias form
	}
	for _, tt := range tests {
		got, ok := AliasToAllowlist(tt.input)
		if ok != tt.wantBool || got != tt.wantOut {
			t.Errorf("AliasToAllowlist(%q) = (%q, %v), want (%q, %v)",
				tt.input, got, ok, tt.wantOut, tt.wantBool)
		}
	}
}

func TestParseSkillToolNames(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    []string
	}{
		{
			name: "15 correct bullet entries",
			content: `
- **engram_mem_capture_passive** — save passive
- **engram_mem_context** — context
- **engram_mem_delete** — delete
- **engram_mem_get_observation** — get
- **engram_mem_merge_projects** — merge
- **engram_mem_save** — save
- **engram_mem_save_prompt** — save prompt
- **engram_mem_search** — search
- **engram_mem_session_end** — end
- **engram_mem_session_start** — start
- **engram_mem_session_summary** — summary
- **engram_mem_stats** — stats
- **engram_mem_suggest_topic_key** — suggest
- **engram_mem_timeline** — timeline
- **engram_mem_update** — update
`,
			want: []string{
				"engram_mem_capture_passive",
				"engram_mem_context",
				"engram_mem_delete",
				"engram_mem_get_observation",
				"engram_mem_merge_projects",
				"engram_mem_save",
				"engram_mem_save_prompt",
				"engram_mem_search",
				"engram_mem_session_end",
				"engram_mem_session_start",
				"engram_mem_session_summary",
				"engram_mem_stats",
				"engram_mem_suggest_topic_key",
				"engram_mem_timeline",
				"engram_mem_update",
			},
		},
		{
			name:    "non-engram bold bullet excluded",
			content: "- **title** — some title\n- **type** — some type\n",
			want:    []string{},
		},
		{
			name:    "duplicate bullet deduplicated",
			content: "- **engram_mem_save** — first\n- **engram_mem_save** — duplicate\n",
			want:    []string{"engram_mem_save"},
		},
		{
			name:    "empty input returns empty",
			content: "",
			want:    []string{},
		},
		{
			name:    "engram_ prefix with digits included",
			content: "- **engram_mem_save2** — numeric suffix\n",
			want:    []string{"engram_mem_save2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseSkillToolNames([]byte(tt.content))
			if len(got) != len(tt.want) {
				t.Fatalf("ParseSkillToolNames() len = %d, want %d\ngot: %v\nwant: %v",
					len(got), len(tt.want), got, tt.want)
			}
			for i, w := range tt.want {
				if got[i] != w {
					t.Errorf("ParseSkillToolNames()[%d] = %q, want %q", i, got[i], w)
				}
			}
		})
	}
}

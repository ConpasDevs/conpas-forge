package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSkillRegistryIncludesEngramMemory(t *testing.T) {
	tests := []struct {
		name        string
		mustContain []string
	}{
		{
			name: "registry includes shipped engram memory skill",
			mustContain: []string{
				"| `engram-memory` |",
				"### judgment-day",
				"### go-testing",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			registryPath := filepath.Join("testdata", "skill-registry.md")
			data, err := os.ReadFile(registryPath)
			if err != nil {
				t.Fatalf("read registry: %v", err)
			}

			content := string(data)
			for _, needle := range tt.mustContain {
				if !strings.Contains(content, needle) {
					t.Fatalf("registry missing %q", needle)
				}
			}
		})
	}
}

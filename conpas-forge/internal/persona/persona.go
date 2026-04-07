package persona

import (
	"fmt"
	"sort"
	"strings"

	"github.com/conpasDEVS/conpas-forge/internal/assets"
)

func LoadPersonaContent(name string) ([]byte, error) {
	data, err := assets.FS.ReadFile("personas/" + name + ".md")
	if err != nil {
		return nil, fmt.Errorf("persona '%s' not found in embedded assets", name)
	}
	return data, nil
}

func ValidNames() []string {
	entries, err := assets.FS.ReadDir("personas")
	if err != nil {
		return nil
	}
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".md") {
			names = append(names, strings.TrimSuffix(e.Name(), ".md"))
		}
	}
	sort.Strings(names)
	return names
}

func IsValid(name string) bool {
	_, err := LoadPersonaContent(name)
	return err == nil
}

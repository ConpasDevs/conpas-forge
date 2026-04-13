package installer

import "github.com/conpasDEVS/conpas-forge/internal/config"

func BuildModules(selectedIDs []string, cfg *config.Config) []Module {
	idSet := make(map[string]bool, len(selectedIDs))
	for _, id := range selectedIDs {
		idSet[id] = true
	}

	var modules []Module
	if idSet["engram"] {
		modules = append(modules, NewEngramInstaller())
	}
	if idSet["gentle-ai"] {
		modules = append(modules, NewGentleAIInstaller())
	}
	if idSet["zoho-deluge"] {
		modules = append(modules, NewConpasAIInstaller())
	}
	modules = append(modules, NewClaudeSettingsInstaller())
	return modules
}

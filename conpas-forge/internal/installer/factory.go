package installer

func BuildModules(selectedIDs []string) []Module {
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
	return modules
}

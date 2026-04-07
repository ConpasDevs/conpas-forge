package models

// CanonicalRoles is the ordered list of all SDD agent roles.
var CanonicalRoles = []string{
	"orchestrator",
	"sdd-explore",
	"sdd-propose",
	"sdd-spec",
	"sdd-design",
	"sdd-tasks",
	"sdd-apply",
	"sdd-verify",
	"sdd-archive",
	"default",
}

// Defaults maps each role to its recommended default LLM model.
var Defaults = map[string]string{
	"orchestrator": "claude-opus-4-6",
	"sdd-explore":  "claude-sonnet-4-6",
	"sdd-propose":  "claude-opus-4-6",
	"sdd-spec":     "claude-sonnet-4-6",
	"sdd-design":   "claude-opus-4-6",
	"sdd-tasks":    "claude-sonnet-4-6",
	"sdd-apply":    "claude-sonnet-4-6",
	"sdd-verify":   "claude-sonnet-4-6",
	"sdd-archive":  "claude-haiku-4-5-20251001",
	"default":      "claude-sonnet-4-6",
}

// ValidPersonas is the list of embedded persona names.
var ValidPersonas = []string{
	"argentino",
	"asturiano",
	"galleguinho",
	"neutra",
	"sargento",
	"tony-stark",
	"yoda",
}

// IsValidRole returns true if role is in CanonicalRoles.
func IsValidRole(role string) bool {
	for _, r := range CanonicalRoles {
		if r == role {
			return true
		}
	}
	return false
}

// IsValidPersona returns true if name is in ValidPersonas.
func IsValidPersona(name string) bool {
	for _, p := range ValidPersonas {
		if p == name {
			return true
		}
	}
	return false
}

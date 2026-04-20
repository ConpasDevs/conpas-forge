package persona

// outputStyleMap maps each persona name to its canonical output-style filename.
var outputStyleMap = map[string]string{
	"argentino":   "argentino.md",
	"asturiano":   "asturiano.md",
	"galleguinho": "galleguinho.md",
	"neutra":      "neutra.md",
	"sargento":    "sargento.md",
	"tony-stark":  "tony-stark.md",
	"yoda":        "yoda.md",
}

// OutputStyleFor returns the output-style filename for the given persona.
// Returns "" if the persona has no mapping (caller should treat as configuration error).
func OutputStyleFor(persona string) string {
	return outputStyleMap[persona]
}

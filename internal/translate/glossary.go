package translate

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Glossary maps source terms to their required translation per target language.
// Loaded from ~/.config/gotranslate/glossary.yaml by default.
type Glossary struct {
	// Terms maps a lowercase source term → per-language preferred translation.
	Terms map[string]map[string]string `yaml:"terms"`
}

// LoadGlossary reads a glossary file. Missing file is not an error — returns
// an empty glossary so callers don't need to check os.Stat first.
func LoadGlossary(path string) (Glossary, error) {
	g := Glossary{Terms: map[string]map[string]string{}}
	if path == "" {
		dir, err := os.UserConfigDir()
		if err != nil {
			return g, nil
		}
		path = filepath.Join(dir, "gotranslate", "glossary.yaml")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return g, nil // absent glossary is fine
	}
	if err := yaml.Unmarshal(data, &g); err != nil {
		return g, fmt.Errorf("parsing glossary %s: %w", path, err)
	}
	if g.Terms == nil {
		g.Terms = map[string]map[string]string{}
	}
	return g, nil
}

// Lookup returns the preferred translation of term into target, or "".
// target may be an ISO code ("es") or English name ("Spanish") — names are
// resolved through the package-level languageCodes in providers; here we match
// against whatever keys the glossary file uses, so users can use either.
func (g Glossary) Lookup(term, target string) string {
	entry, ok := g.Terms[strings.ToLower(strings.TrimSpace(term))]
	if !ok {
		return ""
	}
	if v, ok := entry[strings.ToLower(target)]; ok {
		return v
	}
	return ""
}

// Scan returns all glossary terms found in text (case-insensitive substring).
func (g Glossary) Scan(text string) map[string]string {
	found := map[string]string{}
	lower := strings.ToLower(text)
	for term := range g.Terms {
		if strings.Contains(lower, term) {
			found[term] = term
		}
	}
	return found
}

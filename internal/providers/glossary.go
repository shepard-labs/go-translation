package providers

import (
	"fmt"
	"sort"
	"strings"

	"github.com/shepard-labs/go-translation/internal/translate"
)

// glossarySummary renders the glossary terms relevant to target as prompt
// lines, e.g. `- "account" → "compte"`. Empty when no terms apply.
func glossarySummary(g translate.Glossary, target string) string {
	if len(g.Terms) == 0 {
		return ""
	}
	var lines []string
	for term, perLang := range g.Terms {
		if v, ok := perLang[strings.ToLower(target)]; ok && v != "" {
			lines = append(lines, fmt.Sprintf("- %q → %q", term, v))
		}
	}
	if len(lines) == 0 {
		return ""
	}
	sort.Strings(lines)
	return strings.Join(lines, "\n")
}

package translate

import (
	"fmt"
	"regexp"
	"strings"
)

// Placeholder is a deterministic token that replaces a protected span before
// the text reaches the model. Format: ⟦P000⟧ — characters no natural
// language uses, so the model has no reason to translate or reorder inside.
const placeholderFmt = "⟦P%03d⟧"

var placeholderPattern = regexp.MustCompile(`⟦P(\d{3})⟧`)

// spanRule matches a non-translatable region. Order matters: earlier rules win.
type spanRule struct {
	re *regexp.Regexp
}

// rules protect, in order: fenced code blocks, frontmatter blocks, inline code,
// URLs, HTML/XML tags, i18n placeholders ({{var}}, {0}, %s, %1$s, %@), and
// email addresses.
var rules = []spanRule{
	{regexp.MustCompile(`(?s)` + "```" + `.*?` + "```" + `|~~~.*?~~~`)},
	{regexp.MustCompile(`(?s)\A---\n.*?\n---\n`)},
	{regexp.MustCompile("`[^`\n]+`")},
	{regexp.MustCompile(`https?://[^\s<>\)"']+|www\.[^\s<>\)"']+`)},
	{regexp.MustCompile(`</?[A-Za-z][^>]*>`)},
	{regexp.MustCompile(`\{\{[^}]*\}\}|\{[A-Za-z0-9_.]+\}|%[0-9]*\$?[sdif@]|%\{[^}]*\}`)},
	{regexp.MustCompile(`[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}`)},
}

// Protection holds a protected document and its placeholder→span map.
type Protection struct {
	Text  string
	Spans []string // index-aligned: placeholder i replaces Spans[i]
	Hits  int
}

// Protect replaces every non-translatable span in text with ⟦P000⟧ tokens.
func Protect(text string) Protection {
	var spans []string
	out := text
	// Iterative single pass: scan left to right over original, applying rules
	// at each position. Simplest correct approach: repeatedly find the earliest
	// match among all rules, replace it, continue scanning after it.
	var b strings.Builder
	pos := 0
	for pos < len(out) {
		best := -1
		var bestLoc []int
		for i, r := range rules {
			loc := r.re.FindStringIndex(out[pos:])
			if loc == nil {
				continue
			}
			if bestLoc == nil || loc[0] < bestLoc[0] || (loc[0] == bestLoc[0] && i < best) {
				best = i
				bestLoc = loc
			}
		}
		if bestLoc == nil {
			b.WriteString(out[pos:])
			break
		}
		start, end := pos+bestLoc[0], pos+bestLoc[1]
		b.WriteString(out[pos:start])
		idx := len(spans)
		spans = append(spans, out[start:end])
		b.WriteString(fmt.Sprintf(placeholderFmt, idx))
		pos = end
	}
	return Protection{Text: b.String(), Spans: spans, Hits: len(spans)}
}

// Restore replaces every ⟦P000⟧ token with its original span. Tokens the model
// dropped or reordered are re-appended in original order at the end only if
// strict is false — callers should pass strict=true and surface the error.
func Restore(text string, spans []string, strict bool) (string, error) {
	if len(spans) == 0 {
		return text, nil
	}
	seen := make([]bool, len(spans))
	out := placeholderPattern.ReplaceAllStringFunc(text, func(tok string) string {
		var idx int
		if _, err := fmt.Sscanf(tok, placeholderFmt, &idx); err != nil || idx >= len(spans) {
			return tok
		}
		seen[idx] = true
		return spans[idx]
	})
	var missing []int
	for i, s := range seen {
		if !s {
			missing = append(missing, i)
		}
	}
	if len(missing) > 0 {
		if strict {
			return "", fmt.Errorf("translation dropped %d protected placeholder(s): %v", len(missing), missing)
		}
		// Append missing spans at the end so nothing is silently lost.
		for _, i := range missing {
			out += spans[i]
		}
	}
	return out, nil
}

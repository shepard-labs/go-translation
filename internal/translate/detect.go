package translate

import (
	"regexp"
	"strings"
)

// DetectLanguage returns a best-effort ISO 639-1 code for text without any API
// call. It is a heuristic: script/Unicode-range analysis for non-Latin scripts
// plus a small stopword table for common Latin-script languages. Returns ""
// when the text is too short or ambiguous — callers should treat "" as "auto".
func DetectLanguage(text string) string {
	t := strings.TrimSpace(text)
	if t == "" {
		return ""
	}
	var latin, cjk, cyrillic, arabic, hangul, hebrew, thai, hiragana, katakana int
	for _, r := range t {
		switch {
		case r >= 0x4E00 && r <= 0x9FFF:
			cjk++
		case r >= 0x3040 && r <= 0x309F:
			hiragana++
		case r >= 0x30A0 && r <= 0x30FF:
			katakana++
		case r >= 0xAC00 && r <= 0xD7AF:
			hangul++
		case r >= 0x0400 && r <= 0x04FF:
			cyrillic++
		case r >= 0x0600 && r <= 0x06FF:
			arabic++
		case r >= 0x0590 && r <= 0x05FF:
			hebrew++
		case r >= 0x0E00 && r <= 0x0E7F:
			thai++
		case r < 0x0250:
			latin++
		}
	}
	switch {
	case hiragana+katakana > 0:
		return "ja"
	case hangul > 0:
		return "ko"
	case cjk > 0:
		return "zh" // simplified by default; caller may refine
	case cyrillic > 0:
		return "ru"
	case arabic > 0:
		return "ar"
	case hebrew > 0:
		return "iw"
	case thai > 0:
		return "th"
	case latin > 0:
		return detectLatin(t)
	}
	return ""
}

var latinStopwords = map[string][]string{
	"en": {"the", "and", "is", "are", "you", "of", "to", "in", "that"},
	"es": {"el", "la", "de", "que", "y", "en", "un", "ser", "se", "no", "los", "las", "está", "por"},
	"fr": {"le", "la", "les", "de", "et", "est", "en", "que", "une", "du", "des", "pour", "vous", "nous"},
	"de": {"der", "die", "das", "und", "ist", "in", "zu", "den", "mit", "nicht", "sie", "ich", "auf", "für"},
	"it": {"il", "la", "di", "e", "che", "è", "un", "per", "non", "una", "sono", "del", "della"},
	"pt": {"o", "a", "de", "e", "que", "do", "da", "em", "um", "para", "não", "os", "as", "você"},
	"nl": {"de", "het", "een", "en", "van", "is", "in", "op", "te", "voor", "met", "niet", "zijn"},
}

var wordSplitter = regexp.MustCompile(`[^A-Za-zÀ-ÿ]+`)

func detectLatin(text string) string {
	words := wordSplitter.Split(strings.ToLower(text), -1)
	counts := map[string]int{}
	total := 0
	for _, w := range words {
		if w == "" {
			continue
		}
		total++
		for lang, stops := range latinStopwords {
			for _, s := range stops {
				if w == s {
					counts[lang]++
				}
			}
		}
	}
	best, bestN := "", 0
	for lang, n := range counts {
		if n > bestN {
			best, bestN = lang, n
		}
	}
	if total < 3 || bestN == 0 {
		return ""
	}
	return best
}

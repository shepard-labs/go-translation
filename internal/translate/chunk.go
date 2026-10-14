package translate

import "strings"

func Chunks(text string, maxChars int) []string {
	if maxChars <= 0 || len([]rune(text)) <= maxChars {
		return []string{text}
	}
	paragraphs := strings.SplitAfter(text, "\n\n")
	return splitParts(paragraphs, maxChars)
}

func splitParts(parts []string, maxChars int) []string {
	var chunks []string
	var current strings.Builder
	for _, part := range parts {
		if runeLen(part) > maxChars {
			if current.Len() > 0 {
				chunks = append(chunks, current.String())
				current.Reset()
			}
			chunks = append(chunks, splitLong(part, maxChars)...)
			continue
		}
		if runeLen(current.String())+runeLen(part) > maxChars && current.Len() > 0 {
			chunks = append(chunks, current.String())
			current.Reset()
		}
		current.WriteString(part)
	}
	if current.Len() > 0 {
		chunks = append(chunks, current.String())
	}
	return chunks
}

func splitLong(text string, maxChars int) []string {
	lines := strings.SplitAfter(text, "\n")
	if len(lines) > 1 {
		chunks := splitParts(lines, maxChars)
		var allFit bool = true
		for _, chunk := range chunks {
			if runeLen(chunk) > maxChars {
				allFit = false
				break
			}
		}
		if allFit {
			return chunks
		}
	}

	runes := []rune(text)
	var chunks []string
	for len(runes) > maxChars {
		chunks = append(chunks, string(runes[:maxChars]))
		runes = runes[maxChars:]
	}
	if len(runes) > 0 {
		chunks = append(chunks, string(runes))
	}
	return chunks
}

func runeLen(text string) int { return len([]rune(text)) }

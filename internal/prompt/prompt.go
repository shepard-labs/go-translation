package prompt

import "fmt"

func System(sourceLanguage, targetLanguage string) string {
	source := fmt.Sprintf("Translate the user's content from %s to %s.", sourceLanguage, targetLanguage)
	if sourceLanguage == "" || sourceLanguage == "auto" {
		source = fmt.Sprintf("Detect the source language automatically and translate the user's content to %s.", targetLanguage)
	}
	return "You are a professional translation engine.\n" +
		source + "\n" +
		"Return only the translated content.\n" +
		"Preserve formatting, Markdown structure, code blocks, lists, tables, frontmatter, placeholders, variables, and URLs.\n" +
		"Do not summarize, explain, add commentary, or wrap the output in quotes.\n" +
		"If text should not be translated, preserve it exactly."
}

// SystemAgent is the system prompt used when the provider request carries
// deterministic tools (detect_language, lookup_glossary, placeholder_check,
// restore_spans) and the content has been pre-protected: every code block,
// URL, tag, and placeholder is already replaced by a ⟦PNNN⟧ token.
func SystemAgent(sourceLanguage, targetLanguage string, spanCount int, glossarySummary string) string {
	source := fmt.Sprintf("Translate the user's content from %s to %s.", sourceLanguage, targetLanguage)
	if sourceLanguage == "" || sourceLanguage == "auto" {
		source = fmt.Sprintf("Call detect_language on the content first, then translate from the detected language to %s.", targetLanguage)
	}
	s := "You are a professional translation engine with access to deterministic tools.\n" +
		source + "\n"
	if spanCount > 0 {
		s += fmt.Sprintf("The content contains %d protected placeholder token(s) (⟦P000⟧ …). "+
			"Translate around them; keep each token verbatim and in the correct position. "+
			"When you finish translating, call restore_spans once with your translation to reinstate the original content, "+
			"then return the restored text.\n", spanCount)
	}
	if glossarySummary != "" {
		s += "Glossary terms apply:\n" + glossarySummary + "\n"
	}
	s += "You may call lookup_glossary for any term you are unsure of, and placeholder_check to verify no placeholder was dropped.\n" +
		"Do not summarize, explain, add commentary, or wrap the output in quotes. Return only the final translated content."
	return s
}

func User(text string) string {
	return fmt.Sprintf("Translate this content:\n\n<content>\n%s\n</content>", text)
}

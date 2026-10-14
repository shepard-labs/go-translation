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

func User(text string) string {
	return fmt.Sprintf("Translate this content:\n\n<content>\n%s\n</content>", text)
}

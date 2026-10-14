package translate

import "context"

type Translator interface {
	Translate(ctx context.Context, req TranslationRequest) (TranslationResult, error)
}

type TranslationRequest struct {
	Provider        string
	Model           string
	SourceLanguage  string
	TargetLanguage  string
	Text            string
	MaxOutputTokens int
	Temperature     float64
}

type TranslationResult struct {
	Text     string
	Provider string
	Model    string
	Usage    Usage
}

type Usage struct {
	InputTokens  int
	OutputTokens int
	TotalTokens  int
}

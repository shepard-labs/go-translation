package translate

import (
	"context"
	"strings"
)

type Service struct {
	translator      Translator
	chunkingEnabled bool
	maxChars        int
}

func NewService(translator Translator, chunkingEnabled bool, maxChars int) Service {
	return Service{translator: translator, chunkingEnabled: chunkingEnabled, maxChars: maxChars}
}

func (s Service) Translate(ctx context.Context, req TranslationRequest) (TranslationResult, error) {
	if !s.chunkingEnabled || s.maxChars <= 0 || len([]rune(req.Text)) <= s.maxChars {
		return s.translator.Translate(ctx, req)
	}

	chunks := Chunks(req.Text, s.maxChars)
	var out strings.Builder
	var usage Usage
	for _, chunk := range chunks {
		chunkReq := req
		chunkReq.Text = chunk
		result, err := s.translator.Translate(ctx, chunkReq)
		if err != nil {
			return TranslationResult{}, err
		}
		out.WriteString(result.Text)
		usage.InputTokens += result.Usage.InputTokens
		usage.OutputTokens += result.Usage.OutputTokens
		usage.TotalTokens += result.Usage.TotalTokens
	}
	return TranslationResult{Text: out.String(), Provider: req.Provider, Model: req.Model, Usage: usage}, nil
}

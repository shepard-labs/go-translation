package providers

import (
	"context"
	"strings"

	"github.com/shepard-labs/go-translation/internal/prompt"
	"github.com/shepard-labs/go-translation/internal/translate"
	"github.com/shepard-labs/go-ai-sdk/google"
)

type Gemini struct {
	apiKey       string
	defaultModel string
}

func NewGemini(apiKey, defaultModel string) Gemini {
	return Gemini{apiKey: apiKey, defaultModel: defaultModel}
}

func (g Gemini) Translate(ctx context.Context, req translate.TranslationRequest) (translate.TranslationResult, error) {
	modelID := req.Model
	if modelID == "" {
		modelID = g.defaultModel
	}
	temperature := req.Temperature
	maxTokens := req.MaxOutputTokens
	provider := google.CreateGoogle(google.ProviderSettings{APIKey: g.apiKey})
	if err := provider.Err(); err != nil {
		return translate.TranslationResult{}, err
	}
	result, err := provider.Model(modelID).DoGenerate(ctx, google.GenerateOptions{
		MaxOutputTokens: &maxTokens,
		Temperature:     &temperature,
		Messages: []google.Message{
			google.SystemMessage{Content: prompt.System(req.SourceLanguage, req.TargetLanguage)},
			google.UserMessage{Content: []google.UserContent{google.TextContent{Text: prompt.User(req.Text)}}},
		},
	})
	if err != nil {
		return translate.TranslationResult{}, err
	}
	var out strings.Builder
	for _, content := range result.Content {
		if text, ok := content.(google.TextContent); ok {
			out.WriteString(text.Text)
		}
	}
	return translate.TranslationResult{
		Text:     out.String(),
		Provider: "gemini",
		Model:    modelID,
		Usage: translate.Usage{
			InputTokens:  intPtrValue(result.Usage.InputTokens.Total),
			OutputTokens: intPtrValue(result.Usage.OutputTokens.Total),
			TotalTokens:  intPtrValue(result.Usage.InputTokens.Total) + intPtrValue(result.Usage.OutputTokens.Total),
		},
	}, nil
}

func intPtrValue(v *int) int {
	if v == nil {
		return 0
	}
	return *v
}

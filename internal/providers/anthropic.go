package providers

import (
	"context"
	"strings"

	"github.com/shepard-labs/go-translation/internal/prompt"
	"github.com/shepard-labs/go-translation/internal/translate"
	"github.com/shepard-labs/go-ai-sdk/anthropic"
)

type Anthropic struct {
	apiKey       string
	defaultModel string
}

func NewAnthropic(apiKey, defaultModel string) Anthropic {
	return Anthropic{apiKey: apiKey, defaultModel: defaultModel}
}

func (a Anthropic) Translate(ctx context.Context, req translate.TranslationRequest) (translate.TranslationResult, error) {
	modelID := req.Model
	if modelID == "" {
		modelID = a.defaultModel
	}
	temperature := req.Temperature
	provider := anthropic.CreateAnthropic(anthropic.ProviderSettings{APIKey: a.apiKey})
	if err := provider.Err(); err != nil {
		return translate.TranslationResult{}, err
	}
	result, err := provider.Model(modelID).DoGenerate(ctx, anthropic.GenerateOptions{
		MaxTokens:   req.MaxOutputTokens,
		Temperature: &temperature,
		Messages: []anthropic.Message{
			anthropic.SystemMessage{Content: prompt.System(req.SourceLanguage, req.TargetLanguage)},
			anthropic.UserMessage{Content: []anthropic.UserContent{anthropic.TextContent{Text: prompt.User(req.Text)}}},
		},
	})
	if err != nil {
		return translate.TranslationResult{}, err
	}
	var out strings.Builder
	for _, content := range result.Content {
		if text, ok := content.(anthropic.TextContent); ok {
			out.WriteString(text.Text)
		}
	}
	return translate.TranslationResult{
		Text:     out.String(),
		Provider: "anthropic",
		Model:    modelID,
		Usage: translate.Usage{
			InputTokens:  result.Usage.InputTokens.Total,
			OutputTokens: result.Usage.OutputTokens.Total,
			TotalTokens:  result.Usage.TotalTokens,
		},
	}, nil
}

package providers

import (
	"context"
	"strings"

	"github.com/shepard-labs/go-translation/internal/prompt"
	"github.com/shepard-labs/go-translation/internal/translate"
	"github.com/shepard-labs/go-ai-sdk/openrouter"
)

type OpenRouter struct {
	apiKey       string
	defaultModel string
	appName      string
	appURL       string
}

func NewOpenRouter(apiKey, defaultModel, appName, appURL string) OpenRouter {
	if appName == "" {
		appName = "aitranslate"
	}
	return OpenRouter{apiKey: apiKey, defaultModel: defaultModel, appName: appName, appURL: appURL}
}

func (o OpenRouter) Translate(ctx context.Context, req translate.TranslationRequest) (translate.TranslationResult, error) {
	modelID := req.Model
	if modelID == "" {
		modelID = o.defaultModel
	}
	temperature := req.Temperature
	maxTokens := req.MaxOutputTokens
	provider := openrouter.CreateOpenRouter(openrouter.ProviderSettings{APIKey: o.apiKey, AppName: o.appName, AppURL: o.appURL})
	result, err := provider.Chat(modelID).DoGenerate(ctx, openrouter.GenerateOptions{
		MaxTokens:   &maxTokens,
		Temperature: &temperature,
		Messages: []openrouter.Message{
			openrouter.SystemMessage{Content: prompt.System(req.SourceLanguage, req.TargetLanguage)},
			openrouter.UserMessage{Content: []openrouter.UserContent{openrouter.TextContent{Text: prompt.User(req.Text)}}},
		},
	})
	if err != nil {
		return translate.TranslationResult{}, err
	}
	var out strings.Builder
	for _, content := range result.Content {
		if text, ok := content.(openrouter.TextContent); ok {
			out.WriteString(text.Text)
		}
	}
	return translate.TranslationResult{
		Text:     out.String(),
		Provider: "openrouter",
		Model:    modelID,
		Usage: translate.Usage{
			InputTokens:  result.Usage.InputTokens,
			OutputTokens: result.Usage.OutputTokens,
			TotalTokens:  result.Usage.TotalTokens,
		},
	}, nil
}

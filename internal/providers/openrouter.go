package providers

import (
	"context"

	"github.com/shepard-labs/go-translation/internal/prompt"
	"github.com/shepard-labs/go-translation/internal/translate"

	"github.com/shepard-labs/go-ai-sdk/openrouter"
)

type OpenRouter struct {
	apiKey       string
	defaultModel string
	appName      string
	appURL       string
	glossary     translate.Glossary
}

func NewOpenRouter(apiKey, defaultModel, appName, appURL string) OpenRouter {
	if appName == "" {
		appName = "aitranslate"
	}
	return OpenRouter{apiKey: apiKey, defaultModel: defaultModel, appName: appName, appURL: appURL}
}

func (o OpenRouter) WithGlossary(g translate.Glossary) OpenRouter {
	o.glossary = g
	return o
}

func (o OpenRouter) Translate(ctx context.Context, req translate.TranslationRequest) (translate.TranslationResult, error) {
	modelID := req.Model
	if modelID == "" {
		modelID = o.defaultModel
	}
	provider := openrouter.CreateOpenRouter(openrouter.ProviderSettings{APIKey: o.apiKey, AppName: o.appName, AppURL: o.appURL})

	protected := translate.Protect(req.Text)
	toolset := translate.Toolset{Glossary: o.glossary, Spans: protected.Spans}
	system := prompt.SystemAgent(req.SourceLanguage, req.TargetLanguage, protected.Hits, glossarySummary(o.glossary, req.TargetLanguage))

	sess := newOpenRouterSession(provider.Chat(modelID), system, prompt.User(protected.Text), toolset.Definitions(), req.MaxOutputTokens, req.Temperature)
	res, err := translate.RunAgentLoop(ctx, sess, translate.AgentRequest{
		System:     system,
		User:       prompt.User(protected.Text),
		Toolset:    toolset,
		RestoreEnd: true,
		Spans:      protected.Spans,
	})
	if err != nil {
		return translate.TranslationResult{}, err
	}
	return translate.TranslationResult{
		Text:     res.Text,
		Provider: "openrouter",
		Model:    modelID,
	}, nil
}

package providers

import (
	"context"

	"github.com/shepard-labs/go-ai-sdk/google"
	"github.com/shepard-labs/go-translation/internal/prompt"
	"github.com/shepard-labs/go-translation/internal/translate"
)

type Gemini struct {
	apiKey       string
	defaultModel string
	glossary     translate.Glossary
}

func NewGemini(apiKey, defaultModel string) Gemini {
	return Gemini{apiKey: apiKey, defaultModel: defaultModel}
}

func (g Gemini) WithGlossary(gl translate.Glossary) Gemini {
	g.glossary = gl
	return g
}

func (g Gemini) Translate(ctx context.Context, req translate.TranslationRequest) (translate.TranslationResult, error) {
	modelID := req.Model
	if modelID == "" {
		modelID = g.defaultModel
	}
	provider := google.CreateGoogle(google.ProviderSettings{APIKey: g.apiKey})
	if err := provider.Err(); err != nil {
		return translate.TranslationResult{}, err
	}

	protected := translate.Protect(req.Text)
	toolset := translate.Toolset{Glossary: g.glossary, Spans: protected.Spans}
	system := prompt.SystemAgent(req.SourceLanguage, req.TargetLanguage, protected.Hits, glossarySummary(g.glossary, req.TargetLanguage))

	sess := newGeminiSession(provider.Model(modelID), system, prompt.User(protected.Text), toolset.Definitions(), req.MaxOutputTokens, req.Temperature)
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
		Provider: "gemini",
		Model:    modelID,
	}, nil
}

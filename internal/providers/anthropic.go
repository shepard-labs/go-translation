package providers

import (
	"context"

	"github.com/shepard-labs/go-translation/internal/prompt"
	"github.com/shepard-labs/go-translation/internal/translate"

	"github.com/shepard-labs/go-ai-sdk/anthropic"
)

type Anthropic struct {
	apiKey       string
	defaultModel string
	glossary     translate.Glossary
}

func NewAnthropic(apiKey, defaultModel string) Anthropic {
	return Anthropic{apiKey: apiKey, defaultModel: defaultModel}
}

// WithGlossary attaches a term glossary; the model can query it via the
// lookup_glossary tool.
func (a Anthropic) WithGlossary(g translate.Glossary) Anthropic {
	a.glossary = g
	return a
}

func (a Anthropic) Translate(ctx context.Context, req translate.TranslationRequest) (translate.TranslationResult, error) {
	modelID := req.Model
	if modelID == "" {
		modelID = a.defaultModel
	}
	provider := anthropic.CreateAnthropic(anthropic.ProviderSettings{APIKey: a.apiKey})
	if err := provider.Err(); err != nil {
		return translate.TranslationResult{}, err
	}

	protected := translate.Protect(req.Text)
	toolset := translate.Toolset{Glossary: a.glossary, Spans: protected.Spans}
	system := prompt.SystemAgent(req.SourceLanguage, req.TargetLanguage, protected.Hits, glossarySummary(a.glossary, req.TargetLanguage))

	sess := newAnthropicSession(provider.Model(modelID), system, prompt.User(protected.Text), toolset.Definitions(), req.MaxOutputTokens, req.Temperature)
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
		Provider: "anthropic",
		Model:    modelID,
	}, nil
}

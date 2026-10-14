package providers

import (
	"fmt"

	"github.com/shepard-labs/go-translation/internal/config"
	"github.com/shepard-labs/go-translation/internal/translate"
)

func New(cfg config.Config) (translate.Translator, error) {
	provider := cfg.DefaultProvider
	key := config.APIKey(cfg)
	model := config.ModelFor(cfg)
	glossary, _ := translate.LoadGlossary("")
	switch provider {
	case config.ProviderAnthropic:
		return NewAnthropic(key, model).WithGlossary(glossary), nil
	case config.ProviderGemini:
		return NewGemini(key, model).WithGlossary(glossary), nil
	case config.ProviderOpenRouter:
		pc := cfg.Providers[provider]
		return NewOpenRouter(key, model, pc.AppName, pc.AppURL).WithGlossary(glossary), nil
	case config.ProviderGoogleCloud:
		pc := cfg.Providers[provider]
		return NewGoogleCloud(key, pc.ProjectID, pc.Location, model, pc.MimeType), nil
	default:
		return nil, fmt.Errorf("invalid provider %q: expected anthropic, gemini, openrouter, or googlecloud", provider)
	}
}

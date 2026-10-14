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
	switch provider {
	case config.ProviderAnthropic:
		return NewAnthropic(key, model), nil
	case config.ProviderGemini:
		return NewGemini(key, model), nil
	case config.ProviderOpenRouter:
		pc := cfg.Providers[provider]
		return NewOpenRouter(key, model, pc.AppName, pc.AppURL), nil
	default:
		return nil, fmt.Errorf("invalid provider %q: expected anthropic, gemini, or openrouter", provider)
	}
}

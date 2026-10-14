package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

const (
	ProviderAnthropic  = "anthropic"
	ProviderGemini     = "gemini"
	ProviderOpenRouter = "openrouter"
)

type Config struct {
	DefaultProvider       string                    `yaml:"default_provider"`
	DefaultModel          string                    `yaml:"default_model"`
	DefaultSourceLanguage string                    `yaml:"default_source_language"`
	DefaultTargetLanguage string                    `yaml:"default_target_language"`
	Timeout               time.Duration             `yaml:"timeout"`
	MaxOutputTokens       int                       `yaml:"max_output_tokens"`
	Temperature           float64                   `yaml:"temperature"`
	Chunking              ChunkingConfig            `yaml:"chunking"`
	Providers             map[string]ProviderConfig `yaml:"providers"`
}

type ChunkingConfig struct {
	Enabled  bool `yaml:"enabled"`
	MaxChars int  `yaml:"max_chars"`
}

type ProviderConfig struct {
	APIKeyEnv    string `yaml:"api_key_env"`
	DefaultModel string `yaml:"default_model"`
	BaseURL      string `yaml:"base_url,omitempty"`
	AppName      string `yaml:"app_name,omitempty"`
	AppURL       string `yaml:"app_url,omitempty"`
}

type Overrides struct {
	Provider        string
	Model           string
	SourceLanguage  string
	TargetLanguage  string
	Timeout         time.Duration
	MaxOutputTokens int
	Temperature     *float64
	ConfigPath      string
}

func Default() Config {
	return Config{
		DefaultProvider:       ProviderAnthropic,
		DefaultModel:          "claude-3-5-sonnet-latest",
		DefaultSourceLanguage: "auto",
		DefaultTargetLanguage: "",
		Timeout:               120 * time.Second,
		MaxOutputTokens:       4096,
		Temperature:           0,
		Chunking: ChunkingConfig{
			Enabled:  true,
			MaxChars: 12000,
		},
		Providers: map[string]ProviderConfig{
			ProviderAnthropic: {
				APIKeyEnv:    "ANTHROPIC_API_KEY",
				DefaultModel: "claude-3-5-sonnet-latest",
			},
			ProviderGemini: {
				APIKeyEnv:    "GEMINI_API_KEY",
				DefaultModel: "gemini-1.5-pro",
			},
			ProviderOpenRouter: {
				APIKeyEnv:    "OPENROUTER_API_KEY",
				DefaultModel: "anthropic/claude-3.5-sonnet",
				AppName:      "aitranslate",
			},
		},
	}
}

func DefaultPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "aitranslate", "config.yaml"), nil
}

func Load(path string) (Config, string, error) {
	cfg := Default()
	if path == "" {
		path = os.Getenv("AITRANSLATE_CONFIG")
	}
	if path == "" {
		var err error
		path, err = DefaultPath()
		if err != nil {
			return cfg, "", err
		}
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			ApplyEnv(&cfg)
			return cfg, path, nil
		}
		return cfg, path, err
	}
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return cfg, path, err
	}
	mergeDefaults(&cfg)
	ApplyEnv(&cfg)
	return cfg, path, nil
}

func ApplyOverrides(cfg *Config, overrides Overrides) {
	if overrides.Provider != "" {
		cfg.DefaultProvider = overrides.Provider
	}
	if overrides.Model != "" {
		cfg.DefaultModel = overrides.Model
	}
	if overrides.SourceLanguage != "" {
		cfg.DefaultSourceLanguage = overrides.SourceLanguage
	}
	if overrides.TargetLanguage != "" {
		cfg.DefaultTargetLanguage = overrides.TargetLanguage
	}
	if overrides.Timeout > 0 {
		cfg.Timeout = overrides.Timeout
	}
	if overrides.MaxOutputTokens > 0 {
		cfg.MaxOutputTokens = overrides.MaxOutputTokens
	}
	if overrides.Temperature != nil {
		cfg.Temperature = *overrides.Temperature
	}
}

func ApplyEnv(cfg *Config) {
	if v := os.Getenv("AITRANSLATE_PROVIDER"); v != "" {
		cfg.DefaultProvider = v
	}
	if v := os.Getenv("AITRANSLATE_MODEL"); v != "" {
		cfg.DefaultModel = v
	}
	if v := os.Getenv("AITRANSLATE_FROM"); v != "" {
		cfg.DefaultSourceLanguage = v
	}
	if v := os.Getenv("AITRANSLATE_TO"); v != "" {
		cfg.DefaultTargetLanguage = v
	}
}

func Validate(cfg Config) error {
	if !IsProvider(cfg.DefaultProvider) {
		return fmt.Errorf("invalid provider %q: expected anthropic, gemini, or openrouter", cfg.DefaultProvider)
	}
	if cfg.DefaultTargetLanguage == "" {
		return errors.New("missing target language: pass --to or set default_target_language")
	}
	if cfg.DefaultModel == "" {
		if provider, ok := cfg.Providers[cfg.DefaultProvider]; !ok || provider.DefaultModel == "" {
			return errors.New("missing model: pass --model or configure a provider default_model")
		}
	}
	provider := cfg.Providers[cfg.DefaultProvider]
	if provider.APIKeyEnv == "" {
		return fmt.Errorf("missing credentials config: set providers.%s.api_key_env", cfg.DefaultProvider)
	}
	if os.Getenv(provider.APIKeyEnv) == "" {
		return MissingCredentialsError{Provider: cfg.DefaultProvider, Env: provider.APIKeyEnv}
	}
	return nil
}

func IsProvider(provider string) bool {
	switch provider {
	case ProviderAnthropic, ProviderGemini, ProviderOpenRouter:
		return true
	default:
		return false
	}
}

func ModelFor(cfg Config) string {
	if cfg.DefaultModel != "" {
		return cfg.DefaultModel
	}
	return cfg.Providers[cfg.DefaultProvider].DefaultModel
}

func APIKey(cfg Config) string {
	return os.Getenv(cfg.Providers[cfg.DefaultProvider].APIKeyEnv)
}

func Write(path string, cfg Config, force bool) error {
	if path == "" {
		var err error
		path, err = DefaultPath()
		if err != nil {
			return err
		}
	}
	if !force {
		if _, err := os.Stat(path); err == nil {
			return fmt.Errorf("config file exists: pass --force to overwrite %s", path)
		} else if !errors.Is(err, os.ErrNotExist) {
			return err
		}
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o600)
}

type MissingCredentialsError struct {
	Provider string
	Env      string
}

func (e MissingCredentialsError) Error() string {
	return fmt.Sprintf("missing credentials: set %s or configure providers.%s.api_key_env", e.Env, e.Provider)
}

func mergeDefaults(cfg *Config) {
	def := Default()
	if cfg.DefaultProvider == "" {
		cfg.DefaultProvider = def.DefaultProvider
	}
	if cfg.DefaultSourceLanguage == "" {
		cfg.DefaultSourceLanguage = def.DefaultSourceLanguage
	}
	if cfg.Timeout == 0 {
		cfg.Timeout = def.Timeout
	}
	if cfg.MaxOutputTokens == 0 {
		cfg.MaxOutputTokens = def.MaxOutputTokens
	}
	if cfg.Chunking.MaxChars == 0 {
		cfg.Chunking.MaxChars = def.Chunking.MaxChars
	}
	if cfg.Providers == nil {
		cfg.Providers = map[string]ProviderConfig{}
	}
	for name, provider := range def.Providers {
		current := cfg.Providers[name]
		if current.APIKeyEnv == "" {
			current.APIKeyEnv = provider.APIKeyEnv
		}
		if current.DefaultModel == "" {
			current.DefaultModel = provider.DefaultModel
		}
		if current.AppName == "" {
			current.AppName = provider.AppName
		}
		cfg.Providers[name] = current
	}
}

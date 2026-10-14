package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

// Env keys that the loader reads explicitly. These keys use the GOTRANSLATE_
// prefix and are documented in README.md.
const (
	envKeyProvider       = "GOTRANSLATE_PROVIDER"
	envKeyModel          = "GOTRANSLATE_MODEL"
	envKeySourceLanguage = "GOTRANSLATE_FROM"
	envKeyTargetLanguage = "GOTRANSLATE_TO"
	envKeyConfigPath     = "GOTRANSLATE_CONFIG"
)

// LoaderOption customizes the behavior of LoadWithLoader.
type LoaderOption func(*loader)

// WithViper supplies a preconfigured viper instance. Useful for tests.
func WithViper(v *viper.Viper) LoaderOption {
	return func(l *loader) { l.v = v }
}

// WithSkipDotenv prevents automatic loading of .env files.
func WithSkipDotenv() LoaderOption {
	return func(l *loader) { l.skipDotenv = true }
}

// WithSkipEnv prevents automatic loading of real environment variables.
func WithSkipEnv() LoaderOption {
	return func(l *loader) { l.skipEnv = true }
}

// WithDotenvPath loads a specific .env file path. If empty and loading is
// enabled, godotenv.Load() searches .env, .env.local, .env.<env>, etc.
func WithDotenvPath(path string) LoaderOption {
	return func(l *loader) { l.dotenvPath = path }
}

type loader struct {
	v          *viper.Viper
	skipDotenv bool
	skipEnv    bool
	dotenvPath string
}

// LoadWithLoader resolves configuration using viper and an optional .env file.
//
// Precedence (highest to lowest):
//  1. CLI overrides (ApplyOverrides)
//  2. Real environment variables (GOTRANSLATE_*)
//  3. Config file (resolved path, or GOTRANSLATE_CONFIG)
//  4. .env file (optional, never overwrites existing env vars)
//  5. Built-in defaults
//
// The returned path is the config file path that was used. It may differ from
// the requested path when the requested file does not exist and a default path
// is derived.
func LoadWithLoader(path string, overrides Overrides, opts ...LoaderOption) (Config, string, error) {
	cfg := Default()

	l := &loader{
		v: viper.New(),
	}
	for _, opt := range opts {
		opt(l)
	}
	v := l.v

	v.SetTypeByDefaultValue(true)

	if !l.skipEnv {
		v.SetEnvPrefix("GOTRANSLATE")
		v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
		v.AutomaticEnv()
	}

	if !l.skipDotenv && os.Getenv("CI") == "" && os.Getenv("GOTRANSLATE_NO_DOTENV") == "" {
		if l.dotenvPath != "" {
			_ = godotenv.Load(l.dotenvPath)
		} else {
			_ = godotenv.Load()
		}
	}

	resolvedPath, err := resolveConfigPath(path)
	if err != nil {
		return cfg, "", err
	}

	envProvider := getEnvString(envKeyProvider, "")
	envModel := getEnvString(envKeyModel, "")
	envSourceLanguage := getEnvString(envKeySourceLanguage, "")
	envTargetLanguage := getEnvString(envKeyTargetLanguage, "")

	var loaded Config
	if err := setViperDefaults(v, cfg); err != nil {
		return loaded, resolvedPath, err
	}

	data, err := os.ReadFile(resolvedPath)
	if err == nil {
		v.SetConfigType("yaml")
		if err := v.ReadConfig(strings.NewReader(string(data))); err != nil {
			return loaded, resolvedPath, fmt.Errorf("invalid config file %s: %w", resolvedPath, err)
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return loaded, resolvedPath, err
	}

	loaded.DefaultProvider = cfg.DefaultProvider
	loaded.DefaultModel = cfg.DefaultModel
	loaded.DefaultSourceLanguage = cfg.DefaultSourceLanguage
	loaded.DefaultTargetLanguage = cfg.DefaultTargetLanguage
	loaded.Timeout = cfg.Timeout
	loaded.MaxOutputTokens = cfg.MaxOutputTokens
	loaded.Temperature = cfg.Temperature
	loaded.Chunking = cfg.Chunking
	loaded.Providers = cfg.Providers

	loaded.DefaultProvider = firstNonEmpty(envProvider, getString(v, "default_provider", loaded.DefaultProvider))
	loaded.DefaultModel = firstNonEmpty(envModel, getString(v, "default_model", loaded.DefaultModel))
	loaded.DefaultSourceLanguage = firstNonEmpty(envSourceLanguage, getString(v, "default_source_language", loaded.DefaultSourceLanguage))
	loaded.DefaultTargetLanguage = firstNonEmpty(envTargetLanguage, getString(v, "default_target_language", loaded.DefaultTargetLanguage))
	loaded.Timeout = getDuration(v, "timeout", loaded.Timeout)
	loaded.MaxOutputTokens = getInt(v, "max_output_tokens", loaded.MaxOutputTokens)
	loaded.Temperature = getFloat64(v, "temperature", loaded.Temperature)
	loaded.Chunking = ChunkingConfig{
		Enabled:  getBool(v, "chunking.enabled", loaded.Chunking.Enabled),
		MaxChars: getInt(v, "chunking.max_chars", loaded.Chunking.MaxChars),
	}
	loaded.Providers = make(map[string]ProviderConfig, len(cfg.Providers))
	for name, defProvider := range cfg.Providers {
		loaded.Providers[name] = ProviderConfig{
			APIKeyEnv:    getString(v, "providers."+name+".api_key_env", defProvider.APIKeyEnv),
			DefaultModel: getString(v, "providers."+name+".default_model", defProvider.DefaultModel),
			BaseURL:      getString(v, "providers."+name+".base_url", defProvider.BaseURL),
			AppName:      getString(v, "providers."+name+".app_name", defProvider.AppName),
			AppURL:       getString(v, "providers."+name+".app_url", defProvider.AppURL),
			ProjectID:    getString(v, "providers."+name+".project_id", defProvider.ProjectID),
			Location:     getString(v, "providers."+name+".location", defProvider.Location),
			MimeType:     getString(v, "providers."+name+".mime_type", defProvider.MimeType),
		}
	}

	loaded.Providers = mergeProviderMap(cfg.Providers, loaded.Providers)
	mergeDefaults(&loaded)
	ApplyOverrides(&loaded, overrides)
	return loaded, resolvedPath, nil
}

func resolveConfigPath(path string) (string, error) {
	if path != "" {
		return path, nil
	}
	if v := getEnvString(envKeyConfigPath, ""); v != "" {
		return v, nil
	}
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "gotranslate", "config.yaml"), nil
}

func setViperDefaults(v *viper.Viper, def Config) error {
	v.SetDefault("default_provider", def.DefaultProvider)
	v.SetDefault("default_model", def.DefaultModel)
	v.SetDefault("default_source_language", def.DefaultSourceLanguage)
	v.SetDefault("default_target_language", def.DefaultTargetLanguage)
	v.SetDefault("timeout", def.Timeout)
	v.SetDefault("max_output_tokens", def.MaxOutputTokens)
	v.SetDefault("temperature", def.Temperature)
	v.SetDefault("chunking.enabled", def.Chunking.Enabled)
	v.SetDefault("chunking.max_chars", def.Chunking.MaxChars)

	for name, p := range def.Providers {
		prefix := "providers." + name
		v.SetDefault(prefix+".api_key_env", p.APIKeyEnv)
		v.SetDefault(prefix+".default_model", p.DefaultModel)
		v.SetDefault(prefix+".base_url", p.BaseURL)
		v.SetDefault(prefix+".app_name", p.AppName)
		v.SetDefault(prefix+".app_url", p.AppURL)
		v.SetDefault(prefix+".project_id", p.ProjectID)
		v.SetDefault(prefix+".location", p.Location)
		v.SetDefault(prefix+".mime_type", p.MimeType)
	}

	return nil
}

func mergeProviderMap(defaults, loaded map[string]ProviderConfig) map[string]ProviderConfig {
	if loaded == nil {
		return defaults
	}
	merged := make(map[string]ProviderConfig, len(defaults))
	for k, v := range defaults {
		merged[k] = v
	}
	for k, v := range loaded {
		merged[k] = v
	}
	return merged
}

func getString(v *viper.Viper, key, fallback string) string {
	if val := v.GetString(key); val != "" {
		return val
	}
	return fallback
}

func getBool(v *viper.Viper, key string, fallback bool) bool {
	if v.IsSet(key) {
		return v.GetBool(key)
	}
	return fallback
}

func getInt(v *viper.Viper, key string, fallback int) int {
	if v.IsSet(key) {
		return v.GetInt(key)
	}
	return fallback
}

func getFloat64(v *viper.Viper, key string, fallback float64) float64 {
	if v.IsSet(key) {
		return v.GetFloat64(key)
	}
	return fallback
}

func getDuration(v *viper.Viper, key string, fallback time.Duration) time.Duration {
	if v.IsSet(key) {
		return v.GetDuration(key)
	}
	return fallback
}

func getEnvString(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}

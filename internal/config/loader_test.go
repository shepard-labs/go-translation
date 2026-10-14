package config

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/spf13/viper"
)

func TestLoadWithLoaderDefaults(t *testing.T) {
	cfg, path, err := LoadWithLoader("", Overrides{}, WithSkipDotenv(), WithSkipEnv())
	if err != nil {
		t.Fatalf("LoadWithLoader error: %v", err)
	}
	if cfg.DefaultProvider != ProviderAnthropic {
		t.Fatalf("default provider = %s, want %s", cfg.DefaultProvider, ProviderAnthropic)
	}
	if cfg.Timeout == 0 {
		t.Fatal("timeout not set from defaults")
	}
	if path == "" {
		t.Fatal("expected resolved config path")
	}
}

func TestLoadWithLoaderConfigFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	data := `default_provider: gemini
default_target_language: French
timeout: 60s
`
	if err := os.WriteFile(path, []byte(data), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}

	cfg, _, err := LoadWithLoader(path, Overrides{}, WithSkipDotenv(), WithSkipEnv())
	if err != nil {
		t.Fatalf("LoadWithLoader error: %v", err)
	}
	if cfg.DefaultProvider != ProviderGemini {
		t.Fatalf("provider = %s, want gemini", cfg.DefaultProvider)
	}
	if cfg.DefaultTargetLanguage != "French" {
		t.Fatalf("target = %s, want French", cfg.DefaultTargetLanguage)
	}
	if cfg.Timeout != 60*time.Second {
		t.Fatalf("timeout = %v, want 60s", cfg.Timeout)
	}
}

func TestLoadWithLoaderEnvOverrides(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	data := `default_provider: anthropic
default_target_language: French
`
	if err := os.WriteFile(path, []byte(data), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}

	t.Setenv("GOTRANSLATE_PROVIDER", ProviderGemini)
	t.Setenv("GOTRANSLATE_TO", "German")

	cfg, _, err := LoadWithLoader(path, Overrides{}, WithSkipDotenv())
	if err != nil {
		t.Fatalf("LoadWithLoader error: %v", err)
	}
	if cfg.DefaultProvider != ProviderGemini {
		t.Fatalf("provider = %s, want gemini from env", cfg.DefaultProvider)
	}
	if cfg.DefaultTargetLanguage != "German" {
		t.Fatalf("target = %s, want German from env", cfg.DefaultTargetLanguage)
	}
}

func TestLoadWithLoaderFlagOverrides(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	data := `default_provider: anthropic
default_target_language: French
`
	if err := os.WriteFile(path, []byte(data), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}

	t.Setenv("GOTRANSLATE_TO", "German")

	cfg, _, err := LoadWithLoader(path, Overrides{TargetLanguage: "Spanish"}, WithSkipDotenv())
	if err != nil {
		t.Fatalf("LoadWithLoader error: %v", err)
	}
	if cfg.DefaultTargetLanguage != "Spanish" {
		t.Fatalf("target = %s, want Spanish from overrides", cfg.DefaultTargetLanguage)
	}
}

func TestLoadWithLoaderDotenv(t *testing.T) {
	dir := t.TempDir()
	dotenv := filepath.Join(dir, ".env")
	if err := os.WriteFile(dotenv, []byte("GOTRANSLATE_TO=Italian\n"), 0o600); err != nil {
		t.Fatalf("write .env: %v", err)
	}

	cfg, _, err := LoadWithLoader("", Overrides{}, WithDotenvPath(dotenv), WithSkipEnv())
	if err != nil {
		t.Fatalf("LoadWithLoader error: %v", err)
	}
	if cfg.DefaultTargetLanguage != "Italian" {
		t.Fatalf("target = %s, want Italian from .env", cfg.DefaultTargetLanguage)
	}
}

func TestLoadWithLoaderDotenvDoesNotOverwriteExistingEnv(t *testing.T) {
	dir := t.TempDir()
	dotenv := filepath.Join(dir, ".env")
	if err := os.WriteFile(dotenv, []byte("GOTRANSLATE_TO=Italian\n"), 0o600); err != nil {
		t.Fatalf("write .env: %v", err)
	}

	t.Setenv("GOTRANSLATE_TO", "Spanish")

	cfg, _, err := LoadWithLoader("", Overrides{}, WithDotenvPath(dotenv))
	if err != nil {
		t.Fatalf("LoadWithLoader error: %v", err)
	}
	if cfg.DefaultTargetLanguage != "Spanish" {
		t.Fatalf("target = %s, want Spanish from real env", cfg.DefaultTargetLanguage)
	}
}

func TestLoadWithLoaderMissingFileUsesDefaults(t *testing.T) {
	cfg, path, err := LoadWithLoader(filepath.Join(t.TempDir(), "missing.yaml"), Overrides{}, WithSkipDotenv(), WithSkipEnv())
	if err != nil {
		t.Fatalf("LoadWithLoader error: %v", err)
	}
	if cfg.DefaultProvider != ProviderAnthropic {
		t.Fatalf("expected default provider, got %s", cfg.DefaultProvider)
	}
	if path == "" {
		t.Fatal("expected resolved path even for missing file")
	}
}

func TestLoadWithLoaderInvalidYAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(path, []byte("default_provider: [\n"), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}

	_, _, err := LoadWithLoader(path, Overrides{}, WithSkipDotenv(), WithSkipEnv())
	if err == nil {
		t.Fatal("expected error for invalid YAML")
	}
}

func TestLoadWithLoaderViperInjection(t *testing.T) {
	v := viper.New()
	v.Set("default_provider", ProviderOpenRouter)
	v.Set("default_target_language", "Dutch")

	cfg, _, err := LoadWithLoader("", Overrides{TargetLanguage: "Dutch"}, WithViper(v), WithSkipDotenv(), WithSkipEnv())
	if err != nil {
		t.Fatalf("LoadWithLoader error: %v", err)
	}
	if cfg.DefaultProvider != ProviderOpenRouter {
		t.Fatalf("provider = %s, want openrouter", cfg.DefaultProvider)
	}
	if cfg.DefaultTargetLanguage != "Dutch" {
		t.Fatalf("target = %s, want Dutch", cfg.DefaultTargetLanguage)
	}
}

func TestLoadWithLoaderTemperature(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(path, []byte(fmt.Sprintf("temperature: %v\n", 0.7)), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}

	cfg, _, err := LoadWithLoader(path, Overrides{}, WithSkipDotenv(), WithSkipEnv())
	if err != nil {
		t.Fatalf("LoadWithLoader error: %v", err)
	}
	if cfg.Temperature != 0.7 {
		t.Fatalf("temperature = %v, want 0.7", cfg.Temperature)
	}
}

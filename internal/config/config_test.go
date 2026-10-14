package config

import "testing"

func TestApplyOverrides(t *testing.T) {
	cfg := Default()
	temp := 0.2
	ApplyOverrides(&cfg, Overrides{Provider: ProviderGemini, Model: "gemini-test", TargetLanguage: "French", Temperature: &temp})
	if cfg.DefaultProvider != ProviderGemini {
		t.Fatalf("provider = %s", cfg.DefaultProvider)
	}
	if cfg.DefaultModel != "gemini-test" {
		t.Fatalf("model = %s", cfg.DefaultModel)
	}
	if cfg.DefaultTargetLanguage != "French" {
		t.Fatalf("target = %s", cfg.DefaultTargetLanguage)
	}
	if cfg.Temperature != temp {
		t.Fatalf("temperature = %v", cfg.Temperature)
	}
}

func TestIsProvider(t *testing.T) {
	if !IsProvider(ProviderAnthropic) || !IsProvider(ProviderGemini) || !IsProvider(ProviderOpenRouter) {
		t.Fatal("expected known providers to be valid")
	}
	if IsProvider("bad") {
		t.Fatal("expected bad provider to be invalid")
	}
}

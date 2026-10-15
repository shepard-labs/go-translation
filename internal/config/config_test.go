package config

import (
	"strings"
	"testing"
)

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

func TestGoogleCloudClearsLLMDefaultModel(t *testing.T) {
	// Default() ships an Anthropic model name; googlecloud must not send it
	// to Cloud Translation as a model resource.
	cfg := Default()
	cfg.DefaultProvider = ProviderGoogleCloud
	mergeDefaults(&cfg)
	if cfg.DefaultModel != "" {
		t.Fatalf("DefaultModel = %q, want empty for googlecloud", cfg.DefaultModel)
	}
	if got := ModelFor(cfg); got != "" {
		t.Fatalf("ModelFor = %q, want empty", got)
	}

	// A user-supplied Translation LLM resource name is preserved.
	cfg2 := Default()
	cfg2.DefaultProvider = ProviderGoogleCloud
	cfg2.DefaultModel = "projects/p/locations/global/models/general/translation-llm"
	mergeDefaults(&cfg2)
	if cfg2.DefaultModel != "projects/p/locations/global/models/general/translation-llm" {
		t.Fatalf("DefaultModel = %q, want preserved", cfg2.DefaultModel)
	}

	// Switching provider via overrides clears a stale LLM model.
	cfg3 := Default()
	ApplyOverrides(&cfg3, Overrides{Provider: ProviderGoogleCloud})
	if cfg3.DefaultModel != "" {
		t.Fatalf("DefaultModel = %q after provider override, want empty", cfg3.DefaultModel)
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

func TestValidateRejectsGoogleCloudNMT(t *testing.T) {
	t.Setenv("GOOGLE_APPLICATION_CREDENTIALS", "/tmp/fake-creds.json")
	cfg := Default()
	cfg.DefaultProvider = ProviderGoogleCloud
	cfg.DefaultTargetLanguage = "Spanish"
	cfg.DefaultModel = "projects/p/locations/global/models/general/nmt"
	pc := cfg.Providers[ProviderGoogleCloud]
	pc.ProjectID = "p"
	cfg.Providers[ProviderGoogleCloud] = pc

	err := Validate(cfg)
	if err == nil {
		t.Fatal("expected Validate to reject NMT")
	}
	if !strings.Contains(err.Error(), "NMT is not allowed") {
		t.Fatalf("error = %q, want NMT rejection", err)
	}
}

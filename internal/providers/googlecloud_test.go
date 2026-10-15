package providers

import (
	"reflect"
	"strings"
	"testing"

	"github.com/shepard-labs/go-translation/internal/translate"
)

func TestNewGoogleCloud(t *testing.T) {
	g := NewGoogleCloud("/path/to/creds.json", "proj-123", "us-central1", "model-1", "text/html")
	if g.credentialsPath != "/path/to/creds.json" {
		t.Errorf("credentialsPath = %q, want %q", g.credentialsPath, "/path/to/creds.json")
	}
	if g.projectID != "proj-123" {
		t.Errorf("projectID = %q, want %q", g.projectID, "proj-123")
	}
	if g.location != "us-central1" {
		t.Errorf("location = %q, want %q", g.location, "us-central1")
	}
	if g.defaultModel != "model-1" {
		t.Errorf("defaultModel = %q, want %q", g.defaultModel, "model-1")
	}
	if g.mimeType != "text/html" {
		t.Errorf("mimeType = %q, want %q", g.mimeType, "text/html")
	}
}

func TestNewGoogleCloudDefaults(t *testing.T) {
	g := NewGoogleCloud("/path/to/creds.json", "proj-123", "", "", "")
	if g.location != "global" {
		t.Errorf("location default = %q, want %q", g.location, "global")
	}
	if g.mimeType != "text/plain" {
		t.Errorf("mimeType default = %q, want %q", g.mimeType, "text/plain")
	}
}

func TestTranslationLLMModel(t *testing.T) {
	got := TranslationLLMModel("proj-123", "global")
	want := "projects/proj-123/locations/global/models/general/translation-llm"
	if got != want {
		t.Errorf("TranslationLLMModel = %q, want %q", got, want)
	}
}

func TestIsGoogleCloudNMTModel(t *testing.T) {
	cases := []struct {
		model string
		want  bool
	}{
		{"", true},
		{"nmt", true},
		{"general/nmt", true},
		{"projects/p/locations/global/models/general/nmt", true},
		{"general/translation-llm", false},
		{"projects/p/locations/global/models/general/translation-llm", false},
		{"projects/p/locations/us-central1/models/translation-llm-custom/abc", false},
	}
	for _, tc := range cases {
		if got := IsGoogleCloudNMTModel(tc.model); got != tc.want {
			t.Errorf("IsGoogleCloudNMTModel(%q) = %v, want %v", tc.model, got, tc.want)
		}
	}
}

func TestGoogleCloudBuildRequest(t *testing.T) {
	g := NewGoogleCloud("/path/to/creds.json", "proj-123", "us-central1", "", "text/html")
	llm := TranslationLLMModel("proj-123", "us-central1")
	req := translate.TranslationRequest{
		Provider:       "googlecloud",
		Model:          llm,
		SourceLanguage: "en",
		TargetLanguage: "fr",
		Text:           "hello",
	}

	got, modelID, err := g.BuildRequest(req)
	if err != nil {
		t.Fatalf("BuildRequest: %v", err)
	}

	if got.GetParent() != "projects/proj-123/locations/us-central1" {
		t.Errorf("Parent = %q, want %q", got.GetParent(), "projects/proj-123/locations/us-central1")
	}
	if !reflect.DeepEqual(got.GetContents(), []string{"hello"}) {
		t.Errorf("Contents = %v, want %v", got.GetContents(), []string{"hello"})
	}
	if got.GetTargetLanguageCode() != "fr" {
		t.Errorf("TargetLanguageCode = %q, want %q", got.GetTargetLanguageCode(), "fr")
	}
	if got.GetSourceLanguageCode() != "en" {
		t.Errorf("SourceLanguageCode = %q, want %q", got.GetSourceLanguageCode(), "en")
	}
	if got.GetMimeType() != "text/html" {
		t.Errorf("MimeType = %q, want %q", got.GetMimeType(), "text/html")
	}
	if got.GetModel() != llm || modelID != llm {
		t.Errorf("Model = %q / %q, want %q", got.GetModel(), modelID, llm)
	}
}

func TestGoogleCloudBuildRequestDefaultsToTranslationLLM(t *testing.T) {
	g := NewGoogleCloud("/path/to/creds.json", "proj-123", "global", "", "text/plain")
	req := translate.TranslationRequest{
		SourceLanguage: "auto",
		TargetLanguage: "es",
		Text:           "world",
	}

	got, modelID, err := g.BuildRequest(req)
	if err != nil {
		t.Fatalf("BuildRequest: %v", err)
	}

	want := TranslationLLMModel("proj-123", "global")
	if got.GetSourceLanguageCode() != "" {
		t.Errorf("SourceLanguageCode = %q, want empty for auto", got.GetSourceLanguageCode())
	}
	if got.GetModel() != want || modelID != want {
		t.Errorf("Model = %q / %q, want Translation LLM %q", got.GetModel(), modelID, want)
	}
}

func TestGoogleCloudBuildRequestRejectsNMT(t *testing.T) {
	g := NewGoogleCloud("/path/to/creds.json", "proj-123", "global", "", "text/plain")
	req := translate.TranslationRequest{
		Model:          "projects/proj-123/locations/global/models/general/nmt",
		TargetLanguage: "es",
		Text:           "hello",
	}

	_, _, err := g.BuildRequest(req)
	if err == nil {
		t.Fatal("expected error for NMT model")
	}
	if !strings.Contains(err.Error(), "NMT is not allowed") {
		t.Fatalf("error = %q, want NMT rejection", err)
	}
}

func TestGoogleCloudBuildRequestLanguageNames(t *testing.T) {
	g := NewGoogleCloud("/path/to/creds.json", "proj-123", "global", "", "text/plain")
	req := translate.TranslationRequest{
		SourceLanguage: "English",
		TargetLanguage: "Spanish",
		Text:           "hello",
	}

	got, _, err := g.BuildRequest(req)
	if err != nil {
		t.Fatalf("BuildRequest: %v", err)
	}

	if got.GetSourceLanguageCode() != "en" {
		t.Errorf("SourceLanguageCode = %q, want %q", got.GetSourceLanguageCode(), "en")
	}
	if got.GetTargetLanguageCode() != "es" {
		t.Errorf("TargetLanguageCode = %q, want %q", got.GetTargetLanguageCode(), "es")
	}
}

func TestGoogleCloudBuildRequestLanguageCodePassthrough(t *testing.T) {
	g := NewGoogleCloud("/path/to/creds.json", "proj-123", "global", "", "text/plain")
	req := translate.TranslationRequest{
		SourceLanguage: "auto",
		TargetLanguage: "zh-CN",
		Text:           "hello",
	}

	got, _, err := g.BuildRequest(req)
	if err != nil {
		t.Fatalf("BuildRequest: %v", err)
	}

	if got.GetSourceLanguageCode() != "" {
		t.Errorf("SourceLanguageCode = %q, want empty for auto", got.GetSourceLanguageCode())
	}
	if got.GetTargetLanguageCode() != "zh-CN" {
		t.Errorf("TargetLanguageCode = %q, want %q", got.GetTargetLanguageCode(), "zh-CN")
	}
}

func TestGoogleCloudBuildRequestShortTranslationLLM(t *testing.T) {
	g := NewGoogleCloud("/path/to/creds.json", "proj-123", "global", "", "text/plain")
	req := translate.TranslationRequest{
		Model:          "general/translation-llm",
		TargetLanguage: "de",
		Text:           "hello",
	}

	got, modelID, err := g.BuildRequest(req)
	if err != nil {
		t.Fatalf("BuildRequest: %v", err)
	}

	want := TranslationLLMModel("proj-123", "global")
	if got.GetModel() != want || modelID != want {
		t.Errorf("Model = %q / %q, want %q", got.GetModel(), modelID, want)
	}
}

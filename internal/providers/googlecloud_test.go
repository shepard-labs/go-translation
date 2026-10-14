package providers

import (
	"reflect"
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

func TestGoogleCloudBuildRequest(t *testing.T) {
	g := NewGoogleCloud("/path/to/creds.json", "proj-123", "us-central1", "model-1", "text/html")
	req := translate.TranslationRequest{
		Provider:       "googlecloud",
		Model:          "model-2",
		SourceLanguage: "en",
		TargetLanguage: "fr",
		Text:           "hello",
	}

	got := g.BuildRequest(req)

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
	if got.GetModel() != "model-2" {
		t.Errorf("Model = %q, want %q", got.GetModel(), "model-2")
	}
}

func TestGoogleCloudBuildRequestOmitSourceWhenAuto(t *testing.T) {
	g := NewGoogleCloud("/path/to/creds.json", "proj-123", "global", "", "text/plain")
	req := translate.TranslationRequest{
		SourceLanguage: "auto",
		TargetLanguage: "es",
		Text:           "world",
	}

	got := g.BuildRequest(req)

	if got.GetSourceLanguageCode() != "" {
		t.Errorf("SourceLanguageCode = %q, want empty for auto", got.GetSourceLanguageCode())
	}
	if got.GetModel() != "" {
		t.Errorf("Model = %q, want empty", got.GetModel())
	}
}

func TestGoogleCloudBuildRequestLanguageNames(t *testing.T) {
	g := NewGoogleCloud("/path/to/creds.json", "proj-123", "global", "", "text/plain")
	req := translate.TranslationRequest{
		SourceLanguage: "English",
		TargetLanguage: "Spanish",
		Text:           "hello",
	}

	got := g.BuildRequest(req)

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

	got := g.BuildRequest(req)

	if got.GetSourceLanguageCode() != "" {
		t.Errorf("SourceLanguageCode = %q, want empty for auto", got.GetSourceLanguageCode())
	}
	if got.GetTargetLanguageCode() != "zh-CN" {
		t.Errorf("TargetLanguageCode = %q, want %q", got.GetTargetLanguageCode(), "zh-CN")
	}
}

func TestGoogleCloudBuildRequestDefaultModel(t *testing.T) {
	g := NewGoogleCloud("/path/to/creds.json", "proj-123", "global", "default-model", "text/plain")
	req := translate.TranslationRequest{
		TargetLanguage: "de",
		Text:           "hello",
	}

	got := g.BuildRequest(req)

	if got.GetModel() != "default-model" {
		t.Errorf("Model = %q, want %q", got.GetModel(), "default-model")
	}
}

package providers

import (
	"context"
	"fmt"
	"strings"

	"github.com/shepard-labs/go-translation/internal/translate"

	cloudtranslate "cloud.google.com/go/translate/apiv3"
	cloudtranslatepb "cloud.google.com/go/translate/apiv3/translatepb"
	googleapioption "google.golang.org/api/option"
)

const (
	// googleCloudTranslationLLMID is the only allowed standard Cloud Translation model.
	// Omitting model (or using general/nmt) would fall back to NMT — which is forbidden.
	googleCloudTranslationLLMID = "general/translation-llm"
)

type GoogleCloud struct {
	credentialsPath string
	projectID       string
	location        string
	defaultModel    string
	mimeType        string
}

func NewGoogleCloud(credentialsPath, projectID, location, defaultModel, mimeType string) GoogleCloud {
	if location == "" {
		location = "global"
	}
	if mimeType == "" {
		mimeType = "text/plain"
	}
	return GoogleCloud{
		credentialsPath: credentialsPath,
		projectID:       projectID,
		location:        location,
		defaultModel:    defaultModel,
		mimeType:        mimeType,
	}
}

// TranslationLLMModel returns the full Advanced API resource name for Translation LLM.
func TranslationLLMModel(projectID, location string) string {
	if location == "" {
		location = "global"
	}
	return fmt.Sprintf("projects/%s/locations/%s/models/%s", projectID, location, googleCloudTranslationLLMID)
}

// IsGoogleCloudNMTModel reports whether model selects Cloud Translation NMT
// (including the API default when model is omitted).
func IsGoogleCloudNMTModel(model string) bool {
	m := strings.ToLower(strings.TrimSpace(model))
	if m == "" {
		// Empty model → Cloud Translation uses NMT. Treated as NMT so callers reject it.
		return true
	}
	if m == "nmt" || m == "general/nmt" {
		return true
	}
	return strings.HasSuffix(m, "/models/general/nmt") || strings.Contains(m, "/models/general/nmt/")
}

func (g GoogleCloud) Translate(ctx context.Context, req translate.TranslationRequest) (translate.TranslationResult, error) {
	client, err := cloudtranslate.NewTranslationClient(ctx, googleapioption.WithAuthCredentialsFile(googleapioption.ServiceAccount, g.credentialsPath))
	if err != nil {
		return translate.TranslationResult{}, fmt.Errorf("creating google cloud translation client: %w", err)
	}
	defer client.Close()

	request, modelID, err := g.BuildRequest(req)
	if err != nil {
		return translate.TranslationResult{}, err
	}
	resp, err := client.TranslateText(ctx, request)
	if err != nil {
		return translate.TranslationResult{}, err
	}

	if len(resp.GetTranslations()) == 0 {
		return translate.TranslationResult{}, fmt.Errorf("google cloud returned no translations")
	}

	translation := resp.GetTranslations()[0]
	return translate.TranslationResult{
		Text:     translation.GetTranslatedText(),
		Provider: "googlecloud",
		Model:    modelID,
		Usage:    translate.Usage{},
	}, nil
}

func (g GoogleCloud) BuildRequest(req translate.TranslationRequest) (*cloudtranslatepb.TranslateTextRequest, string, error) {
	modelID, err := g.resolveModel(req.Model)
	if err != nil {
		return nil, "", err
	}

	source := languageCode(req.SourceLanguage)
	if source == "auto" {
		source = ""
	}

	parent := fmt.Sprintf("projects/%s/locations/%s", g.projectID, g.location)

	request := &cloudtranslatepb.TranslateTextRequest{
		Parent:             parent,
		Contents:           []string{req.Text},
		TargetLanguageCode: languageCode(req.TargetLanguage),
		MimeType:           g.mimeType,
		// Always set — omitting Model makes the API use NMT.
		Model: modelID,
	}

	if source != "" {
		request.SourceLanguageCode = source
	}

	return request, modelID, nil
}

// resolveModel returns a non-NMT Cloud Translation model resource name.
// Empty request/default model resolves to Translation LLM for this project/location.
func (g GoogleCloud) resolveModel(requested string) (string, error) {
	modelID := strings.TrimSpace(requested)
	if modelID == "" {
		modelID = strings.TrimSpace(g.defaultModel)
	}

	switch strings.ToLower(modelID) {
	case "", googleCloudTranslationLLMID, "translation-llm":
		if g.projectID == "" {
			return "", fmt.Errorf("googlecloud: project_id is required to use Translation LLM")
		}
		return TranslationLLMModel(g.projectID, g.location), nil
	}

	if IsGoogleCloudNMTModel(modelID) {
		return "", fmt.Errorf("googlecloud: NMT is not allowed (model %q); use Translation LLM (%s)", modelID, googleCloudTranslationLLMID)
	}

	// Short custom ids are expanded; full projects/... paths pass through if non-NMT.
	if !strings.HasPrefix(modelID, "projects/") {
		if g.projectID == "" {
			return "", fmt.Errorf("googlecloud: project_id is required for model %q", modelID)
		}
		modelID = fmt.Sprintf("projects/%s/locations/%s/models/%s", g.projectID, g.location, modelID)
		if IsGoogleCloudNMTModel(modelID) {
			return "", fmt.Errorf("googlecloud: NMT is not allowed (model %q); use Translation LLM (%s)", modelID, googleCloudTranslationLLMID)
		}
	}

	return modelID, nil
}

// languageCode maps common English language names to ISO 639-1 codes so the
// same --to values that work for LLM providers ("Spanish", "French") work for
// Cloud Translation. Values that already look like codes pass through.
func languageCode(lang string) string {
	if code, ok := languageCodes[strings.ToLower(strings.TrimSpace(lang))]; ok {
		return code
	}
	return lang
}

var languageCodes = map[string]string{
	"afrikaans": "af", "albanian": "sq", "arabic": "ar", "armenian": "hy",
	"azerbaijani": "az", "basque": "eu", "belarusian": "be", "bengali": "bn",
	"bosnian": "bs", "bulgarian": "bg", "catalan": "ca", "chinese": "zh",
	"chinese (simplified)": "zh-CN", "chinese (traditional)": "zh-TW",
	"croatian": "hr", "czech": "cs", "danish": "da", "dutch": "nl",
	"english": "en", "estonian": "et", "finnish": "fi", "french": "fr",
	"galician": "gl", "georgian": "ka", "german": "de", "greek": "el",
	"gujarati": "gu", "haitian creole": "ht", "hebrew": "iw", "hindi": "hi",
	"hungarian": "hu", "icelandic": "is", "indonesian": "id", "irish": "ga",
	"italian": "it", "japanese": "ja", "kannada": "kn", "kazakh": "kk",
	"korean": "ko", "latvian": "lv", "lithuanian": "lt", "macedonian": "mk",
	"malay": "ms", "malayalam": "ml", "maltese": "mt", "marathi": "mr",
	"norwegian": "no", "persian": "fa", "polish": "pl", "portuguese": "pt",
	"punjabi": "pa", "romanian": "ro", "russian": "ru", "serbian": "sr",
	"slovak": "sk", "slovenian": "sl", "spanish": "es", "swahili": "sw",
	"swedish": "sv", "tagalog": "tl", "tamil": "ta", "telugu": "te",
	"thai": "th", "turkish": "tr", "ukrainian": "uk", "urdu": "ur",
	"vietnamese": "vi", "welsh": "cy",
}

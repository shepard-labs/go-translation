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

func (g GoogleCloud) Translate(ctx context.Context, req translate.TranslationRequest) (translate.TranslationResult, error) {
	client, err := cloudtranslate.NewTranslationClient(ctx, googleapioption.WithCredentialsFile(g.credentialsPath))
	if err != nil {
		return translate.TranslationResult{}, fmt.Errorf("creating google cloud translation client: %w", err)
	}
	defer client.Close()

	request := g.BuildRequest(req)
	resp, err := client.TranslateText(ctx, request)
	if err != nil {
		return translate.TranslationResult{}, err
	}

	if len(resp.GetTranslations()) == 0 {
		return translate.TranslationResult{}, fmt.Errorf("google cloud returned no translations")
	}

	translation := resp.GetTranslations()[0]
	modelID := req.Model
	if modelID == "" {
		modelID = g.defaultModel
	}
	return translate.TranslationResult{
		Text:     translation.GetTranslatedText(),
		Provider: "googlecloud",
		Model:    modelID,
		Usage:    translate.Usage{},
	}, nil
}

func (g GoogleCloud) BuildRequest(req translate.TranslationRequest) *cloudtranslatepb.TranslateTextRequest {
	modelID := req.Model
	if modelID == "" {
		modelID = g.defaultModel
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
	}

	if source != "" {
		request.SourceLanguageCode = source
	}
	if modelID != "" {
		request.Model = modelID
	}

	return request
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

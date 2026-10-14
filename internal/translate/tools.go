package translate

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
)

// Tools builds the deterministic tool set offered to LLM providers during a
// translation. The model uses these instead of hallucinating: language
// detection, protected-span bookkeeping, glossary lookups, and placeholder
// self-verification.
//
// Each tool's dispatch function is pure Go — no network calls.
type Toolset struct {
	Glossary Glossary
	// Spans are the protected spans for the *current* request; populated when
	// the request carries pre-protected text (see WithProtection).
	Spans []string
}

type ToolCall struct {
	Name  string
	Input json.RawMessage
}

type ToolResult struct {
	Output  string
	IsError bool
}

// ToolDef describes a tool for a provider request (provider-agnostic shape —
// each provider adapter maps it to its own Tool type).
type ToolDef struct {
	Name        string
	Description string
	InputSchema map[string]any
}

// Definitions returns the tool schemas exposed to the model.
func (t Toolset) Definitions() []ToolDef {
	return []ToolDef{
		{
			Name:        "detect_language",
			Description: "Detect the ISO 639-1 language code of a text sample. Use before translating when the source language is 'auto'.",
			InputSchema: obj(map[string]any{"text": str("text sample to identify")}, "text"),
		},
		{
			Name:        "lookup_glossary",
			Description: "Look up the preferred translation of a term for the target language. Returns an empty string when no glossary entry exists.",
			InputSchema: obj(map[string]any{
				"term":            str("source term"),
				"target_language": str("target language code or name"),
			}, "term", "target_language"),
		},
		{
			Name:        "placeholder_check",
			Description: "Verify that every ⟦PNNN⟧ placeholder in the source appears in your translation. Returns the list of missing placeholder ids — empty means all preserved.",
			InputSchema: obj(map[string]any{
				"translation": str("your translated output to verify"),
			}, "translation"),
		},
		{
			Name:        "restore_spans",
			Description: "Replace all ⟦PNNN⟧ placeholders in your final translation with the original protected content (code, URLs, tags). Call once, at the end, then return the restored text.",
			InputSchema: obj(map[string]any{
				"translation": str("translated text containing ⟦PNNN⟧ placeholders"),
			}, "translation"),
		},
	}
}

// Dispatch executes a tool call. Unknown tools return an error result rather
// than failing the whole request.
func (t Toolset) Dispatch(ctx context.Context, call ToolCall) ToolResult {
	switch call.Name {
	case "detect_language":
		var in struct {
			Text string `json:"text"`
		}
		if err := json.Unmarshal(call.Input, &in); err != nil {
			return errResult(err)
		}
		code := DetectLanguage(in.Text)
		if code == "" {
			return okResult(`{"language":"auto","confidence":"low"}`)
		}
		return okResult(fmt.Sprintf(`{"language":%q}`, code))

	case "lookup_glossary":
		var in struct {
			Term   string `json:"term"`
			Target string `json:"target_language"`
		}
		if err := json.Unmarshal(call.Input, &in); err != nil {
			return errResult(err)
		}
		return okResult(fmt.Sprintf(`{"translation":%q}`, t.Glossary.Lookup(in.Term, in.Target)))

	case "placeholder_check":
		var in struct {
			Translation string `json:"translation"`
		}
		if err := json.Unmarshal(call.Input, &in); err != nil {
			return errResult(err)
		}
		missing := missingPlaceholders(in.Translation, len(t.Spans))
		if missing == nil {
			missing = []string{}
		}
		out, _ := json.Marshal(map[string]any{"missing": missing})
		return okResult(string(out))

	case "restore_spans":
		var in struct {
			Translation string `json:"translation"`
		}
		if err := json.Unmarshal(call.Input, &in); err != nil {
			return errResult(err)
		}
		restored, err := Restore(in.Translation, t.Spans, true)
		if err != nil {
			return errResult(err)
		}
		return okResult(fmt.Sprintf(`{"translation":%q}`, restored))

	default:
		return errResult(fmt.Errorf("unknown tool %q", call.Name))
	}
}

func missingPlaceholders(text string, n int) []string {
	var missing []string
	for i := 0; i < n; i++ {
		tok := fmt.Sprintf(placeholderFmt, i)
		if !regexp.MustCompile(regexp.QuoteMeta(tok)).MatchString(text) {
			missing = append(missing, tok)
		}
	}
	return missing
}

func str(desc string) map[string]any { return map[string]any{"type": "string", "description": desc} }
func obj(props map[string]any, required ...string) map[string]any {
	return map[string]any{"type": "object", "properties": props, "required": required}
}
func okResult(s string) ToolResult { return ToolResult{Output: s} }
func errResult(err error) ToolResult {
	return ToolResult{Output: fmt.Sprintf(`{"error":%q}`, err.Error()), IsError: true}
}

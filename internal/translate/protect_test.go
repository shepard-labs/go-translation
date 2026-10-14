package translate

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
)

func TestProtectFencedCode(t *testing.T) {
	in := "Hello ```go\nfmt.Println()\n``` world"
	p := Protect(in)
	if p.Hits != 1 {
		t.Fatalf("hits = %d, want 1", p.Hits)
	}
	if strings.Contains(p.Text, "Println") {
		t.Fatalf("code leaked into protected text: %q", p.Text)
	}
	if !strings.Contains(p.Text, "⟦P000⟧") {
		t.Fatalf("missing placeholder: %q", p.Text)
	}
}

func TestProtectMultipleSpans(t *testing.T) {
	in := "Visit https://example.com or `code` and {{name}}"
	p := Protect(in)
	if p.Hits != 3 {
		t.Fatalf("hits = %d, want 3: %q spans=%v", p.Hits, p.Text, p.Spans)
	}
}

func TestProtectFrontmatter(t *testing.T) {
	in := "---\ntitle: hi\n---\nBody text"
	p := Protect(in)
	if p.Hits != 1 || !strings.HasPrefix(p.Spans[0], "---") {
		t.Fatalf("frontmatter not protected: %v", p.Spans)
	}
}

func TestRestoreRoundTrip(t *testing.T) {
	in := "Hello `x` and https://a.b"
	p := Protect(in)
	// Simulate a "translation" that moves tokens.
	translated := strings.ReplaceAll(p.Text, "Hello", "Bonjour")
	out, err := Restore(translated, p.Spans, true)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "`x`") || !strings.Contains(out, "https://a.b") || !strings.Contains(out, "Bonjour") {
		t.Fatalf("bad restore: %q", out)
	}
}

func TestRestoreMissingStrict(t *testing.T) {
	p := Protect("go to `x`")
	// "Translation" drops the token.
	_, err := Restore("aller à", p.Spans, true)
	if err == nil {
		t.Fatal("expected strict restore to fail")
	}
	// Non-strict appends the missing span.
	out, err := Restore("aller à", p.Spans, false)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "`x`") {
		t.Fatalf("expected appended span, got %q", out)
	}
}

func TestDetectLanguage(t *testing.T) {
	cases := map[string]string{
		"Hello, how are you?":                                 "en",
		"Bonjour, comment allez-vous? le chat et la maison":   "fr",
		"Hola, ¿cómo estás? el perro y la casa":               "es",
		"Guten Tag, wie geht es Ihnen? der Hund und das Haus": "de",
		"こんにちは、お元気ですか":                                        "ja",
		"안녕하세요":                                               "ko",
		"Привет, как дела":                                    "ru",
	}
	for text, want := range cases {
		if got := DetectLanguage(text); got != want {
			t.Errorf("DetectLanguage(%q) = %q, want %q", text, got, want)
		}
	}
}

func TestDetectLanguageEmpty(t *testing.T) {
	if got := DetectLanguage("   "); got != "" {
		t.Fatalf("want empty, got %q", got)
	}
}

func TestGlossaryLookupAndScan(t *testing.T) {
	g := Glossary{Terms: map[string]map[string]string{
		"account": {"es": "cuenta", "fr": "compte"},
		"save":    {"es": "guardar"},
	}}
	if got := g.Lookup("Account", "es"); got != "cuenta" {
		t.Fatalf("lookup = %q", got)
	}
	if got := g.Lookup("account", "de"); got != "" {
		t.Fatalf("missing target should be empty, got %q", got)
	}
	found := g.Scan("Save your Account settings")
	if len(found) != 2 {
		t.Fatalf("scan = %v", found)
	}
}

func TestToolsetDispatch(t *testing.T) {
	ts := Toolset{
		Glossary: Glossary{Terms: map[string]map[string]string{"hello": {"es": "hola"}}},
		Spans:    []string{"`code`"},
	}
	ctx := context.Background()

	r := ts.Dispatch(ctx, ToolCall{Name: "detect_language", Input: mustJSON(t, map[string]string{"text": "the cat and the dog are"})})
	if r.IsError || !strings.Contains(r.Output, `"en"`) {
		t.Fatalf("detect_language: %+v", r)
	}

	r = ts.Dispatch(ctx, ToolCall{Name: "lookup_glossary", Input: mustJSON(t, map[string]string{"term": "hello", "target_language": "es"})})
	if !strings.Contains(r.Output, "hola") {
		t.Fatalf("lookup_glossary: %s", r.Output)
	}

	r = ts.Dispatch(ctx, ToolCall{Name: "placeholder_check", Input: mustJSON(t, map[string]string{"translation": "hola ⟦P000⟧"})})
	if !strings.Contains(r.Output, `"missing":[]`) {
		t.Fatalf("placeholder_check: %s", r.Output)
	}
	r = ts.Dispatch(ctx, ToolCall{Name: "placeholder_check", Input: mustJSON(t, map[string]string{"translation": "hola"})})
	if !strings.Contains(r.Output, "⟦P000⟧") {
		t.Fatalf("expected missing placeholder, got %s", r.Output)
	}

	r = ts.Dispatch(ctx, ToolCall{Name: "restore_spans", Input: mustJSON(t, map[string]string{"translation": "hola ⟦P000⟧ mundo"})})
	if !strings.Contains(r.Output, "`code`") {
		t.Fatalf("restore_spans: %s", r.Output)
	}

	r = ts.Dispatch(ctx, ToolCall{Name: "nope", Input: mustJSON(t, map[string]string{})})
	if !r.IsError {
		t.Fatal("expected unknown tool error")
	}
}

// fakeSession simulates a model that calls restore_spans once, then returns text.
type fakeSession struct {
	turns int
}

func (f *fakeSession) Generate(ctx context.Context) (AssistantTurn, error) {
	f.turns++
	if f.turns == 1 {
		input, _ := json.Marshal(map[string]string{"translation": "Bonjour ⟦P000⟧"})
		return AssistantTurn{ToolCalls: []ToolCallRequest{
			{ID: "t1", Name: "restore_spans", Input: input},
		}}, nil
	}
	return AssistantTurn{Text: "Bonjour `x`"}, nil
}

func (f *fakeSession) AddToolResults(results []ToolResult) {}

func mustJSON(t *testing.T, v any) []byte {
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func TestRunAgentLoopUsesRestore(t *testing.T) {
	sess := &fakeSession{}
	res, err := RunAgentLoop(context.Background(), sess, AgentRequest{
		Toolset:    Toolset{Spans: []string{"`x`"}},
		Spans:      []string{"`x`"},
		RestoreEnd: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !res.Restored || res.Text != "Bonjour `x`" {
		t.Fatalf("unexpected: %+v", res)
	}
}

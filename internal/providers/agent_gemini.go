package providers

import (
	"context"
	"encoding/json"

	"github.com/shepard-labs/go-translation/internal/translate"

	"github.com/shepard-labs/go-ai-sdk/google"
)

// geminiSession adapts one Gemini conversation to translate.ModelSession.
type geminiSession struct {
	model    google.LanguageModel
	base     google.GenerateOptions
	messages []google.Message
	tools    []google.Tool
}

func newGeminiSession(model google.LanguageModel, system, user string, defs []translate.ToolDef, maxTokens int, temperature float64) *geminiSession {
	s := &geminiSession{
		model: model,
		base: google.GenerateOptions{
			MaxOutputTokens: &maxTokens,
			Temperature:     &temperature,
		},
		messages: []google.Message{
			google.SystemMessage{Content: system},
			google.UserMessage{Content: []google.UserContent{google.TextContent{Text: user}}},
		},
	}
	tools := make([]google.Tool, len(defs))
	for i, d := range defs {
		tools[i] = google.Tool{Type: "function", Name: d.Name, Description: d.Description, InputSchema: d.InputSchema}
	}
	s.tools = tools
	return s
}

func (s *geminiSession) Generate(ctx context.Context) (translate.AssistantTurn, error) {
	opts := s.base
	opts.Messages = s.messages
	opts.Tools = s.tools
	res, err := s.model.DoGenerate(ctx, opts)
	if err != nil {
		return translate.AssistantTurn{}, err
	}
	var turn translate.AssistantTurn
	assistant := make([]google.AssistantContent, 0, len(res.Content))
	for _, c := range res.Content {
		if ac, ok := c.(google.AssistantContent); ok {
			assistant = append(assistant, ac)
		}
		switch v := c.(type) {
		case google.TextContent:
			turn.Text += v.Text
		case google.ToolCallContent:
			turn.ToolCalls = append(turn.ToolCalls, translate.ToolCallRequest{ID: v.ToolCallID, Name: v.ToolName, Input: v.Input})
		}
	}
	s.messages = append(s.messages, google.AssistantMessage{Content: assistant})
	return turn, nil
}

func (s *geminiSession) AddToolResults(results []translate.ToolResult) {
	var pending []google.ToolCallContent
	for i := len(s.messages) - 1; i >= 0; i-- {
		if am, ok := s.messages[i].(google.AssistantMessage); ok {
			for _, c := range am.Content {
				if tc, ok := c.(google.ToolCallContent); ok {
					pending = append(pending, tc)
				}
			}
			break
		}
	}
	parts := make([]google.ToolContent, len(results))
	for i, r := range results {
		id := ""
		if i < len(pending) {
			id = pending[i].ToolCallID
		}
		var out google.ToolResultOutput
		var v any
		if err := json.Unmarshal([]byte(r.Output), &v); err == nil {
			out = google.ToolResultOutput{Type: "json", Value: v}
		} else {
			out = google.ToolResultOutput{Type: "text", Value: r.Output}
		}
		parts[i] = google.ToolResultContent{ToolCallID: id, Output: out}
	}
	s.messages = append(s.messages, google.ToolMessage{Content: parts})
}

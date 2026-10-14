package providers

import (
	"context"
	"encoding/json"

	"github.com/shepard-labs/go-translation/internal/translate"

	"github.com/shepard-labs/go-ai-sdk/openrouter"
)

// openrouterSession adapts one OpenRouter conversation to translate.ModelSession.
type openrouterSession struct {
	model    openrouter.LanguageModel
	base     openrouter.GenerateOptions
	messages []openrouter.Message
	tools    []openrouter.Tool
}

func newOpenRouterSession(model openrouter.LanguageModel, system, user string, defs []translate.ToolDef, maxTokens int, temperature float64) *openrouterSession {
	s := &openrouterSession{
		model: model,
		base: openrouter.GenerateOptions{
			MaxTokens:   &maxTokens,
			Temperature: &temperature,
		},
		messages: []openrouter.Message{
			openrouter.SystemMessage{Content: system},
			openrouter.UserMessage{Content: []openrouter.UserContent{openrouter.TextContent{Text: user}}},
		},
	}
	tools := make([]openrouter.Tool, len(defs))
	for i, d := range defs {
		tools[i] = openrouter.Tool{Name: d.Name, Description: d.Description, InputSchema: d.InputSchema}
	}
	s.tools = tools
	return s
}

func (s *openrouterSession) Generate(ctx context.Context) (translate.AssistantTurn, error) {
	opts := s.base
	opts.Messages = s.messages
	opts.Tools = s.tools
	res, err := s.model.DoGenerate(ctx, opts)
	if err != nil {
		return translate.AssistantTurn{}, err
	}
	var turn translate.AssistantTurn
	assistant := make([]openrouter.AssistantContent, 0, len(res.Content))
	for _, c := range res.Content {
		if ac, ok := c.(openrouter.AssistantContent); ok {
			assistant = append(assistant, ac)
		}
		switch v := c.(type) {
		case openrouter.TextContent:
			turn.Text += v.Text
		case openrouter.ToolCallContent:
			raw, _ := json.Marshal(v.Input)
			turn.ToolCalls = append(turn.ToolCalls, translate.ToolCallRequest{ID: v.ToolCallID, Name: v.ToolName, Input: raw})
		}
	}
	s.messages = append(s.messages, openrouter.AssistantMessage{Content: assistant})
	return turn, nil
}

func (s *openrouterSession) AddToolResults(results []translate.ToolResult) {
	var pending []openrouter.ToolCallContent
	for i := len(s.messages) - 1; i >= 0; i-- {
		if am, ok := s.messages[i].(openrouter.AssistantMessage); ok {
			for _, c := range am.Content {
				if tc, ok := c.(openrouter.ToolCallContent); ok {
					pending = append(pending, tc)
				}
			}
			break
		}
	}
	parts := make([]openrouter.ToolContent, len(results))
	for i, r := range results {
		id, name := "", ""
		if i < len(pending) {
			id, name = pending[i].ToolCallID, pending[i].ToolName
		}
		parts[i] = openrouter.ToolResultContent{ToolCallID: id, ToolName: name, Output: r.Output, IsError: r.IsError}
	}
	s.messages = append(s.messages, openrouter.ToolMessage{Content: parts})
}

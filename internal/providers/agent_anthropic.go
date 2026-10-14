package providers

import (
	"context"
	"encoding/json"

	"github.com/shepard-labs/go-translation/internal/translate"

	"github.com/shepard-labs/go-ai-sdk/anthropic"
)

// anthropicSession adapts one Anthropic conversation to translate.ModelSession.
type anthropicSession struct {
	model    anthropic.LanguageModel
	base     anthropic.GenerateOptions
	messages []anthropic.Message
	tools    []anthropic.Tool
}

func newAnthropicSession(model anthropic.LanguageModel, system, user string, defs []translate.ToolDef, maxTokens int, temperature float64) *anthropicSession {
	s := &anthropicSession{
		model: model,
		base: anthropic.GenerateOptions{
			MaxTokens:   maxTokens,
			Temperature: &temperature,
		},
		messages: []anthropic.Message{
			anthropic.SystemMessage{Content: system},
			anthropic.UserMessage{Content: []anthropic.UserContent{anthropic.TextContent{Text: user}}},
		},
	}
	tools := make([]anthropic.Tool, len(defs))
	for i, d := range defs {
		tools[i] = anthropic.Tool{Name: d.Name, Description: d.Description, InputSchema: d.InputSchema}
	}
	s.tools = tools
	return s
}

func (s *anthropicSession) Generate(ctx context.Context) (translate.AssistantTurn, error) {
	opts := s.base
	opts.Messages = s.messages
	opts.Tools = s.tools
	res, err := s.model.DoGenerate(ctx, opts)
	if err != nil {
		return translate.AssistantTurn{}, err
	}
	var turn translate.AssistantTurn
	assistant := make([]anthropic.AssistantContent, 0, len(res.Content))
	for _, c := range res.Content {
		if ac, ok := c.(anthropic.AssistantContent); ok {
			assistant = append(assistant, ac)
		}
		switch v := c.(type) {
		case anthropic.TextContent:
			turn.Text += v.Text
		case anthropic.ToolCallContent:
			raw, _ := json.Marshal(v.Input)
			turn.ToolCalls = append(turn.ToolCalls, translate.ToolCallRequest{ID: v.ToolCallID, Name: v.ToolName, Input: raw})
		}
	}
	s.messages = append(s.messages, anthropic.AssistantMessage{Content: assistant})
	return turn, nil
}

func (s *anthropicSession) AddToolResults(results []translate.ToolResult) {
	// Results are paired positionally to the previous turn's tool calls; IDs
	// live in the pending calls tracked by the provider. We rebuild them from
	// the last assistant message's tool calls.
	var pending []anthropic.ToolCallContent
	for i := len(s.messages) - 1; i >= 0; i-- {
		if am, ok := s.messages[i].(anthropic.AssistantMessage); ok {
			for _, c := range am.Content {
				if tc, ok := c.(anthropic.ToolCallContent); ok {
					pending = append(pending, tc)
				}
			}
			break
		}
	}
	parts := make([]anthropic.UserContent, len(results))
	for i, r := range results {
		id := ""
		if i < len(pending) {
			id = pending[i].ToolCallID
		}
		parts[i] = anthropic.ToolResultContent{
			ToolCallID: id,
			Result:     []anthropic.ToolResultPart{anthropic.ToolResultText{Text: r.Output}},
			IsError:    r.IsError,
		}
	}
	s.messages = append(s.messages, anthropic.UserMessage{Content: parts})
}

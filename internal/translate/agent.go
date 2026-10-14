package translate

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

// AgentRequest bundles everything needed for one tool-driven translation.
type AgentRequest struct {
	System     string   // system prompt (built by prompt.SystemAgent)
	User       string   // user prompt wrapping the protected text
	Toolset    Toolset  // tool definitions + dispatch state
	MaxRounds  int      // safety cap on generate→tool cycles; 0 → default 6
	RestoreEnd bool     // strip remaining placeholders locally if model didn't call restore_spans
	Spans      []string // protected spans for RestoreEnd
}

// AgentResult is the normalized output of a tool-driven translation.
type AgentResult struct {
	Text      string
	ToolCalls int
	Restored  bool // true when restore_spans ran (model or fallback)
}

// ModelSession abstracts one provider's conversation state so the loop is
// provider-agnostic. Each provider adapter implements this once.
type ModelSession interface {
	// Generate sends the current conversation to the model and returns the
	// assistant content blocks (text and/or tool calls).
	Generate(ctx context.Context) (AssistantTurn, error)
	// AddToolResults appends tool outputs to the conversation for the next turn.
	AddToolResults(results []ToolResult)
}

// AssistantTurn is one model response: optional text plus zero or more tool calls.
type AssistantTurn struct {
	Text      string
	ToolCalls []ToolCallRequest
}

// ToolCallRequest is a model-emitted tool invocation.
type ToolCallRequest struct {
	ID    string
	Name  string
	Input json.RawMessage
}

// RunAgentLoop drives a generate→tool→generate cycle until the model stops
// calling tools. Tool defs are provider-mapped once by the caller's session.
func RunAgentLoop(ctx context.Context, sess ModelSession, req AgentRequest) (AgentResult, error) {
	max := req.MaxRounds
	if max <= 0 {
		max = 6
	}
	var res AgentResult
	var lastText string
	for round := 0; round < max; round++ {
		turn, err := sess.Generate(ctx)
		if err != nil {
			return res, err
		}
		if turn.Text != "" {
			lastText = turn.Text
		}
		if len(turn.ToolCalls) == 0 {
			break
		}
		res.ToolCalls += len(turn.ToolCalls)
		results := make([]ToolResult, len(turn.ToolCalls))
		for i, tc := range turn.ToolCalls {
			results[i] = req.Toolset.Dispatch(ctx, ToolCall{Name: tc.Name, Input: tc.Input})
			if tc.Name == "restore_spans" && !results[i].IsError {
				var out struct {
					Translation string `json:"translation"`
				}
				if err := json.Unmarshal([]byte(results[i].Output), &out); err == nil && out.Translation != "" {
					res.Text = out.Translation
					res.Restored = true
				}
			}
		}
		sess.AddToolResults(results)
		if round == max-1 {
			return res, errors.New("agent loop exhausted max rounds without final answer")
		}
	}
	// Prefer restore_spans output; otherwise use last assistant text.
	if res.Text == "" {
		res.Text = lastText
	}
	res.Text = strings.TrimSpace(res.Text)
	// Safety net: if placeholders remain (model skipped restore_spans), restore locally.
	if req.RestoreEnd && len(req.Spans) > 0 && placeholderPattern.MatchString(res.Text) {
		if restored, err := Restore(res.Text, req.Spans, false); err == nil {
			res.Text = restored
			res.Restored = true
		}
	}
	if res.Text == "" {
		return res, fmt.Errorf("model returned no text")
	}
	return res, nil
}

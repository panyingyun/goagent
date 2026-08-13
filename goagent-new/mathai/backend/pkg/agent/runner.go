package agent

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"mathai/pkg/llm"
)

// Runner 驱动 LLM ↔ 工具循环,直到产出最终文本或超过轮次上限。
type Runner struct {
	LLM       llm.Client
	Executor  *Executor
	Observer  Observer
	MaxRounds int
}

func (r *Runner) Run(ctx context.Context, messages []llm.Message) (string, []llm.Message, error) {
	if r.LLM == nil || r.Executor == nil {
		return "", messages, errors.New("runner requires LLM and Executor")
	}
	maxRounds := r.MaxRounds
	if maxRounds <= 0 {
		maxRounds = 8
	}
	history := append([]llm.Message(nil), messages...)
	for round := 0; round < maxRounds; round++ {
		started := time.Now()
		resp, err := r.LLM.Complete(ctx, llm.Request{Messages: history, Tools: r.Executor.Definitions()})
		if r.Observer != nil {
			event := Event{Kind: "llm", Status: "ok", Duration: time.Since(started), InputTokens: resp.Usage.PromptTokens, OutputTokens: resp.Usage.CompletionTokens}
			if err != nil {
				event.Status = "error"
				event.Error = err.Error()
			}
			r.Observer.Record(ctx, event)
		}
		if err != nil {
			return "", history, err
		}
		history = append(history, resp.Message)
		if len(resp.Message.ToolCalls) == 0 {
			if resp.Message.Content == "" {
				return "", history, errors.New("llm returned neither content nor tool calls")
			}
			return resp.Message.Content, history, nil
		}
		results := r.Executor.ExecuteBatch(ctx, resp.Message.ToolCalls)
		for _, result := range results {
			content := result.Out
			if result.Err != nil {
				if errors.Is(result.Err, context.Canceled) || errors.Is(result.Err, context.DeadlineExceeded) {
					return "", history, result.Err
				}
				encoded, _ := json.Marshal(map[string]string{"error": result.Err.Error()})
				content = string(encoded)
			}
			history = append(history, llm.Message{Role: llm.RoleTool, ToolCallID: result.Call.ID, Content: content})
		}
	}
	return "", history, fmt.Errorf("agent exceeded maximum rounds (%d)", maxRounds)
}


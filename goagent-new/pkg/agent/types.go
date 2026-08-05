package agent

import (
	"context"
	"encoding/json"
	"time"

	"goagent-new/pkg/llm"
)

type SideEffect string

const (
	SideEffectNone     SideEffect = "none"
	SideEffectRead     SideEffect = "read"
	SideEffectWrite    SideEffect = "write"
	SideEffectExternal SideEffect = "external"
)

type Tool struct {
	Name             string
	Description      string
	Parameters       map[string]any
	SideEffect       SideEffect
	Cacheable        bool
	Idempotent       bool
	ParallelSafe     bool
	RequiresApproval bool
	Timeout          time.Duration
	Run              func(context.Context, json.RawMessage) (string, error)
}

func (t Tool) Definition() llm.ToolDefinition {
	return llm.ToolDefinition{
		Type: "function",
		Function: llm.FunctionDefinition{
			Name:        t.Name,
			Description: t.Description,
			Parameters:  t.Parameters,
		},
	}
}

type ApprovalRequest struct {
	ToolName      string
	Arguments     string
	SideEffect    SideEffect
	IdempotencyKey string
}

type Approver interface {
	Approve(ctx context.Context, req ApprovalRequest) (bool, error)
}

type ApproverFunc func(context.Context, ApprovalRequest) (bool, error)

func (f ApproverFunc) Approve(ctx context.Context, req ApprovalRequest) (bool, error) {
	return f(ctx, req)
}

type Event struct {
	Kind           string
	Tool           string
	Status         string
	Duration       time.Duration
	InputTokens    int
	OutputTokens   int
	IdempotencyKey string
	Error          string
}

type Observer interface {
	Record(ctx context.Context, event Event)
}

type ObserverFunc func(context.Context, Event)

func (f ObserverFunc) Record(ctx context.Context, event Event) { f(ctx, event) }

type ToolResult struct {
	Call llm.ToolCall
	Out  string
	Err  error
}


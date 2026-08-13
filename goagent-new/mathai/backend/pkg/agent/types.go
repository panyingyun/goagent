package agent

import (
	"context"
	"encoding/json"
	"time"

	"mathai/pkg/llm"
)

// SideEffect 描述工具对外部世界的影响级别,用于审批与并行策略。
type SideEffect string

const (
	SideEffectNone     SideEffect = "none"
	SideEffectRead     SideEffect = "read"
	SideEffectWrite    SideEffect = "write"
	SideEffectExternal SideEffect = "external"
)

// Tool 是可注册到 Executor 的工具定义与实现。
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

// ApprovalRequest 是需要人工/策略审批时提交的工具调用信息。
type ApprovalRequest struct {
	ToolName       string
	Arguments      string
	SideEffect     SideEffect
	IdempotencyKey string
}

// Approver 决定是否允许执行一次有副作用的工具调用。
type Approver interface {
	Approve(ctx context.Context, req ApprovalRequest) (bool, error)
}

// ApproverFunc 将函数适配为 Approver。
type ApproverFunc func(context.Context, ApprovalRequest) (bool, error)

func (f ApproverFunc) Approve(ctx context.Context, req ApprovalRequest) (bool, error) {
	return f(ctx, req)
}

// Event 是 Agent 运行时的可观测事件(LLM 调用或工具执行)。
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

// Observer 接收运行时事件,用于 trace、指标或审计。
type Observer interface {
	Record(ctx context.Context, event Event)
}

// ObserverFunc 将函数适配为 Observer。
type ObserverFunc func(context.Context, Event)

func (f ObserverFunc) Record(ctx context.Context, event Event) { f(ctx, event) }

// ToolResult 是单次工具调用的执行结果。
type ToolResult struct {
	Call llm.ToolCall
	Out  string
	Err  error
}

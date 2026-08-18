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
// Name / Description / Parameters 会经 Definition() 发给 LLM；
// 其余字段是运行时的安全与调度元数据，模型看不到。
type Tool struct {
	// Name 工具唯一名，LLM 用此名字发起调用（如 "apply_graph_patch"）。
	Name string
	// Description 自然语言说明，帮助模型决定何时调用该工具。
	Description string
	// Parameters JSON Schema，约束调用参数的形状与必填字段。
	Parameters map[string]any
	// SideEffect 对外部世界的影响级别（none/read/write/external），
	// 用于审批与并行策略；write/external 注册时会强制审批、禁缓存与并行。
	SideEffect SideEffect
	// Cacheable 相同参数是否可直接返回缓存结果；写类工具通常为 false。
	Cacheable bool
	// Idempotent 同样调用执行两次结果是否一致（幂等）。
	Idempotent bool
	// ParallelSafe 是否可与其他工具并行执行；写画布等共享状态的工具应为 false。
	ParallelSafe bool
	// RequiresApproval 执行前是否必须经过 Approver；write/external 注册时会被强制设为 true。
	RequiresApproval bool
	// Timeout 单次执行的最长时间，超时由带 deadline 的 context 取消。
	Timeout time.Duration
	// Run 真正执行工具逻辑：入参为模型给出的 JSON 参数，返回结果字符串或错误。
	Run func(context.Context, json.RawMessage) (string, error)
}

// Definition 转为发给模型的工具定义。
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
	// Approve 决定是否允许执行该工具调用。
	Approve(ctx context.Context, req ApprovalRequest) (bool, error)
}

// ApproverFunc 将函数适配为 Approver。
type ApproverFunc func(context.Context, ApprovalRequest) (bool, error)

// Approve 将函数调用转发给底层 ApproverFunc。
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
	// Record 接收一条运行时事件。
	Record(ctx context.Context, event Event)
}

// ObserverFunc 将函数适配为 Observer。
type ObserverFunc func(context.Context, Event)

// Record 将函数调用转发给底层 ObserverFunc。
func (f ObserverFunc) Record(ctx context.Context, event Event) { f(ctx, event) }

// ToolResult 是单次工具调用的执行结果。
type ToolResult struct {
	Call llm.ToolCall
	Out  string
	Err  error
}

package llm

import "context"

// Role 表示对话消息角色,对应 Chat Completions 的 role 字段。
type Role string

const (
	RoleSystem    Role = "system"
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
	RoleTool      Role = "tool"
)

// FunctionCall 是模型发起的一次函数调用(名称 + JSON 参数字符串)。
type FunctionCall struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

// ToolCall 是 assistant 消息中的单条工具调用,含调用 ID 与函数体。
type ToolCall struct {
	ID       string       `json:"id"`
	Type     string       `json:"type"`
	Function FunctionCall `json:"function"`
}

// Message 是一轮对话中的单条消息(含可选的 tool_calls / tool_call_id)。
type Message struct {
	Role       Role       `json:"role"`
	Content    string     `json:"content,omitempty"`
	ToolCallID string     `json:"tool_call_id,omitempty"`
	ToolCalls  []ToolCall `json:"tool_calls,omitempty"`
}

// FunctionDefinition 描述可供模型调用的函数(名称、说明与 JSON Schema 参数)。
type FunctionDefinition struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Parameters  map[string]any `json:"parameters"`
}

// ToolDefinition 是发给模型的工具定义包装;Type 通常为 "function"。
type ToolDefinition struct {
	Type     string             `json:"type"`
	Function FunctionDefinition `json:"function"`
}

// ResponseFormat 请求供应商以指定格式返回(如 json_object)。
// DeepSeek / OpenAI-compatible 常用 Type: "json_object"；应用端仍须严格解码与业务校验。
type ResponseFormat struct {
	Type string `json:"type"`
}

// Request 是一次补全请求的入参(消息、可选工具与采样温度)。
type Request struct {
	Messages       []Message
	Tools          []ToolDefinition
	Temperature    *float64
	ResponseFormat *ResponseFormat
}

// Usage 记录本次调用的 token 用量。
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// Response 是一次补全的结果(消息、结束原因、用量与请求 ID)。
type Response struct {
	Message      Message
	FinishReason string
	Usage        Usage
	RequestID    string
}

// Client 抽象 LLM 补全能力,便于用假实现做单测。
type Client interface {
	// Complete 发起一次补全请求。
	Complete(ctx context.Context, req Request) (Response, error)
}

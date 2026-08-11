package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// DeepSeek 提供 OpenAI 兼容的 Chat Completions API,
// 所以这里用标准库 http 直接调用即可,不需要任何第三方 SDK。
//
// 截至 2026-08,DeepSeek 的模型名是 deepseek-v4-flash / deepseek-v4-pro
// (旧的 deepseek-chat / deepseek-reasoner 已于 2026-07-24 退役)。
const defaultBaseURL = "https://api.deepseek.com"

// Role 是对话消息的角色。
type Role string

const (
	RoleSystem    Role = "system"
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
	RoleTool      Role = "tool"
)

// ToolCall 是模型在 function calling 模式下请求调用工具时返回的结构。
type ToolCall struct {
	ID       string           `json:"id"`
	Type     string           `json:"type"`
	Function ToolCallFunction `json:"function"`
}

type ToolCallFunction struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"` // 参数是 JSON 字符串
}

// Message 是发给/来自模型的单条消息。
type Message struct {
	Role       Role       `json:"role"`
	Content    string     `json:"content,omitempty"`
	ToolCalls  []ToolCall `json:"tool_calls,omitempty"`
	ToolCallID string     `json:"tool_call_id,omitempty"`
	Name       string     `json:"name,omitempty"`
}

// ToolFunction 描述一个工具的元信息(名字、说明、参数 JSON Schema)。
type ToolFunction struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Parameters  map[string]any `json:"parameters"`
}

type ToolDef struct {
	Type     string       `json:"type"`
	Function ToolFunction `json:"function"`
}

type chatRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
	Tools    []ToolDef `json:"tools,omitempty"`
	Stream   bool      `json:"stream"`
}

type chatResponse struct {
	Choices []struct {
		Message Message `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

// LLMClient 封装对 DeepSeek Chat Completions 的调用。
type LLMClient struct {
	APIKey  string
	BaseURL string
	Model   string
	HTTP    *http.Client
}

func NewLLMClient() *LLMClient {
	return &LLMClient{
		APIKey:  os.Getenv("DEEPSEEK_API_KEY"),
		BaseURL: envOr("DEEPSEEK_BASE_URL", defaultBaseURL),
		Model:   envOr("DEEPSEEK_MODEL", "deepseek-v4-flash"),
		HTTP:    &http.Client{Timeout: 120 * time.Second},
	}
}

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

// Chat 发送一轮对话,返回模型回复。
// tools 非空时启用 function calling 模式。
func (c *LLMClient) Chat(ctx context.Context, messages []Message, tools []ToolDef) (Message, error) {
	body, err := json.Marshal(chatRequest{
		Model:    c.Model,
		Messages: messages,
		Tools:    tools,
	})
	if err != nil {
		return Message{}, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return Message{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.APIKey)

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return Message{}, fmt.Errorf("请求 DeepSeek 失败: %w", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return Message{}, err
	}
	if resp.StatusCode != http.StatusOK {
		return Message{}, fmt.Errorf("DeepSeek 返回 %d: %s", resp.StatusCode, string(data))
	}

	var out chatResponse
	if err := json.Unmarshal(data, &out); err != nil {
		return Message{}, fmt.Errorf("解析响应失败: %w", err)
	}
	if out.Error != nil {
		return Message{}, fmt.Errorf("DeepSeek 错误: %s", out.Error.Message)
	}
	if len(out.Choices) == 0 {
		return Message{}, fmt.Errorf("响应中没有 choices")
	}
	return out.Choices[0].Message, nil
}

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

// 带 function calling 的 LLM 客户端(与 ch05 相同)。
const defaultBaseURL = "https://api.deepseek.com"

type Message struct {
	Role       string     `json:"role"`
	Content    string     `json:"content,omitempty"`
	ToolCalls  []ToolCall `json:"tool_calls,omitempty"`
	ToolCallID string     `json:"tool_call_id,omitempty"`
	Name       string     `json:"name,omitempty"`
}

type ToolCall struct {
	ID       string `json:"id"`
	Type     string `json:"type"`
	Function struct {
		Name      string `json:"name"`
		Arguments string `json:"arguments"`
	} `json:"function"`
}

type ToolFunction struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Parameters  map[string]any `json:"parameters"`
}

type ToolDef struct {
	Type     string       `json:"type"`
	Function ToolFunction `json:"function"`
}

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

func (c *LLMClient) Chat(ctx context.Context, messages []Message, tools []ToolDef) (Message, error) {
	body, _ := json.Marshal(map[string]any{
		"model":    c.Model,
		"messages": messages,
		"tools":    tools,
		"stream":   false,
	})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return Message{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.APIKey)
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return Message{}, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return Message{}, fmt.Errorf("API 返回 %d: %s", resp.StatusCode, string(data))
	}
	var out struct {
		Choices []struct {
			Message Message `json:"message"`
		} `json:"choices"`
		Error *struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(data, &out); err != nil {
		return Message{}, err
	}
	if out.Error != nil {
		return Message{}, fmt.Errorf("API 错误: %s", out.Error.Message)
	}
	if len(out.Choices) == 0 {
		return Message{}, fmt.Errorf("没有 choices")
	}
	return out.Choices[0].Message, nil
}

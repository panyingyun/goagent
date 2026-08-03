package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// 本课时的核心:流式(streaming)调用。
// 和普通调用唯一的区别是请求里加 "stream": true,
// 响应变成 SSE(Server-Sent Events)格式,一行一行地"吐"内容。
const defaultBaseURL = "https://api.deepseek.com"

type Role string

const (
	RoleSystem    Role = "system"
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
)

type Message struct {
	Role    Role   `json:"role"`
	Content string `json:"content"`
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

// StreamChat 发送流式请求,每收到一段增量内容就回调 onDelta。
func (c *LLMClient) StreamChat(ctx context.Context, messages []Message, onDelta func(string)) error {
	body, err := json.Marshal(map[string]any{
		"model":    c.Model,
		"messages": messages,
		"stream":   true, // 关键开关
	})
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.APIKey)

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return fmt.Errorf("请求 DeepSeek 失败: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		data, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("DeepSeek 返回 %d: %s", resp.StatusCode, string(data))
	}

	// SSE 格式:每行以 "data: " 开头,内容是一段 JSON;空行分隔消息;
	// 结束标志是 data: [DONE]。
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if !strings.HasPrefix(line, "data:") {
			continue
		}
		data := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
		if data == "[DONE]" {
			return nil
		}
		var chunk struct {
			Choices []struct {
				Delta struct {
					Content string `json:"content"`
				} `json:"delta"`
			} `json:"choices"`
		}
		if err := json.Unmarshal([]byte(data), &chunk); err != nil {
			return fmt.Errorf("解析流式分片失败: %w", err)
		}
		if len(chunk.Choices) > 0 {
			onDelta(chunk.Choices[0].Delta.Content)
		}
	}
	return scanner.Err()
}

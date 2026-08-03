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

// DeepSeekClient 封装 DeepSeek Chat Completions 调用(与 ch01 相同的模式,
// 但抽象成独立 service,不散落在 handler 里——设计稿 §10)。
type DeepSeekClient struct {
	APIKey  string
	BaseURL string
	Model   string
	HTTP    *http.Client
}

func NewDeepSeekClient() *DeepSeekClient {
	return &DeepSeekClient{
		APIKey:  os.Getenv("DEEPSEEK_API_KEY"),
		BaseURL: envOr("DEEPSEEK_BASE_URL", "https://api.deepseek.com"),
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

type llmMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// Chat 发送一轮对话,返回模型文本。temperature 用 0.1 保证输出稳定(设计稿 §10)。
func (c *DeepSeekClient) Chat(ctx context.Context, messages []llmMessage) (string, error) {
	body, err := json.Marshal(map[string]any{
		"model":       c.Model,
		"messages":    messages,
		"temperature": 0.1,
		"stream":      false,
	})
	if err != nil {
		return "", err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.APIKey)
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return "", fmt.Errorf("请求 DeepSeek 失败: %w", err)
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("DeepSeek 返回 %d: %s", resp.StatusCode, string(data))
	}
	var out struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Error *struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(data, &out); err != nil {
		return "", fmt.Errorf("解析响应失败: %w", err)
	}
	if out.Error != nil {
		return "", fmt.Errorf("DeepSeek 错误: %s", out.Error.Message)
	}
	if len(out.Choices) == 0 {
		return "", fmt.Errorf("响应中没有 choices")
	}
	return out.Choices[0].Message.Content, nil
}

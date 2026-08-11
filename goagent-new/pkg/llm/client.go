package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const maxResponseBytes = 4 << 20

// HTTPClient 通过 OpenAI-compatible HTTP API 发起 Chat Completions 请求。
type HTTPClient struct {
	APIKey  string
	BaseURL string
	Model   string
	HTTP    *http.Client
}

func NewHTTPClient(apiKey, baseURL, model string) *HTTPClient {
	return &HTTPClient{
		APIKey:  apiKey,
		BaseURL: strings.TrimRight(baseURL, "/"),
		Model:   model,
		HTTP:    &http.Client{Timeout: 90 * time.Second},
	}
}

// HTTPError 表示非 2xx 的 LLM HTTP 响应,便于按状态码判断是否可重试。
type HTTPError struct {
	StatusCode int
	Body       string
	RequestID  string
}

func (e *HTTPError) Error() string {
	return fmt.Sprintf("llm http status %d (request_id=%q): %s", e.StatusCode, e.RequestID, e.Body)
}

func (e *HTTPError) Retryable() bool {
	return e.StatusCode == http.StatusTooManyRequests || e.StatusCode >= 500
}

func (c *HTTPClient) Complete(ctx context.Context, in Request) (Response, error) {
	if c.APIKey == "" || c.BaseURL == "" || c.Model == "" {
		return Response{}, errors.New("缺少 DEEPSEEK_API_KEY(或 LLM_API_KEY)")
	}
	payload := struct {
		Model          string           `json:"model"`
		Messages       []Message        `json:"messages"`
		Tools          []ToolDefinition `json:"tools,omitempty"`
		Temperature    *float64         `json:"temperature,omitempty"`
		ResponseFormat *ResponseFormat  `json:"response_format,omitempty"`
		Stream         bool             `json:"stream"`
	}{c.Model, in.Messages, in.Tools, in.Temperature, in.ResponseFormat, false}
	body, err := json.Marshal(payload)
	if err != nil {
		return Response{}, fmt.Errorf("encode llm request: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return Response{}, fmt.Errorf("build llm request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.APIKey)
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return Response{}, fmt.Errorf("send llm request: %w", err)
	}
	defer resp.Body.Close()
	requestID := responseRequestID(resp.Header)
	limited := io.LimitReader(resp.Body, maxResponseBytes+1)
	data, err := io.ReadAll(limited)
	if err != nil {
		return Response{}, fmt.Errorf("read llm response: %w", err)
	}
	if len(data) > maxResponseBytes {
		return Response{}, errors.New("llm response exceeds 4 MiB limit")
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return Response{}, &HTTPError{StatusCode: resp.StatusCode, Body: truncate(string(data), 1000), RequestID: requestID}
	}
	var decoded struct {
		Choices []struct {
			Message      Message `json:"message"`
			FinishReason string  `json:"finish_reason"`
		} `json:"choices"`
		Usage Usage `json:"usage"`
	}
	if err := json.Unmarshal(data, &decoded); err != nil {
		return Response{}, fmt.Errorf("decode llm response: %w", err)
	}
	if len(decoded.Choices) == 0 {
		return Response{}, errors.New("llm response contains no choices")
	}
	return Response{
		Message:      decoded.Choices[0].Message,
		FinishReason: decoded.Choices[0].FinishReason,
		Usage:        decoded.Usage,
		RequestID:    requestID,
	}, nil
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}

// responseRequestID 从响应头提取供应商请求追踪 ID。
// DeepSeek 使用 X-Ds-Trace-Id；其它 OpenAI-compatible 供应商常见 X-Request-Id。
func responseRequestID(h http.Header) string {
	for _, key := range []string{"X-Ds-Trace-Id", "X-Request-Id"} {
		if id := h.Get(key); id != "" {
			return id
		}
	}
	return ""
}

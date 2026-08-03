package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"time"
)

// Embedder 抽象"文本 -> 向量"的能力。这样换提供商只改实现,不动业务代码。
type Embedder interface {
	Embed(ctx context.Context, texts []string) ([][]float64, error)
}

// OpenAIEmbedder 用 OpenAI text-embedding-3-small 做向量化(路线B)。
// DeepSeek 官方 API 目前没有 embedding 接口,所以向量化交给 OpenAI,
// "想"仍然由 DeepSeek 负责。
const defaultEmbedBaseURL = "https://api.openai.com/v1"
const defaultEmbedModel = "text-embedding-3-small"

type OpenAIEmbedder struct {
	APIKey  string
	Model   string
	BaseURL string
	HTTP    *http.Client
}

func NewOpenAIEmbedder() *OpenAIEmbedder {
	return &OpenAIEmbedder{
		APIKey:  os.Getenv("OPENAI_API_KEY"),
		Model:   envOr("OPENAI_EMBED_MODEL", defaultEmbedModel),
		BaseURL: envOr("OPENAI_BASE_URL", defaultEmbedBaseURL),
		HTTP:    &http.Client{Timeout: 60 * time.Second},
	}
}

func (e *OpenAIEmbedder) Embed(ctx context.Context, texts []string) ([][]float64, error) {
	if e.APIKey == "" {
		return nil, fmt.Errorf("缺少 OPENAI_API_KEY 环境变量(本课时用 OpenAI text-embedding-3-small 做向量化)")
	}
	body, err := json.Marshal(map[string]any{
		"model": e.Model,
		"input": texts,
	})
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, e.BaseURL+"/embeddings", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+e.APIKey)

	resp, err := e.HTTP.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求 OpenAI embeddings 失败: %w", err)
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("OpenAI 返回 %d: %s", resp.StatusCode, string(data))
	}
	var out struct {
		Data []struct {
			Index     int       `json:"index"`
			Embedding []float64 `json:"embedding"`
		} `json:"data"`
		Error *struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(data, &out); err != nil {
		return nil, fmt.Errorf("解析 embeddings 响应失败: %w", err)
	}
	if out.Error != nil {
		return nil, fmt.Errorf("OpenAI 错误: %s", out.Error.Message)
	}
	// 按 index 排好序,保证和输入顺序一致
	sort.Slice(out.Data, func(i, j int) bool { return out.Data[i].Index < out.Data[j].Index })
	result := make([][]float64, 0, len(out.Data))
	for _, d := range out.Data {
		result = append(result, d.Embedding)
	}
	return result, nil
}

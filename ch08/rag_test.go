package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// fakeEmbedder 用固定向量模拟 embedding,让测试离线可跑。
type fakeEmbedder struct {
	vectors map[string][]float64
}

func (f fakeEmbedder) Embed(_ context.Context, texts []string) ([][]float64, error) {
	out := make([][]float64, len(texts))
	for i, t := range texts {
		out[i] = f.vectors[t]
	}
	return out, nil
}

func TestChunkText(t *testing.T) {
	text := strings.Repeat("第一段内容", 100) + "\n" + strings.Repeat("第二段内容", 100) + "\n\n第三段"
	chunks := chunkText(text, 300)
	if len(chunks) < 2 {
		t.Fatalf("至少应有 2 个片段,实际 %d", len(chunks))
	}
	for i, c := range chunks {
		if len([]rune(c)) > 330 {
			t.Errorf("片段 %d 超过 330 字符: %d", i, len([]rune(c)))
		}
	}
}

func TestCosineSimilarity(t *testing.T) {
	a := []float64{1, 0}
	b := []float64{0, 1}
	if s := cosineSimilarity(a, a); s < 0.99 {
		t.Errorf("相同向量相似度应接近 1: %v", s)
	}
	if s := cosineSimilarity(a, b); s > 0.01 {
		t.Errorf("正交向量相似度应接近 0: %v", s)
	}
}

func TestSearchTopK(t *testing.T) {
	emb := fakeEmbedder{vectors: map[string][]float64{
		"苹果":  {1, 0, 0},
		"香蕉":  {0, 1, 0},
		"苹果汁": {0.9, 0.1, 0},
	}}
	idx, err := BuildIndex(context.Background(), emb, []string{"苹果", "香蕉", "苹果汁"})
	if err != nil {
		t.Fatal(err)
	}
	hits, err := idx.Search(context.Background(), emb, "苹果", 2)
	if err != nil {
		t.Fatal(err)
	}
	if len(hits) != 2 || hits[0] != "苹果" {
		t.Fatalf("最相关片段应为苹果,实际: %v", hits)
	}
}

// OpenAIEmbedder 的请求/响应解析,用本地 mock 验证(不真正调 OpenAI)。
func TestOpenAIEmbedderWithMock(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/embeddings" {
			t.Errorf("意外路径: %s", r.URL.Path)
		}
		var req struct {
			Model string   `json:"model"`
			Input []string `json:"input"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("解析请求失败: %v", err)
		}
		if req.Model != "text-embedding-3-small" {
			t.Errorf("模型应为 text-embedding-3-small: %q", req.Model)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"data": []any{
				map[string]any{"index": 1, "embedding": []float64{0, 1}},
				map[string]any{"index": 0, "embedding": []float64{1, 0}},
			},
		})
	}))
	defer srv.Close()

	e := &OpenAIEmbedder{
		APIKey:  "test",
		Model:   "text-embedding-3-small",
		BaseURL: srv.URL,
		HTTP:    srv.Client(),
	}
	vecs, err := e.Embed(context.Background(), []string{"a", "b"})
	if err != nil {
		t.Fatal(err)
	}
	if len(vecs) != 2 || vecs[0][0] != 1 || vecs[1][1] != 1 {
		t.Fatalf("应按 index 排序: %v", vecs)
	}
}

package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestTokenEstimate(t *testing.T) {
	if tokenEstimate("") != 0 {
		t.Error("空字符串应为 0")
	}
	en := tokenEstimate(strings.Repeat("a", 100))
	zh := tokenEstimate(strings.Repeat("你", 100))
	if zh <= en {
		t.Errorf("中文估算应大于英文估算: en=%d zh=%d", en, zh)
	}
}

func mockLLM(t *testing.T, reply string) (*httptest.Server, *LLMClient) {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"choices": []any{map[string]any{"message": map[string]any{"role": "assistant", "content": reply}}},
		})
	}))
	client := NewLLMClient()
	client.APIKey = "test"
	client.BaseURL = srv.URL
	return srv, client
}

// 历史超过预算时,旧消息被压成摘要,并且摘要会出现在后续请求里。
func TestCompressIfNeeded(t *testing.T) {
	srv, llm := mockLLM(t, "用户喜欢喝咖啡,正在写周报")
	defer srv.Close()

	h := &History{System: "test"}
	for i := 0; i < 40; i++ {
		h.Add(Message{Role: RoleUser, Content: strings.Repeat("对话内容", 100)})
		h.Add(Message{Role: RoleAssistant, Content: strings.Repeat("回复内容", 100)})
	}
	if err := h.CompressIfNeeded(context.Background(), llm); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(h.Summary, "咖啡") {
		t.Fatalf("摘要应包含 mock 返回内容,实际: %q", h.Summary)
	}
	if len(h.Messages) >= 80 {
		t.Fatalf("历史应被压缩,剩余 %d 条", len(h.Messages))
	}
	msgs := h.SystemMessages()
	found := false
	for _, m := range msgs {
		if strings.Contains(m.Content, "早期对话摘要") {
			found = true
		}
	}
	if !found {
		t.Error("后续请求应包含摘要消息")
	}
}

func TestPreferencesRoundtrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "memory.json")
	p, err := LoadPreferences(path)
	if err != nil {
		t.Fatal(err)
	}
	p.Remember("称呼", "小明")
	if err := p.Save(); err != nil {
		t.Fatal(err)
	}
	p2, err := LoadPreferences(path)
	if err != nil {
		t.Fatal(err)
	}
	if p2.data["称呼"] != "小明" {
		t.Fatalf("持久化失败: %v", p2.data)
	}
	_ = os.Remove(path)
}

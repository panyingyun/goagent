package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// mockStreamServer 模拟 DeepSeek 的流式响应:
// 每个 data: 分片带一个汉字,最后以 [DONE] 结束。
func mockStreamServer(t *testing.T, words []string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		fl, _ := w.(http.Flusher)
		for _, word := range words {
			chunk := map[string]any{
				"choices": []any{map[string]any{
					"delta": map[string]any{"content": word},
				}},
			}
			data, _ := json.Marshal(chunk)
			fmt.Fprintf(w, "data: %s\n\n", data)
			if fl != nil {
				fl.Flush()
			}
		}
		fmt.Fprint(w, "data: [DONE]\n\n")
		if fl != nil {
			fl.Flush()
		}
	}))
}

// 流式内容应能完整拼接,并且 onDelta 按分片逐个触发。
func TestStreamChat(t *testing.T) {
	words := []string{"你", "好", ",", "世", "界"}
	srv := mockStreamServer(t, words)
	defer srv.Close()

	client := NewLLMClient()
	client.APIKey = "test"
	client.BaseURL = srv.URL

	var got []string
	err := client.StreamChat(context.Background(), []Message{
		{Role: RoleUser, Content: "你好"},
	}, func(delta string) {
		got = append(got, delta)
	})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Join(got, "") != "你好,世界" {
		t.Fatalf("拼接结果不符: %q", strings.Join(got, ""))
	}
	if len(got) != len(words) {
		t.Fatalf("onDelta 触发次数应为 %d,实际 %d", len(words), len(got))
	}
}

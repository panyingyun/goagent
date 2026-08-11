// 课时25:LangServe 风格服务
//
// LangServe 是 LangChain 的部署框架:把 Chain/Agent 发布成 REST API,
// 自带 /invoke、/stream、/playground、/docs。本课用 Go 复刻这套接口。
//
// 运行(已配置 DEEPSEEK_API_KEY):
//
//	go run .
//
// 浏览器打开 http://localhost:8080/playground
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

const defaultBaseURL = "https://api.deepseek.com"

func main() {
	apiKey := os.Getenv("DEEPSEEK_API_KEY")
	if apiKey == "" {
		fmt.Fprintln(os.Stderr, "缺少 DEEPSEEK_API_KEY")
		os.Exit(1)
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, map[string]any{"status": "ok"})
	})
	mux.HandleFunc("/invoke", func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Input string `json:"input"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Input == "" {
			http.Error(w, "input 必填", http.StatusBadRequest)
			return
		}
		answer, err := invoke(apiKey, req.Input)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		writeJSON(w, map[string]any{"output": answer})
	})
	mux.HandleFunc("/chat/stream", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query().Get("q")
		if q == "" {
			http.Error(w, "缺少 q", http.StatusBadRequest)
			return
		}
		fl, _ := w.(http.Flusher)
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		fl.Flush()
		err := stream(apiKey, q, func(delta string) {
			fmt.Fprintf(w, "data: %s\n\n", delta)
			fl.Flush()
		})
		if err != nil {
			fmt.Fprintf(w, "data: [ERROR] %v\n\n", err)
		}
	})
	mux.HandleFunc("/playground", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "web/playground.html")
	})
	mux.HandleFunc("/docs", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, map[string]any{
			"endpoints": []string{"/invoke", "/chat/stream", "/playground", "/docs", "/health"},
			"examples": map[string]any{
				"invoke": `{"input": "你好"}`,
			},
		})
	})
	fmt.Println("LangServe 风格服务: http://localhost:8080/playground")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}

func invoke(apiKey, input string) (string, error) {
	body, _ := json.Marshal(map[string]any{
		"model":    "deepseek-v4-flash",
		"messages": []any{map[string]any{"role": "user", "content": input}},
		"stream":   false,
	})
	req, _ := http.NewRequestWithContext(context.Background(), http.MethodPost,
		defaultBaseURL+"/chat/completions", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)
	resp, err := (&http.Client{Timeout: 120 * time.Second}).Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API 返回 %d: %s", resp.StatusCode, string(data))
	}
	var out struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(data, &out); err != nil {
		return "", err
	}
	if len(out.Choices) == 0 {
		return "", fmt.Errorf("没有 choices")
	}
	return out.Choices[0].Message.Content, nil
}

func stream(apiKey, q string, onDelta func(string)) error {
	body, _ := json.Marshal(map[string]any{
		"model":    "deepseek-v4-flash",
		"messages": []any{map[string]any{"role": "user", "content": q}},
		"stream":   true,
	})
	req, _ := http.NewRequestWithContext(context.Background(), http.MethodPost,
		defaultBaseURL+"/chat/completions", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)
	resp, err := (&http.Client{Timeout: 120 * time.Second}).Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
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
			return err
		}
		if len(chunk.Choices) > 0 {
			onDelta(chunk.Choices[0].Delta.Content)
		}
	}
	return scanner.Err()
}

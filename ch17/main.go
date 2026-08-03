// 课时17:Agent Web 界面
//
// 给 Agent 套一个浏览器聊天界面:
// GET /            -> 网页
// POST /chat       -> JSON 问答
// GET /chat/stream -> SSE 流式问答
//
// 运行(已配置 DEEPSEEK_API_KEY):
//
//	go run .
//
// 然后浏览器打开 http://localhost:8080
package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

type server struct {
	llm ChatLLM
}

func main() {
	llm := NewLLMClient()
	if llm.APIKey == "" {
		fmt.Fprintln(os.Stderr, "缺少 DEEPSEEK_API_KEY")
		os.Exit(1)
	}
	s := &server{llm: llm}
	http.HandleFunc("/", s.handleIndex)
	http.HandleFunc("/chat", s.handleChat)
	http.HandleFunc("/chat/stream", s.handleStream)
	fmt.Println("打开 http://localhost:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func (s *server) handleIndex(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "web/index.html")
}

func (s *server) handleChat(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		Question string `json:"question"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Question == "" {
		http.Error(w, "question 必填", http.StatusBadRequest)
		return
	}
	reply, err := s.llm.Chat(r.Context(), []Message{
		{Role: "system", Content: "你是乐于助人的中文助手。"},
		{Role: "user", Content: req.Question},
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"answer": reply.Content})
}

// handleStream 用 SSE 逐字推送回复。
func (s *server) handleStream(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	if q == "" {
		http.Error(w, "缺少参数 q", http.StatusBadRequest)
		return
	}
	fl, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "不支持流式", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.WriteHeader(http.StatusOK)
	fl.Flush()
	err := s.llm.StreamChat(r.Context(), []Message{
		{Role: "system", Content: "你是乐于助人的中文助手。"},
		{Role: "user", Content: q},
	}, func(delta string) {
		fmt.Fprintf(w, "data: %s\n\n", delta)
		fl.Flush()
	})
	if err != nil {
		fmt.Fprintf(w, "data: [ERROR] %v\n\n", err)
	}
}

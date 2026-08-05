// 课时12:评估、安全与工程化
//
// 三个能力:
// 1. 评估:go run . -eval 用 eval/cases.json 回归测试 Agent;
// 2. 安全:Guard 检测提示注入并注入防御指令;
// 3. 工程化:把 Agent 发布成 HTTP 服务 + 结构化日志 + 自动重试。
//
// 运行(已配置 DEEPSEEK_API_KEY):
//
//	go run . -eval                     # 跑离线评估
//	go run .                           # 启动 HTTP 服务
//	Invoke-RestMethod -Method Post -Uri http://localhost:8080/chat -ContentType application/json -Body '{"question":"12 * 8 + 4 等于多少?"}'
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)

func main() {
	evalMode := flag.Bool("eval", false, "运行离线评估(eval/cases.json)")
	port := flag.String("port", "8080", "HTTP 服务端口")
	flag.Parse()

	llm := NewLLMClient()
	if llm.APIKey == "" {
		fmt.Fprintln(os.Stderr, "缺少 DEEPSEEK_API_KEY 环境变量")
		os.Exit(1)
	}
	reg := NewRegistry()
	registerDefaultTools(reg)
	logger := NewLogger(os.Stderr)

	if *evalMode {
		cases, err := loadCases("eval/cases.json")
		if err != nil {
			fmt.Fprintln(os.Stderr, "加载用例失败:", err)
			os.Exit(1)
		}
		passed, total := RunEval(context.Background(), llm, reg, cases, logger)
		fmt.Printf("评估结果: %d/%d 通过\n", passed, total)
		if passed < total {
			os.Exit(1)
		}
		return
	}

	s := &server{llm: llm, reg: reg, logger: logger}
	http.HandleFunc("/health", s.handleHealth)
	http.HandleFunc("/chat", s.handleChat)
	fmt.Printf("Agent HTTP 服务已启动: http://localhost:%s\n", *port)
	if err := http.ListenAndServe(":"+*port, nil); err != nil {
		fmt.Fprintln(os.Stderr, "服务启动失败:", err)
		os.Exit(1)
	}
}

type server struct {
	llm    ChatLLM
	reg    *Registry
	logger *Logger
}

type chatReq struct {
	Question string `json:"question"`
}

type chatResp struct {
	Answer string `json:"answer"`
	Guard  bool   `json:"guard,omitempty"`
}

func (s *server) handleHealth(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"status": "ok"})
}

func (s *server) handleChat(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req chatReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || strings.TrimSpace(req.Question) == "" {
		http.Error(w, "question 必填", http.StatusBadRequest)
		return
	}
	start := time.Now()
	guardMsg := Guard(req.Question)

	// 网络类错误自动重试 3 次(指数退避)
	var answer string
	err := Retry(r.Context(), 3, 500*time.Millisecond, func() error {
		a, err := runReAct(r.Context(), s.llm, s.reg, req.Question, guardMsg)
		if err != nil {
			return &RetryableError{Err: err}
		}
		answer = a
		return nil
	})
	s.logger.Log("info", "chat", map[string]any{
		"question": truncateStr(req.Question, 80),
		"guard":    guardMsg != "",
		"duration": time.Since(start).String(),
	})
	if err != nil {
		s.logger.Log("error", "chat_failed", map[string]any{"err": err.Error()})
		http.Error(w, "agent 执行失败: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(chatResp{Answer: answer, Guard: guardMsg != ""})
}

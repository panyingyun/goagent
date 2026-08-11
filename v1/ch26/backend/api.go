package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
)

// Server 聚合 Store + DeepSeek 客户端,提供 HTTP API。
type Server struct {
	store *Store
	llm   *DeepSeekClient
}

func (s *Server) routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/sessions", s.handleListSessions)
	mux.HandleFunc("POST /api/sessions", s.handleCreateSession)
	mux.HandleFunc("GET /api/sessions/{id}", s.handleGetSession)
	mux.HandleFunc("PATCH /api/sessions/{id}", s.handlePatchSession)
	mux.HandleFunc("DELETE /api/sessions/{id}", s.handleDeleteSession)
	mux.HandleFunc("POST /api/chat", s.handleChat)
	// 前端构建产物(可选):go build 后把 React 的 dist 放到 ../frontend/dist
	if _, err := os.Stat("../frontend/dist"); err == nil {
		mux.Handle("/", http.FileServer(http.Dir("../frontend/dist")))
	} else {
		mux.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
			writeJSON(w, map[string]any{"name": "MathGraph AI API", "docs": "/api/sessions"})
		})
	}
	return cors(mux)
}

func cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET,POST,PATCH,DELETE,OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}

func decodeJSON(w http.ResponseWriter, r *http.Request, v any) error {
	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		http.Error(w, "请求体不是合法 JSON", http.StatusBadRequest)
		return err
	}
	return nil
}

func (s *Server) handleListSessions(w http.ResponseWriter, _ *http.Request) {
	list, err := s.store.ListSessions()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, list)
}

func (s *Server) handleCreateSession(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Title string `json:"title"`
	}
	if err := decodeJSON(w, r, &req); err != nil {
		return
	}
	if req.Title == "" {
		req.Title = "新建绘图"
	}
	sess, err := s.store.CreateSession(req.Title)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, sess)
}

func (s *Server) handleGetSession(w http.ResponseWriter, r *http.Request) {
	sess, err := s.store.GetSession(r.PathValue("id"))
	if errors.Is(err, ErrNotFound) {
		http.Error(w, "会话不存在", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, sess)
}

func (s *Server) handlePatchSession(w http.ResponseWriter, r *http.Request) {
	var p sessionPatch
	if err := decodeJSON(w, r, &p); err != nil {
		return
	}
	sess, err := s.store.PatchSession(r.PathValue("id"), p)
	if errors.Is(err, ErrNotFound) {
		http.Error(w, "会话不存在", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, sess)
}

func (s *Server) handleDeleteSession(w http.ResponseWriter, r *http.Request) {
	if err := s.store.DeleteSession(r.PathValue("id")); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, map[string]any{"ok": true})
}

type chatReq struct {
	SessionID string `json:"sessionId"`
	Message   string `json:"message"`
}

type chatResp struct {
	Message    Message    `json:"message"`
	GraphState GraphState `json:"graphState"`
}

// handleChat 是核心流程(设计稿 §9.2):
// 保存用户消息 -> 组装上下文 -> 调 DeepSeek -> 解析校验 -> 更新 graphState -> 保存并返回。
func (s *Server) handleChat(w http.ResponseWriter, r *http.Request) {
	var req chatReq
	if err := decodeJSON(w, r, &req); err != nil {
		return
	}
	if req.SessionID == "" || strings.TrimSpace(req.Message) == "" {
		http.Error(w, "sessionId 和 message 必填", http.StatusBadRequest)
		return
	}
	sess, err := s.store.GetSession(req.SessionID)
	if errors.Is(err, ErrNotFound) {
		http.Error(w, "会话不存在", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if s.llm.APIKey == "" {
		http.Error(w, "服务未就绪:缺少 DEEPSEEK_API_KEY", http.StatusServiceUnavailable)
		return
	}

	// 1. 保存用户消息
	userMsg := &Message{ID: newID("msg"), Role: "user", Content: req.Message, CreatedAt: now(), Status: "success"}
	if err := s.store.AddMessage(sess.ID, userMsg); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	sess.Messages = append(sess.Messages, *userMsg)

	// 2. 组装最近上下文(最近 10 条)
	recent := make([]llmMessage, 0, 10)
	start := len(sess.Messages) - 10
	if start < 0 {
		start = 0
	}
	for _, m := range sess.Messages[start:] {
		recent = append(recent, llmMessage{Role: m.Role, Content: m.Content})
	}
	userPrompt := buildUserPrompt(req.Message, &sess.GraphState, recent)

	// 3. 调 DeepSeek
	raw, err := s.llm.Chat(r.Context(), []llmMessage{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: userPrompt},
	})
	if err != nil {
		http.Error(w, "AI 调用失败: "+err.Error(), http.StatusBadGateway)
		return
	}
	if os.Getenv("DEEPSEEK_DEBUG") == "1" {
		fmt.Fprintf(os.Stderr, "[deepseek-debug] msg=%q raw=%s\n", req.Message, raw)
	}

	// 4. 解析校验
	res, err := ParseStructuredResult(raw)
	if err != nil {
		http.Error(w, "AI 返回格式异常,请重试", http.StatusUnprocessableEntity)
		return
	}
	if os.Getenv("DEEPSEEK_DEBUG") == "1" {
		fmt.Fprintf(os.Stderr, "[deepseek-debug] parsed_intent=%s result_equations=%d\n", res.Intent, len(res.Equations))
	}
	// 4.5 意图兜底:用户说"再加"但模型返回 plot 时,强制改成追加
	CoerceIntent(res, req.Message, &sess.GraphState)
	if os.Getenv("DEEPSEEK_DEBUG") == "1" {
		fmt.Fprintf(os.Stderr, "[deepseek-debug] after_coerce=%s\n", res.Intent)
	}

	// 4.6 兜底重试:模型把"解释/分析"误判为 unknown 时,带方程上下文重试一次
	if res.Intent == "unknown" && wantsAnalysis(req.Message) && len(sess.GraphState.Equations) > 0 {
		retryRaw, retryErr := s.llm.Chat(r.Context(), []llmMessage{
			{Role: "system", Content: systemPrompt +
				"\n用户想分析当前图像。当前图像方程: " + equationsSummary(&sess.GraphState) +
				"。请返回 analyze 或 explain 意图,analysis 字段使用指定 schema。"},
			{Role: "user", Content: req.Message},
		})
		if retryErr == nil {
			if retry, pErr := ParseStructuredResult(retryRaw); pErr == nil {
				res = retry
			}
		}
	}

	// 5. 应用更新,生成 assistant 文本
	content := ApplyResult(&sess.GraphState, res)

	// 6. 保存 assistant 消息 + graphState
	assistantMsg := &Message{
		ID:               newID("msg"),
		Role:             "assistant",
		Content:          content,
		StructuredResult: res,
		CreatedAt:        now(),
		Status:           "success",
	}
	if err := s.store.AddMessage(sess.ID, assistantMsg); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := s.store.SaveGraphState(sess.ID, &sess.GraphState); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	sess.Messages = append(sess.Messages, *assistantMsg)

	writeJSON(w, chatResp{Message: *assistantMsg, GraphState: sess.GraphState})
}

// wantsAnalysis 判断用户消息是否在要求"解释/分析"。
func wantsAnalysis(msg string) bool {
	keywords := []string{"解释", "分析", "单调", "顶点", "零点", "对称", "极值", "图像特征"}
	for _, k := range keywords {
		if strings.Contains(msg, k) {
			return true
		}
	}
	return false
}

func equationsSummary(gs *GraphState) string {
	labels := make([]string, 0, len(gs.Equations))
	for _, e := range gs.Equations {
		labels = append(labels, e.Expression)
	}
	return strings.Join(labels, ", ")
}

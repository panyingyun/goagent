package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// mockLLM 用本地 HTTP 服务模拟 DeepSeek,离线也能测试 ReAct 循环。
func mockLLM(t *testing.T, responses []string, after func(idx int, req chatRequest)) (*httptest.Server, *LLMClient) {
	t.Helper()
	var calls int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req chatRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("解析请求失败: %v", err)
			return
		}
		idx := calls
		calls++
		msg := Message{Role: RoleAssistant}
		if idx < len(responses) {
			msg.Content = responses[idx]
		} else {
			msg.Content = "Final Answer: 模拟回复"
		}
		if after != nil {
			after(idx, req)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"choices": []any{map[string]any{"message": msg}},
		})
	}))
	client := NewLLMClient()
	client.APIKey = "test-key"
	client.BaseURL = srv.URL
	return srv, client
}

// 模型先要调用计算器,拿到 Observation 后给出最终答案。
func TestReActLoopUsesCalculator(t *testing.T) {
	srv, llm := mockLLM(t, []string{
		"Thought: 需要计算。\nAction: calculator\nAction Input: {\"expression\": \"12 * 8 + 4\"}",
		"Thought: 得到了 100。\nFinal Answer: 12 × 8 + 4 = 100",
	}, func(idx int, req chatRequest) {
		if idx == 1 {
			last := req.Messages[len(req.Messages)-1]
			if !strings.Contains(last.Content, "Observation: 100") {
				t.Errorf("第 2 次请求应包含 Observation: 100,实际: %q", last.Content)
			}
		}
	})
	defer srv.Close()

	reg := NewRegistry()
	registerDefaultTools(reg)
	answer, err := runReAct(context.Background(), llm, reg, "12 * 8 + 4 等于多少?")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(answer, "100") {
		t.Fatalf("期望答案包含 100,得到: %q", answer)
	}
}

func TestReActDirectAnswer(t *testing.T) {
	srv, llm := mockLLM(t, []string{"Thought: 简单问候。\nFinal Answer: 你好,我是 DemoAgent。"}, nil)
	defer srv.Close()

	reg := NewRegistry()
	registerDefaultTools(reg)
	answer, err := runReAct(context.Background(), llm, reg, "你好")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(answer, "DemoAgent") {
		t.Fatalf("期望答案包含 DemoAgent,得到: %q", answer)
	}
}

func TestCalculator(t *testing.T) {
	cases := map[string]float64{
		"12 * 8 + 4":  100,
		"(1 + 2) * 3": 9,
		"10 / 4":      2.5,
		"-5 + 3":      -2,
	}
	for expr, want := range cases {
		got, err := evalCalc(expr)
		if err != nil {
			t.Errorf("%q 计算失败: %v", expr, err)
			continue
		}
		if got != want {
			t.Errorf("%q = %v,期望 %v", expr, got, want)
		}
	}
}

func TestCalculatorErrors(t *testing.T) {
	for _, expr := range []string{"1 / 0", "1 +", "(1 + 2", "abc"} {
		if _, err := evalCalc(expr); err == nil {
			t.Errorf("%q 应当报错但没有", expr)
		}
	}
}

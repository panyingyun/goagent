package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

// mockLLM 用本地 HTTP 服务模拟 DeepSeek 的 /chat/completions,
// 按调用次数返回脚本化回复,让我们不联网也能测试 Agent 循环。
// after 在每次请求处理完后回调,便于断言请求内容。
func mockLLM(t *testing.T, responses []string, toolCalls []ToolCall, after func(idx int, req chatRequest)) (*httptest.Server, *LLMClient) {
	t.Helper()
	var calls int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/chat/completions" {
			t.Errorf("意外的路径: %s", r.URL.Path)
			return
		}
		var req chatRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("解析请求失败: %v", err)
			return
		}
		idx := calls
		calls++
		if len(req.Messages) > 0 {
			last := req.Messages[len(req.Messages)-1]
			t.Logf("第 %d 次请求,最后消息 role=%s content=%q", idx+1, last.Role, truncateStr(last.Content, 80))
		}

		var msg Message
		if idx < len(responses) {
			msg = Message{Role: RoleAssistant, Content: responses[idx]}
		}
		if idx < len(toolCalls) {
			msg.ToolCalls = []ToolCall{toolCalls[idx]}
		}
		if msg.Content == "" && len(msg.ToolCalls) == 0 {
			msg = Message{Role: RoleAssistant, Content: "Final Answer: 模拟回复"}
		}
		if after != nil {
			after(idx, req)
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(map[string]any{
			"choices": []any{map[string]any{"message": msg}},
		}); err != nil {
			t.Errorf("写响应失败: %v", err)
		}
	}))
	client := NewLLMClient()
	client.APIKey = "test-key"
	client.BaseURL = srv.URL
	return srv, client
}

func truncateStr(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}

// 测试文本 ReAct 循环:模型先要求调用计算器,拿到 Observation 后给出最终答案。
func TestReActLoopUsesCalculator(t *testing.T) {
	srv, llm := mockLLM(t, []string{
		"Thought: 用户需要计算 12*8+4,我该用计算器。\nAction: calculator\nAction Input: {\"expression\": \"12 * 8 + 4\"}",
		"Thought: 得到了 100,任务完成。\nFinal Answer: 12 × 8 + 4 = 100",
	}, nil, func(idx int, req chatRequest) {
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
	answer, err := runReAct(context.Background(), llm, reg, nil, "12 * 8 + 4 等于多少?")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(answer, "100") {
		t.Fatalf("期望答案包含 100,得到: %q", answer)
	}
}

// 模型直接给出最终答案,不需要调用工具。
func TestReActDirectAnswer(t *testing.T) {
	srv, llm := mockLLM(t, []string{"Thought: 简单问候。\nFinal Answer: 你好,我是 DemoAgent。"}, nil, nil)
	defer srv.Close()

	reg := NewRegistry()
	registerDefaultTools(reg)
	answer, err := runReAct(context.Background(), llm, reg, nil, "你好")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(answer, "DemoAgent") {
		t.Fatalf("期望答案包含 DemoAgent,得到: %q", answer)
	}
}

// 测试原生 function calling 模式:模型返回结构化 tool_calls。
func TestToolCallMode(t *testing.T) {
	tc := ToolCall{
		ID:   "call_1",
		Type: "function",
		Function: ToolCallFunction{
			Name:      "calculator",
			Arguments: `{"expression": "2 + 3 * 4"}`,
		},
	}
	srv, llm := mockLLM(t, []string{"", "Final Answer: 结果是 14"}, []ToolCall{tc}, nil)
	defer srv.Close()

	reg := NewRegistry()
	registerDefaultTools(reg)
	answer, err := runToolCall(context.Background(), llm, reg, nil, "2 + 3 * 4 等于多少?")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(answer, "14") {
		t.Fatalf("期望答案包含 14,得到: %q", answer)
	}
}

func TestCalculator(t *testing.T) {
	cases := map[string]float64{
		"12 * 8 + 4":  100,
		"(1 + 2) * 3": 9,
		"10 / 4":      2.5,
		"-5 + 3":      -2,
		"2.5 * 2":     5,
		"100 - 3 * 7": 79,
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

// TestMain 先编译 MCP 服务器,供下面的 MCP 测试使用。
var mcpBin string

func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "mcp-test")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	mcpBin = filepath.Join(dir, "mcp-server")
	if runtime.GOOS == "windows" {
		mcpBin += ".exe"
	}
	// -buildvcs=false:测试不依赖 git 仓库状态(某些环境 git 不可用时也能跑)
	cmd := exec.Command("go", "build", "-buildvcs=false", "-o", mcpBin, "./mcp/server")
	if out, err := cmd.CombinedOutput(); err != nil {
		fmt.Fprintf(os.Stderr, "构建 MCP 服务器失败: %v\n%s\n", err, out)
		os.Exit(1)
	}
	code := m.Run()
	_ = os.RemoveAll(dir)
	os.Exit(code)
}

// 最小 MCP 客户端:握手、调用 echo 和 get_time。
func TestMCPClient(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	c, err := StartMCPClient(ctx, mcpBin)
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()

	text, err := c.CallTool(ctx, "echo", map[string]any{"text": "hello mcp"})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(text, "hello mcp") {
		t.Fatalf("echo 结果不符: %q", text)
	}

	text, err = c.CallTool(ctx, "get_time", map[string]any{})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(text, ":") {
		t.Fatalf("时间格式不符: %q", text)
	}
}

// 把 MCP 服务器注册成 Agent 的一个工具,验证转发链路。
func TestMCPToolRegistered(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	c, err := StartMCPClient(ctx, mcpBin)
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()

	reg := NewRegistry()
	registerDefaultTools(reg)
	registerMCPTool(reg, c)

	got, err := reg.Run(ctx, "mcp", `{"tool": "echo", "arguments": {"text": "hi"}}`)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, "hi") {
		t.Fatalf("MCP 工具结果不符: %q", got)
	}
}

// 课时24:LangSmith 风格追踪
//
// LangSmith 的核心是 run 树:每次 LLM 调用、工具调用都是一个 run,
// 带父子关系、属性(如 token)和耗时。本课实现同样的结构,
// 并演示把真实 Agent 的每一步包进 run 树。
//
// 运行(已配置 DEEPSEEK_API_KEY):
//
//	go run . "12 * 8 + 4 等于多少?"
package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
)

func main() {
	llm := NewLLMClient()
	if llm.APIKey == "" {
		fmt.Fprintln(os.Stderr, "缺少 DEEPSEEK_API_KEY")
		os.Exit(1)
	}
	question := strings.Join(os.Args[1:], " ")
	if question == "" {
		data, _ := io.ReadAll(os.Stdin)
		question = strings.TrimSpace(string(data))
	}
	if question == "" {
		fmt.Fprintln(os.Stderr, "用法: go run . \"问题\"")
		os.Exit(1)
	}

	tracer := NewTracer()
	ctx := context.Background()

	agentRun := tracer.StartRun("agent")
	thinkRun := tracer.StartRun("llm.think")
	reply, err := llm.Chat(ctx, []Message{
		{Role: "system", Content: "判断是否需要工具,回答 JSON: {\"need_tool\": true, \"expression\": \"...\"}"},
		{Role: "user", Content: question},
	})
	thinkRun.Attrs["tokens"] = estimateTokens(question) + estimateTokens(reply.Content)
	if err != nil {
		tracer.FailRun(thinkRun, err)
		fmt.Fprintln(os.Stderr, "调用失败:", err)
		os.Exit(1)
	}
	tracer.EndRun(thinkRun)

	toolRun := tracer.StartRun("tool.calculator")
	obs := "未使用工具"
	if strings.Contains(reply.Content, "true") {
		expr := extractExpr(reply.Content)
		obs = evalCalc(expr)
		toolRun.Attrs["expression"] = expr
	}
	tracer.EndRun(toolRun)

	answerRun := tracer.StartRun("llm.answer")
	answer, err := llm.Chat(ctx, []Message{
		{Role: "system", Content: "根据 Observation 组织最终答案。"},
		{Role: "user", Content: "Observation: " + obs + "\n问题: " + question},
	})
	answerRun.Attrs["tokens"] = estimateTokens(obs) + estimateTokens(answer.Content)
	if err != nil {
		tracer.FailRun(answerRun, err)
		os.Exit(1)
	}
	tracer.EndRun(answerRun)
	tracer.EndRun(agentRun)

	fmt.Println("=== 回答 ===")
	fmt.Println(answer.Content)
	fmt.Println("\n=== run 树(JSON) ===")
	fmt.Println(tracer.ExportJSON())
}

func estimateTokens(s string) int {
	return len([]rune(s))
}

func extractExpr(s string) string {
	// 极简提取:找双引号里的内容
	start := strings.IndexByte(s, '"')
	if start < 0 {
		return ""
	}
	rest := s[start+1:]
	end := strings.IndexByte(rest, '"')
	if end < 0 {
		return ""
	}
	return rest[:end]
}

func evalCalc(expr string) string {
	// 演示工具:支持 "a op b"
	parts := strings.Fields(expr)
	if len(parts) != 3 {
		return "无法解析: " + expr
	}
	var a, b float64
	if _, err := fmt.Sscanf(parts[0], "%g", &a); err != nil {
		return "解析失败"
	}
	if _, err := fmt.Sscanf(parts[2], "%g", &b); err != nil {
		return "解析失败"
	}
	switch parts[1] {
	case "+":
		return fmt.Sprintf("%v", a+b)
	case "*":
		return fmt.Sprintf("%v", a*b)
	}
	return "不支持的运算符"
}

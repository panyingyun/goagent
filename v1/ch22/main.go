// 课时22:迷你 LangChain——用 Go 复刻 LangChain 的核心抽象
//
// 我们手写的这些组件,和 LangChain 的组件一一对应:
//
//	Chain / LLMChain / SequentialChain / Memory / Tool / AgentExecutor
//
// 运行(已配置 DEEPSEEK_API_KEY):
//
//	go run . "把下面这句话翻译成英文: 你好,世界"
package main

import (
	"context"
	"encoding/json"
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
	input := strings.Join(os.Args[1:], " ")
	if input == "" {
		data, _ := io.ReadAll(os.Stdin)
		input = strings.TrimSpace(string(data))
	}
	if input == "" {
		fmt.Fprintln(os.Stderr, "用法: go run . \"输入\"")
		os.Exit(1)
	}
	ctx := context.Background()

	// 1. SequentialChain:翻译 -> 总结
	translate := &LLMChain{LLM: llm, System: "你是翻译,只输出译文。"}
	summarize := &LLMChain{LLM: llm, System: "你是总结,把输入压缩成一句话。"}
	seq := &SequentialChain{Chains: []Chain{translate, summarize}}
	out, err := seq.Call(ctx, input)
	if err != nil {
		fmt.Fprintln(os.Stderr, "链执行失败:", err)
		os.Exit(1)
	}
	fmt.Println("=== SequentialChain 输出 ===")
	fmt.Println(out)

	// 2. AgentExecutor:带工具的 Agent
	executor := &AgentExecutor{
		LLM: llm,
		Tools: []Tool{
			{Name: "calculator", Description: "计算数学表达式", Run: runCalculator},
		},
		MaxSteps: 8,
	}
	ans, err := executor.Call(ctx, "12 * 8 + 4 等于多少?")
	if err != nil {
		fmt.Fprintln(os.Stderr, "Agent 执行失败:", err)
		os.Exit(1)
	}
	fmt.Println("\n=== AgentExecutor 输出 ===")
	fmt.Println(ans)
}

// runCalculator 简单计算(演示用,支持 + - * / 括号)。
func runCalculator(_ context.Context, jsonArgs string) (string, error) {
	var args struct {
		Expression string `json:"expression"`
	}
	if err := json.Unmarshal([]byte(jsonArgs), &args); err != nil || args.Expression == "" {
		return "", fmt.Errorf("缺少 expression")
	}
	return evalSimple(args.Expression)
}

func evalSimple(expr string) (string, error) {
	// 极简:只处理 "a op b"
	parts := strings.Fields(expr)
	if len(parts) != 3 {
		return "", fmt.Errorf("仅支持 a op b 形式")
	}
	var a, b float64
	if _, err := fmt.Sscanf(parts[0], "%g", &a); err != nil {
		return "", err
	}
	if _, err := fmt.Sscanf(parts[2], "%g", &b); err != nil {
		return "", err
	}
	switch parts[1] {
	case "+":
		return fmt.Sprintf("%v", a+b), nil
	case "-":
		return fmt.Sprintf("%v", a-b), nil
	case "*":
		return fmt.Sprintf("%v", a*b), nil
	case "/":
		if b == 0 {
			return "", fmt.Errorf("除数为 0")
		}
		return fmt.Sprintf("%v", a/b), nil
	}
	return "", fmt.Errorf("不支持的运算符: %s", parts[1])
}

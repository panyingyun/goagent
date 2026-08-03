// 课时19:评估进阶——LLM-as-Judge
//
// 用"评审员模型"给"答题模型"的回答打分,解决主观题自动化评估。
//
// 运行(已配置 DEEPSEEK_API_KEY):
//
//	go run . "Go 的 goroutine 是什么?"
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
	ctx := context.Background()

	// 1. 答题模型回答
	answer, err := llm.Chat(ctx, []Message{
		{Role: "system", Content: "你是知识助手,回答简洁准确。"},
		{Role: "user", Content: question},
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, "答题失败:", err)
		os.Exit(1)
	}
	fmt.Println("=== 回答 ===")
	fmt.Println(answer.Content)

	// 2. 评审模型打分
	judge := &Judge{LLM: llm, Criteria: []string{
		"准确性:没有事实错误",
		"完整性:覆盖问题要点",
		"清晰度:语言通俗易懂",
	}}
	score, reason, err := judge.Score(ctx, question, answer.Content)
	if err != nil {
		fmt.Fprintln(os.Stderr, "评分失败:", err)
		os.Exit(1)
	}
	fmt.Printf("\n=== 评审(LLM-as-Judge) ===\n得分: %.1f/10\n理由: %s\n", score, reason)
}

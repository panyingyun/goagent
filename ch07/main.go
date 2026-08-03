// 课时7:流式输出
//
// 让模型回复"边生成边显示",而不是干等十几秒。
//
// 运行(已配置 DEEPSEEK_API_KEY):
//
//	go run . "用三句话介绍 Go 语言"
//	go run . "写一首关于夏天的短诗"
package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"time"
)

func main() {
	llm := NewLLMClient()
	if llm.APIKey == "" {
		fmt.Fprintln(os.Stderr, "缺少 DEEPSEEK_API_KEY 环境变量")
		os.Exit(1)
	}

	question := strings.Join(os.Args[1:], " ")
	if question == "" {
		data, _ := io.ReadAll(os.Stdin)
		question = strings.TrimSpace(string(data))
	}
	if question == "" {
		fmt.Fprintln(os.Stderr, "用法: go run . \"你的问题\"")
		os.Exit(1)
	}

	messages := []Message{
		{Role: RoleSystem, Content: "你是一个乐于助人的中文助手。"},
		{Role: RoleUser, Content: question},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	fmt.Print("Agent> ")
	start := time.Now()
	err := llm.StreamChat(ctx, messages, func(delta string) {
		fmt.Print(delta)
		// 立即刷出,否则终端会等缓冲区满了才显示
		_ = os.Stdout.Sync()
	})
	fmt.Println()
	if err != nil {
		fmt.Fprintln(os.Stderr, "\n流式调用失败:", err)
		os.Exit(1)
	}
	fmt.Printf("(耗时 %.1f 秒)\n", time.Since(start).Seconds())
}

// 课时4:Go 实现 ReAct
//
// ReAct = Reasoning(推理) + Acting(行动):让 LLM 边想边动,
// 需要信息就调用工具,看到结果继续想,直到给出最终答案。
//
// 运行(已配置好 DEEPSEEK_API_KEY):
//   go run . "12 * 8 + 4 等于多少?"
//   go run . "现在北京时间几点?"
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
		fmt.Fprintln(os.Stderr, "缺少 DEEPSEEK_API_KEY 环境变量")
		os.Exit(1)
	}

	reg := NewRegistry()
	registerDefaultTools(reg)

	input := strings.Join(os.Args[1:], " ")
	if input == "" {
		data, _ := io.ReadAll(os.Stdin)
		input = strings.TrimSpace(string(data))
	}
	if input == "" {
		fmt.Fprintln(os.Stderr, "用法: go run . \"你的问题\"")
		os.Exit(1)
	}

	answer, err := runReAct(context.Background(), llm, reg, input)
	if err != nil {
		fmt.Fprintln(os.Stderr, "运行失败:", err)
		os.Exit(1)
	}
	fmt.Println("\n=== 最终回答 ===")
	fmt.Println(answer)
}

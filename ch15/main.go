// 课时15:多 Agent 协作
//
// 用三个角色协作完成任务:
// 规划者(拆解任务) -> 执行者(执行) -> 审查者(质检),不合格就返工。
//
// 运行(已配置 DEEPSEEK_API_KEY):
//
//	go run . "写一段产品介绍,重点突出 AI 能力"
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
	task := strings.Join(os.Args[1:], " ")
	if task == "" {
		data, _ := io.ReadAll(os.Stdin)
		task = strings.TrimSpace(string(data))
	}
	if task == "" {
		fmt.Fprintln(os.Stderr, "用法: go run . \"任务描述\"")
		os.Exit(1)
	}

	orch := &Orchestra{
		Planner:   &Agent{Name: "planner", Role: "你是规划者:把任务拆成清晰的执行计划,输出 2~3 步。", LLM: llm},
		Executor:  &Agent{Name: "executor", Role: "你是执行者:严格按照计划产出最终内容,不要解释过程。", LLM: llm},
		Reviewer:  &Agent{Name: "reviewer", Role: "你是审查者:检查内容质量,发现问题就说改进建议,合格就回复\"通过\"。", LLM: llm},
		MaxRounds: 3,
	}
	result, err := orch.Run(context.Background(), task)
	if err != nil {
		fmt.Fprintln(os.Stderr, "失败:", err)
		os.Exit(1)
	}
	fmt.Println("\n=== 最终交付 ===")
	fmt.Println(result)
}

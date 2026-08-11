// 课时23:迷你 LangGraph——从"循环"到"图"
//
// LangGraph 的核心:节点(Node) + 边(Edge) + 共享状态(State) + 检查点(Checkpoint)。
// 本课实现一个规划->执行->审查的图,审查不通过就沿 __next__ 回到执行节点。
//
// 运行(已配置 DEEPSEEK_API_KEY):
//
//	go run . "写一段产品介绍,突出 AI 能力"
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
		fmt.Fprintln(os.Stderr, "用法: go run . \"任务\"")
		os.Exit(1)
	}

	g := BuildWorkflowGraph(llm, 2)
	state := map[string]any{"task": task, "rounds": 0}
	final, err := g.Invoke(context.Background(), "plan", state, "checkpoint.json")
	if err != nil {
		fmt.Fprintln(os.Stderr, "图执行失败:", err)
		os.Exit(1)
	}
	fmt.Println("=== 最终结果 ===")
	fmt.Println(final["result"])
	fmt.Println("\n=== 审查意见 ===")
	fmt.Println(final["review"])
	fmt.Println("\n(执行轮次:", final["rounds"], "检查点已保存到 checkpoint.json)")
}

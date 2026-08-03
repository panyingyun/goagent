// 课时14:Agent 工作流与任务规划
//
// 把"单循环"升级成"多步骤工作流":
// 规划(拆解) -> 执行(逐步骤) -> 检查点(断点续跑)
//
// 运行(已配置 DEEPSEEK_API_KEY):
//
//	go run . -plan "帮我准备一次新员工入职介绍"
//	go run . -run      # 执行 checkpoint.json 里的工作流
//	go run . -resume   # 从断点继续(模拟中断后恢复)
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
)

const checkpointFile = "checkpoint.json"

func main() {
	plan := flag.String("plan", "", "输入目标,生成工作流")
	run := flag.Bool("run", false, "执行 checkpoint.json 中的工作流")
	resume := flag.Bool("resume", false, "从断点继续执行")
	flag.Parse()

	llm := NewLLMClient()
	if llm.APIKey == "" {
		fmt.Fprintln(os.Stderr, "缺少 DEEPSEEK_API_KEY")
		os.Exit(1)
	}
	ctx := context.Background()

	switch {
	case *plan != "":
		w := &Workflow{ID: "wf-1", Goal: *plan, Checkpoint: checkpointFile}
		if err := w.Plan(ctx, llm); err != nil {
			fmt.Fprintln(os.Stderr, "规划失败:", err)
			os.Exit(1)
		}
		fmt.Printf("规划完成,共 %d 步,已保存到 %s\n", len(w.Steps), checkpointFile)
		for _, s := range w.Steps {
			fmt.Printf("  %d. %s\n", s.ID+1, s.Desc)
		}
	case *run || *resume:
		w, err := LoadWorkflow(checkpointFile)
		if err != nil {
			fmt.Fprintln(os.Stderr, "加载工作流失败:", err)
			os.Exit(1)
		}
		if err := w.Resume(ctx, llm); err != nil {
			fmt.Fprintln(os.Stderr, "执行中断:", err)
			os.Exit(1)
		}
		fmt.Println("\n=== 全部步骤完成 ===")
		for _, s := range w.Steps {
			fmt.Printf("%d. %s\n   -> %s\n", s.ID+1, s.Desc, s.Output)
		}
	default:
		fmt.Fprintln(os.Stderr, "用法: go run . -plan \"目标\" / -run / -resume")
		os.Exit(1)
	}
}

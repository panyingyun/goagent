// 附录示例:对标旧课 ch04 的综合 Agent(薄入口)。
//
// 功能走共享框架:
//   - Agent 循环 → pkg/agent.Runner + Executor
//   - 工具       → pkg/agent.NewCalculatorTool / NewGetTimeTool
//   - LLM        → pkg/llm
//   - 观测       → pkg/trace
//
//	go run ./appendices/complete-agent "12 * 8 + 4 等于多少?"
package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"goagent-new/pkg/agent"
	"goagent-new/pkg/llm"
	"goagent-new/pkg/trace"
)

const systemPrompt = `你是一个运行在 Go Agent 框架中的助手 DemoAgent。
需要信息时调用工具;数学题必须用 calculator,时间题必须用 get_time。
不要编造工具结果。`

func main() {
	client := llm.NewHTTPClientFromEnv()
	if !client.Configured() {
		fmt.Fprintln(os.Stderr, "请设置 DEEPSEEK_API_KEY")
		os.Exit(1)
	}

	recorder := trace.NewRecorder()
	executor := agent.NewExecutor()
	executor.Observer = recorder
	for _, tool := range []agent.Tool{agent.NewCalculatorTool(), agent.NewGetTimeTool()} {
		if err := executor.Register(tool); err != nil {
			fatal(err)
		}
	}

	input := strings.Join(os.Args[1:], " ")
	if input == "" {
		data, _ := io.ReadAll(os.Stdin)
		input = strings.TrimSpace(string(data))
	}
	if input == "" {
		fmt.Fprintln(os.Stderr, `用法: go run ./appendices/complete-agent "12 * 8 + 4 等于多少?"`)
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	runner := agent.Runner{LLM: client, Executor: executor, Observer: recorder, MaxRounds: 10}
	answer, history, err := runner.Run(ctx, []llm.Message{
		{Role: llm.RoleSystem, Content: systemPrompt},
		{Role: llm.RoleUser, Content: input},
	})
	if err != nil {
		fatal(err)
	}
	// 跟踪轨迹:assistant 的 tool_calls(含参数) → tool 结果 → 最终回答
	for i, m := range history {
		switch m.Role {
		case llm.RoleAssistant:
			if len(m.ToolCalls) > 0 {
				fmt.Fprintf(os.Stderr, "\n--- 第 %d 条: LLM 决定调工具 ---\n", i)
				for _, tc := range m.ToolCalls {
					fmt.Fprintf(os.Stderr, "call_id=%s\n  tool=%s\n  arguments=%s\n", tc.ID, tc.Function.Name, tc.Function.Arguments)
				}
			} else if strings.TrimSpace(m.Content) != "" {
				fmt.Fprintf(os.Stderr, "\n--- 第 %d 条: LLM 最终文本 ---\n%s\n", i, m.Content)
			}
		case llm.RoleTool:
			fmt.Fprintf(os.Stderr, "\n--- 第 %d 条: 工具结果 ---\ncall_id=%s\n  result=%s\n", i, m.ToolCallID, m.Content)
		}
	}
	fmt.Println("\n=== 最终回答 ===")
	fmt.Println(answer)
	_ = recorder.Export(os.Stderr) // 可选:再导出 observer 事件(耗时/token)
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}

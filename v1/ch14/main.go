// 课时14:可观测性与日志
//
// 给 Agent 加"黑匣子":调用链追踪(span 树) + 指标(计数/耗时)。
// 本课用模拟 Agent 演示,把真实 LLM 调用包在 span 里即可。
//
// 运行:
//
//	go run .
package main

import (
	"fmt"
	"os"
	"time"
)

func main() {
	tracer := NewTracer()
	metrics := NewMetrics()

	// 模拟一次 Agent 任务:llm 思考 -> 调用工具 -> 再次 llm 回答
	root := tracer.Start("agent.task")
	metrics.Inc("task_started")
	start := time.Now()

	llmSpan := tracer.Start("llm.think")
	time.Sleep(20 * time.Millisecond)
	tracer.End(llmSpan)
	metrics.Observe("llm_duration_ms", float64(llmSpan.End.Sub(llmSpan.Start).Milliseconds()))

	toolSpan := tracer.Start("tool.calculator")
	time.Sleep(10 * time.Millisecond)
	toolSpan.Attrs["expression"] = "12 * 8 + 4"
	tracer.End(toolSpan)
	metrics.Observe("tool_duration_ms", float64(toolSpan.End.Sub(toolSpan.Start).Milliseconds()))

	llmSpan2 := tracer.Start("llm.answer")
	time.Sleep(15 * time.Millisecond)
	tracer.End(llmSpan2)
	metrics.Observe("llm_duration_ms", float64(llmSpan2.End.Sub(llmSpan2.Start).Milliseconds()))

	tracer.End(root)
	metrics.Inc("task_finished")
	metrics.Observe("task_duration_ms", float64(time.Since(start).Milliseconds()))

	fmt.Println("=== 调用链追踪(span 树) ===")
	if err := tracer.Export(os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
	fmt.Println("\n=== 指标快照 ===")
	for k, v := range metrics.Snapshot() {
		fmt.Printf("%s = %v\n", k, v)
	}
}

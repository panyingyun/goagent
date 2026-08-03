// 课时13:工具进阶与并行调用
//
// 三个升级:
// 1. 参数校验:必填字段缺失直接报错,不执行
// 2. 结果缓存:同样的调用 60 秒内直接返回缓存
// 3. 并行执行:一次多个 tool_calls,并发执行、按序回填
//
// 运行(已配置 DEEPSEEK_API_KEY):
//
//	go run . "帮我同时查几个问题:1+1=? 2+2=? 3+3=?"
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
		fmt.Fprintln(os.Stderr, "缺少 DEEPSEEK_API_KEY")
		os.Exit(1)
	}
	reg := NewRegistry()
	registerDemoTools(reg)
	ex := NewExecutor(reg)

	input := strings.Join(os.Args[1:], " ")
	if input == "" {
		data, _ := io.ReadAll(os.Stdin)
		input = strings.TrimSpace(string(data))
	}
	if input == "" {
		fmt.Fprintln(os.Stderr, "用法: go run . \"你的问题\"")
		os.Exit(1)
	}
	answer, err := runToolCallParallel(context.Background(), llm, ex, input)
	if err != nil {
		fmt.Fprintln(os.Stderr, "失败:", err)
		os.Exit(1)
	}
	fmt.Println("\n=== 最终回答 ===")
	fmt.Println(answer)
}

// registerDemoTools 注册几个"慢工具",方便观察并行效果。
func registerDemoTools(r *Registry) {
	mk := func(name, desc string, fn func(a, b float64) float64) {
		r.Register(Tool{
			Name:        name,
			Description: desc,
			Required:    []string{"a", "b"},
			Run: func(ctx context.Context, args map[string]any) (string, error) {
				select {
				case <-ctx.Done():
					return "", ctx.Err()
				case <-time.After(300 * time.Millisecond): // 模拟耗时工具
				}
				a, ok1 := toFloat(args["a"])
				b, ok2 := toFloat(args["b"])
				if !ok1 || !ok2 {
					return "", fmt.Errorf("参数 a、b 必须是数字")
				}
				return fmt.Sprintf("%v", fn(a, b)), nil
			},
		})
	}
	mk("add", "两个数相加", func(a, b float64) float64 { return a + b })
	mk("mul", "两个数相乘", func(a, b float64) float64 { return a * b })
	mk("sub", "两个数相减", func(a, b float64) float64 { return a - b })
}

func toFloat(v any) (float64, bool) {
	switch n := v.(type) {
	case float64:
		return n, true
	case string:
		var f float64
		if _, err := fmt.Sscanf(n, "%g", &f); err == nil {
			return f, true
		}
	}
	return 0, false
}

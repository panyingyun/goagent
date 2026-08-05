package main

import (
	"fmt"

	"goagent-new/pkg/contextkit"
)

func buildDemoContext(query string) (contextkit.Bundle, error) {
	sources := []contextkit.Source{
		{ID: "system-policy", Kind: "policy", Content: "修改代码后必须运行测试。不得执行未授权的外部写操作。", Trust: contextkit.Trusted, Priority: 100, Required: true},
		{ID: "go-guide", Kind: "project-doc", Content: "Go HTTP handler 必须传播 request context，并限制请求体。", Trust: contextkit.Untrusted, Priority: 20},
		{ID: "ui-guide", Kind: "project-doc", Content: "前端使用 React 和响应式三栏布局。", Trust: contextkit.Untrusted, Priority: 10},
		{ID: "old-log", Kind: "tool-output", Content: "旧任务构建成功。", Trust: contextkit.Untrusted, Priority: 1},
	}
	return (contextkit.Assembler{MaxRunes: 420, MaxItems: 3}).Build(query, sources)
}

func main() {
	bundle, err := buildDemoContext("修复 Go HTTP handler 的取消问题")
	if err != nil {
		panic(err)
	}
	fmt.Printf("selected=%d dropped=%v used=%d/%d\n%s\n", len(bundle.Items), bundle.DroppedIDs, bundle.UsedRunes, bundle.BudgetRunes, bundle.Text)
}

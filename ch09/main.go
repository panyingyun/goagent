// 课时9:真实工具——让 Agent 真的"干活"
//
// 给 Agent 装上:联网(http_get)、读写文件(read_file/write_file,限 docs 目录)、
// 执行命令(run_command,白名单)、本地键值存储(kv_get/kv_set)。
// 重点讲"工具安全":路径穿越、命令白名单、超时。
//
// 运行(已配置 DEEPSEEK_API_KEY):
//
//	go run . "访问 https://httpbin.org/get 看看返回什么"
//	go run . "把 hello agent 写到 docs/note.txt"
//	go run . "执行 go version 看看环境"
package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	llm := NewLLMClient()
	if llm.APIKey == "" {
		fmt.Fprintln(os.Stderr, "缺少 DEEPSEEK_API_KEY 环境变量")
		os.Exit(1)
	}
	if err := os.MkdirAll(docRoot, 0o755); err != nil {
		fmt.Fprintln(os.Stderr, "创建 docs 目录失败:", err)
		os.Exit(1)
	}
	// 首次运行放一个示例文件,方便演示 read_file
	note := filepath.Join(docRoot, "notes.txt")
	if _, err := os.Stat(note); os.IsNotExist(err) {
		_ = os.WriteFile(note, []byte("这是 Agent 可以读取的示例笔记:项目代号 moon,每周五发周报。"), 0o644)
	}

	reg := NewRegistry()
	registerDefaultTools(reg)

	input := strings.Join(os.Args[1:], " ")
	if input == "" {
		data, _ := io.ReadAll(os.Stdin)
		input = strings.TrimSpace(string(data))
	}
	if input == "" {
		fmt.Fprintln(os.Stderr, "用法: go run . \"你的指令\"")
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

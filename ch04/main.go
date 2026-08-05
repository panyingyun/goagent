package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
)

func main() {
	var (
		mode      = flag.String("mode", "react", "运行模式: react(文本 ReAct)或 tool(原生函数调用)")
		mcpAddr   = flag.String("mcp-addr", "127.0.0.1:8765", "MCP 服务器地址;传 -mcp-addr= 可关闭")
		skillsDir = flag.String("skills", "skills", "SKILL 文件目录(存放 *.md)")
	)
	flag.Parse()

	llm := NewLLMClient()
	if llm.APIKey == "" {
		fmt.Fprintln(os.Stderr, "缺少 DEEPSEEK_API_KEY 环境变量")
		fmt.Fprintln(os.Stderr, `示例: $env:DEEPSEEK_API_KEY="sk-..."; go run . "12 * 8 + 4 等于多少?"`)
		os.Exit(1)
	}

	reg := NewRegistry()
	registerDefaultTools(reg)

	skills, err := loadSkills(*skillsDir)
	if err != nil {
		fmt.Fprintln(os.Stderr, "警告: 读取技能失败:", err)
	} else if len(skills) == 0 {
		fmt.Fprintln(os.Stderr, "警告: skills 目录为空,未加载技能")
	} else {
		fmt.Fprintf(os.Stderr, "已加载 %d 个 SKILL: %s\n", len(skills), skillNames(skills))
	}

	if *mcpAddr != "" {
		client, err := ConnectMCPClient(context.Background(), *mcpAddr)
		if err != nil {
			fmt.Fprintln(os.Stderr, "MCP 连接失败:", err)
			os.Exit(1)
		}
		defer client.Close()
		registerMCPTool(reg, client)
		fmt.Fprintf(os.Stderr, "已连接 MCP 服务器 %s(工具名: mcp)\n", *mcpAddr)
	}

	input := strings.Join(flag.Args(), " ")
	if input == "" {
		data, _ := io.ReadAll(os.Stdin)
		input = strings.TrimSpace(string(data))
	}
	if input == "" {
		fmt.Fprintln(os.Stderr, "请输入一个问题,例如: go run . \"12 * 8 + 4 等于多少?\"")
		os.Exit(1)
	}

	var answer string
	switch *mode {
	case "react":
		answer, err = runReAct(context.Background(), llm, reg, skills, input)
	case "tool":
		answer, err = runToolCall(context.Background(), llm, reg, skills, input)
	default:
		fmt.Fprintf(os.Stderr, "未知模式: %s(可选 react / tool)\n", *mode)
		os.Exit(1)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "运行失败:", err)
		os.Exit(1)
	}

	fmt.Println("\n=== 最终回答 ===")
	fmt.Println(answer)
}

func skillNames(skills []Skill) string {
	names := make([]string, 0, len(skills))
	for _, s := range skills {
		names = append(names, s.Name)
	}
	return strings.Join(names, ", ")
}

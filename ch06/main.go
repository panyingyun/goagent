// 课时6:多轮对话与记忆
//
// 解决两个问题:
// 1. 聊久了历史太长 -> 超过预算就交给 LLM 压缩成摘要(记忆);
// 2. 跨会话记住用户偏好 -> 持久化到 memory.json。
//
// 运行(已配置 DEEPSEEK_API_KEY):
//
//	go run .
//
// 交互命令: /exit 退出, /reset 清空对话, /remember 内容 记住偏好,
//
//	/mem 查看偏好, /help 帮助
package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
)

const systemPrompt = `你是一个有记忆的中文助手。对话历史里可能包含一段"早期对话摘要",
请把它当作你之前的记忆。回答保持简洁清晰。`

func main() {
	llm := NewLLMClient()
	if llm.APIKey == "" {
		fmt.Fprintln(os.Stderr, "缺少 DEEPSEEK_API_KEY 环境变量")
		os.Exit(1)
	}

	prefs, err := LoadPreferences("memory.json")
	if err != nil {
		fmt.Fprintln(os.Stderr, "读取记忆失败:", err)
		os.Exit(1)
	}

	sys := rebuildSystem(prefs)
	h := &History{System: sys}

	ctx := context.Background()
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("对话开始。输入 /exit 退出,/reset 清空对话,/remember 内容 记住偏好,/mem 查看偏好,/help 帮助")

	for {
		fmt.Print("你> ")
		if !scanner.Scan() {
			break
		}
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		switch {
		case line == "/exit":
			return
		case line == "/help":
			fmt.Println("/exit /reset /remember <内容> /mem /help")
			continue
		case line == "/reset":
			h = &History{System: sys}
			fmt.Println("已清空对话记忆。")
			continue
		case line == "/mem":
			if s := prefs.String(); s != "" {
				fmt.Println(s)
			} else {
				fmt.Println("(暂无偏好)")
			}
			continue
		case strings.HasPrefix(line, "/remember "):
			v := strings.TrimSpace(strings.TrimPrefix(line, "/remember "))
			if v == "" {
				fmt.Println("用法: /remember <内容>")
				continue
			}
			prefs.Remember("pref", v)
			_ = prefs.Save()
			h.System = rebuildSystem(prefs) // 立即生效,不用等下次启动
			fmt.Println("记住了:", v)
			continue
		}

		// 顺带支持"记住..."的自然语言句式
		if i := strings.Index(line, "记住"); i >= 0 {
			if v := strings.TrimSpace(line[i+len("记住"):]); v != "" {
				prefs.Remember("pref", v)
				_ = prefs.Save()
				h.System = rebuildSystem(prefs) // 立即生效
				fmt.Println("已自动记住:", v)
			}
		}

		h.Add(Message{Role: RoleUser, Content: line})
		if err := h.CompressIfNeeded(ctx, llm); err != nil {
			fmt.Fprintln(os.Stderr, "压缩失败(继续对话):", err)
		}
		reply, err := llm.Chat(ctx, h.SystemMessages())
		if err != nil {
			fmt.Fprintln(os.Stderr, "调用失败:", err)
			continue
		}
		fmt.Println("Agent>", reply.Content)
		h.Add(reply)
	}
}

// rebuildSystem 把当前偏好拼进系统提示词。
func rebuildSystem(prefs *Preferences) string {
	sys := systemPrompt
	if s := prefs.String(); s != "" {
		sys += "\n\n【用户偏好】\n" + s
	}
	return sys
}

// 课时13:模型路由与成本优化
//
// 演示:不同任务选不同模型(flash/pro),带预算控制;相同问题命中缓存,不再花钱。
//
// 运行(已配置 DEEPSEEK_API_KEY):
//
//	go run . "1+1 等于多少?"
//	go run . "帮我分析一下 Go 和 Rust 的优缺点"
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

func main() {
	apiKey := os.Getenv("DEEPSEEK_API_KEY")
	if apiKey == "" {
		fmt.Fprintln(os.Stderr, "缺少 DEEPSEEK_API_KEY")
		os.Exit(1)
	}
	question := strings.Join(os.Args[1:], " ")
	if question == "" {
		data, _ := io.ReadAll(os.Stdin)
		question = strings.TrimSpace(string(data))
	}
	if question == "" {
		fmt.Fprintln(os.Stderr, "用法: go run . \"问题\"")
		os.Exit(1)
	}

	router := NewRouter()
	// 第一次提问
	askAndPrint(router, apiKey, question)
	// 第二次相同问题 -> 缓存命中
	askAndPrint(router, apiKey, question)

	fmt.Printf("\n本次会话花费估算: $%.4f\n", router.Spent())
}

func askAndPrint(router *Router, apiKey, question string) {
	if ans, ok := router.cache.Get(question); ok {
		fmt.Println(">>> 缓存命中,不再调用模型")
		fmt.Println(ans)
		return
	}
	model := router.Choose(question)
	fmt.Printf(">>> 选择模型: %s(复杂度: %s)\n", model.Name, model.Capability)
	answer, err := chatWithModel(apiKey, model.Name, question)
	if err != nil {
		fmt.Fprintln(os.Stderr, "调用失败:", err)
		return
	}
	// 粗略 token 估算:中文 1 字约 1 token
	router.Record(model, len([]rune(question)), len([]rune(answer)))
	router.cache.Set(question, answer)
	fmt.Println(answer)
}

func chatWithModel(apiKey, model, question string) (string, error) {
	body, _ := json.Marshal(map[string]any{
		"model": model,
		"messages": []any{
			map[string]any{"role": "system", "content": "回答简洁。"},
			map[string]any{"role": "user", "content": question},
		},
		"stream": false,
	})
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost,
		"https://api.deepseek.com/chat/completions", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)
	resp, err := (&http.Client{Timeout: 120 * time.Second}).Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API 返回 %d: %s", resp.StatusCode, string(data))
	}
	var out struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(data, &out); err != nil {
		return "", err
	}
	if len(out.Choices) == 0 {
		return "", fmt.Errorf("没有 choices")
	}
	return out.Choices[0].Message.Content, nil
}

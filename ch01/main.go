// 课时1:Go 调用 LLM(DeepSeek)
//
// LLM 本质上是一个"只会说话"的文本生成器:给它消息,它返回续写。
// DeepSeek 提供 OpenAI 兼容的 Chat Completions API,一个 POST 请求就够了,
// 不需要任何第三方 SDK。
//
// 运行(已配置好 DEEPSEEK_API_KEY):
//   go run . "用一句话介绍你自己"
//   go run . "12 * 8 + 4 等于多少?"
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

// 截至 2026-08,DeepSeek 的模型名是 deepseek-v4-flash / deepseek-v4-pro
// (旧的 deepseek-chat / deepseek-reasoner 已于 2026-07-24 退役)。
const defaultBaseURL = "https://api.deepseek.com"
const defaultModel = "deepseek-v4-flash"

// Message 是对话中的一条消息。role 决定这句话是谁说的:
// system(系统设定)、user(用户)、assistant(模型)。
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

func main() {
	// 从命令行参数或标准输入读取问题
	question := strings.Join(os.Args[1:], " ")
	if question == "" {
		data, _ := io.ReadAll(os.Stdin)
		question = strings.TrimSpace(string(data))
	}
	if question == "" {
		fmt.Fprintln(os.Stderr, "用法: go run . \"你的问题\"")
		os.Exit(1)
	}

	apiKey := os.Getenv("DEEPSEEK_API_KEY")
	if apiKey == "" {
		fmt.Fprintln(os.Stderr, "缺少 DEEPSEEK_API_KEY 环境变量")
		os.Exit(1)
	}

	// 对话历史:system 设定角色,user 提问
	messages := []Message{
		{Role: "system", Content: "你是一个乐于助人的中文助手,回答简洁清晰。"},
		{Role: "user", Content: question},
	}

	reply, err := chat(apiKey, messages)
	if err != nil {
		fmt.Fprintln(os.Stderr, "调用失败:", err)
		os.Exit(1)
	}
	fmt.Println("模型回复:")
	fmt.Println(reply)
}

// chat 把消息数组 POST 给 DeepSeek,返回模型回复的文本。
func chat(apiKey string, messages []Message) (string, error) {
	body, err := json.Marshal(map[string]any{
		"model":    envOr("DEEPSEEK_MODEL", defaultModel),
		"messages": messages,
		"stream":   false, // 非流式,一次拿完整回复
	})
	if err != nil {
		return "", err
	}

	url := envOr("DEEPSEEK_BASE_URL", defaultBaseURL) + "/chat/completions"
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{Timeout: 120 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("请求失败(请检查网络): %w", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API 返回 %d: %s", resp.StatusCode, string(data))
	}

	// 解析响应:choices[0].message.content 就是模型说的话
	var out struct {
		Choices []struct {
			Message Message `json:"message"`
		} `json:"choices"`
		Error *struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(data, &out); err != nil {
		return "", fmt.Errorf("解析响应失败: %w", err)
	}
	if out.Error != nil {
		return "", fmt.Errorf("API 错误: %s", out.Error.Message)
	}
	if len(out.Choices) == 0 {
		return "", fmt.Errorf("响应中没有 choices")
	}
	fmt.Println("out.Choices = ", out.Choices)
	return out.Choices[0].Message.Content, nil
}

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

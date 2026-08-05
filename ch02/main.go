// 课时2:提示词工程
//
// 同样的模型,提示词决定上限。本课演示三个最实用的技巧:
// 1. 系统提示词:角色 + 规则
// 2. few-shot:给几个例子,让模型"照葫芦画瓢"
// 3. 结构化输出:要求 JSON,并做解析校验
//
// 运行(已配置 DEEPSEEK_API_KEY):
//
//	go run . "这家店的服务很一般"
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

const defaultBaseURL = "https://api.deepseek.com"

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// 技巧1:角色 + 规则
const rolePrompt = `你是一位资深情感分析师。规则:
1. 只判断"正面/负面/中性"三种结果
2. 先给结论,再给一句理由
3. 理由不超过 20 字`

func main() {
	apiKey := os.Getenv("DEEPSEEK_API_KEY")
	if apiKey == "" {
		fmt.Fprintln(os.Stderr, "缺少 DEEPSEEK_API_KEY 环境变量")
		os.Exit(1)
	}
	question := strings.Join(os.Args[1:], " ")
	if question == "" {
		data, _ := io.ReadAll(os.Stdin)
		question = strings.TrimSpace(string(data))
	}
	if question == "" {
		fmt.Fprintln(os.Stderr, "用法: go run . \"要分析的文本\"")
		os.Exit(1)
	}

	fmt.Println("=== 技巧1:系统提示词(角色+规则) ===")
	r1, err := chat(apiKey, []Message{{Role: "system", Content: rolePrompt}, {Role: "user", Content: question}})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	} else {
		fmt.Println(r1)
	}

	fmt.Println("\n=== 技巧2:few-shot(给例子) ===")
	examples := []string{
		"这家店超好吃 -> 正面",
		"太难吃了,再也不来 -> 负面",
		"一般般,没什么特别的 -> 中性",
	}
	r2, err := chat(apiKey, []Message{{Role: "user", Content: buildFewShot("判断评论情感,只输出 正面/负面/中性", examples, question)}})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	} else {
		fmt.Println(r2)
	}

	fmt.Println("\n=== 技巧3:结构化输出(JSON + 校验) ===")
	r3, err := chat(apiKey, []Message{{
		Role: "user",
		Content: "判断下面评论的情感,只输出 JSON:\n" +
			`{"sentiment": "正面|负面|中性", "confidence": 0~1, "reason": "一句话理由"}` + "\n评论: " + question,
	}})

	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	} else {
		fmt.Println("原始回复:", r3)
		parsed, err := extractJSON(r3)
		if err != nil {
			fmt.Fprintln(os.Stderr, "JSON 解析失败:", err)
		} else if err := validateResult(parsed); err != nil {
			fmt.Fprintln(os.Stderr, "校验失败:", err)
		} else {
			fmt.Printf("解析结果: 情感=%v 置信度=%v 理由=%v\n", parsed["sentiment"], parsed["confidence"], parsed["reason"])
		}
	}
}

// buildFewShot 把"指令 + 例子 + 问题"拼成一个提示词。
func buildFewShot(instruction string, examples []string, question string) string {
	var b strings.Builder
	b.WriteString(instruction + "\n\n")
	for _, e := range examples {
		b.WriteString(e + "\n")
	}
	b.WriteString("\n" + question)
	return b.String()
}

// extractJSON 从回复里抽出第一个完整的 JSON 对象(兼容代码块包裹)。
func extractJSON(s string) (map[string]any, error) {
	start := strings.IndexByte(s, '{')
	if start < 0 {
		return nil, errors.New("没有找到 {")
	}
	depth := 0
	for i := start; i < len(s); i++ {
		switch s[i] {
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				var m map[string]any
				if err := json.Unmarshal([]byte(s[start:i+1]), &m); err != nil {
					return nil, err
				}
				return m, nil
			}
		}
	}
	return nil, errors.New("JSON 括号不匹配")
}

func validateResult(m map[string]any) error {
	s, _ := m["sentiment"].(string)
	if s != "正面" && s != "负面" && s != "中性" {
		return fmt.Errorf("sentiment 必须是 正面/负面/中性,得到 %q", s)
	}
	if _, ok := m["confidence"].(float64); !ok {
		return errors.New("confidence 必须是数字")
	}
	return nil
}

func chat(apiKey string, messages []Message) (string, error) {
	body, _ := json.Marshal(map[string]any{
		"model":    envOr("DEEPSEEK_MODEL", "deepseek-v4-flash"),
		"messages": messages,
		"stream":   false,
	})
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost,
		envOr("DEEPSEEK_BASE_URL", defaultBaseURL)+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)
	client := &http.Client{Timeout: 120 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API 返回 %d: %s", resp.StatusCode, string(data))
	}
	var out struct {
		Choices []struct {
			Message Message `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(data, &out); err != nil {
		return "", err
	}
	if len(out.Choices) == 0 {
		return "", errors.New("响应中没有 choices")
	}
	return out.Choices[0].Message.Content, nil
}

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

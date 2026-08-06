// 课时5:Go 实现 SKILL 调用
//
// SKILL(技能)= 打包成文件的行为说明书。本质上是"系统提示词管理":
// 启动时把 skills/*.md 读进来,拼进 system 消息,模型就会遵守其中的规范。
//
// 运行(已配置好 DEEPSEEK_API_KEY):
//
//	go run . "12 * 8 + 4 等于多少?"
//	go run . "把 Hello world 翻译成中文"
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const defaultBaseURL = "https://api.deepseek.com"
const defaultModel = "deepseek-v4-flash"

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// Skill 是一个技能:名字 + 一段行为规范。
type Skill struct {
	Name         string
	Instructions string
}

func main() {
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

	// 1. 加载 skills/ 目录下的所有技能
	skills, err := loadSkills("skills")

	//fmt.Printf("[YYPAN]  skills = [[[\n   %v   \n]]]\n", skills)

	if err != nil {
		fmt.Fprintln(os.Stderr, "加载技能失败:", err)
		os.Exit(1)
	}
	fmt.Fprintf(os.Stderr, "已加载 %d 个 SKILL:\n", len(skills))
	for _, s := range skills {
		fmt.Fprintf(os.Stderr, " - %s\n", s.Name)
	}

	// 2. 把技能拼进 system 消息
	system := "你是一个乐于助人的中文助手。\n\n## 你掌握的技能(SKILL)\n"
	for _, s := range skills {
		system += fmt.Sprintf("\n### 技能: %s\n%s\n", s.Name, s.Instructions)
	}

	// 3. 正常对话
	reply, err := chat(apiKey, []Message{
		{Role: "system", Content: system},
		{Role: "user", Content: question},
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, "调用失败:", err)
		os.Exit(1)
	}
	fmt.Println("模型回复:")
	fmt.Println(reply)
}

// loadSkills 读取 skills/*.md,约定第一行是 "# 技能名",其余是技能内容。
func loadSkills(dir string) ([]Skill, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var skills []Skill
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			return nil, err
		}
		text := strings.TrimSpace(string(data))
		if text == "" {
			continue
		}
		lines := strings.SplitN(text, "\n", 2)
		name := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(lines[0]), "#"))
		instructions := ""
		if len(lines) == 2 {
			instructions = strings.TrimSpace(lines[1])
		}
		skills = append(skills, Skill{Name: name, Instructions: instructions})
	}
	return skills, nil
}

// chat 与课时1 完全相同,只负责发请求、拿回复。
func chat(apiKey string, messages []Message) (string, error) {
	body, err := json.Marshal(map[string]any{
		"model":    envOr("DEEPSEEK_MODEL", defaultModel),
		"messages": messages,
		"stream":   false,
	})
	if err != nil {
		return "", err
	}
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
	return out.Choices[0].Message.Content, nil
}

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

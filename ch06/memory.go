package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// summaryBudget 是历史消息的估算 token 预算:超过就触发压缩。
const summaryBudget = 3000

// History 管理对话历史:截断 + 摘要压缩,解决"聊久了爆上下文"的问题。
type History struct {
	System   string
	Messages []Message // 只存 user/assistant 对话
	Summary  string    // 被压缩掉的旧对话摘要
}

func (h *History) Add(m Message) {
	h.Messages = append(h.Messages, m)
}

// SystemMessages 返回真正发给 LLM 的消息:system + 摘要 + 窗口内历史。
func (h *History) SystemMessages() []Message {
	msgs := []Message{{Role: RoleSystem, Content: h.System}}
	if h.Summary != "" {
		msgs = append(msgs, Message{Role: RoleSystem, Content: "【早期对话摘要】" + h.Summary})
	}
	msgs = append(msgs, h.Messages...)
	return msgs
}

func (h *History) estimatedTokens() int {
	total := 0
	for _, m := range h.Messages {
		total += tokenEstimate(m.Content)
	}
	return total
}

// CompressIfNeeded 当历史超过预算时,把前半段交给 LLM 压成摘要。
// 压缩失败不会中断对话(把旧消息放回去继续聊)。
func (h *History) CompressIfNeeded(ctx context.Context, llm *LLMClient) error {
	if h.estimatedTokens() <= summaryBudget {
		return nil
	}
	split := len(h.Messages) / 2
	old := h.Messages[:split]
	h.Messages = h.Messages[split:]

	var b strings.Builder
	for _, m := range old {
		b.WriteString(string(m.Role) + ": " + m.Content + "\n")
	}
	resp, err := llm.Chat(ctx, []Message{
		{Role: RoleSystem, Content: "你是一个记忆压缩器。把下面的对话压缩成一段不超过 50 字的中文摘要,保留关键事实、用户偏好和未完成的任务。"},
		{Role: RoleUser, Content: b.String()},
	})
	if err != nil {
		h.Messages = append(old, h.Messages...)
		return err
	}
	if h.Summary != "" {
		h.Summary = h.Summary + ";" + resp.Content
	} else {
		h.Summary = resp.Content
	}
	return nil
}

// tokenEstimate 粗略估算 token 数:中文 1 字约 1 token,ASCII 约 0.3,加 20% 余量。
// 只是估算,用于决定何时压缩,不需要精确。
func tokenEstimate(s string) int {
	if s == "" {
		return 0
	}
	var n float64
	for _, r := range s {
		if r > 0x7F {
			n++
		} else {
			n += 0.3
		}
	}
	return int(n*1.2) + 1
}

// Preferences 是最简的"长期记忆":键值对持久化到 JSON 文件。
type Preferences struct {
	path string
	data map[string]string
}

func LoadPreferences(path string) (*Preferences, error) {
	p := &Preferences{path: path, data: map[string]string{}}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return p, nil
		}
		return nil, err
	}
	_ = json.Unmarshal(data, &p.data)
	return p, nil
}

func (p *Preferences) Remember(key, value string) {
	p.data[key] = value
}

func (p *Preferences) Save() error {
	data, err := json.MarshalIndent(p.data, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(p.path, data, 0o644)
}

func (p *Preferences) String() string {
	var b strings.Builder
	for k, v := range p.data {
		b.WriteString(fmt.Sprintf("- %s: %s\n", k, v))
	}
	return strings.TrimSpace(b.String())
}

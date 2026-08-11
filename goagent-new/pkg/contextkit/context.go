package contextkit

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"unicode/utf8"
)

// Trust 标记上下文来源是否可信(影响注入与渲染策略)。
type Trust string

const (
	Trusted   Trust = "trusted"
	Untrusted Trust = "untrusted"
)

// Source 是一条待组装进提示词的上下文来源。
type Source struct {
	ID       string `json:"id"`
	Kind     string `json:"kind"`
	Content  string `json:"content"`
	Trust    Trust  `json:"trust"`
	Priority int    `json:"priority"`
	Required bool   `json:"required"`
}

// Item 是选中后的上下文条目(含相关度与渲染文本)。
type Item struct {
	Source    Source `json:"source"`
	Relevance int    `json:"relevance"`
	Rendered  string `json:"rendered"`
}

// Bundle 是按预算裁剪后的上下文组装结果。
type Bundle struct {
	Text        string   `json:"text"`
	Items       []Item   `json:"items"`
	DroppedIDs  []string `json:"dropped_ids"`
	UsedRunes   int      `json:"used_runes"`
	BudgetRunes int      `json:"budget_runes"`
}

// Assembler 按相关度与预算从 Source 列表组装 Bundle。
type Assembler struct {
	MaxRunes int
	MaxItems int
}

var wordPattern = regexp.MustCompile(`[\p{L}\p{N}_]+`)

func (a Assembler) Build(query string, sources []Source) (Bundle, error) {
	if a.MaxRunes <= 0 || a.MaxItems <= 0 {
		return Bundle{}, errors.New("positive context rune and item budgets are required")
	}
	seen := make(map[string]bool, len(sources))
	queryWords := words(query)
	candidates := make([]Item, 0, len(sources))
	for _, source := range sources {
		if strings.TrimSpace(source.ID) == "" || strings.TrimSpace(source.Kind) == "" {
			return Bundle{}, errors.New("source id and kind are required")
		}
		if seen[source.ID] {
			return Bundle{}, fmt.Errorf("duplicate context source %q", source.ID)
		}
		seen[source.ID] = true
		if source.Trust == "" {
			source.Trust = Untrusted
		}
		if source.Trust != Trusted && source.Trust != Untrusted {
			return Bundle{}, fmt.Errorf("unsupported trust level %q", source.Trust)
		}
		item := Item{Source: source, Relevance: relevance(queryWords, words(source.Content))}
		item.Rendered = render(source)
		candidates = append(candidates, item)
	}
	sort.SliceStable(candidates, func(i, j int) bool {
		left, right := candidates[i], candidates[j]
		if left.Source.Required != right.Source.Required {
			return left.Source.Required
		}
		if left.Relevance != right.Relevance {
			return left.Relevance > right.Relevance
		}
		if left.Source.Priority != right.Source.Priority {
			return left.Source.Priority > right.Source.Priority
		}
		return left.Source.ID < right.Source.ID
	})

	bundle := Bundle{Items: []Item{}, DroppedIDs: []string{}, BudgetRunes: a.MaxRunes}
	parts := make([]string, 0, len(candidates))
	for _, candidate := range candidates {
		runes := utf8.RuneCountInString(candidate.Rendered)
		separatorRunes := 0
		if len(parts) > 0 {
			separatorRunes = 1
		}
		fits := len(bundle.Items) < a.MaxItems && bundle.UsedRunes+separatorRunes+runes <= a.MaxRunes
		if !fits {
			if candidate.Source.Required {
				return Bundle{}, fmt.Errorf("required source %q exceeds the context budget", candidate.Source.ID)
			}
			bundle.DroppedIDs = append(bundle.DroppedIDs, candidate.Source.ID)
			continue
		}
		bundle.Items = append(bundle.Items, candidate)
		bundle.UsedRunes += separatorRunes + runes
		parts = append(parts, candidate.Rendered)
	}
	bundle.Text = strings.Join(parts, "\n")
	return bundle, nil
}

func render(source Source) string {
	content, _ := json.Marshal(source.Content)
	return fmt.Sprintf("<SOURCE id=%q kind=%q trust=%q>\nCONTENT_JSON: %s\n</SOURCE>", source.ID, source.Kind, source.Trust, content)
}

func words(text string) map[string]bool {
	result := make(map[string]bool)
	for _, word := range wordPattern.FindAllString(strings.ToLower(text), -1) {
		result[word] = true
	}
	return result
}

func relevance(query, content map[string]bool) int {
	score := 0
	for word := range query {
		if content[word] {
			score++
		}
	}
	return score
}

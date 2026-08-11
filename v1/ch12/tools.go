package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	_ "time/tzdata"
)

type Tool struct {
	Name        string
	Description string
	Run         func(ctx context.Context, args map[string]any) (string, error)
}

type Registry struct {
	tools map[string]Tool
}

func NewRegistry() *Registry {
	return &Registry{tools: map[string]Tool{}}
}

func (r *Registry) Register(t Tool) {
	r.tools[t.Name] = t
}

func (r *Registry) Names() []string {
	names := make([]string, 0, len(r.tools))
	for n := range r.tools {
		names = append(names, n)
	}
	sort.Strings(names)
	return names
}

func (r *Registry) Run(ctx context.Context, name, inputJSON string) (string, error) {
	t, ok := r.tools[name]
	if !ok {
		return "", fmt.Errorf("未知工具: %s(可用: %s)", name, strings.Join(r.Names(), ", "))
	}
	var args map[string]any
	if strings.TrimSpace(inputJSON) != "" {
		if err := json.Unmarshal([]byte(inputJSON), &args); err != nil {
			return "", fmt.Errorf("Action Input 不是合法 JSON: %w", err)
		}
	}
	return t.Run(ctx, args)
}

func registerDefaultTools(r *Registry) {
	r.Register(Tool{
		Name:        "calculator",
		Description: "计算数学表达式,支持 + - * / 和括号。示例: (12 + 4) * 8",
		Run: func(_ context.Context, args map[string]any) (string, error) {
			expr, _ := args["expression"].(string)
			if strings.TrimSpace(expr) == "" {
				return "", errors.New("缺少 expression 参数")
			}
			v, err := evalCalc(expr)
			if err != nil {
				return "", err
			}
			return fmt.Sprintf("%v", v), nil
		},
	})
	r.Register(Tool{
		Name:        "get_time",
		Description: "获取当前日期和时间(北京时间 Asia/Shanghai)。",
		Run: func(_ context.Context, _ map[string]any) (string, error) {
			loc, err := time.LoadLocation("Asia/Shanghai")
			if err != nil {
				return "", err
			}
			return time.Now().In(loc).Format("2006-01-02 15:04:05 (Monday)"), nil
		},
	})
}

func truncateStr(s string, n int) string {
	runes := []rune(s)
	if len(runes) <= n {
		return s
	}
	return string(runes[:n]) + "..."
}

// ---- 极简表达式求值器 ----

type calcParser struct {
	s   string
	pos int
}

func evalCalc(expr string) (float64, error) {
	p := &calcParser{s: expr}
	v, err := p.parseExpr()
	if err != nil {
		return 0, err
	}
	p.skipSpace()
	if p.pos != len(p.s) {
		return 0, fmt.Errorf("表达式末尾有无法解析的内容: %q", p.s[p.pos:])
	}
	return v, nil
}

func (p *calcParser) skipSpace() {
	for p.pos < len(p.s) && (p.s[p.pos] == ' ' || p.s[p.pos] == '\t') {
		p.pos++
	}
}

func (p *calcParser) parseExpr() (float64, error) {
	v, err := p.parseTerm()
	if err != nil {
		return 0, err
	}
	for {
		p.skipSpace()
		if p.pos >= len(p.s) {
			return v, nil
		}
		switch p.s[p.pos] {
		case '+':
			p.pos++
			rhs, err := p.parseTerm()
			if err != nil {
				return 0, err
			}
			v += rhs
		case '-':
			p.pos++
			rhs, err := p.parseTerm()
			if err != nil {
				return 0, err
			}
			v -= rhs
		default:
			return v, nil
		}
	}
}

func (p *calcParser) parseTerm() (float64, error) {
	v, err := p.parseFactor()
	if err != nil {
		return 0, err
	}
	for {
		p.skipSpace()
		if p.pos >= len(p.s) {
			return v, nil
		}
		switch p.s[p.pos] {
		case '*':
			p.pos++
			rhs, err := p.parseFactor()
			if err != nil {
				return 0, err
			}
			v *= rhs
		case '/':
			p.pos++
			rhs, err := p.parseFactor()
			if err != nil {
				return 0, err
			}
			if rhs == 0 {
				return 0, errors.New("除数不能为 0")
			}
			v /= rhs
		default:
			return v, nil
		}
	}
}

func (p *calcParser) parseFactor() (float64, error) {
	p.skipSpace()
	if p.pos < len(p.s) && (p.s[p.pos] == '+' || p.s[p.pos] == '-') {
		sign := 1.0
		if p.s[p.pos] == '-' {
			sign = -1
		}
		p.pos++
		v, err := p.parseFactor()
		if err != nil {
			return 0, err
		}
		return sign * v, nil
	}
	return p.parsePrimary()
}

func (p *calcParser) parsePrimary() (float64, error) {
	p.skipSpace()
	if p.pos >= len(p.s) {
		return 0, errors.New("表达式不完整")
	}
	if p.s[p.pos] == '(' {
		p.pos++
		v, err := p.parseExpr()
		if err != nil {
			return 0, err
		}
		p.skipSpace()
		if p.pos >= len(p.s) || p.s[p.pos] != ')' {
			return 0, errors.New("缺少右括号 )")
		}
		p.pos++
		return v, nil
	}
	start := p.pos
	for p.pos < len(p.s) && ((p.s[p.pos] >= '0' && p.s[p.pos] <= '9') || p.s[p.pos] == '.') {
		p.pos++
	}
	if start == p.pos {
		return 0, fmt.Errorf("无法解析的字符: %q", p.s[p.pos])
	}
	raw := p.s[start:p.pos]
	var v float64
	if _, err := fmt.Sscanf(raw, "%g", &v); err != nil {
		return 0, fmt.Errorf("无效数字 %q", raw)
	}
	return v, nil
}

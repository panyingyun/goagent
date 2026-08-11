package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	_ "time/tzdata" // 内嵌时区数据库,保证 Windows 上也能解析 Asia/Shanghai
)

// Tool 是 Agent 可以调用的一个能力:名字、说明、参数 Schema、执行函数。
type Tool struct {
	Name        string
	Description string
	Parameters  map[string]any // JSON Schema,function calling 模式会原样发给模型
	Run         func(ctx context.Context, args map[string]any) (string, error)
}

// Registry 是工具注册表:Agent 与具体工具解耦,
// 新增一个工具只需要 Register 一次。
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

// Run 解析 Action Input 的 JSON,找到工具并执行。
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

// ToolDefs 生成发给模型的工具定义(function calling 模式用)。
func (r *Registry) ToolDefs() []ToolDef {
	defs := make([]ToolDef, 0, len(r.tools))
	for _, n := range r.Names() {
		t := r.tools[n]
		defs = append(defs, ToolDef{
			Type: "function",
			Function: ToolFunction{
				Name:        t.Name,
				Description: t.Description,
				Parameters:  t.Parameters,
			},
		})
	}
	return defs
}

// registerDefaultTools 注册两个演示用内置工具。
func registerDefaultTools(r *Registry) {
	r.Register(Tool{
		Name:        "calculator",
		Description: "计算数学表达式,支持 + - * / 和括号。示例: (12 + 4) * 8",
		Parameters: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"expression": map[string]any{
					"type":        "string",
					"description": "要计算的数学表达式",
				},
			},
			"required": []string{"expression"},
		},
		Run: func(ctx context.Context, args map[string]any) (string, error) {
			expr, _ := args["expression"].(string)
			expr = strings.TrimSpace(expr)
			if expr == "" {
				return "", errors.New("缺少 expression 参数")
			}
			v, err := evalCalc(expr)
			if err != nil {
				return "", err
			}
			return strconv.FormatFloat(v, 'f', -1, 64), nil
		},
	})

	r.Register(Tool{
		Name:        "get_time",
		Description: "获取当前日期和时间(北京时间 Asia/Shanghai)。",
		Parameters: map[string]any{
			"type":       "object",
			"properties": map[string]any{},
		},
		Run: func(ctx context.Context, args map[string]any) (string, error) {
			loc, err := time.LoadLocation("Asia/Shanghai")
			if err != nil {
				return "", err
			}
			return time.Now().In(loc).Format("2006-01-02 15:04:05 (Monday)"), nil
		},
	})
}

// ---- 一个极简的表达式求值器(递归下降解析) ----
// 只为演示"工具内部也可以有逻辑",生产环境可以直接用第三方表达式库。

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
	v, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return 0, fmt.Errorf("无效数字 %q: %w", raw, err)
	}
	return v, nil
}

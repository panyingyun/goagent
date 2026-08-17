package agent

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode"

	_ "time/tzdata"
)

// NewCalculatorTool 返回仅支持四则运算与括号的计算器工具。
func NewCalculatorTool() Tool {
	return Tool{
		Name:         "calculator",
		Description:  "计算只包含数字、括号和 + - * / 的表达式。",
		Parameters:   objectSchema(map[string]any{"expression": map[string]any{"type": "string"}}, "expression"),
		SideEffect:   SideEffectNone,
		Cacheable:    true,
		Idempotent:   true,
		ParallelSafe: true,
		Run: func(_ context.Context, raw json.RawMessage) (string, error) {
			var in struct {
				Expression string `json:"expression"`
			}
			if err := decodeStrict(raw, &in); err != nil {
				return "", err
			}
			if strings.TrimSpace(in.Expression) == "" {
				return "", errors.New("expression is required")
			}
			value, err := evalExpression(in.Expression)
			if err != nil {
				return "", err
			}
			return strconv.FormatFloat(value, 'f', -1, 64), nil
		},
	}
}

// NewGetTimeTool 返回获取北京时间的工具。
func NewGetTimeTool() Tool {
	return Tool{
		Name:         "get_time",
		Description:  "获取当前日期和时间(北京时间 Asia/Shanghai)。",
		Parameters:   objectSchema(nil),
		SideEffect:   SideEffectNone,
		Cacheable:    false,
		Idempotent:   true,
		ParallelSafe: true,
		Timeout:      2 * time.Second,
		Run: func(ctx context.Context, _ json.RawMessage) (string, error) {
			if err := ctx.Err(); err != nil {
				return "", err
			}
			loc, err := time.LoadLocation("Asia/Shanghai")
			if err != nil {
				return "", err
			}
			return time.Now().In(loc).Format("2006-01-02 15:04:05 (Monday)"), nil
		},
	}
}

// NewListFilesTool 返回列出 root 下 .txt 文件的工具。
func NewListFilesTool(root string) Tool {
	return Tool{
		Name:         "list_documents",
		Description:  "列出工作区根目录中的文本文件。",
		Parameters:   objectSchema(nil),
		SideEffect:   SideEffectRead,
		Cacheable:    false,
		Idempotent:   true,
		ParallelSafe: true,
		Run: func(ctx context.Context, _ json.RawMessage) (string, error) {
			if err := ctx.Err(); err != nil {
				return "", err
			}
			entries, err := os.ReadDir(root)
			if err != nil {
				return "", err
			}
			names := make([]string, 0, len(entries))
			for _, entry := range entries {
				if !entry.IsDir() && strings.EqualFold(filepath.Ext(entry.Name()), ".txt") {
					names = append(names, entry.Name())
				}
			}
			sort.Strings(names)
			b, _ := json.Marshal(names)
			return string(b), nil
		},
	}
}

// NewReadFileTool 返回读取 root 下 .txt 文件的工具（最多 64 KiB）。
func NewReadFileTool(root string) Tool {
	return Tool{
		Name:         "read_document",
		Description:  "读取工作区中的一个 .txt 文档，最多返回 64 KiB。",
		Parameters:   objectSchema(map[string]any{"path": map[string]any{"type": "string"}}, "path"),
		SideEffect:   SideEffectRead,
		Cacheable:    true,
		Idempotent:   true,
		ParallelSafe: true,
		Run: func(ctx context.Context, raw json.RawMessage) (string, error) {
			var in struct {
				Path string `json:"path"`
			}
			if err := decodeStrict(raw, &in); err != nil {
				return "", err
			}
			path, err := existingSafePath(root, in.Path)
			if err != nil {
				return "", err
			}
			if !strings.EqualFold(filepath.Ext(path), ".txt") {
				return "", errors.New("only .txt documents are allowed")
			}
			file, err := os.Open(path)
			if err != nil {
				return "", err
			}
			defer file.Close()
			data, err := io.ReadAll(io.LimitReader(file, (64<<10)+1))
			if err != nil {
				return "", err
			}
			if len(data) > 64<<10 {
				return "", errors.New("document exceeds 64 KiB limit")
			}
			if err := ctx.Err(); err != nil {
				return "", err
			}
			return string(data), nil
		},
	}
}

// NewWriteFileTool 返回需审批后写入 root 下 .txt 笔记的工具。
func NewWriteFileTool(root string) Tool {
	return Tool{
		Name:             "write_note",
		Description:      "经用户批准后，在工作区中创建或覆盖一个 .txt 笔记。",
		Parameters:       objectSchema(map[string]any{"path": map[string]any{"type": "string"}, "content": map[string]any{"type": "string"}}, "path", "content"),
		SideEffect:       SideEffectWrite,
		Idempotent:       true,
		RequiresApproval: true,
		Run: func(ctx context.Context, raw json.RawMessage) (string, error) {
			var in struct {
				Path    string `json:"path"`
				Content string `json:"content"`
			}
			if err := decodeStrict(raw, &in); err != nil {
				return "", err
			}
			if len(in.Content) > 64<<10 {
				return "", errors.New("content exceeds 64 KiB limit")
			}
			path, err := newSafePath(root, in.Path)
			if err != nil {
				return "", err
			}
			if !strings.EqualFold(filepath.Ext(path), ".txt") {
				return "", errors.New("only .txt notes are allowed")
			}
			if err := ctx.Err(); err != nil {
				return "", err
			}
			if err := os.WriteFile(path, []byte(in.Content), 0o600); err != nil {
				return "", fmt.Errorf("write %s: %w", path, err)
			}
			return "saved " + filepath.Base(path), nil
		},
	}
}

func objectSchema(properties map[string]any, required ...string) map[string]any {
	if properties == nil {
		properties = map[string]any{}
	}
	// 无必填字段时也要发 []，不能发 null；部分供应商会拒绝 required:null。
	if required == nil {
		required = []string{}
	}
	return map[string]any{
		"type":                 "object",
		"properties":           properties,
		"required":             required,
		"additionalProperties": false,
	}
}

func decodeStrict(raw json.RawMessage, dst any) error {
	dec := json.NewDecoder(strings.NewReader(string(raw)))
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		return fmt.Errorf("decode arguments: %w", err)
	}
	var extra any
	if err := dec.Decode(&extra); !errors.Is(err, io.EOF) {
		return errors.New("arguments contain multiple JSON values")
	}
	return nil
}

func existingSafePath(root, name string) (string, error) {
	base, candidate, err := lexicalSafePath(root, name)
	if err != nil {
		return "", err
	}
	realBase, err := filepath.EvalSymlinks(base)
	if err != nil {
		return "", fmt.Errorf("resolve workspace root %s: %w", base, err)
	}
	realCandidate, err := filepath.EvalSymlinks(candidate)
	if err != nil {
		return "", fmt.Errorf("resolve document path %s: %w", candidate, err)
	}
	if !within(realBase, realCandidate) {
		return "", errors.New("path escapes workspace through a link")
	}
	return realCandidate, nil
}

func newSafePath(root, name string) (string, error) {
	base, candidate, err := lexicalSafePath(root, name)
	if err != nil {
		return "", err
	}
	realBase, err := filepath.EvalSymlinks(base)
	if err != nil {
		return "", fmt.Errorf("resolve workspace root %s: %w", base, err)
	}
	realParent, err := filepath.EvalSymlinks(filepath.Dir(candidate))
	if err != nil {
		return "", fmt.Errorf("resolve document parent %s: %w", filepath.Dir(candidate), err)
	}
	if !within(realBase, realParent) {
		return "", errors.New("path escapes workspace through a link")
	}
	return filepath.Join(realParent, filepath.Base(candidate)), nil
}

func lexicalSafePath(root, name string) (string, string, error) {
	if strings.TrimSpace(name) == "" || filepath.IsAbs(name) {
		return "", "", errors.New("path must be a non-empty relative path")
	}
	base, err := filepath.Abs(root)
	if err != nil {
		return "", "", err
	}
	candidate := filepath.Clean(filepath.Join(base, name))
	if !within(base, candidate) {
		return "", "", errors.New("path escapes workspace")
	}
	return base, candidate, nil
}

func within(base, target string) bool {
	rel, err := filepath.Rel(base, target)
	return err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) && !filepath.IsAbs(rel)
}

// expressionParser 是 calculator 工具的简易算术表达式解析器。
type expressionParser struct {
	s   string
	pos int
}

func evalExpression(input string) (float64, error) {
	p := &expressionParser{s: input}
	value, err := p.parseExpr()
	if err != nil {
		return 0, err
	}
	p.space()
	if p.pos != len(p.s) {
		return 0, fmt.Errorf("unexpected input at %q", p.s[p.pos:])
	}
	return value, nil
}

func (p *expressionParser) space() {
	for p.pos < len(p.s) && unicode.IsSpace(rune(p.s[p.pos])) {
		p.pos++
	}
}

func (p *expressionParser) parseExpr() (float64, error) {
	left, err := p.parseTerm()
	if err != nil {
		return 0, err
	}
	for {
		p.space()
		if p.pos >= len(p.s) || (p.s[p.pos] != '+' && p.s[p.pos] != '-') {
			return left, nil
		}
		op := p.s[p.pos]
		p.pos++
		right, err := p.parseTerm()
		if err != nil {
			return 0, err
		}
		if op == '+' {
			left += right
		} else {
			left -= right
		}
	}
}

func (p *expressionParser) parseTerm() (float64, error) {
	left, err := p.parseFactor()
	if err != nil {
		return 0, err
	}
	for {
		p.space()
		if p.pos >= len(p.s) || (p.s[p.pos] != '*' && p.s[p.pos] != '/') {
			return left, nil
		}
		op := p.s[p.pos]
		p.pos++
		right, err := p.parseFactor()
		if err != nil {
			return 0, err
		}
		if op == '*' {
			left *= right
		} else {
			if right == 0 {
				return 0, errors.New("division by zero")
			}
			left /= right
		}
	}
}

func (p *expressionParser) parseFactor() (float64, error) {
	p.space()
	if p.pos >= len(p.s) {
		return 0, errors.New("unexpected end of expression")
	}
	if p.s[p.pos] == '(' {
		p.pos++
		value, err := p.parseExpr()
		if err != nil {
			return 0, err
		}
		p.space()
		if p.pos >= len(p.s) || p.s[p.pos] != ')' {
			return 0, errors.New("missing closing parenthesis")
		}
		p.pos++
		return value, nil
	}
	start := p.pos
	if p.s[p.pos] == '+' || p.s[p.pos] == '-' {
		p.pos++
	}
	for p.pos < len(p.s) && ((p.s[p.pos] >= '0' && p.s[p.pos] <= '9') || p.s[p.pos] == '.') {
		p.pos++
	}
	if start == p.pos {
		return 0, fmt.Errorf("expected number at %q", p.s[p.pos:])
	}
	value, err := strconv.ParseFloat(p.s[start:p.pos], 64)
	if err != nil {
		return 0, fmt.Errorf("invalid number: %w", err)
	}
	return value, nil
}

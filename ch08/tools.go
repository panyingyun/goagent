package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"

	_ "time/tzdata"
)

// Tool 与课时3 相同:Agent 可调用的能力,注册进 Registry。
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

// registerDefaultTools 注册内置工具:计算器、时钟 + 本课时的"真实工具"。
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
	r.Register(Tool{
		Name:        "http_get",
		Description: "请求一个 http/https URL,返回响应正文(最多 2000 字符)。",
		Run: func(ctx context.Context, args map[string]any) (string, error) {
			raw, _ := args["url"].(string)
			u, err := url.Parse(raw)
			if err != nil || (u.Scheme != "http" && u.Scheme != "https") {
				return "", errors.New("url 必须是 http/https 地址")
			}
			ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
			defer cancel()
			req, _ := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				return "", err
			}
			defer resp.Body.Close()
			body, _ := ioReadAllLimited(resp.Body, 2000)
			return fmt.Sprintf("状态码: %d\n%s", resp.StatusCode, body), nil
		},
	})
	r.Register(Tool{
		Name:        "read_file",
		Description: "读取 docs 目录下的文本文件,返回内容。",
		Run: func(_ context.Context, args map[string]any) (string, error) {
			fmt.Println("[YYPAN] ========== read_file ==========")
			name, _ := args["filename"].(string)
			path, err := safePath(docRoot, name)
			if err != nil {
				return "", err
			}
			data, err := os.ReadFile(path)
			if err != nil {
				return "", err
			}
			return truncateStr(string(data), 2000), nil
		},
	})
	r.Register(Tool{
		Name:        "write_file",
		Description: "把内容写入 docs 目录下的文件(覆盖)。",
		Run: func(_ context.Context, args map[string]any) (string, error) {
			fmt.Println("[YYPAN] ========== write_file ========== ")
			name, _ := args["filename"].(string)
			content, _ := args["content"].(string)
			path, err := safePath(docRoot, name)
			if err != nil {
				return "", err
			}
			if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
				return "", err
			}
			return "已写入 " + path, nil
		},
	})
	r.Register(Tool{
		Name:        "run_command",
		Description: "执行白名单内的命令,如 go version / echo <文本> / dir / ls",
		Run: func(ctx context.Context, args map[string]any) (string, error) {
			fmt.Println("[YYPAN] ========== run_command ========== ")
			cmd, _ := args["command"].(string)
			cmd = strings.TrimSpace(cmd)
			if !allowedCommand(cmd) {
				return "", fmt.Errorf("命令不在白名单内,拒绝执行: %q", cmd)
			}
			ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()
			var out []byte
			var err error
			if runtime.GOOS == "windows" {
				out, err = exec.CommandContext(ctx, "cmd", "/C", cmd).CombinedOutput()
			} else {
				out, err = exec.CommandContext(ctx, "sh", "-c", cmd).CombinedOutput()
			}
			if err != nil {
				return "", fmt.Errorf("执行失败: %v", err)
			}
			return truncateStr(string(out), 2000), nil
		},
	})
	r.Register(Tool{
		Name:        "kv_get",
		Description: "从本地键值存储读取一个值。",
		Run: func(_ context.Context, args map[string]any) (string, error) {
			key, _ := args["key"].(string)
			return kvStore().get(key)
		},
	})
	r.Register(Tool{
		Name:        "kv_set",
		Description: "向本地键值存储写入 key/value。",
		Run: func(_ context.Context, args map[string]any) (string, error) {
			key, _ := args["key"].(string)
			value, _ := args["value"].(string)
			if key == "" {
				return "", errors.New("缺少 key")
			}
			if err := kvStore().set(key, value); err != nil {
				return "", err
			}
			return "已保存 " + key, nil
		},
	})
}

const docRoot = "docs"

// safePath 把文件名限制在 root 目录内,防止 ../ 路径穿越。
func safePath(root, name string) (string, error) {
	if strings.TrimSpace(name) == "" {
		return "", errors.New("缺少 filename")
	}
	base, err := filepath.Abs(root)
	if err != nil {
		return "", err
	}
	p := filepath.Join(base, name)
	if p != base && !strings.HasPrefix(p, base+string(filepath.Separator)) {
		return "", errors.New("路径越界:只允许访问 docs 目录")
	}
	return p, nil
}

// allowedCommand 命令白名单:只有明确的命令前缀才允许执行。
func allowedCommand(cmd string) bool {
	allowed := []string{"go version", "echo ", "dir", "ls"}
	for _, a := range allowed {
		if strings.HasPrefix(cmd, a) {
			return true
		}
	}
	return false
}

func ioReadAllLimited(r io.Reader, limit int) (string, error) {
	buf := make([]byte, 0, 4096)
	tmp := make([]byte, 1024)
	for len(buf) < limit {
		n, err := r.Read(tmp)
		buf = append(buf, tmp[:n]...)
		if err != nil {
			break
		}
	}
	return truncateStr(string(buf), limit), nil
}

func truncateStr(s string, n int) string {
	runes := []rune(s)
	if len(runes) <= n {
		return s
	}
	return string(runes[:n]) + "..."
}

// ---- 持久化 KV 存储(data/kv.json) ----

type KVStore struct {
	mu   sync.Mutex
	path string
	data map[string]string
}

var kvOnce sync.Once
var kv *KVStore

func kvStore() *KVStore {
	kvOnce.Do(func() { kv = newKV("data/kv.json") })
	return kv
}

func newKV(path string) *KVStore {
	k := &KVStore{path: path, data: map[string]string{}}
	if b, err := os.ReadFile(path); err == nil {
		_ = json.Unmarshal(b, &k.data)
	}
	return k
}

func (k *KVStore) get(key string) (string, error) {
	k.mu.Lock()
	defer k.mu.Unlock()
	if v, ok := k.data[key]; ok {
		return v, nil
	}
	return "", fmt.Errorf("没有这个键: %s", key)
}

func (k *KVStore) set(key, value string) error {
	k.mu.Lock()
	defer k.mu.Unlock()
	k.data[key] = value
	if err := os.MkdirAll(filepath.Dir(k.path), 0o755); err != nil {
		return err
	}
	b, err := json.MarshalIndent(k.data, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(k.path, b, 0o644)
}

// ---- 极简表达式求值器(来自课时3) ----

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

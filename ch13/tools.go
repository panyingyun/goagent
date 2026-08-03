package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"
)

type Tool struct {
	Name        string
	Description string
	Required    []string // 必填参数
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

func (r *Registry) Get(name string) (Tool, bool) {
	t, ok := r.tools[name]
	return t, ok
}

// Executor 在 Registry 外加了三层能力:参数校验、结果缓存、并行执行。
type Executor struct {
	reg   *Registry
	cache *Cache
}

func NewExecutor(reg *Registry) *Executor {
	return &Executor{reg: reg, cache: NewCache(60 * time.Second)}
}

// Execute 校验 -> 查缓存 -> 执行 -> 写缓存。
func (e *Executor) Execute(ctx context.Context, name, argsJSON string) (string, error) {
	tool, ok := e.reg.Get(name)
	if !ok {
		return "", fmt.Errorf("未知工具: %s", name)
	}
	var args map[string]any
	if strings.TrimSpace(argsJSON) != "" {
		if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
			return "", fmt.Errorf("参数不是合法 JSON: %w", err)
		}
	}
	for _, f := range tool.Required {
		if v, ok := args[f]; !ok || v == "" {
			return "", fmt.Errorf("缺少必填参数: %s", f)
		}
	}
	key := name + "|" + argsJSON
	if v, ok := e.cache.Get(key); ok {
		return v, nil
	}
	out, err := tool.Run(ctx, args)
	if err != nil {
		return "", err
	}
	e.cache.Set(key, out)
	return out, nil
}

// ToolResult 是并行执行一个工具调用的结果。
type ToolResult struct {
	ID   string
	Name string
	Out  string
	Err  error
}

// ExecuteParallel 并发执行多个工具调用(保持输入顺序),用固定 worker 池。
func (e *Executor) ExecuteParallel(ctx context.Context, calls []ToolCall) []ToolResult {
	results := make([]ToolResult, len(calls))
	if len(calls) == 0 {
		return results
	}
	jobs := make(chan int)
	var wg sync.WaitGroup
	workers := len(calls)
	if workers > 4 {
		workers = 4
	}
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := range jobs {
				out, err := e.Execute(ctx, calls[i].Function.Name, calls[i].Function.Arguments)
				results[i] = ToolResult{ID: calls[i].ID, Name: calls[i].Function.Name, Out: out, Err: err}
			}
		}()
	}
	go func() {
		defer close(jobs)
		for i := range calls {
			jobs <- i
		}
	}()
	wg.Wait()
	return results
}

// Cache 是带 TTL 的简单结果缓存。
type Cache struct {
	mu      sync.Mutex
	ttl     time.Duration
	entries map[string]cacheEntry
}

type cacheEntry struct {
	value   string
	expires time.Time
}

func NewCache(ttl time.Duration) *Cache {
	return &Cache{ttl: ttl, entries: map[string]cacheEntry{}}
}

func (c *Cache) Get(key string) (string, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	e, ok := c.entries[key]
	if !ok || time.Now().After(e.expires) {
		delete(c.entries, key)
		return "", false
	}
	return e.value, true
}

func (c *Cache) Set(key, value string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries[key] = cacheEntry{value: value, expires: time.Now().Add(c.ttl)}
}

var errNoParams = errors.New("无参数")

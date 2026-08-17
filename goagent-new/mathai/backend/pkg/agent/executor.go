package agent

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sort"
	"strings"
	"sync"
	"time"

	"mathai/pkg/llm"
)

// Executor 注册并执行工具,负责审批、缓存、幂等与并行安全调度。
type Executor struct {
	mu         sync.Mutex
	tools      map[string]Tool
	cache      map[string]string
	idempotent map[string]string
	Approver   Approver
	Observer   Observer
}

// NewExecutor 创建空的工具执行器。
func NewExecutor() *Executor {
	return &Executor{
		tools:      make(map[string]Tool),
		cache:      make(map[string]string),
		idempotent: make(map[string]string),
	}
}

// Register 注册工具；写操作会强制审批且禁止缓存与并行。
func (e *Executor) Register(tool Tool) error {
	if strings.TrimSpace(tool.Name) == "" || tool.Run == nil {
		return errors.New("tool name and Run are required")
	}
	if tool.Timeout <= 0 {
		tool.Timeout = 10 * time.Second
	}
	if tool.SideEffect == "" {
		tool.SideEffect = SideEffectNone
	}
	if tool.SideEffect == SideEffectWrite || tool.SideEffect == SideEffectExternal {
		tool.RequiresApproval = true
		tool.Cacheable = false
		tool.ParallelSafe = false
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	if _, exists := e.tools[tool.Name]; exists {
		return fmt.Errorf("tool %q already registered", tool.Name)
	}
	e.tools[tool.Name] = tool
	return nil
}

// Definitions 返回按名称排序的工具定义，供 LLM 请求使用。
func (e *Executor) Definitions() []llm.ToolDefinition {
	e.mu.Lock()
	names := make([]string, 0, len(e.tools))
	for name := range e.tools {
		names = append(names, name)
	}
	sort.Strings(names)
	defs := make([]llm.ToolDefinition, 0, len(names))
	for _, name := range names {
		defs = append(defs, e.tools[name].Definition())
	}
	e.mu.Unlock()
	return defs
}

// Execute 执行单次工具调用，处理审批、缓存与幂等。
func (e *Executor) Execute(ctx context.Context, call llm.ToolCall) ToolResult {
	started := time.Now()
	e.mu.Lock()
	tool, ok := e.tools[call.Function.Name]
	e.mu.Unlock()
	if !ok {
		return e.finish(ctx, ToolResult{Call: call, Err: fmt.Errorf("unknown tool %q", call.Function.Name)}, started, "")
	}
	canonical, err := canonicalJSON(call.Function.Arguments)
	if err != nil {
		return e.finish(ctx, ToolResult{Call: call, Err: fmt.Errorf("invalid tool arguments: %w", err)}, started, "")
	}
	key := idempotencyKey(tool.Name, canonical)

	e.mu.Lock()
	if tool.Idempotent {
		if out, exists := e.idempotent[key]; exists {
			e.mu.Unlock()
			return e.finish(ctx, ToolResult{Call: call, Out: out}, started, key)
		}
	}
	if tool.Cacheable {
		if out, exists := e.cache[key]; exists {
			e.mu.Unlock()
			return e.finish(ctx, ToolResult{Call: call, Out: out}, started, key)
		}
	}
	e.mu.Unlock()

	if tool.RequiresApproval {
		if e.Approver == nil {
			return e.finish(ctx, ToolResult{Call: call, Err: errors.New("tool requires approval but no approver is configured")}, started, key)
		}
		approved, approveErr := e.Approver.Approve(ctx, ApprovalRequest{
			ToolName: tool.Name, Arguments: canonical, SideEffect: tool.SideEffect, IdempotencyKey: key,
		})
		if approveErr != nil {
			return e.finish(ctx, ToolResult{Call: call, Err: fmt.Errorf("approval failed: %w", approveErr)}, started, key)
		}
		if !approved {
			return e.finish(ctx, ToolResult{Call: call, Err: errors.New("tool call denied by approver")}, started, key)
		}
	}

	toolCtx, cancel := context.WithTimeout(ctx, tool.Timeout)
	defer cancel()
	out, err := tool.Run(toolCtx, json.RawMessage(canonical))
	result := ToolResult{Call: call, Out: out, Err: err}
	if err == nil {
		e.mu.Lock()
		if tool.Idempotent {
			e.idempotent[key] = out
		}
		if tool.Cacheable {
			e.cache[key] = out
		}
		e.mu.Unlock()
	}
	return e.finish(ctx, result, started, key)
}

// ExecuteBatch 执行一批工具调用；全部只读且并行安全时并发执行。
func (e *Executor) ExecuteBatch(ctx context.Context, calls []llm.ToolCall) []ToolResult {
	results := make([]ToolResult, len(calls))
	if len(calls) == 0 {
		return results
	}
	parallel := true
	e.mu.Lock()
	for _, call := range calls {
		tool, ok := e.tools[call.Function.Name]
		if !ok || !tool.ParallelSafe || tool.SideEffect == SideEffectWrite || tool.SideEffect == SideEffectExternal {
			parallel = false
			break
		}
	}
	e.mu.Unlock()
	if !parallel {
		for i, call := range calls {
			results[i] = e.Execute(ctx, call)
		}
		return results
	}
	var wg sync.WaitGroup
	for i, call := range calls {
		wg.Add(1)
		go func(i int, call llm.ToolCall) {
			defer wg.Done()
			results[i] = e.Execute(ctx, call)
		}(i, call)
	}
	wg.Wait()
	return results
}

func (e *Executor) finish(ctx context.Context, result ToolResult, started time.Time, key string) ToolResult {
	if e.Observer != nil {
		event := Event{Kind: "tool", Tool: result.Call.Function.Name, Status: "ok", Duration: time.Since(started), IdempotencyKey: key}
		if result.Err != nil {
			event.Status = "error"
			event.Error = result.Err.Error()
		}
		e.Observer.Record(ctx, event)
	}
	return result
}

func canonicalJSON(input string) (string, error) {
	if strings.TrimSpace(input) == "" {
		input = "{}"
	}
	var value any
	dec := json.NewDecoder(strings.NewReader(input))
	dec.UseNumber()
	if err := dec.Decode(&value); err != nil {
		return "", err
	}
	var extra any
	if err := dec.Decode(&extra); !errors.Is(err, io.EOF) {
		return "", errors.New("multiple JSON values")
	}
	b, err := json.Marshal(value)
	return string(b), err
}

func idempotencyKey(name, canonical string) string {
	sum := sha256.Sum256([]byte(name + "\x00" + canonical))
	return hex.EncodeToString(sum[:])
}

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestExecuteParallelPreservesOrder(t *testing.T) {
	reg := NewRegistry()
	var calls atomic.Int64
	reg.Register(Tool{
		Name: "slow",
		Run: func(ctx context.Context, args map[string]any) (string, error) {
			calls.Add(1)
			time.Sleep(50 * time.Millisecond)
			return "ok", nil
		},
	})
	ex := NewExecutor(reg)
	mk := func(id, name, args string) ToolCall {
		var tc ToolCall
		tc.ID = id
		tc.Function.Name = name
		tc.Function.Arguments = args
		return tc
	}
	callsList := []ToolCall{mk("1", "slow", "{}"), mk("2", "slow", "{}"), mk("3", "slow", "{}")}
	start := time.Now()
	results := ex.ExecuteParallel(context.Background(), callsList)
	elapsed := time.Since(start)
	if elapsed > 150*time.Millisecond {
		t.Fatalf("并行应明显快于串行,耗时 %v", elapsed)
	}
	for i, r := range results {
		if r.ID != callsList[i].ID {
			t.Fatalf("结果顺序应与输入一致: %d", i)
		}
	}
	if calls.Load() != 3 {
		t.Fatalf("应执行 3 次,实际 %d", calls.Load())
	}
}

func TestCacheAvoidsRerun(t *testing.T) {
	reg := NewRegistry()
	var calls atomic.Int64
	reg.Register(Tool{
		Name:     "add",
		Required: []string{"a", "b"},
		Run: func(ctx context.Context, args map[string]any) (string, error) {
			calls.Add(1)
			return "42", nil
		},
	})
	ex := NewExecutor(reg)
	if _, err := ex.Execute(context.Background(), "add", `{"a":1,"b":2}`); err != nil {
		t.Fatal(err)
	}
	if _, err := ex.Execute(context.Background(), "add", `{"a":1,"b":2}`); err != nil {
		t.Fatal(err)
	}
	if calls.Load() != 1 {
		t.Fatalf("缓存后不应重复执行,实际 %d 次", calls.Load())
	}
}

func TestRequiredFieldValidation(t *testing.T) {
	reg := NewRegistry()
	reg.Register(Tool{
		Name:     "add",
		Required: []string{"a", "b"},
		Run: func(ctx context.Context, args map[string]any) (string, error) {
			return "ok", nil
		},
	})
	ex := NewExecutor(reg)
	if _, err := ex.Execute(context.Background(), "add", `{"a":1}`); err == nil {
		t.Error("缺少必填参数应报错")
	}
	if _, err := ex.Execute(context.Background(), "add", `{bad json`); err == nil {
		t.Error("非法 JSON 应报错")
	}
}

func TestToolDefsJSON(t *testing.T) {
	reg := NewRegistry()
	registerDemoTools(reg)
	defs := toolDefs(reg)
	if len(defs) != 3 {
		t.Fatalf("应有 3 个工具,实际 %d", len(defs))
	}
	data, _ := json.Marshal(defs)
	if !strings.Contains(string(data), "add") {
		t.Error("JSON 序列化应包含工具名")
	}
	_ = fmt.Sprint()
}

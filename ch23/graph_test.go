package main

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestLinearGraph(t *testing.T) {
	g := NewGraph()
	var order []string
	g.AddNode("a", func(ctx context.Context, s map[string]any) (map[string]any, error) {
		order = append(order, "a")
		return map[string]any{"x": 1}, nil
	})
	g.AddNode("b", func(ctx context.Context, s map[string]any) (map[string]any, error) {
		order = append(order, "b")
		return map[string]any{"y": s["x"].(int) + 1}, nil
	})
	g.AddEdge("a", "b")
	state, err := g.Invoke(context.Background(), "a", map[string]any{}, "")
	if err != nil {
		t.Fatal(err)
	}
	if len(order) != 2 || order[0] != "a" || order[1] != "b" {
		t.Fatalf("执行顺序错误: %v", order)
	}
	if state["y"] != 2 {
		t.Fatalf("状态应传递并更新: %v", state)
	}
}

func TestLoopViaNext(t *testing.T) {
	g := NewGraph()
	rounds := 0
	g.AddNode("loop", func(ctx context.Context, s map[string]any) (map[string]any, error) {
		rounds++
		if rounds < 3 {
			return map[string]any{"__next__": "loop"}, nil
		}
		return map[string]any{"__next__": ""}, nil
	})
	if _, err := g.Invoke(context.Background(), "loop", map[string]any{}, ""); err != nil {
		t.Fatal(err)
	}
	if rounds != 3 {
		t.Fatalf("应循环 3 次,实际 %d", rounds)
	}
}

func TestCheckpoint(t *testing.T) {
	g := NewGraph()
	g.AddNode("set", func(ctx context.Context, s map[string]any) (map[string]any, error) {
		return map[string]any{"done": true}, nil
	})
	path := filepath.Join(t.TempDir(), "c.json")
	if _, err := g.Invoke(context.Background(), "set", map[string]any{}, path); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) == "" {
		t.Fatal("检查点文件不应为空")
	}
}

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
)

// NodeFunc 是一个图节点:输入共享状态,返回要合并进状态的更新。
// 如果更新里带 "__next__",就路由到该节点(支持循环);否则走第一条出边。
type NodeFunc func(ctx context.Context, state map[string]any) (map[string]any, error)

// Graph 是 LangGraph 的最小实现:节点 + 边 + 共享状态 + 检查点。
type Graph struct {
	nodes map[string]NodeFunc
	edges [][2]string
}

func NewGraph() *Graph {
	return &Graph{nodes: map[string]NodeFunc{}}
}

func (g *Graph) AddNode(name string, fn NodeFunc) {
	g.nodes[name] = fn
}

func (g *Graph) AddEdge(from, to string) {
	g.edges = append(g.edges, [2]string{from, to})
}

// Invoke 从 start 节点开始执行,直到没有下一步;结果保存到检查点文件。
func (g *Graph) Invoke(ctx context.Context, start string, state map[string]any, checkpoint string) (map[string]any, error) {
	current := start
	for step := 0; step < 100; step++ {
		fn, ok := g.nodes[current]
		if !ok {
			return state, fmt.Errorf("未知节点: %s", current)
		}
		updates, err := fn(ctx, state)
		if err != nil {
			return state, err
		}
		state["last_node"] = current
		for k, v := range updates {
			state[k] = v
		}
		next, _ := updates["__next__"].(string)
		if next == "" {
			for _, e := range g.edges {
				if e[0] == current {
					next = e[1]
					break
				}
			}
		}
		if next == "" {
			break
		}
		current = next
	}
	if checkpoint != "" {
		data, _ := json.MarshalIndent(state, "", "  ")
		if err := os.WriteFile(checkpoint, data, 0o644); err != nil {
			return state, err
		}
	}
	return state, nil
}

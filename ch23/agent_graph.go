package main

import (
	"context"
	"fmt"
)

// BuildWorkflowGraph 把"规划->执行->审查"画成一张图:
// 审查不通过就通过 __next__ 回到执行节点(最多返工 maxRework 次)。
func BuildWorkflowGraph(llm ChatLLM, maxRework int) *Graph {
	g := NewGraph()
	g.AddNode("plan", func(ctx context.Context, state map[string]any) (map[string]any, error) {
		reply, err := llm.Chat(ctx, []Message{{Role: "system", Content: "你是规划者,输出 2~3 步计划。"}, {Role: "user", Content: state["task"].(string)}})
		if err != nil {
			return nil, err
		}
		return map[string]any{"plan": reply.Content}, nil
	})
	g.AddNode("execute", func(ctx context.Context, state map[string]any) (map[string]any, error) {
		reply, err := llm.Chat(ctx, []Message{
			{Role: "system", Content: "你是执行者,按计划输出结果。"},
			{Role: "user", Content: fmt.Sprintf("计划: %v", state["plan"])},
		})
		if err != nil {
			return nil, err
		}
		return map[string]any{"result": reply.Content}, nil
	})
	g.AddNode("review", func(ctx context.Context, state map[string]any) (map[string]any, error) {
		reply, err := llm.Chat(ctx, []Message{
			{Role: "system", Content: "你是审查者:合格回复\"通过\",否则给出改进意见。"},
			{Role: "user", Content: fmt.Sprintf("结果: %v", state["result"])},
		})
		if err != nil {
			return nil, err
		}
		rounds, _ := state["rounds"].(int)
		if containsCN(reply.Content, "通过") || rounds >= maxRework {
			return map[string]any{"review": reply.Content, "__next__": ""}, nil
		}
		return map[string]any{"review": reply.Content, "rounds": rounds + 1, "__next__": "execute"}, nil
	})
	g.AddEdge("plan", "execute")
	g.AddEdge("execute", "review")
	return g
}

func containsCN(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}

package main

import (
	"testing"

	"goagent-new/pkg/llm"
)

func TestTrimHistoryDoesNotLeaveOrphanToolMessage(t *testing.T) {
	messages := []llm.Message{{Role: llm.RoleSystem, Content: "system"}}
	for i := 0; i < 10; i++ {
		messages = append(messages,
			llm.Message{Role: llm.RoleUser, Content: "question"},
			llm.Message{Role: llm.RoleAssistant, ToolCalls: []llm.ToolCall{{ID: "call"}}},
			llm.Message{Role: llm.RoleTool, ToolCallID: "call", Content: "result"},
			llm.Message{Role: llm.RoleAssistant, Content: "answer"},
		)
	}
	got := trimHistory(messages, 12)
	if len(got) < 2 || got[0].Role != llm.RoleSystem || got[1].Role != llm.RoleUser {
		t.Fatalf("history starts in the middle of a turn: %+v", got)
	}
}


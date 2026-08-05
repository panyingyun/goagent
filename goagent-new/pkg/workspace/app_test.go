package workspace

import (
	"context"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"goagent-new/pkg/llm"
)

type fakeClient struct {
	mu        sync.Mutex
	responses []llm.Response
}

func (f *fakeClient) Complete(_ context.Context, _ llm.Request) (llm.Response, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	response := f.responses[0]
	f.responses = f.responses[1:]
	return response, nil
}

func TestWorkspaceUsesRealToolLoop(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "guide.txt"), []byte("hello"), 0o600); err != nil {
		t.Fatal(err)
	}
	client := &fakeClient{responses: []llm.Response{
		{Message: llm.Message{Role: llm.RoleAssistant, ToolCalls: []llm.ToolCall{{ID: "list-1", Type: "function", Function: llm.FunctionCall{Name: "list_documents", Arguments: `{}`}}}}},
		{Message: llm.Message{Role: llm.RoleAssistant, Content: "guide.txt"}},
	}}
	app, err := New(root, client, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	answer, history, err := app.Chat(context.Background(), nil, "列出文档")
	if err != nil {
		t.Fatal(err)
	}
	if answer != "guide.txt" || len(history) != 5 || history[3].Role != llm.RoleTool {
		t.Fatalf("unexpected result: %q %+v", answer, history)
	}
}

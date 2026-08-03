package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHTTPMCPRoundtrip(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(HTTPMCPHandler))
	defer ts.Close()
	c := &HTTPMCPClient{URL: ts.URL + "/mcp", HTTP: ts.Client()}
	ctx := context.Background()
	tools, err := c.ListTools(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(tools) == 0 {
		t.Fatal("tools/list 应返回演示工具")
	}
	out, err := c.CallTool(ctx, "echo", map[string]any{"text": "hi"})
	if err != nil {
		t.Fatal(err)
	}
	if out != "echo: hi" {
		t.Fatalf("echo 结果不符: %q", out)
	}
}

type fakeProvider struct {
	calls []string
}

func (f *fakeProvider) ListTools(context.Context) ([]ToolInfo, error) {
	return []ToolInfo{{Name: "ping"}}, nil
}

func (f *fakeProvider) CallTool(_ context.Context, name string, _ map[string]any) (string, error) {
	f.calls = append(f.calls, name)
	return "pong:" + name, nil
}

func TestRouterPrefixRouting(t *testing.T) {
	a := &fakeProvider{}
	b := &fakeProvider{}
	r := NewRouter()
	r.Register("srvA", a)
	r.Register("srvB", b)
	ctx := context.Background()
	if _, err := r.Call(ctx, "srvA.ping", nil); err != nil {
		t.Fatal(err)
	}
	if _, err := r.Call(ctx, "srvB.ping", nil); err != nil {
		t.Fatal(err)
	}
	if len(a.calls) != 1 || len(b.calls) != 1 {
		t.Fatalf("路由应分发到各自服务器: %v %v", a.calls, b.calls)
	}
	if _, err := r.Call(ctx, "missing.ping", nil); err == nil {
		t.Error("未知服务器应报错")
	}
	if _, err := r.Call(ctx, "noprefix", nil); err == nil {
		t.Error("缺少服务器前缀应报错")
	}
}

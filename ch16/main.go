// 课时16:MCP 进阶
//
// 三个升级:
// 1. HTTP 传输:JSON-RPC 走 POST /mcp,不再只能 stdio;
// 2. 多服务器路由:Router 把多个 MCP 服务器合成一个命名空间;
// 3. 工具名规范:服务器名.工具名(如 echo.get_time)。
//
// 运行(不需要 API Key):
//
//	go run .
package main

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"time"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 1. 起一个 HTTP 传输的 MCP 服务器
	ts := httptest.NewServer(http.HandlerFunc(HTTPMCPHandler))
	defer ts.Close()

	// 2. 用 Router 挂上两个"服务器"(这里演示 HTTP + 一个内存假服务器)
	router := NewRouter()
	httpClient := &HTTPMCPClient{URL: ts.URL + "/mcp", HTTP: ts.Client()}
	router.Register("echo", httpClient)
	router.Register("mem", &memServer{})

	// 3. 通过路由调用
	fmt.Println("可用服务器:", router.Names())
	out, err := router.Call(ctx, "echo.echo", map[string]any{"text": "你好,HTTP MCP!"})
	if err != nil {
		fmt.Fprintln(os.Stderr, "调用失败:", err)
		return
	}
	fmt.Println("echo.echo ->", out)

	out, err = router.Call(ctx, "mem.add", map[string]any{"a": 1, "b": 2})
	if err != nil {
		fmt.Fprintln(os.Stderr, "调用失败:", err)
		return
	}
	fmt.Println("mem.add ->", out)
}

// memServer 是一个"假"内存服务器,演示多服务器路由。
type memServer struct{}

func (m *memServer) ListTools(context.Context) ([]ToolInfo, error) {
	return []ToolInfo{{Name: "add", Description: "两数相加"}}, nil
}

func (m *memServer) CallTool(_ context.Context, name string, args map[string]any) (string, error) {
	if name != "add" {
		return "", fmt.Errorf("未知工具: %s", name)
	}
	a, _ := toF64(args["a"])
	b, _ := toF64(args["b"])
	return fmt.Sprintf("%v", a+b), nil
}

func toF64(v any) (float64, bool) {
	switch n := v.(type) {
	case float64:
		return n, true
	case int:
		return float64(n), true
	}
	return 0, false
}

package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

// ToolProvider 抽象"能提供工具的服务器",stdio 和 HTTP 两种传输都实现它。
type ToolProvider interface {
	ListTools(ctx context.Context) ([]ToolInfo, error)
	CallTool(ctx context.Context, name string, args map[string]any) (string, error)
}

type ToolInfo struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// Router 把多个 MCP 服务器的工具合成一个命名空间:
// 工具名格式 "服务器名.工具名",例如 weather.get_weather。
type Router struct {
	servers map[string]ToolProvider
}

func NewRouter() *Router {
	return &Router{servers: map[string]ToolProvider{}}
}

func (r *Router) Register(name string, p ToolProvider) {
	r.servers[name] = p
}

func (r *Router) Names() []string {
	names := make([]string, 0, len(r.servers))
	for n := range r.servers {
		names = append(names, n)
	}
	return names
}

// Call 按 "server.tool" 路由到对应服务器。
func (r *Router) Call(ctx context.Context, fullName string, args map[string]any) (string, error) {
	i := strings.IndexByte(fullName, '.')
	if i < 0 {
		return "", fmt.Errorf("工具名必须是 服务器名.工具名: %q", fullName)
	}
	server, tool := fullName[:i], fullName[i+1:]
	p, ok := r.servers[server]
	if !ok {
		return "", fmt.Errorf("未知服务器: %s(可用: %s)", server, strings.Join(r.Names(), ", "))
	}
	return p.CallTool(ctx, tool, args)
}

// ---- JSON-RPC 消息结构(stdio 与 HTTP 共用) ----

type rpcRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params"`
}

type rpcResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *rpcError       `json:"error,omitempty"`
}

type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// handleRPC 处理一个请求,返回响应;通知(id 为空)返回 nil。
func handleRPC(req rpcRequest) *rpcResponse {
	if len(req.ID) == 0 {
		return nil // 通知
	}
	switch req.Method {
	case "initialize":
		return mkResult(req.ID, map[string]any{
			"protocolVersion": "2025-06-18",
			"capabilities":    map[string]any{"tools": map[string]any{"listChanged": false}},
			"serverInfo":      map[string]any{"name": "ch20-server", "version": "0.1.0"},
		})
	case "tools/list":
		return mkResult(req.ID, map[string]any{"tools": []any{
			map[string]any{
				"name":        "echo",
				"description": "把输入原样返回",
				"inputSchema": map[string]any{"type": "object", "properties": map[string]any{"text": map[string]any{"type": "string"}}},
			},
			map[string]any{
				"name":        "get_time",
				"description": "返回当前时间",
				"inputSchema": map[string]any{"type": "object", "properties": map[string]any{}},
			},
		}})
	case "tools/call":
		var p struct {
			Name      string         `json:"name"`
			Arguments map[string]any `json:"arguments"`
		}
		if err := json.Unmarshal(req.Params, &p); err != nil {
			return mkError(req.ID, -32602, "参数错误: "+err.Error())
		}
		text, err := callServerTool(p.Name, p.Arguments)
		if err != nil {
			return mkResult(req.ID, map[string]any{
				"content": []any{map[string]any{"type": "text", "text": err.Error()}},
				"isError": true,
			})
		}
		return mkResult(req.ID, map[string]any{
			"content": []any{map[string]any{"type": "text", "text": text}},
			"isError": false,
		})
	default:
		return mkError(req.ID, -32601, "未知方法: "+req.Method)
	}
}

func mkResult(id json.RawMessage, result any) *rpcResponse {
	b, _ := json.Marshal(result)
	return &rpcResponse{JSONRPC: "2.0", ID: id, Result: b}
}

func mkError(id json.RawMessage, code int, msg string) *rpcResponse {
	return &rpcResponse{JSONRPC: "2.0", ID: id, Error: &rpcError{Code: code, Message: msg}}
}

// callServerTool 是服务器真正的工具逻辑(演示:echo 和 get_time)。
func callServerTool(name string, args map[string]any) (string, error) {
	switch name {
	case "echo":
		text, _ := args["text"].(string)
		return "echo: " + text, nil
	case "get_time":
		return time.Now().Format("2006-01-02 15:04:05"), nil
	default:
		return "", errors.New("未知工具: " + name)
	}
}

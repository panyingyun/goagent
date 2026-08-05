// 一个最小的 MCP(Model Context Protocol)服务器。
//
// MCP 服务器通过标准输入输出(stdin/stdout)与客户端通信:
// 每一行是一个 JSON-RPC 2.0 消息(换行分隔的 JSON)。
// 实现 initialize、tools/list、tools/call 三个方法就够 Agent 使用了。
//
// 运行方式: go run ./mcp/server
package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"time"
)

const protocolVersion = "2025-06-18"

type request struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id"` // 通知消息没有 id
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params"`
}

type response struct {
	JSONRPC string    `json:"jsonrpc"`
	ID      any       `json:"id"`
	Result  any       `json:"result,omitempty"`
	Error   *rpcError `json:"error,omitempty"`
}

type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type callParams struct {
	Name      string         `json:"name"`
	Arguments map[string]any `json:"arguments"`
}

func main() {
	fmt.Fprintln(os.Stderr, "[mcp-server] 已启动,等待 JSON-RPC 请求...")
	scanner := bufio.NewScanner(os.Stdin)
	enc := json.NewEncoder(os.Stdout)
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(trimSpace(line)) == 0 {
			continue
		}
		var req request
		if err := json.Unmarshal(line, &req); err != nil {
			sendError(enc, nil, -32700, "解析失败: "+err.Error())
			continue
		}
		handle(enc, req)
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "[mcp-server] stdin 读取错误:", err)
	}
}

func handle(enc *json.Encoder, req request) {
	id := req.ID
	switch req.Method {
	case "initialize":
		sendResult(enc, id, map[string]any{
			"protocolVersion": protocolVersion,
			"capabilities": map[string]any{
				"tools": map[string]any{"listChanged": false},
			},
			"serverInfo": map[string]any{"name": "demo-mcp-server", "version": "0.1.0"},
		})

	case "notifications/initialized":
		// 通知:无需回复

	case "tools/list":
		sendResult(enc, id, map[string]any{
			"tools": []any{
				map[string]any{
					"name":        "echo",
					"description": "把输入原样返回,用于测试 MCP 链路。",
					"inputSchema": map[string]any{
						"type": "object",
						"properties": map[string]any{
							"text": map[string]any{"type": "string"},
						},
					},
				},
				map[string]any{
					"name":        "get_time",
					"description": "返回服务器所在机器的当前时间。",
					"inputSchema": map[string]any{"type": "object", "properties": map[string]any{}},
				},
			},
		})

	case "tools/call":
		var p callParams
		if err := json.Unmarshal(req.Params, &p); err != nil {
			sendError(enc, id, -32602, "参数解析失败: "+err.Error())
			return
		}
		text, err := callTool(p)
		if err != nil {
			sendResult(enc, id, map[string]any{
				"content": []any{map[string]any{"type": "text", "text": err.Error()}},
				"isError": true,
			})
			return
		}
		sendResult(enc, id, map[string]any{
			"content": []any{map[string]any{"type": "text", "text": text}},
			"isError": false,
		})

	default:
		sendError(enc, id, -32601, "未知方法: "+req.Method)
	}
}

func callTool(p callParams) (string, error) {
	switch p.Name {
	case "echo":
		text, _ := p.Arguments["text"].(string)
		return "echo: " + text, nil
	case "get_time":
		return time.Now().Format("2006-01-02 15:04:05"), nil
	default:
		return "", fmt.Errorf("未知工具: %s", p.Name)
	}
}

func sendResult(enc *json.Encoder, id any, result any) {
	if id == nil {
		return // 通知不回复
	}
	_ = enc.Encode(response{JSONRPC: "2.0", ID: id, Result: result})
}

func sendError(enc *json.Encoder, id any, code int, msg string) {
	if id == nil {
		return
	}
	_ = enc.Encode(response{JSONRPC: "2.0", ID: id, Error: &rpcError{Code: code, Message: msg}})
}

func trimSpace(b []byte) []byte {
	for len(b) > 0 && (b[0] == ' ' || b[0] == '\t' || b[0] == '\r' || b[0] == '\n') {
		b = b[1:]
	}
	for len(b) > 0 && (b[len(b)-1] == ' ' || b[len(b)-1] == '\t' || b[len(b)-1] == '\r' || b[len(b)-1] == '\n') {
		b = b[:len(b)-1]
	}
	return b
}

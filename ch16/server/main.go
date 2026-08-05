// 课时16:MCP 服务器(Server)
//
// MCP(Model Context Protocol)是"AI 世界的 USB-C":服务器把工具暴露出来,
// 任何支持 MCP 的客户端都能即插即用。本课时用标准库实现最小服务器:
// 通过 stdin/stdout 通信,每一行是一个 JSON-RPC 2.0 消息。
//
// 运行:
//   go run ./server
// 也可以直接把 JSON 请求喂给它,体验协议本身:
//   '{"jsonrpc":"2.0","id":1,"method":"tools/list"}' | go run ./server
package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
)

const protocolVersion = "2025-06-18"

// request 是客户端发来的 JSON-RPC 请求。
// 注意:通知(notification)没有 id,所以我们用 json.RawMessage 来区分"有/无 id"。
type request struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id"`
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

// callParams 是 tools/call 的参数:{name: 工具名, arguments: 参数对象}
type callParams struct {
	Name      string         `json:"name"`
	Arguments map[string]any `json:"arguments"`
}

func main() {
	fmt.Fprintln(os.Stderr, "[MCP 服务器] 已启动,从 stdin 读取 JSON-RPC 请求...")
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
}

// handle 按方法分发。MCP 最小闭环只需要三个方法:
// initialize(握手)、tools/list(列出工具)、tools/call(调用工具)。
func handle(enc *json.Encoder, req request) {
	id := req.ID
	switch req.Method {
	case "initialize":
		// 握手:告诉客户端协议版本、能力、服务器信息
		sendResult(enc, id, map[string]any{
			"protocolVersion": protocolVersion,
			"capabilities":    map[string]any{"tools": map[string]any{"listChanged": false}},
			"serverInfo":      map[string]any{"name": "ch16-mcp-server", "version": "0.1.0"},
		})

	case "notifications/initialized":
		// 通知:不需要回复

	case "tools/list":
		sendResult(enc, id, map[string]any{
			"tools": []any{
				toolDef("echo", "把输入原样返回", map[string]any{
					"type": "object",
					"properties": map[string]any{
						"text": map[string]any{"type": "string"},
					},
				}),
				toolDef("add", "计算两个数字相加", map[string]any{
					"type": "object",
					"properties": map[string]any{
						"a": map[string]any{"type": "number"},
						"b": map[string]any{"type": "number"},
					},
					"required": []string{"a", "b"},
				}),
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

// toolDef 生成 tools/list 里一个工具的描述结构。
func toolDef(name, desc string, schema map[string]any) map[string]any {
	return map[string]any{"name": name, "description": desc, "inputSchema": schema}
}

// callTool 是服务器真正的"干活"逻辑:根据工具名分发到具体函数。
func callTool(p callParams) (string, error) {
	switch p.Name {
	case "echo":
		text, _ := p.Arguments["text"].(string)
		return "echo: " + text, nil
	case "add":
		a, okA := toFloat(p.Arguments["a"])
		b, okB := toFloat(p.Arguments["b"])
		if !okA || !okB {
			return "", fmt.Errorf("参数 a、b 必须是数字")
		}
		return strconv.FormatFloat(a+b, 'f', -1, 64), nil
	default:
		return "", fmt.Errorf("未知工具: %s", p.Name)
	}
}

// toFloat 兼容 JSON 数字(默认 float64)和字符串形式的数字。
func toFloat(v any) (float64, bool) {
	switch n := v.(type) {
	case float64:
		return n, true
	case string:
		f, err := strconv.ParseFloat(n, 64)
		return f, err == nil
	default:
		return 0, false
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

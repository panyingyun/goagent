package main

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
)

// MCPClient 是一个最小的 MCP(Model Context Protocol)客户端:
// 通过 JSON-RPC 2.0(每行一个 JSON)与 MCP 服务器通信。
//
// 日常用法:你先手动启动服务器(TCP),再 ConnectMCPClient 连上去。
// 测试用法:StartMCPClient 仍可按 stdio 拉起子进程(见 react_test.go)。
const mcpProtocolVersion = "2025-06-18"

type MCPClient struct {
	cmd     *exec.Cmd // 仅 StartMCPClient 使用;手动连接时为 nil
	conn    net.Conn  // 仅 ConnectMCPClient 使用
	stdin   io.WriteCloser
	scanner *bufio.Scanner
	nextID  int
	mu      sync.Mutex
	pending map[int]chan json.RawMessage
}

type mcpResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      int             `json:"id"`
	Result  json.RawMessage `json:"result"`
	Error   *struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

// ConnectMCPClient 连接到已手动启动的 MCP 服务器。
// addr 形如 "127.0.0.1:8765"(服务器需用 -addr 监听)。
func ConnectMCPClient(ctx context.Context, addr string) (*MCPClient, error) {
	var d net.Dialer
	conn, err := d.DialContext(ctx, "tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("连接 MCP 服务器 %s 失败: %w(请先手动启动: go run ./mcp/server)", addr, err)
	}
	c := &MCPClient{
		conn:    conn,
		stdin:   conn,
		scanner: bufio.NewScanner(conn),
		pending: map[int]chan json.RawMessage{},
	}
	go c.readLoop()
	if err := c.handshake(ctx); err != nil {
		c.Close()
		return nil, err
	}
	return c, nil
}

// StartMCPClient 把 MCP 服务器当作子进程启动(stdio)。仅测试使用。
func StartMCPClient(ctx context.Context, binary string) (*MCPClient, error) {
	cmd := exec.CommandContext(ctx, binary, "-addr", "stdio")
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("启动 MCP 服务器失败: %w", err)
	}

	c := &MCPClient{
		cmd:     cmd,
		stdin:   stdin,
		scanner: bufio.NewScanner(stdout),
		pending: map[int]chan json.RawMessage{},
	}
	go c.readLoop()
	if err := c.handshake(ctx); err != nil {
		c.Close()
		return nil, err
	}
	return c, nil
}

func (c *MCPClient) handshake(ctx context.Context) error {
	if _, err := c.call(ctx, "initialize", map[string]any{
		"protocolVersion": mcpProtocolVersion,
		"capabilities":    map[string]any{},
		"clientInfo":      map[string]any{"name": "agent-demo", "version": "0.1.0"},
	}); err != nil {
		return err
	}
	c.notify("notifications/initialized", map[string]any{})
	return nil
}

// readLoop 持续读服务器输出,把响应分发到等待的 channel。
func (c *MCPClient) readLoop() {
	for c.scanner.Scan() {
		line := c.scanner.Bytes()
		if len(line) == 0 {
			continue
		}
		var resp mcpResponse
		if err := json.Unmarshal(line, &resp); err != nil {
			fmt.Fprintf(os.Stderr, "[mcp-client] 无法解析响应: %v\n", err)
			continue
		}
		c.mu.Lock()
		ch, ok := c.pending[resp.ID]
		if ok {
			delete(c.pending, resp.ID)
		}
		c.mu.Unlock()
		if ok {
			ch <- append([]byte(nil), line...) // 复制,避免 scanner 复用缓冲区
		}
	}
}

// call 发送一个请求并同步等待响应。
func (c *MCPClient) call(ctx context.Context, method string, params any) (json.RawMessage, error) {
	c.mu.Lock()
	c.nextID++
	id := c.nextID
	ch := make(chan json.RawMessage, 1)
	c.pending[id] = ch
	c.mu.Unlock()

	msg := map[string]any{"jsonrpc": "2.0", "id": id, "method": method, "params": params}
	data, err := json.Marshal(msg)
	if err != nil {
		return nil, err
	}
	if _, err := c.stdin.Write(append(data, '\n')); err != nil {
		return nil, err
	}

	select {
	case raw := <-ch:
		var resp mcpResponse
		if err := json.Unmarshal(raw, &resp); err != nil {
			return nil, err
		}
		if resp.Error != nil {
			return nil, fmt.Errorf("MCP 错误(%d): %s", resp.Error.Code, resp.Error.Message)
		}
		return resp.Result, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-time.After(30 * time.Second):
		return nil, errors.New("MCP 调用超时")
	}
}

func (c *MCPClient) notify(method string, params any) {
	msg := map[string]any{"jsonrpc": "2.0", "method": method, "params": params}
	if data, err := json.Marshal(msg); err == nil {
		_, _ = c.stdin.Write(append(data, '\n'))
	}
}

// CallTool 调用 MCP 服务器上的一个工具,返回文本结果。
func (c *MCPClient) CallTool(ctx context.Context, name string, args map[string]any) (string, error) {
	raw, err := c.call(ctx, "tools/call", map[string]any{"name": name, "arguments": args})
	if err != nil {
		return "", err
	}
	var result struct {
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
		IsError bool `json:"isError"`
	}
	if err := json.Unmarshal(raw, &result); err != nil {
		return "", err
	}
	var texts []string
	for _, item := range result.Content {
		texts = append(texts, item.Text)
	}
	if result.IsError {
		return "", errors.New(strings.Join(texts, "\n"))
	}
	return strings.Join(texts, "\n"), nil
}

func (c *MCPClient) Close() {
	if c.conn != nil {
		_ = c.conn.Close()
	} else if c.stdin != nil {
		_ = c.stdin.Close()
	}
	if c.cmd != nil && c.cmd.Process != nil {
		_ = c.cmd.Process.Kill()
	}
	if c.cmd != nil {
		_ = c.cmd.Wait()
	}
}

// registerMCPTool 把 MCP 服务器暴露成一个普通工具 "mcp",
// 这样 Agent 只需要一个工具入口,就能转发到任意 MCP 服务器。
func registerMCPTool(r *Registry, mcp *MCPClient) {
	r.Register(Tool{
		Name:        "mcp",
		Description: `通过 MCP 协议调用外部工具服务器上的工具。输入格式: {"tool": "工具名", "arguments": {参数对象}}。当前服务器可用工具: echo(参数 {"text": "要回显的字符串"})、get_time(无参数,arguments 可传 {})。`,
		Parameters: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"tool": map[string]any{
					"type":        "string",
					"description": "MCP 服务器上要调用的工具名,例如 echo 或 get_time",
				},
				"arguments": map[string]any{
					"type":        "object",
					"description": "传给该工具的参数。echo 必须用 {\"text\": \"...\"},不要用 content/message 等其它键名;get_time 传 {}",
				},
			},
			"required": []string{"tool"},
		},
		Run: func(ctx context.Context, args map[string]any) (string, error) {
			name, _ := args["tool"].(string)
			if name == "" {
				return "", errors.New("缺少 tool 参数")
			}
			argMap, _ := args["arguments"].(map[string]any)
			return mcp.CallTool(ctx, name, argMap)
		},
	})
}

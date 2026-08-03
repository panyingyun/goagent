// 课时2:MCP 客户端(Client)
//
// 客户端把 MCP 服务器当作子进程启动,通过 stdio 用 JSON-RPC 通信,
// 完成三步:initialize 握手 -> tools/list 看有哪些工具 -> tools/call 调用工具。
//
// 运行(先编译服务器):
//   go build -o server.exe ./server
//   go run ./client -server server.exe
package main

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
)

const mcpProtocolVersion = "2025-06-18"

func main() {
	serverBin := flag.String("server", "server.exe", "MCP 服务器可执行文件路径")
	flag.Parse()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client, err := StartMCPClient(ctx, *serverBin)
	if err != nil {
		fmt.Fprintln(os.Stderr, "连接 MCP 服务器失败:", err)
		os.Exit(1)
	}
	defer client.Close()

	// 第二步:查看服务器提供了哪些工具
	tools, err := client.ListTools(ctx)
	if err != nil {
		fmt.Fprintln(os.Stderr, "tools/list 失败:", err)
		os.Exit(1)
	}
	fmt.Println("服务器提供的工具:")
	for _, t := range tools {
		fmt.Printf(" - %s: %s\n", t.Name, t.Description)
	}

	// 第三步:调用工具
	text, err := client.CallTool(ctx, "echo", map[string]any{"text": "你好,MCP!"})
	if err != nil {
		fmt.Fprintln(os.Stderr, "echo 失败:", err)
	} else {
		fmt.Println("\n调用 echo:", text)
	}

	text, err = client.CallTool(ctx, "add", map[string]any{"a": 12, "b": 8})
	if err != nil {
		fmt.Fprintln(os.Stderr, "add 失败:", err)
	} else {
		fmt.Println("调用 add(12, 8):", text)
	}
}

// ---- 下面是最小 MCP 客户端实现 ----

// MCPClient 持有与服务器子进程的 stdin/stdout 连接。
type MCPClient struct {
	cmd     *exec.Cmd
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

// ToolInfo 是 tools/list 返回的一个工具描述。
type ToolInfo struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

func StartMCPClient(ctx context.Context, binary string) (*MCPClient, error) {
	cmd := exec.CommandContext(ctx, binary)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	cmd.Stderr = os.Stderr // 服务器的日志走 stderr,不污染协议通道
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

	// 第一步:initialize 握手 + initialized 通知
	if _, err := c.call(ctx, "initialize", map[string]any{
		"protocolVersion": mcpProtocolVersion,
		"capabilities":    map[string]any{},
		"clientInfo":      map[string]any{"name": "ch02-client", "version": "0.1.0"},
	}); err != nil {
		c.Close()
		return nil, err
	}
	c.notify("notifications/initialized", map[string]any{})
	return c, nil
}

// readLoop 持续读服务器输出,把响应分发到等待中的 channel。
func (c *MCPClient) readLoop() {
	for c.scanner.Scan() {
		line := c.scanner.Bytes()
		if len(line) == 0 {
			continue
		}
		var resp mcpResponse
		if err := json.Unmarshal(line, &resp); err != nil {
			fmt.Fprintf(os.Stderr, "[客户端] 无法解析响应: %v\n", err)
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

// call 发送一个 JSON-RPC 请求并同步等待响应。
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
		c.stdin.Write(append(data, '\n'))
	}
}

func (c *MCPClient) ListTools(ctx context.Context) ([]ToolInfo, error) {
	raw, err := c.call(ctx, "tools/list", map[string]any{})
	if err != nil {
		return nil, err
	}
	var result struct {
		Tools []ToolInfo `json:"tools"`
	}
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, err
	}
	return result.Tools, nil
}

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
	for _, c := range result.Content {
		texts = append(texts, c.Text)
	}
	if result.IsError {
		return "", errors.New(strings.Join(texts, "\n"))
	}
	return strings.Join(texts, "\n"), nil
}

func (c *MCPClient) Close() {
	if c.stdin != nil {
		c.stdin.Close()
	}
	if c.cmd != nil && c.cmd.Process != nil {
		_ = c.cmd.Process.Kill()
	}
	if c.cmd != nil {
		_ = c.cmd.Wait()
	}
}

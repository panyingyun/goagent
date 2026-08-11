package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type mockLLM struct {
	chatReply  string
	streamWord []string
}

func (m *mockLLM) Chat(_ context.Context, _ []Message) (Message, error) {
	return Message{Role: "assistant", Content: m.chatReply}, nil
}

func (m *mockLLM) StreamChat(_ context.Context, _ []Message, onDelta func(string)) error {
	for _, w := range m.streamWord {
		onDelta(w)
	}
	return nil
}

func TestIndexPage(t *testing.T) {
	s := &server{llm: &mockLLM{}}
	ts := httptest.NewServer(http.HandlerFunc(s.handleIndex))
	defer ts.Close()
	resp, err := http.Get(ts.URL)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Fatalf("首页应 200,实际 %d", resp.StatusCode)
	}
}

func TestChatJSON(t *testing.T) {
	s := &server{llm: &mockLLM{chatReply: "你好,我是测试 Agent"}}
	ts := httptest.NewServer(http.HandlerFunc(s.handleChat))
	defer ts.Close()
	body, _ := json.Marshal(map[string]string{"question": "你好"})
	resp, err := http.Post(ts.URL, "application/json", strings.NewReader(string(body)))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	var out map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out["answer"], "测试 Agent") {
		t.Fatalf("回答不符: %q", out["answer"])
	}
}

func TestStreamSSE(t *testing.T) {
	s := &server{llm: &mockLLM{streamWord: []string{"你", "好"}}}
	ts := httptest.NewServer(http.HandlerFunc(s.handleStream))
	defer ts.Close()
	resp, err := http.Get(ts.URL + "?q=hi")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	buf := make([]byte, 1024)
	n, _ := resp.Body.Read(buf)
	got := string(buf[:n])
	if !strings.Contains(got, "data: 你") || !strings.Contains(got, "data: 好") {
		t.Fatalf("SSE 输出不符: %q", got)
	}
}

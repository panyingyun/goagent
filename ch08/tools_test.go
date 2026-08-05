package main

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestHTTPSGetTool(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "你好,来自测试服务器")
	}))
	defer srv.Close()

	reg := NewRegistry()
	registerDefaultTools(reg)
	out, err := reg.Run(context.Background(), "http_get", fmt.Sprintf(`{"url": %q}`, srv.URL))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "测试服务器") {
		t.Fatalf("http_get 结果不符: %q", out)
	}
}

// 路径穿越必须被拦截。
func TestSafePathRejectsTraversal(t *testing.T) {
	if _, err := safePath("docs", "../secret.txt"); err == nil {
		t.Error("路径穿越应当被拒绝")
	}
	if _, err := safePath("docs", "notes.txt"); err != nil {
		t.Errorf("docs 内合法路径不应报错: %v", err)
	}
}

func TestAllowedCommand(t *testing.T) {
	if allowedCommand("echo hello") == false {
		t.Error("echo 应在白名单内")
	}
	if allowedCommand("rm -rf /") == true {
		t.Error("危险命令必须被拒绝")
	}
	if allowedCommand("del C:\\*") == true {
		t.Error("危险命令必须被拒绝")
	}
}

func TestKVStoreRoundtrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "kv.json")
	k := newKV(path)
	if err := k.set("theme", "dark"); err != nil {
		t.Fatal(err)
	}
	v, err := k.get("theme")
	if err != nil || v != "dark" {
		t.Fatalf("读取失败: %q %v", v, err)
	}
	if _, err := k.get("missing"); err == nil {
		t.Error("不存在的键应报错")
	}
	_ = os.Remove(path)
}

func TestWriteFileTool(t *testing.T) {
	if err := os.MkdirAll(docRoot, 0o755); err != nil {
		t.Fatal(err)
	}
	p, err := safePath(docRoot, "test.txt")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	defer os.Remove(p)
}

package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestStreamReadsUntilDone(t *testing.T) {
	server := httptest.NewServer(newStreamHandler(0))
	defer server.Close()
	response, err := http.Get(server.URL + "?q=你好")
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatal(err)
	}
	text := string(body)
	if !strings.Contains(text, `"data":"你"`) || !strings.Contains(text, `"data":"好"`) || !strings.Contains(text, "event: done") {
		t.Fatalf("incomplete stream: %q", text)
	}
}


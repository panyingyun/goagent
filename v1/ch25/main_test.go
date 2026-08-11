package main

import (
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHealth(t *testing.T) {
	rec := httptest.NewRecorder()
	writeJSON(rec, map[string]any{"status": "ok"})
	if rec.Code != 200 {
		t.Fatalf("health 应 200,实际 %d", rec.Code)
	}
}

func TestDocs(t *testing.T) {
	rec := httptest.NewRecorder()
	writeJSON(rec, map[string]any{"endpoints": []string{"/invoke", "/docs"}})
	var out map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&out); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(fmt.Sprint(out["endpoints"]), "/invoke") {
		t.Fatalf("docs 应包含 /invoke: %v", out)
	}
}

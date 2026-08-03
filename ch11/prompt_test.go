package main

import (
	"strings"
	"testing"
)

func TestBuildFewShot(t *testing.T) {
	p := buildFewShot("判断情感", []string{"A -> 正面", "B -> 负面"}, "C")
	if !strings.Contains(p, "判断情感") || !strings.Contains(p, "A -> 正面") || !strings.Contains(p, "C") {
		t.Fatalf("few-shot 拼接不完整: %q", p)
	}
}

func TestExtractJSON(t *testing.T) {
	m, err := extractJSON(`好的,结果如下:
{"sentiment": "正面", "confidence": 0.9, "reason": "夸赞"}
结束`)
	if err != nil {
		t.Fatal(err)
	}
	if m["sentiment"] != "正面" {
		t.Fatalf("sentiment 解析错误: %v", m)
	}
	if _, err := extractJSON("没有任何 JSON"); err == nil {
		t.Error("应当解析失败")
	}
}

func TestValidateResult(t *testing.T) {
	if err := validateResult(map[string]any{"sentiment": "正面", "confidence": 0.9}); err != nil {
		t.Fatal(err)
	}
	if err := validateResult(map[string]any{"sentiment": "胡说", "confidence": 0.9}); err == nil {
		t.Error("非法情感值应校验失败")
	}
}

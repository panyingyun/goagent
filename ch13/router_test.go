package main

import (
	"testing"
	"time"
)

func TestChooseHeuristic(t *testing.T) {
	r := NewRouter()
	if m := r.Choose("1+1 等于多少?"); m.Name != "deepseek-v4-flash" {
		t.Errorf("简单问题应选 flash,实际 %s", m.Name)
	}
	if m := r.Choose("帮我分析 Go 和 Rust 的优缺点"); m.Name != "deepseek-v4-pro" {
		t.Errorf("复杂问题应选 pro,实际 %s", m.Name)
	}
}

func TestBudgetDowngrade(t *testing.T) {
	r := NewRouter()
	// 把预算花掉一大半
	r.mu.Lock()
	r.spent = r.DailyBudget * 0.8
	r.mu.Unlock()
	if m := r.Choose("帮我分析 Go 和 Rust 的优缺点"); m.Name != "deepseek-v4-flash" {
		t.Errorf("预算不足时应降级到 flash,实际 %s", m.Name)
	}
}

func TestCacheHitAndExpire(t *testing.T) {
	c := NewRespCache(50 * time.Millisecond)
	if _, ok := c.Get("你好"); ok {
		t.Fatal("空缓存不应命中")
	}
	c.Set("你好  ", "回答")
	if ans, ok := c.Get("你好"); !ok || ans != "回答" {
		t.Fatalf("归一化后应命中: %q %v", ans, ok)
	}
	time.Sleep(80 * time.Millisecond)
	if _, ok := c.Get("你好"); ok {
		t.Fatal("过期缓存不应命中")
	}
}

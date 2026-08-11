package main

import (
	"strings"
	"sync"
	"time"
)

// Model 是一个可用的模型及其成本(粗略,美元/千 token)。
type Model struct {
	Name       string
	CostPer1K  float64
	Capability string // low / high
}

// Router 负责:按任务复杂度选模型 + 预算控制 + 响应缓存。
type Router struct {
	Flash       Model
	Pro         Model
	DailyBudget float64
	spent       float64
	mu          sync.Mutex
	cache       *RespCache
}

func NewRouter() *Router {
	return &Router{
		Flash:       Model{Name: "deepseek-v4-flash", CostPer1K: 0.0005, Capability: "low"},
		Pro:         Model{Name: "deepseek-v4-pro", CostPer1K: 0.002, Capability: "high"},
		DailyBudget: 2.0, // 每天最多花 2 美元
		cache:       NewRespCache(10 * time.Minute),
	}
}

// Choose 启发式选模型:复杂任务用 pro,简单任务用 flash;预算不足强制降级。
func (r *Router) Choose(question string) Model {
	complex := r.isComplex(question)
	r.mu.Lock()
	defer r.mu.Unlock()
	if complex && r.spent < r.DailyBudget*0.5 {
		return r.Pro
	}
	return r.Flash
}

func (r *Router) isComplex(q string) bool {
	keywords := []string{"分析", "总结", "写代码", "翻译", "多步", "推理", "规划", "对比"}
	for _, k := range keywords {
		if strings.Contains(q, k) {
			return true
		}
	}
	return len([]rune(q)) > 80
}

// Record 记账:估算 token 数与花费。
func (r *Router) Record(model Model, inTokens, outTokens int) {
	cost := float64(inTokens+outTokens) / 1000 * model.CostPer1K
	r.mu.Lock()
	r.spent += cost
	r.mu.Unlock()
}

func (r *Router) Spent() float64 {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.spent
}

// RespCache 按问题缓存回答。
type RespCache struct {
	mu      sync.Mutex
	ttl     time.Duration
	entries map[string]cacheEntry
}

type cacheEntry struct {
	answer  string
	expires time.Time
}

func NewRespCache(ttl time.Duration) *RespCache {
	return &RespCache{ttl: ttl, entries: map[string]cacheEntry{}}
}

func (c *RespCache) Get(question string) (string, bool) {
	key := normalize(question)
	c.mu.Lock()
	defer c.mu.Unlock()
	e, ok := c.entries[key]
	if !ok || time.Now().After(e.expires) {
		delete(c.entries, key)
		return "", false
	}
	return e.answer, true
}

func (c *RespCache) Set(question, answer string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries[normalize(question)] = cacheEntry{answer: answer, expires: time.Now().Add(c.ttl)}
}

// normalize 把问题归一化(去空白、转小写),相似问题命中同一缓存。
func normalize(q string) string {
	return strings.ToLower(strings.Join(strings.Fields(q), " "))
}

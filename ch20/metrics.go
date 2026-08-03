package main

import (
	"sync"
)

// Metrics 是最小指标集:计数器 + 直方图(用于平均耗时等)。
type Metrics struct {
	mu         sync.Mutex
	counters   map[string]int64
	histograms map[string][]float64
}

func NewMetrics() *Metrics {
	return &Metrics{counters: map[string]int64{}, histograms: map[string][]float64{}}
}

func (m *Metrics) Inc(name string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.counters[name]++
}

func (m *Metrics) Observe(name string, v float64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.histograms[name] = append(m.histograms[name], v)
}

// Snapshot 汇总:计数器原样输出,直方图输出计数与平均值。
func (m *Metrics) Snapshot() map[string]any {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := map[string]any{}
	for k, v := range m.counters {
		out["counter_"+k] = v
	}
	for k, vals := range m.histograms {
		if len(vals) == 0 {
			continue
		}
		var sum float64
		for _, v := range vals {
			sum += v
		}
		out["hist_"+k+"_count"] = len(vals)
		out["hist_"+k+"_avg"] = sum / float64(len(vals))
	}
	return out
}

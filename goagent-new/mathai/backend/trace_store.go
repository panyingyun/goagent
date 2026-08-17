package main

import "sync"

type TraceStore struct {
	mu      sync.RWMutex
	maximum int
	order   []string
	records map[string]TraceRecord
}

// NewTraceStore 创建最多保留 maximum 条记录的内存追踪存储。
func NewTraceStore(maximum int) *TraceStore {
	if maximum < 1 {
		maximum = 100
	}
	return &TraceStore{maximum: maximum, records: make(map[string]TraceRecord)}
}

// Put 写入一条追踪记录，超出容量时淘汰最旧的。
func (s *TraceStore) Put(record TraceRecord) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.records[record.ID]; !exists {
		s.order = append(s.order, record.ID)
	}
	s.records[record.ID] = record
	for len(s.order) > s.maximum {
		delete(s.records, s.order[0])
		s.order = s.order[1:]
	}
}

// Get 按 ID 读取追踪记录。
func (s *TraceStore) Get(id string) (TraceRecord, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	record, ok := s.records[id]
	return record, ok
}

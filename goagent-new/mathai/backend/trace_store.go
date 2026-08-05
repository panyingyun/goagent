package main

import "sync"

type TraceStore struct {
	mu      sync.RWMutex
	maximum int
	order   []string
	records map[string]TraceRecord
}

func NewTraceStore(maximum int) *TraceStore {
	if maximum < 1 {
		maximum = 100
	}
	return &TraceStore{maximum: maximum, records: make(map[string]TraceRecord)}
}

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

func (s *TraceStore) Get(id string) (TraceRecord, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	record, ok := s.records[id]
	return record, ok
}

package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

var (
	ErrNotFound = errors.New("session not found")
	ErrConflict = errors.New("session version conflict")
)

type Store struct {
	mu       sync.RWMutex
	path     string
	sessions map[string]Session
}

type storeFile struct {
	Sessions map[string]Session `json:"sessions"`
}

func OpenStore(path string) (*Store, error) {
	store := &Store{path: path, sessions: make(map[string]Session)}
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return store, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read session store: %w", err)
	}
	var file storeFile
	if err := json.Unmarshal(data, &file); err != nil {
		return nil, fmt.Errorf("decode session store: %w", err)
	}
	if file.Sessions != nil {
		store.sessions = file.Sessions
	}
	return store, nil
}

func (s *Store) Create(title string) (Session, error) {
	now := time.Now().UTC()
	title = strings.TrimSpace(title)
	if title == "" {
		title = "新的数学探索"
	}
	titleRunes := []rune(title)
	if len(titleRunes) > 80 {
		title = string(titleRunes[:80])
	}
	session := Session{ID: newID("session"), Title: title, Version: 1, Messages: []Message{}, GraphState: DefaultGraphState(), CreatedAt: now, UpdatedAt: now}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sessions[session.ID] = session
	if err := s.persistLocked(); err != nil {
		delete(s.sessions, session.ID)
		return Session{}, err
	}
	return cloneSession(session), nil
}

func (s *Store) List() []SessionSummary {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]SessionSummary, 0, len(s.sessions))
	for _, session := range s.sessions {
		result = append(result, SessionSummary{ID: session.ID, Title: session.Title, Version: session.Version, UpdatedAt: session.UpdatedAt})
	}
	sort.Slice(result, func(i, j int) bool { return result[i].UpdatedAt.After(result[j].UpdatedAt) })
	return result
}

func (s *Store) Get(id string) (Session, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	session, ok := s.sessions[id]
	if !ok {
		return Session{}, ErrNotFound
	}
	return cloneSession(session), nil
}

func (s *Store) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	session, ok := s.sessions[id]
	if !ok {
		return ErrNotFound
	}
	delete(s.sessions, id)
	if err := s.persistLocked(); err != nil {
		s.sessions[id] = session
		return err
	}
	return nil
}

func (s *Store) CommitChat(id string, expectedVersion int64, user, assistant Message, state GraphState) (Session, error) {
	if err := validateGraphState(state); err != nil {
		return Session{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	session, ok := s.sessions[id]
	if !ok {
		return Session{}, ErrNotFound
	}
	if session.Version != expectedVersion {
		return Session{}, ErrConflict
	}
	previous := cloneSession(session)
	session.Messages = append(session.Messages, user, assistant)
	session.GraphState = cloneGraphState(state)
	session.Version++
	session.UpdatedAt = time.Now().UTC()
	s.sessions[id] = session
	if err := s.persistLocked(); err != nil {
		s.sessions[id] = previous
		return Session{}, err
	}
	return cloneSession(session), nil
}

func (s *Store) persistLocked() error {
	directory := filepath.Dir(s.path)
	if err := os.MkdirAll(directory, 0o755); err != nil {
		return fmt.Errorf("create store directory: %w", err)
	}
	data, err := json.MarshalIndent(storeFile{Sessions: s.sessions}, "", "  ")
	if err != nil {
		return fmt.Errorf("encode session store: %w", err)
	}
	temporary, err := os.CreateTemp(directory, ".mathai-*.json")
	if err != nil {
		return fmt.Errorf("create temporary store: %w", err)
	}
	temporaryName := temporary.Name()
	defer os.Remove(temporaryName)
	if err := temporary.Chmod(0o600); err != nil {
		temporary.Close()
		return err
	}
	if _, err := temporary.Write(data); err != nil {
		temporary.Close()
		return err
	}
	if err := temporary.Sync(); err != nil {
		temporary.Close()
		return err
	}
	if err := temporary.Close(); err != nil {
		return err
	}
	if err := os.Rename(temporaryName, s.path); err != nil {
		return fmt.Errorf("replace session store: %w", err)
	}
	return nil
}

func cloneSession(session Session) Session {
	session.Messages = append([]Message(nil), session.Messages...)
	session.GraphState = cloneGraphState(session.GraphState)
	return session
}

package graphstore

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"goagent-new/pkg/graph"
)

type FileStore struct {
	mu             sync.Mutex
	CheckpointPath string
	EventPath      string
}

func (s *FileStore) Save(state graph.State) error {
	if s.CheckpointPath == "" {
		return errors.New("checkpoint path is required")
	}
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	return replaceFile(s.CheckpointPath, data)
}

func (s *FileStore) Append(event graph.Event) error {
	if s.EventPath == "" {
		return errors.New("event path is required")
	}
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := os.MkdirAll(filepath.Dir(s.EventPath), 0o700); err != nil {
		return err
	}
	file, err := os.OpenFile(s.EventPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		return err
	}
	if _, err := file.Write(append(data, '\n')); err != nil {
		file.Close()
		return err
	}
	if err := file.Sync(); err != nil {
		file.Close()
		return err
	}
	return file.Close()
}

func (s *FileStore) Load() (*graph.State, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	data, err := os.ReadFile(s.CheckpointPath)
	if err != nil {
		return nil, err
	}
	var state graph.State
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("decode graph checkpoint: %w", err)
	}
	if state.Data == nil {
		state.Data = make(map[string]string)
	}
	if state.Outputs == nil {
		state.Outputs = make(map[string]string)
	}
	if state.CompletedKeys == nil {
		state.CompletedKeys = make(map[string]string)
	}
	return &state, nil
}

func (s *FileStore) ReadEvents() ([]graph.Event, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	file, err := os.Open(s.EventPath)
	if errors.Is(err, os.ErrNotExist) {
		return []graph.Event{}, nil
	}
	if err != nil {
		return nil, err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 64<<10), 1<<20)
	events := make([]graph.Event, 0)
	for scanner.Scan() {
		var event graph.Event
		if err := json.Unmarshal(scanner.Bytes(), &event); err != nil {
			return nil, fmt.Errorf("decode graph event: %w", err)
		}
		events = append(events, event)
	}
	return events, scanner.Err()
}

type Migration func(*graph.State) error

func Migrate(state *graph.State, targetVersion int, migrations map[int]Migration) error {
	if state == nil || targetVersion <= 0 {
		return errors.New("state and positive target version are required")
	}
	if state.GraphVersion > targetVersion {
		return errors.New("checkpoint is newer than the running graph")
	}
	for state.GraphVersion < targetVersion {
		migration, ok := migrations[state.GraphVersion]
		if !ok {
			return fmt.Errorf("missing migration from graph version %d", state.GraphVersion)
		}
		if err := migration(state); err != nil {
			return fmt.Errorf("migrate graph version %d: %w", state.GraphVersion, err)
		}
		state.GraphVersion++
	}
	return nil
}

func replaceFile(path string, data []byte) error {
	directory := filepath.Dir(path)
	if err := os.MkdirAll(directory, 0o700); err != nil {
		return err
	}
	temporary, err := os.CreateTemp(directory, ".graph-*")
	if err != nil {
		return err
	}
	name := temporary.Name()
	defer os.Remove(name)
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
	return os.Rename(name, path)
}

package memory

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"
)

// Preference 是一条可持久化的用户偏好(键值 + 来源 + 更新时间)。
type Preference struct {
	Key       string    `json:"key"`
	Value     string    `json:"value"`
	Source    string    `json:"source"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Store 是基于本地 JSON 文件的偏好存储器。
type Store struct {
	mu    sync.Mutex
	path  string
	items map[string]Preference
}

func Open(path string) (*Store, error) {
	s := &Store{path: path, items: make(map[string]Preference)}
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return s, nil
	}
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(data, &s.items); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *Store) Set(key, value, source string) error {
	if key == "" || source == "" {
		return errors.New("preference key and source are required")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.items[key] = Preference{Key: key, Value: value, Source: source, UpdatedAt: time.Now().UTC()}
	return s.saveLocked()
}

func (s *Store) Get(key string) (Preference, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	p, ok := s.items[key]
	return p, ok
}

func (s *Store) List() []Preference {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]Preference, 0, len(s.items))
	for _, p := range s.items {
		out = append(out, p)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Key < out[j].Key })
	return out
}

func (s *Store) Delete(key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.items, key)
	return s.saveLocked()
}

func (s *Store) DeleteAll() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.items = make(map[string]Preference)
	return s.saveLocked()
}

func (s *Store) saveLocked() error {
	data, err := json.MarshalIndent(s.items, "", "  ")
	if err != nil {
		return err
	}
	return writeFile(s.path, data, 0o600)
}

func writeFile(path string, data []byte, perm os.FileMode) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), ".memory-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)
	if err := tmp.Chmod(perm); err != nil {
		tmp.Close()
		return err
	}
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmpName, path)
}

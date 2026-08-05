package learning

import (
	"crypto/rand"
	"encoding/hex"
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

type Status string

const (
	Candidate Status = "candidate"
	Verified  Status = "verified"
	Promoted  Status = "promoted"
	Retired   Status = "retired"
)

type Risk string

const (
	LowRisk  Risk = "low"
	HighRisk Risk = "high"
)

type Feedback struct {
	TaskID    string    `json:"task_id"`
	Outcome   string    `json:"outcome"`
	Score     float64   `json:"score"`
	Comment   string    `json:"comment,omitempty"`
	Source    string    `json:"source"`
	CreatedAt time.Time `json:"created_at"`
}

type Experience struct {
	ID         string     `json:"id"`
	TaskType   string     `json:"task_type"`
	Situation  string     `json:"situation"`
	Action     string     `json:"action"`
	Outcome    string     `json:"outcome"`
	Lesson     string     `json:"lesson"`
	Tags       []string   `json:"tags,omitempty"`
	Scope      string     `json:"scope"`
	Risk       Risk       `json:"risk"`
	Confidence float64    `json:"confidence"`
	Evidence   []string   `json:"evidence"`
	Status     Status     `json:"status"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
	ExpiresAt  *time.Time `json:"expires_at,omitempty"`
}

type Review struct {
	Passed      int `json:"passed"`
	Total       int `json:"total"`
	Regressions int `json:"regressions"`
}

func (r Review) PassRate() float64 {
	if r.Total == 0 {
		return 0
	}
	return float64(r.Passed) / float64(r.Total)
}

type Gate struct {
	MinEvidence   int
	MinConfidence float64
	MinPassRate   float64
}

type Store struct {
	mu    sync.RWMutex
	path  string
	items map[string]Experience
	now   func() time.Time
}

func Open(path string) (*Store, error) {
	store := &Store{path: path, items: make(map[string]Experience), now: time.Now}
	if path == "" {
		return store, nil
	}
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return store, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read experience store: %w", err)
	}
	if err := json.Unmarshal(data, &store.items); err != nil {
		return nil, fmt.Errorf("decode experience store: %w", err)
	}
	return store, nil
}

func (s *Store) Propose(experience Experience) (Experience, error) {
	if strings.TrimSpace(experience.TaskType) == "" || strings.TrimSpace(experience.Situation) == "" || strings.TrimSpace(experience.Lesson) == "" {
		return Experience{}, errors.New("task type, situation and lesson are required")
	}
	if experience.Confidence < 0 || experience.Confidence > 1 {
		return Experience{}, errors.New("confidence must be between 0 and 1")
	}
	if experience.Scope == "" {
		experience.Scope = "global"
	}
	if experience.Risk == "" {
		experience.Risk = LowRisk
	}
	if experience.Risk != LowRisk && experience.Risk != HighRisk {
		return Experience{}, errors.New("unsupported risk level")
	}
	now := s.now().UTC()
	experience.ID = randomID("exp")
	experience.Status = Candidate
	experience.CreatedAt = now
	experience.UpdatedAt = now
	experience.Tags = normalizeStrings(experience.Tags)
	experience.Evidence = normalizeStrings(experience.Evidence)
	s.mu.Lock()
	defer s.mu.Unlock()
	s.items[experience.ID] = experience
	if err := s.saveLocked(); err != nil {
		delete(s.items, experience.ID)
		return Experience{}, err
	}
	return cloneExperience(experience), nil
}

func (s *Store) Verify(id string, gate Gate, review Review) (Experience, error) {
	if gate.MinEvidence <= 0 || gate.MinConfidence <= 0 || gate.MinConfidence > 1 || gate.MinPassRate <= 0 || gate.MinPassRate > 1 {
		return Experience{}, errors.New("promotion gate thresholds must use positive counts and rates between 0 and 1")
	}
	if review.Total <= 0 || review.Passed < 0 || review.Passed > review.Total || review.Regressions < 0 {
		return Experience{}, errors.New("review counts are invalid")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	experience, ok := s.items[id]
	if !ok {
		return Experience{}, errors.New("experience not found")
	}
	if experience.Status != Candidate {
		return Experience{}, fmt.Errorf("experience in status %q cannot be verified", experience.Status)
	}
	if len(experience.Evidence) < gate.MinEvidence {
		return Experience{}, errors.New("experience has insufficient independent evidence")
	}
	if experience.Confidence < gate.MinConfidence {
		return Experience{}, errors.New("experience confidence is below the promotion threshold")
	}
	if review.Regressions > 0 || review.PassRate() < gate.MinPassRate {
		return Experience{}, errors.New("regression gate failed")
	}
	previous := experience
	experience.Status = Verified
	experience.UpdatedAt = s.now().UTC()
	s.items[id] = experience
	if err := s.saveLocked(); err != nil {
		s.items[id] = previous
		return Experience{}, err
	}
	return cloneExperience(experience), nil
}

func (s *Store) Promote(id string, humanApproved bool) (Experience, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	experience, ok := s.items[id]
	if !ok {
		return Experience{}, errors.New("experience not found")
	}
	if experience.Status != Verified {
		return Experience{}, fmt.Errorf("experience in status %q cannot be promoted", experience.Status)
	}
	if experience.Risk == HighRisk && !humanApproved {
		return Experience{}, errors.New("high-risk experience requires human approval")
	}
	previous := experience
	experience.Status = Promoted
	experience.UpdatedAt = s.now().UTC()
	s.items[id] = experience
	if err := s.saveLocked(); err != nil {
		s.items[id] = previous
		return Experience{}, err
	}
	return cloneExperience(experience), nil
}

func (s *Store) Retire(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	experience, ok := s.items[id]
	if !ok {
		return errors.New("experience not found")
	}
	previous := experience
	experience.Status = Retired
	experience.UpdatedAt = s.now().UTC()
	s.items[id] = experience
	if err := s.saveLocked(); err != nil {
		s.items[id] = previous
		return err
	}
	return nil
}

func (s *Store) Search(taskType, query, scope string, limit int) []Experience {
	if limit <= 0 {
		return nil
	}
	now := s.now()
	queryWords := splitWords(query)
	type scored struct {
		experience Experience
		score      float64
	}
	s.mu.RLock()
	matches := make([]scored, 0, len(s.items))
	for _, experience := range s.items {
		if experience.Status != Promoted || (experience.ExpiresAt != nil && !experience.ExpiresAt.After(now)) {
			continue
		}
		if experience.Scope != "global" && experience.Scope != scope {
			continue
		}
		score := experience.Confidence
		if experience.TaskType == taskType {
			score += 3
		}
		haystack := splitWords(strings.Join(append(append([]string{experience.Situation, experience.Lesson}, experience.Tags...), experience.Action), " "))
		for word := range queryWords {
			if haystack[word] {
				score++
			}
		}
		if score > experience.Confidence {
			matches = append(matches, scored{experience: cloneExperience(experience), score: score})
		}
	}
	s.mu.RUnlock()
	sort.Slice(matches, func(i, j int) bool {
		if matches[i].score != matches[j].score {
			return matches[i].score > matches[j].score
		}
		return matches[i].experience.ID < matches[j].experience.ID
	})
	if len(matches) > limit {
		matches = matches[:limit]
	}
	result := make([]Experience, len(matches))
	for index := range matches {
		result[index] = matches[index].experience
	}
	return result
}

func (s *Store) List() []Experience {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]Experience, 0, len(s.items))
	for _, item := range s.items {
		result = append(result, cloneExperience(item))
	}
	sort.Slice(result, func(i, j int) bool { return result[i].CreatedAt.Before(result[j].CreatedAt) })
	return result
}

func (s *Store) saveLocked() error {
	if s.path == "" {
		return nil
	}
	data, err := json.MarshalIndent(s.items, "", "  ")
	if err != nil {
		return err
	}
	directory := filepath.Dir(s.path)
	if err := os.MkdirAll(directory, 0o700); err != nil {
		return err
	}
	temporary, err := os.CreateTemp(directory, ".experience-*")
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
	return os.Rename(name, s.path)
}

func normalizeStrings(values []string) []string {
	seen := make(map[string]bool)
	result := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" && !seen[value] {
			seen[value] = true
			result = append(result, value)
		}
	}
	sort.Strings(result)
	return result
}

func splitWords(text string) map[string]bool {
	result := make(map[string]bool)
	for _, word := range strings.FieldsFunc(strings.ToLower(text), func(char rune) bool {
		return !(char >= 'a' && char <= 'z') && !(char >= '0' && char <= '9') && !(char >= '\u4e00' && char <= '\u9fff')
	}) {
		result[word] = true
	}
	return result
}

func cloneExperience(experience Experience) Experience {
	experience.Tags = append([]string(nil), experience.Tags...)
	experience.Evidence = append([]string(nil), experience.Evidence...)
	return experience
}

func randomID(prefix string) string {
	buffer := make([]byte, 8)
	if _, err := rand.Read(buffer); err != nil {
		panic("crypto/rand unavailable: " + err.Error())
	}
	return prefix + "_" + hex.EncodeToString(buffer)
}

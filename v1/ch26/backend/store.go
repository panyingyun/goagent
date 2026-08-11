package main

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"time"

	_ "modernc.org/sqlite"
)

var ErrNotFound = errors.New("session not found")

// Store 封装 SQLite(modernc.org/sqlite,纯 Go 无 CGO)。
// 表结构严格按设计稿 §8。
type Store struct {
	db *sql.DB
}

func OpenStore(path string) (*Store, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS sessions (
		id TEXT PRIMARY KEY,
		title TEXT NOT NULL,
		graph_state TEXT NOT NULL,
		is_favorite INTEGER DEFAULT 0,
		created_at TEXT NOT NULL,
		updated_at TEXT NOT NULL
	)`); err != nil {
		return nil, err
	}
	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS messages (
		id TEXT PRIMARY KEY,
		session_id TEXT NOT NULL,
		role TEXT NOT NULL,
		content TEXT NOT NULL,
		structured_result TEXT,
		status TEXT DEFAULT 'success',
		created_at TEXT NOT NULL,
		FOREIGN KEY(session_id) REFERENCES sessions(id)
	)`); err != nil {
		return nil, err
	}
	return &Store{db: db}, nil
}

func (s *Store) Close() error { return s.db.Close() }

func now() string { return time.Now().Format(time.RFC3339) }

// newID 生成 "前缀_16位hex" 的 ID。
func newID(prefix string) string {
	b := make([]byte, 6)
	_, _ = rand.Read(b)
	return prefix + "_" + hex.EncodeToString(b)
}

func (s *Store) CreateSession(title string) (*Session, error) {
	sess := &Session{
		ID:         newID("session"),
		Title:      title,
		Messages:   []Message{},
		GraphState: defaultGraphState(),
		CreatedAt:  now(),
		UpdatedAt:  now(),
	}
	gs, _ := json.Marshal(sess.GraphState)
	if _, err := s.db.Exec(
		"INSERT INTO sessions (id, title, graph_state, is_favorite, created_at, updated_at) VALUES (?,?,?,?,?,?)",
		sess.ID, sess.Title, string(gs), boolToInt(sess.IsFavorite), sess.CreatedAt, sess.UpdatedAt,
	); err != nil {
		return nil, err
	}
	return sess, nil
}

func (s *Store) ListSessions() ([]Session, error) {
	rows, err := s.db.Query("SELECT id, title, is_favorite, created_at, updated_at FROM sessions ORDER BY updated_at DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := []Session{}
	for rows.Next() {
		var sess Session
		var fav int
		if err := rows.Scan(&sess.ID, &sess.Title, &fav, &sess.CreatedAt, &sess.UpdatedAt); err != nil {
			return nil, err
		}
		sess.IsFavorite = fav == 1
		list = append(list, sess)
	}
	return list, rows.Err()
}

func (s *Store) GetSession(id string) (*Session, error) {
	var sess Session
	var gsJSON string
	var fav int
	err := s.db.QueryRow(
		"SELECT id, title, graph_state, is_favorite, created_at, updated_at FROM sessions WHERE id = ?", id,
	).Scan(&sess.ID, &sess.Title, &gsJSON, &fav, &sess.CreatedAt, &sess.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	sess.IsFavorite = fav == 1
	if err := json.Unmarshal([]byte(gsJSON), &sess.GraphState); err != nil {
		return nil, err
	}
	msgs, err := s.listMessages(id)
	if err != nil {
		return nil, err
	}
	sess.Messages = msgs
	return &sess, nil
}

func (s *Store) listMessages(sessionID string) ([]Message, error) {
	rows, err := s.db.Query(
		"SELECT id, role, content, structured_result, status, created_at FROM messages WHERE session_id = ? ORDER BY created_at ASC",
		sessionID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	msgs := []Message{}
	for rows.Next() {
		var m Message
		var sr string
		var srPtr *string
		if err := rows.Scan(&m.ID, &m.Role, &m.Content, &sr, &m.Status, &m.CreatedAt); err != nil {
			return nil, err
		}
		if sr != "" {
			srPtr = &sr
		}
		if srPtr != nil {
			var res LLMStructuredResult
			if err := json.Unmarshal([]byte(*srPtr), &res); err == nil {
				m.StructuredResult = &res
			}
		}
		msgs = append(msgs, m)
	}
	return msgs, rows.Err()
}

type sessionPatch struct {
	Title      *string     `json:"title"`
	GraphState *GraphState `json:"graphState"`
	IsFavorite *bool       `json:"isFavorite"`
}

func (s *Store) PatchSession(id string, p sessionPatch) (*Session, error) {
	sess, err := s.GetSession(id)
	if err != nil {
		return nil, err
	}
	if p.Title != nil {
		sess.Title = *p.Title
	}
	if p.GraphState != nil {
		sess.GraphState = *p.GraphState
	}
	if p.IsFavorite != nil {
		sess.IsFavorite = *p.IsFavorite
	}
	gs, _ := json.Marshal(sess.GraphState)
	sess.UpdatedAt = now()
	if _, err := s.db.Exec(
		"UPDATE sessions SET title=?, graph_state=?, is_favorite=?, updated_at=? WHERE id=?",
		sess.Title, string(gs), boolToInt(sess.IsFavorite), sess.UpdatedAt, id,
	); err != nil {
		return nil, err
	}
	return sess, nil
}

func (s *Store) SaveGraphState(id string, gs *GraphState) error {
	data, _ := json.Marshal(gs)
	_, err := s.db.Exec("UPDATE sessions SET graph_state=?, updated_at=? WHERE id=?", string(data), now(), id)
	return err
}

func (s *Store) AddMessage(sessionID string, m *Message) error {
	srJSON := ""
	if m.StructuredResult != nil {
		b, _ := json.Marshal(m.StructuredResult)
		srJSON = string(b)
	}
	if _, err := s.db.Exec(
		"INSERT INTO messages (id, session_id, role, content, structured_result, status, created_at) VALUES (?,?,?,?,?,?,?)",
		m.ID, sessionID, m.Role, m.Content, srJSON, m.Status, m.CreatedAt,
	); err != nil {
		return err
	}
	_, err := s.db.Exec("UPDATE sessions SET updated_at=? WHERE id=?", now(), sessionID)
	return err
}

func (s *Store) DeleteSession(id string) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err := tx.Exec("DELETE FROM messages WHERE session_id=?", id); err != nil {
		return err
	}
	if _, err := tx.Exec("DELETE FROM sessions WHERE id=?", id); err != nil {
		return err
	}
	return tx.Commit()
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

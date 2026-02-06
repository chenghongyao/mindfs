package session

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
)

type Store struct {
	managedDir string
	mu         sync.Mutex
}

func NewStore(managedDir string) (*Store, error) {
	if managedDir == "" {
		return nil, errors.New("managed dir required")
	}
	if err := os.MkdirAll(filepath.Join(managedDir, "sessions"), 0o755); err != nil {
		return nil, err
	}
	return &Store{managedDir: managedDir}, nil
}

func (s *Store) Create(session *Session) error {
	if session == nil {
		return errors.New("session required")
	}
	path, err := s.sessionPath(session.Key)
	if err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, err := os.Stat(path); err == nil {
		return fmt.Errorf("session already exists: %s", session.Key)
	}
	return writeJSON(path, session)
}

func (s *Store) Save(session *Session) error {
	if session == nil {
		return errors.New("session required")
	}
	path, err := s.sessionPath(session.Key)
	if err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	return writeJSON(path, session)
}

func (s *Store) Get(key string) (*Session, error) {
	path, err := s.sessionPath(key)
	if err != nil {
		return nil, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	payload, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var session Session
	if err := json.Unmarshal(payload, &session); err != nil {
		return nil, err
	}
	return &session, nil
}

func (s *Store) List() ([]*Session, error) {
	sessionsDir := filepath.Join(s.managedDir, "sessions")
	entries, err := os.ReadDir(sessionsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []*Session{}, nil
		}
		return nil, err
	}
	items := make([]*Session, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasPrefix(name, "session-") || !strings.HasSuffix(name, ".json") {
			continue
		}
		payload, err := os.ReadFile(filepath.Join(sessionsDir, name))
		if err != nil {
			continue
		}
		var session Session
		if err := json.Unmarshal(payload, &session); err != nil {
			continue
		}
		items = append(items, &session)
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].UpdatedAt.After(items[j].UpdatedAt)
	})
	return items, nil
}

func (s *Store) sessionPath(key string) (string, error) {
	if key == "" {
		return "", errors.New("session key required")
	}
	if strings.Contains(key, "..") || strings.ContainsRune(key, filepath.Separator) || strings.Contains(key, "/") {
		return "", fmt.Errorf("invalid session key: %s", key)
	}
	name := fmt.Sprintf("session-%s.json", key)
	return filepath.Join(s.managedDir, "sessions", name), nil
}

func writeJSON(path string, value any) error {
	payload, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, payload, 0o644)
}

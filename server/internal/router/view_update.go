package router

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// ViewStore caches view.json for a managed directory.
type ViewStore struct {
	mu         sync.RWMutex
	managedDir string
	view       map[string]any
	updatedAt  time.Time
	modTime    time.Time
}

// NewViewStore creates a ViewStore for a managed directory.
func NewViewStore(managedDir string) (*ViewStore, error) {
	if managedDir == "" {
		return nil, errors.New("managed dir required")
	}
	return &ViewStore{managedDir: managedDir}, nil
}

// Load reads view.json from disk and refreshes the cache.
func (s *ViewStore) Load() (map[string]any, error) {
	path := filepath.Join(s.managedDir, "view.json")
	payload, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	var view map[string]any
	if err := json.Unmarshal(payload, &view); err != nil {
		return nil, err
	}
	s.mu.Lock()
	s.view = view
	s.updatedAt = time.Now().UTC()
	s.modTime = info.ModTime()
	s.mu.Unlock()
	return view, nil
}

// Get returns the cached view if available.
func (s *ViewStore) Get() (map[string]any, time.Time, time.Time, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.view == nil {
		return nil, time.Time{}, time.Time{}, false
	}
	return s.view, s.updatedAt, s.modTime, true
}

// ViewStoreManager manages per-root view caches.
type ViewStoreManager struct {
	mu     sync.Mutex
	stores map[string]*ViewStore
}

func NewViewStoreManager() *ViewStoreManager {
	return &ViewStoreManager{stores: make(map[string]*ViewStore)}
}

func (m *ViewStoreManager) Get(managedDir string) (*ViewStore, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if store, ok := m.stores[managedDir]; ok {
		return store, nil
	}
	store, err := NewViewStore(managedDir)
	if err != nil {
		return nil, err
	}
	m.stores[managedDir] = store
	return store, nil
}

package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

// Preferences stores user preferences for agent selection per mode
type Preferences struct {
	ModeAgents map[string]string `json:"mode_agents"` // mode -> agent name
}

// PreferencesStore manages user preferences
type PreferencesStore struct {
	mu       sync.RWMutex
	prefs    *Preferences
	filePath string
}

// NewPreferencesStore creates a new preferences store
func NewPreferencesStore() (*PreferencesStore, error) {
	configDir, err := MindFSConfigDir()
	if err != nil {
		return nil, err
	}

	if err := os.MkdirAll(configDir, 0o755); err != nil {
		return nil, err
	}

	filePath := filepath.Join(configDir, "preferences.json")
	store := &PreferencesStore{
		filePath: filePath,
		prefs:    defaultPreferences(),
	}

	// Load existing preferences
	if err := store.load(); err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	return store, nil
}

// defaultPreferences returns default preferences
func defaultPreferences() *Preferences {
	return &Preferences{
		ModeAgents: map[string]string{
			"chat": "claude",
			"view": "claude",
		},
	}
}

// load reads preferences from file
func (s *PreferencesStore) load() error {
	data, err := os.ReadFile(s.filePath)
	if err != nil {
		return err
	}

	var prefs Preferences
	if err := json.Unmarshal(data, &prefs); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Merge with defaults
	if prefs.ModeAgents == nil {
		prefs.ModeAgents = make(map[string]string)
	}
	for mode, agent := range defaultPreferences().ModeAgents {
		if _, ok := prefs.ModeAgents[mode]; !ok {
			prefs.ModeAgents[mode] = agent
		}
	}

	s.prefs = &prefs
	return nil
}

// save writes preferences to file
func (s *PreferencesStore) save() error {
	s.mu.RLock()
	data, err := json.MarshalIndent(s.prefs, "", "  ")
	s.mu.RUnlock()

	if err != nil {
		return err
	}

	return os.WriteFile(s.filePath, data, 0644)
}

// GetAgentForMode returns the preferred agent for a mode
func (s *PreferencesStore) GetAgentForMode(mode string) string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if agent, ok := s.prefs.ModeAgents[mode]; ok {
		return agent
	}
	return "claude" // default
}

// SetAgentForMode sets the preferred agent for a mode
func (s *PreferencesStore) SetAgentForMode(mode, agent string) error {
	s.mu.Lock()
	s.prefs.ModeAgents[mode] = agent
	s.mu.Unlock()

	return s.save()
}

// GetAll returns all preferences
func (s *PreferencesStore) GetAll() *Preferences {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Return a copy
	copy := &Preferences{
		ModeAgents: make(map[string]string),
	}
	for k, v := range s.prefs.ModeAgents {
		copy.ModeAgents[k] = v
	}
	return copy
}

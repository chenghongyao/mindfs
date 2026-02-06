package router

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// ViewPreference stores user's view preferences
type ViewPreference struct {
	RouteID    string    `json:"route_id"`
	Version    string    `json:"version,omitempty"`
	SelectedAt time.Time `json:"selected_at"`
}

// ViewPreferences manages user view preferences
type ViewPreferences struct {
	mu          sync.RWMutex
	managedDir  string
	preferences map[string]ViewPreference // path -> preference
}

// NewViewPreferences creates a new preferences manager
func NewViewPreferences(managedDir string) (*ViewPreferences, error) {
	p := &ViewPreferences{
		managedDir:  managedDir,
		preferences: make(map[string]ViewPreference),
	}

	if err := p.load(); err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	return p, nil
}

// preferencesPath returns the path to preferences file
func (p *ViewPreferences) preferencesPath() string {
	return filepath.Join(p.managedDir, "view-preferences.json")
}

// load reads preferences from file
func (p *ViewPreferences) load() error {
	data, err := os.ReadFile(p.preferencesPath())
	if err != nil {
		return err
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	return json.Unmarshal(data, &p.preferences)
}

// save writes preferences to file
func (p *ViewPreferences) save() error {
	p.mu.RLock()
	data, err := json.MarshalIndent(p.preferences, "", "  ")
	p.mu.RUnlock()

	if err != nil {
		return err
	}

	return os.WriteFile(p.preferencesPath(), data, 0644)
}

// Get returns the preference for a path
func (p *ViewPreferences) Get(path string) *ViewPreference {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if pref, ok := p.preferences[path]; ok {
		return &pref
	}
	return nil
}

// Set saves a preference for a path
func (p *ViewPreferences) Set(path string, routeID string, version string) error {
	p.mu.Lock()
	p.preferences[path] = ViewPreference{
		RouteID:    routeID,
		Version:    version,
		SelectedAt: time.Now(),
	}
	p.mu.Unlock()

	return p.save()
}

// SetRoute saves route preference (without version)
func (p *ViewPreferences) SetRoute(path, routeID string) error {
	p.mu.Lock()
	existing := p.preferences[path]
	p.preferences[path] = ViewPreference{
		RouteID:    routeID,
		Version:    existing.Version, // preserve version
		SelectedAt: time.Now(),
	}
	p.mu.Unlock()

	return p.save()
}

// SetVersion saves version preference (without changing route)
func (p *ViewPreferences) SetVersion(path, version string) error {
	p.mu.Lock()
	existing := p.preferences[path]
	p.preferences[path] = ViewPreference{
		RouteID:    existing.RouteID, // preserve route
		Version:    version,
		SelectedAt: time.Now(),
	}
	p.mu.Unlock()

	return p.save()
}

// Delete removes preference for a path
func (p *ViewPreferences) Delete(path string) error {
	p.mu.Lock()
	delete(p.preferences, path)
	p.mu.Unlock()

	return p.save()
}

// GetAll returns all preferences
func (p *ViewPreferences) GetAll() map[string]ViewPreference {
	p.mu.RLock()
	defer p.mu.RUnlock()

	result := make(map[string]ViewPreference)
	for k, v := range p.preferences {
		result[k] = v
	}
	return result
}

// Clear removes all preferences
func (p *ViewPreferences) Clear() error {
	p.mu.Lock()
	p.preferences = make(map[string]ViewPreference)
	p.mu.Unlock()

	return p.save()
}

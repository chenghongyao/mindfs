package router

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// ViewStatus tracks view state
type ViewStatus struct {
	ActiveVersions map[string]string `json:"active_versions"` // route_id -> version
	LastSelected   map[string]string `json:"last_selected"`   // path -> route_id
	Pending        *PendingView      `json:"pending,omitempty"`
}

// PendingView represents a view being generated
type PendingView struct {
	RouteID   string `json:"route_id"`
	Version   string `json:"version"`
	SessionID string `json:"session_id,omitempty"`
	StartedAt string `json:"started_at"`
}

// LoadViewStatus loads view status from .mindfs/view-status.json
func LoadViewStatus(managedDir string) (*ViewStatus, error) {
	statusPath := filepath.Join(managedDir, "view-status.json")
	data, err := os.ReadFile(statusPath)
	if err != nil {
		if os.IsNotExist(err) {
			return &ViewStatus{
				ActiveVersions: make(map[string]string),
				LastSelected:   make(map[string]string),
			}, nil
		}
		return nil, err
	}

	var status ViewStatus
	if err := json.Unmarshal(data, &status); err != nil {
		return nil, err
	}

	// Initialize maps if nil
	if status.ActiveVersions == nil {
		status.ActiveVersions = make(map[string]string)
	}
	if status.LastSelected == nil {
		status.LastSelected = make(map[string]string)
	}

	return &status, nil
}

// SaveViewStatus saves view status
func SaveViewStatus(managedDir string, status *ViewStatus) error {
	statusPath := filepath.Join(managedDir, "view-status.json")
	data, err := json.MarshalIndent(status, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(statusPath, data, 0644)
}

// SetActiveVersion sets the active version for a route
func (s *ViewStatus) SetActiveVersion(routeID, version string) {
	s.ActiveVersions[routeID] = version
}

// GetActiveVersion returns the active version for a route
func (s *ViewStatus) GetActiveVersion(routeID string) string {
	return s.ActiveVersions[routeID]
}

// SetLastSelected sets the last selected route for a path
func (s *ViewStatus) SetLastSelected(path, routeID string) {
	s.LastSelected[path] = routeID
}

// GetLastSelected returns the last selected route for a path
func (s *ViewStatus) GetLastSelected(path string) string {
	return s.LastSelected[path]
}

// SetPending sets the pending view
func (s *ViewStatus) SetPending(pending *PendingView) {
	s.Pending = pending
}

// ClearPending clears the pending view
func (s *ViewStatus) ClearPending() {
	s.Pending = nil
}

// HasPending returns true if there's a pending view
func (s *ViewStatus) HasPending() bool {
	return s.Pending != nil
}

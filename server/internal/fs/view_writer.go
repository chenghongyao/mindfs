package fs

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"time"
)

const (
	viewFileName       = "view.json"
	viewPreviousName   = "view.prev.json"
	viewStatusFileName = "view.status.json"
)

// ViewStatus tracks whether the current view is pending acceptance.
type ViewStatus struct {
	Pending    bool      `json:"pending"`
	CurrentID  string    `json:"current_id"`
	PreviousID string    `json:"previous_id,omitempty"`
	UpdatedAt  time.Time `json:"updated_at"`
}

func writeJSONAtomic(path string, payload any) error {
	if path == "" {
		return errors.New("path required")
	}
	data, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return err
	}
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(dir, ".tmp-*.json")
	if err != nil {
		return err
	}
	defer func() {
		_ = os.Remove(tmp.Name())
	}()
	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmp.Name(), path)
}

func copyFile(src, dst string) error {
	payload, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(dst), ".tmp-*.json")
	if err != nil {
		return err
	}
	defer func() {
		_ = os.Remove(tmp.Name())
	}()
	if _, err := tmp.Write(payload); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmp.Name(), dst)
}

// LoadViewStatus reads view.status.json if present.
func LoadViewStatus(managedDir string) (ViewStatus, error) {
	statusPath := filepath.Join(managedDir, viewStatusFileName)
	payload, err := os.ReadFile(statusPath)
	if err != nil {
		if os.IsNotExist(err) {
			return ViewStatus{}, nil
		}
		return ViewStatus{}, err
	}
	var status ViewStatus
	if err := json.Unmarshal(payload, &status); err != nil {
		return ViewStatus{}, err
	}
	return status, nil
}

// WriteView writes view.json and updates status, keeping a previous snapshot if present.
func WriteView(managedDir string, view map[string]any, pending bool) (ViewStatus, error) {
	if managedDir == "" {
		return ViewStatus{}, errors.New("managed dir required")
	}
	if err := os.MkdirAll(managedDir, 0o755); err != nil {
		return ViewStatus{}, err
	}
	viewPath := filepath.Join(managedDir, viewFileName)
	prevPath := filepath.Join(managedDir, viewPreviousName)
	if _, err := os.Stat(viewPath); err == nil {
		_ = copyFile(viewPath, prevPath)
	}
	if err := writeJSONAtomic(viewPath, view); err != nil {
		return ViewStatus{}, err
	}
	status := ViewStatus{
		Pending:   pending,
		CurrentID: time.Now().UTC().Format(time.RFC3339Nano),
		UpdatedAt: time.Now().UTC(),
	}
	if prev, err := os.Stat(prevPath); err == nil && prev.Mode().IsRegular() {
		status.PreviousID = "previous"
	}
	if err := writeJSONAtomic(filepath.Join(managedDir, viewStatusFileName), status); err != nil {
		return ViewStatus{}, err
	}
	return status, nil
}

// AcceptView marks the current view as accepted.
func AcceptView(managedDir string) (ViewStatus, error) {
	status, err := LoadViewStatus(managedDir)
	if err != nil {
		return ViewStatus{}, err
	}
	status.Pending = false
	status.UpdatedAt = time.Now().UTC()
	if err := writeJSONAtomic(filepath.Join(managedDir, viewStatusFileName), status); err != nil {
		return ViewStatus{}, err
	}
	return status, nil
}

// RevertView restores the previous view snapshot if present.
func RevertView(managedDir string) (ViewStatus, error) {
	viewPath := filepath.Join(managedDir, viewFileName)
	prevPath := filepath.Join(managedDir, viewPreviousName)
	if _, err := os.Stat(prevPath); err != nil {
		return ViewStatus{}, err
	}
	if err := copyFile(prevPath, viewPath); err != nil {
		return ViewStatus{}, err
	}
	status := ViewStatus{
		Pending:   false,
		CurrentID: time.Now().UTC().Format(time.RFC3339Nano),
		UpdatedAt: time.Now().UTC(),
	}
	if err := writeJSONAtomic(filepath.Join(managedDir, viewStatusFileName), status); err != nil {
		return ViewStatus{}, err
	}
	return status, nil
}

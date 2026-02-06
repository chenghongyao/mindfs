package fs

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

// ViewVersionStore manages versioned views for a route
type ViewVersionStore struct {
	managedDir string
	routeID    string
}

// ViewVersionMeta contains metadata for a view version
type ViewVersionMeta struct {
	Version   string    `json:"version"`
	Prompt    string    `json:"prompt,omitempty"`
	Agent     string    `json:"agent,omitempty"`
	Parent    string    `json:"parent,omitempty"`
	SessionID string    `json:"session_id,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

// NewViewVersionStore creates a new version store for a route
func NewViewVersionStore(managedDir, routeID string) *ViewVersionStore {
	return &ViewVersionStore{
		managedDir: managedDir,
		routeID:    routeID,
	}
}

// viewsDir returns the views directory path
func (s *ViewVersionStore) viewsDir() string {
	return filepath.Join(s.managedDir, "views", s.routeID)
}

// versionPath returns the path for a version file
func (s *ViewVersionStore) versionPath(version string) string {
	return filepath.Join(s.viewsDir(), version+".json")
}

// metaPath returns the path for a version meta file
func (s *ViewVersionStore) metaPath(version string) string {
	return filepath.Join(s.viewsDir(), version+".meta.json")
}

// ListVersions returns all available versions sorted
func (s *ViewVersionStore) ListVersions() ([]string, error) {
	entries, err := os.ReadDir(s.viewsDir())
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var versions []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.HasSuffix(name, ".json") && !strings.HasSuffix(name, ".meta.json") {
			version := strings.TrimSuffix(name, ".json")
			versions = append(versions, version)
		}
	}

	// Sort versions (v1, v2, v3... or by number)
	sort.Slice(versions, func(i, j int) bool {
		// Try numeric sort first
		ni := extractVersionNumber(versions[i])
		nj := extractVersionNumber(versions[j])
		if ni > 0 && nj > 0 {
			return ni < nj
		}
		return versions[i] < versions[j]
	})

	return versions, nil
}

// extractVersionNumber extracts number from version string like "v1", "v2"
func extractVersionNumber(version string) int {
	v := strings.TrimPrefix(version, "v")
	n, err := strconv.Atoi(v)
	if err != nil {
		return 0
	}
	return n
}

// GetLatestVersion returns the latest version
func (s *ViewVersionStore) GetLatestVersion() (string, error) {
	versions, err := s.ListVersions()
	if err != nil {
		return "", err
	}
	if len(versions) == 0 {
		return "", nil
	}
	return versions[len(versions)-1], nil
}

// NextVersion returns the next version number
func (s *ViewVersionStore) NextVersion() string {
	latest, _ := s.GetLatestVersion()
	if latest == "" {
		return "v1"
	}
	n := extractVersionNumber(latest)
	if n > 0 {
		return fmt.Sprintf("v%d", n+1)
	}
	return fmt.Sprintf("%s-1", latest)
}

// Load loads a specific version
func (s *ViewVersionStore) Load(version string) (map[string]any, error) {
	data, err := os.ReadFile(s.versionPath(version))
	if err != nil {
		return nil, err
	}

	var view map[string]any
	if err := json.Unmarshal(data, &view); err != nil {
		return nil, err
	}

	return view, nil
}

// LoadMeta loads metadata for a version
func (s *ViewVersionStore) LoadMeta(version string) (*ViewVersionMeta, error) {
	data, err := os.ReadFile(s.metaPath(version))
	if err != nil {
		if os.IsNotExist(err) {
			return &ViewVersionMeta{Version: version}, nil
		}
		return nil, err
	}

	var meta ViewVersionMeta
	if err := json.Unmarshal(data, &meta); err != nil {
		return nil, err
	}

	return &meta, nil
}

// Save saves a new version with metadata
func (s *ViewVersionStore) Save(version string, view map[string]any, meta *ViewVersionMeta) error {
	// Ensure directory exists
	if err := os.MkdirAll(s.viewsDir(), 0755); err != nil {
		return err
	}

	// Save view
	if err := writeJSONAtomic(s.versionPath(version), view); err != nil {
		return err
	}

	// Save metadata
	if meta != nil {
		meta.Version = version
		if meta.CreatedAt.IsZero() {
			meta.CreatedAt = time.Now().UTC()
		}
		if err := writeJSONAtomic(s.metaPath(version), meta); err != nil {
			return err
		}
	}

	return nil
}

// SaveNew saves a new version with auto-generated version number
func (s *ViewVersionStore) SaveNew(view map[string]any, meta *ViewVersionMeta) (string, error) {
	version := s.NextVersion()
	if err := s.Save(version, view, meta); err != nil {
		return "", err
	}
	return version, nil
}

// Delete removes a version
func (s *ViewVersionStore) Delete(version string) error {
	_ = os.Remove(s.versionPath(version))
	_ = os.Remove(s.metaPath(version))
	return nil
}

// GetVersionInfo returns version info with metadata
func (s *ViewVersionStore) GetVersionInfo(version string) (*ViewVersionMeta, error) {
	meta, err := s.LoadMeta(version)
	if err != nil {
		return nil, err
	}

	// If no meta file, get info from file stat
	if meta.CreatedAt.IsZero() {
		if info, err := os.Stat(s.versionPath(version)); err == nil {
			meta.CreatedAt = info.ModTime()
		}
	}

	return meta, nil
}

// ListVersionsWithMeta returns all versions with their metadata
func (s *ViewVersionStore) ListVersionsWithMeta() ([]ViewVersionMeta, error) {
	versions, err := s.ListVersions()
	if err != nil {
		return nil, err
	}

	var result []ViewVersionMeta
	for _, v := range versions {
		meta, err := s.GetVersionInfo(v)
		if err != nil {
			meta = &ViewVersionMeta{Version: v}
		}
		result = append(result, *meta)
	}

	return result, nil
}

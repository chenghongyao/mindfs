package fs

import (
	"os"
	"path/filepath"
	"sort"
)

// Entry represents a filesystem entry for UI listings.
type Entry struct {
	Name  string `json:"name"`
	Path  string `json:"path"`
	IsDir bool   `json:"is_dir"`
}

// ListEntries returns entries for a directory with paths relative to root.
func ListEntries(root, dir string) ([]Entry, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	result := make([]Entry, 0, len(entries))
	for _, entry := range entries {
		name := entry.Name()
		absPath := filepath.Join(dir, name)
		relPath, err := filepath.Rel(root, absPath)
		if err != nil {
			return nil, err
		}
		result = append(result, Entry{
			Name:  name,
			Path:  relPath,
			IsDir: entry.IsDir(),
		})
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].IsDir != result[j].IsDir {
			return result[i].IsDir
		}
		return result[i].Name < result[j].Name
	})
	return result, nil
}

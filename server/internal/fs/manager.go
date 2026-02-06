package fs

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"time"
)

const (
	managedDirName  = ".mindfs"
	managedMetaFile = "managed.json"
)

// ManagedMeta captures minimal metadata about a managed directory.
type ManagedMeta struct {
	RootPath  string    `json:"root_path"`
	CreatedAt time.Time `json:"created_at"`
}

// EnsureManagedDir creates the managed directory metadata folder and file if needed.
func EnsureManagedDir(rootPath string) (string, error) {
	if rootPath == "" {
		return "", errors.New("root path required")
	}
	metaDir := filepath.Join(rootPath, managedDirName)
	if err := os.MkdirAll(metaDir, 0o755); err != nil {
		return "", err
	}
	metaPath := filepath.Join(metaDir, managedMetaFile)
	if _, err := os.Stat(metaPath); err == nil {
		return metaDir, nil
	}
	meta := ManagedMeta{RootPath: rootPath, CreatedAt: time.Now().UTC()}
	payload, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return "", err
	}
	if err := os.WriteFile(metaPath, payload, 0o644); err != nil {
		return "", err
	}
	return metaDir, nil
}

// IsManagedDir checks whether a directory is already managed.
func IsManagedDir(rootPath string) (bool, error) {
	if rootPath == "" {
		return false, errors.New("root path required")
	}
	metaPath := filepath.Join(rootPath, managedDirName, managedMetaFile)
	_, err := os.Stat(metaPath)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

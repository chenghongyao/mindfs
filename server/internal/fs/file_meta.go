package fs

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"time"
)

type FileMetaEntry struct {
	SourceSession string    `json:"source_session"`
	SessionName   string    `json:"session_name,omitempty"`
	Agent         string    `json:"agent,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at,omitempty"`
	CreatedBy     string    `json:"created_by"`
}

type FileMeta map[string]FileMetaEntry

func LoadFileMeta(managedDir string) (FileMeta, error) {
	if managedDir == "" {
		return nil, errors.New("managed dir required")
	}
	path := filepath.Join(managedDir, "file-meta.json")
	payload, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return FileMeta{}, nil
		}
		return nil, err
	}
	var meta FileMeta
	if err := json.Unmarshal(payload, &meta); err != nil {
		return nil, err
	}
	if meta == nil {
		meta = FileMeta{}
	}
	return meta, nil
}

func SaveFileMeta(managedDir string, meta FileMeta) error {
	if managedDir == "" {
		return errors.New("managed dir required")
	}
	if meta == nil {
		meta = FileMeta{}
	}
	payload, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(managedDir, 0o755); err != nil {
		return err
	}
	path := filepath.Join(managedDir, "file-meta.json")
	return os.WriteFile(path, payload, 0o644)
}

func UpdateFileMeta(managedDir, relativePath, sessionKey, createdBy string) error {
	if relativePath == "" {
		return errors.New("path required")
	}
	meta, err := LoadFileMeta(managedDir)
	if err != nil {
		return err
	}
	meta[relativePath] = FileMetaEntry{
		SourceSession: sessionKey,
		CreatedAt:     time.Now().UTC(),
		CreatedBy:     createdBy,
	}
	return SaveFileMeta(managedDir, meta)
}

// UpdateFileMetaFull updates file metadata with full session info
func UpdateFileMetaFull(managedDir, relativePath, sessionKey, sessionName, agent, createdBy string) error {
	if relativePath == "" {
		return errors.New("path required")
	}
	meta, err := LoadFileMeta(managedDir)
	if err != nil {
		return err
	}

	now := time.Now().UTC()
	entry := FileMetaEntry{
		SourceSession: sessionKey,
		SessionName:   sessionName,
		Agent:         agent,
		CreatedAt:     now,
		UpdatedAt:     now,
		CreatedBy:     createdBy,
	}

	// Preserve original creation time if updating
	if existing, ok := meta[relativePath]; ok {
		entry.CreatedAt = existing.CreatedAt
	}

	meta[relativePath] = entry
	return SaveFileMeta(managedDir, meta)
}

// GetFileMeta returns metadata for a specific file
func GetFileMeta(managedDir, relativePath string) (*FileMetaEntry, error) {
	meta, err := LoadFileMeta(managedDir)
	if err != nil {
		return nil, err
	}

	if entry, ok := meta[relativePath]; ok {
		return &entry, nil
	}
	return nil, nil
}

// GetFilesBySession returns all files created by a session
func GetFilesBySession(managedDir, sessionKey string) ([]string, error) {
	meta, err := LoadFileMeta(managedDir)
	if err != nil {
		return nil, err
	}

	var files []string
	for path, entry := range meta {
		if entry.SourceSession == sessionKey {
			files = append(files, path)
		}
	}
	return files, nil
}

// DeleteFileMeta removes metadata for a file
func DeleteFileMeta(managedDir, relativePath string) error {
	meta, err := LoadFileMeta(managedDir)
	if err != nil {
		return err
	}

	delete(meta, relativePath)
	return SaveFileMeta(managedDir, meta)
}

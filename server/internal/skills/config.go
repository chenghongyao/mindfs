package skills

import (
	"encoding/json"
	"os"
	"path/filepath"
)

const defaultConfigName = "config.json"

// DirConfig stores per-directory preferences.
type DirConfig struct {
	DefaultAgent    string `json:"defaultAgent"`
	UserDescription string `json:"userDescription"`
}

// LoadDirConfig reads .mindfs/config.json if present.
func LoadDirConfig(managedDir string) (DirConfig, error) {
	if managedDir == "" {
		return DirConfig{}, nil
	}
	path := filepath.Join(managedDir, defaultConfigName)
	payload, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return DirConfig{}, nil
		}
		return DirConfig{}, err
	}
	var cfg DirConfig
	if err := json.Unmarshal(payload, &cfg); err != nil {
		return DirConfig{}, err
	}
	return cfg, nil
}

// SaveDirConfig writes .mindfs/config.json with provided settings.
func SaveDirConfig(managedDir string, cfg DirConfig) error {
	if managedDir == "" {
		return nil
	}
	path := filepath.Join(managedDir, defaultConfigName)
	payload, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(managedDir, 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, payload, 0o644)
}

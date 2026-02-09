package config

import (
	"os"
	"path/filepath"
)

// MindFSConfigDir returns the user-level config directory for MindFS.
// Example: ~/.config/mindfs (Linux/macOS), %AppData%/mindfs (Windows).
func MindFSConfigDir() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, "mindfs"), nil
}

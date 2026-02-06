package router

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Shortcut defines a quick action button
type Shortcut struct {
	ID       string         `json:"id"`
	Label    string         `json:"label"`
	Action   string         `json:"action"`
	Position string         `json:"position,omitempty"` // left, center, right
	Type     string         `json:"type,omitempty"`     // button, text
	Icon     string         `json:"icon,omitempty"`
	Params   map[string]any `json:"params,omitempty"`
	Disabled bool           `json:"disabled,omitempty"`
	Hidden   bool           `json:"hidden,omitempty"`
}

// ShortcutsConfig holds shortcuts from view.json
type ShortcutsConfig struct {
	Shortcuts []Shortcut `json:"shortcuts"`
}

// LoadShortcuts loads shortcuts from view.json
func LoadShortcuts(managedDir string) ([]Shortcut, error) {
	viewPath := filepath.Join(managedDir, "view.json")
	data, err := os.ReadFile(viewPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var view map[string]any
	if err := json.Unmarshal(data, &view); err != nil {
		return nil, err
	}

	// Extract shortcuts from view
	shortcutsRaw, ok := view["shortcuts"]
	if !ok {
		return nil, nil
	}

	// Re-marshal and unmarshal to get typed shortcuts
	shortcutsData, err := json.Marshal(shortcutsRaw)
	if err != nil {
		return nil, err
	}

	var shortcuts []Shortcut
	if err := json.Unmarshal(shortcutsData, &shortcuts); err != nil {
		return nil, err
	}

	// Set defaults
	for i := range shortcuts {
		if shortcuts[i].Position == "" {
			shortcuts[i].Position = "center"
		}
		if shortcuts[i].Type == "" {
			shortcuts[i].Type = "button"
		}
	}

	return shortcuts, nil
}

// LoadShortcutsFromView extracts shortcuts from a view map
func LoadShortcutsFromView(view map[string]any) []Shortcut {
	shortcutsRaw, ok := view["shortcuts"]
	if !ok {
		return nil
	}

	shortcutsData, err := json.Marshal(shortcutsRaw)
	if err != nil {
		return nil
	}

	var shortcuts []Shortcut
	if err := json.Unmarshal(shortcutsData, &shortcuts); err != nil {
		return nil
	}

	// Set defaults
	for i := range shortcuts {
		if shortcuts[i].Position == "" {
			shortcuts[i].Position = "center"
		}
		if shortcuts[i].Type == "" {
			shortcuts[i].Type = "button"
		}
	}

	return shortcuts
}

// GroupShortcutsByPosition groups shortcuts by their position
func GroupShortcutsByPosition(shortcuts []Shortcut) map[string][]Shortcut {
	result := map[string][]Shortcut{
		"left":   {},
		"center": {},
		"right":  {},
	}

	for _, s := range shortcuts {
		if s.Hidden {
			continue
		}
		pos := s.Position
		if pos != "left" && pos != "right" {
			pos = "center"
		}
		result[pos] = append(result[pos], s)
	}

	return result
}

package router

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
)

// ViewRoute defines a view routing rule
type ViewRoute struct {
	ID       string     `json:"id"`
	Name     string     `json:"name"`
	Match    MatchRule  `json:"match"`
	View     string     `json:"view,omitempty"`     // view file path or inline
	ViewData any        `json:"view_data,omitempty"` // inline view data
	Priority int        `json:"priority"`           // higher = more specific
	Default  bool       `json:"default,omitempty"`  // is default view
}

// MatchRule defines how to match files/paths
type MatchRule struct {
	Path string      `json:"path,omitempty"` // glob pattern
	Ext  string      `json:"ext,omitempty"`  // extension (e.g., ".md")
	Mime string      `json:"mime,omitempty"` // mime type pattern
	Name string      `json:"name,omitempty"` // filename pattern
	Any  []MatchRule `json:"any,omitempty"`  // OR combination
	All  []MatchRule `json:"all,omitempty"`  // AND combination
}

// ViewRouterConfig holds all view routes
type ViewRouterConfig struct {
	Routes []ViewRoute `json:"routes"`
}

// LoadViewRouterConfig loads view router config from .mindfs/view-routes.json
func LoadViewRouterConfig(managedDir string) (*ViewRouterConfig, error) {
	configPath := filepath.Join(managedDir, "view-routes.json")
	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Return default config
			return defaultViewRouterConfig(), nil
		}
		return nil, err
	}

	var config ViewRouterConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

// SaveViewRouterConfig saves view router config
func SaveViewRouterConfig(managedDir string, config *ViewRouterConfig) error {
	configPath := filepath.Join(managedDir, "view-routes.json")
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(configPath, data, 0644)
}

// defaultViewRouterConfig returns default view routes
func defaultViewRouterConfig() *ViewRouterConfig {
	return &ViewRouterConfig{
		Routes: []ViewRoute{
			{
				ID:       "_default",
				Name:     "默认视图",
				Match:    MatchRule{Path: "**/*"},
				Priority: 0,
				Default:  true,
			},
		},
	}
}

// GetMatchingRoutes returns all routes that match the given path
func (c *ViewRouterConfig) GetMatchingRoutes(path string) []ViewRoute {
	var matches []ViewRoute

	for _, route := range c.Routes {
		if matchesRule(path, route.Match) {
			matches = append(matches, route)
		}
	}

	// Sort by priority (higher first)
	sort.Slice(matches, func(i, j int) bool {
		return matches[i].Priority > matches[j].Priority
	})

	return matches
}

// GetRouteByID returns a route by its ID
func (c *ViewRouterConfig) GetRouteByID(id string) *ViewRoute {
	for i := range c.Routes {
		if c.Routes[i].ID == id {
			return &c.Routes[i]
		}
	}
	return nil
}

// AddRoute adds a new route
func (c *ViewRouterConfig) AddRoute(route ViewRoute) {
	c.Routes = append(c.Routes, route)
}

// UpdateRoute updates an existing route
func (c *ViewRouterConfig) UpdateRoute(route ViewRoute) bool {
	for i := range c.Routes {
		if c.Routes[i].ID == route.ID {
			c.Routes[i] = route
			return true
		}
	}
	return false
}

// RemoveRoute removes a route by ID
func (c *ViewRouterConfig) RemoveRoute(id string) bool {
	for i := range c.Routes {
		if c.Routes[i].ID == id {
			c.Routes = append(c.Routes[:i], c.Routes[i+1:]...)
			return true
		}
	}
	return false
}

package router

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
)

// ResolvedView represents a resolved view with its metadata
type ResolvedView struct {
	RouteID   string         `json:"route_id"`
	RouteName string         `json:"route_name"`
	Priority  int            `json:"priority"`
	IsDefault bool           `json:"is_default"`
	ViewData  map[string]any `json:"view_data,omitempty"`
	Versions  []string       `json:"versions,omitempty"`
	Active    string         `json:"active,omitempty"`
}

// ViewResolver resolves views for paths
type ViewResolver struct {
	managedDir string
	config     *ViewRouterConfig
	status     *ViewStatus
}

// NewViewResolver creates a new view resolver
func NewViewResolver(managedDir string) (*ViewResolver, error) {
	config, err := LoadViewRouterConfig(managedDir)
	if err != nil {
		return nil, err
	}

	status, err := LoadViewStatus(managedDir)
	if err != nil {
		// Use empty status if not found
		status = &ViewStatus{
			ActiveVersions: make(map[string]string),
			LastSelected:   make(map[string]string),
		}
	}

	return &ViewResolver{
		managedDir: managedDir,
		config:     config,
		status:     status,
	}, nil
}

// ResolveViews returns all matching views for a path, sorted by priority
func (r *ViewResolver) ResolveViews(path string) []ResolvedView {
	routes := r.config.GetMatchingRoutes(path)
	var views []ResolvedView

	for _, route := range routes {
		view := ResolvedView{
			RouteID:   route.ID,
			RouteName: route.Name,
			Priority:  route.Priority,
			IsDefault: route.Default,
		}

		// Load view data if specified
		if route.ViewData != nil {
			if data, ok := route.ViewData.(map[string]any); ok {
				view.ViewData = data
			}
		} else if route.View != "" {
			viewData, _ := r.loadViewFile(route.View)
			view.ViewData = viewData
		}

		// Get versions for this route
		view.Versions = r.getVersions(route.ID)
		view.Active = r.status.ActiveVersions[route.ID]

		views = append(views, view)
	}

	return views
}

// ResolvePreferred returns the preferred view for a path
func (r *ViewResolver) ResolvePreferred(path string) *ResolvedView {
	views := r.ResolveViews(path)
	if len(views) == 0 {
		return nil
	}

	// Check if user has a preference for this path
	if preferred, ok := r.status.LastSelected[path]; ok {
		for i := range views {
			if views[i].RouteID == preferred {
				return &views[i]
			}
		}
	}

	// Return highest priority view
	return &views[0]
}

// SetPreference saves user's view preference for a path
func (r *ViewResolver) SetPreference(path, routeID string) error {
	r.status.LastSelected[path] = routeID
	return SaveViewStatus(r.managedDir, r.status)
}

// SetActiveVersion sets the active version for a route
func (r *ViewResolver) SetActiveVersion(routeID, version string) error {
	r.status.ActiveVersions[routeID] = version
	return SaveViewStatus(r.managedDir, r.status)
}

// loadViewFile loads view data from a file
func (r *ViewResolver) loadViewFile(viewPath string) (map[string]any, error) {
	fullPath := viewPath
	if !filepath.IsAbs(viewPath) {
		fullPath = filepath.Join(r.managedDir, "views", viewPath)
	}

	data, err := os.ReadFile(fullPath)
	if err != nil {
		return nil, err
	}

	var viewData map[string]any
	if err := json.Unmarshal(data, &viewData); err != nil {
		return nil, err
	}

	return viewData, nil
}

// getVersions returns available versions for a route
func (r *ViewResolver) getVersions(routeID string) []string {
	versionsDir := filepath.Join(r.managedDir, "views", routeID)
	entries, err := os.ReadDir(versionsDir)
	if err != nil {
		return nil
	}

	var versions []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		ext := filepath.Ext(name)
		if ext == ".json" && !filepath.HasPrefix(name, ".") {
			version := name[:len(name)-len(ext)]
			if version != "" && !filepath.HasPrefix(version, "_") {
				versions = append(versions, version)
			}
		}
	}

	// Sort versions (v1, v2, v3...)
	sort.Strings(versions)
	return versions
}

// Reload reloads config and status
func (r *ViewResolver) Reload() error {
	config, err := LoadViewRouterConfig(r.managedDir)
	if err != nil {
		return err
	}
	r.config = config

	status, err := LoadViewStatus(r.managedDir)
	if err != nil {
		status = &ViewStatus{
			ActiveVersions: make(map[string]string),
			LastSelected:   make(map[string]string),
		}
	}
	r.status = status

	return nil
}

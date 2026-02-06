package api

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"mindfs/server/internal/fs"
	"mindfs/server/internal/router"
)

// ViewHandler handles view-related API requests
type ViewHandler struct {
	Root     string
	Registry *fs.Registry
}

// Routes returns the view routes
func (h *ViewHandler) Routes() http.Handler {
	r := chi.NewRouter()
	r.Get("/routes", h.handleRoutes)
	r.Post("/switch", h.handleSwitch)
	r.Post("/preference", h.handlePreference)
	r.Get("/versions/{routeId}", h.handleVersions)
	r.Post("/generate", h.handleGenerate)
	return r
}

// handleRoutes returns matching view routes for a path
func (h *ViewHandler) handleRoutes(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	rootID := r.URL.Query().Get("root")
	resolved, err := resolveRoot(rootID, h.Root, h.Registry)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]any{"error": err.Error()})
		return
	}

	path := r.URL.Query().Get("path")

	resolver, err := router.NewViewResolver(resolved.ManagedDir)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]any{"error": err.Error()})
		return
	}

	var views []router.ResolvedView
	if path != "" {
		views = resolver.ResolveViews(path)
	} else {
		// Return all routes if no path specified
		config, _ := router.LoadViewRouterConfig(resolved.ManagedDir)
		if config != nil {
			for _, route := range config.Routes {
				views = append(views, router.ResolvedView{
					RouteID:   route.ID,
					RouteName: route.Name,
					Priority:  route.Priority,
					IsDefault: route.Default,
				})
			}
		}
	}

	// Convert to response format
	routes := make([]map[string]any, 0, len(views))
	for _, v := range views {
		routes = append(routes, map[string]any{
			"route_id":   v.RouteID,
			"route_name": v.RouteName,
			"priority":   v.Priority,
			"is_default": v.IsDefault,
			"versions":   v.Versions,
			"active":     v.Active,
		})
	}

	_ = json.NewEncoder(w).Encode(map[string]any{"routes": routes})
}

// handleSwitch switches view route or version
func (h *ViewHandler) handleSwitch(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req struct {
		RootID  string `json:"root_id"`
		RouteID string `json:"route_id"`
		Version string `json:"version"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]any{"error": "invalid json"})
		return
	}

	resolved, err := resolveRoot(req.RootID, h.Root, h.Registry)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]any{"error": err.Error()})
		return
	}

	resolver, err := router.NewViewResolver(resolved.ManagedDir)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]any{"error": err.Error()})
		return
	}

	// Set active version if specified
	if req.Version != "" {
		if err := resolver.SetActiveVersion(req.RouteID, req.Version); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(map[string]any{"error": err.Error()})
			return
		}
	}

	_ = json.NewEncoder(w).Encode(map[string]any{"status": "ok"})
}

// handlePreference saves user view preference
func (h *ViewHandler) handlePreference(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req struct {
		RootID  string `json:"root_id"`
		Path    string `json:"path"`
		RouteID string `json:"route_id"`
		Version string `json:"version"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]any{"error": "invalid json"})
		return
	}

	resolved, err := resolveRoot(req.RootID, h.Root, h.Registry)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]any{"error": err.Error()})
		return
	}

	prefs, err := router.NewViewPreferences(resolved.ManagedDir)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]any{"error": err.Error()})
		return
	}

	if err := prefs.Set(req.Path, req.RouteID, req.Version); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]any{"error": err.Error()})
		return
	}

	_ = json.NewEncoder(w).Encode(map[string]any{"status": "ok"})
}

// handleVersions returns versions for a route
func (h *ViewHandler) handleVersions(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	routeID := chi.URLParam(r, "routeId")
	rootID := r.URL.Query().Get("root")

	resolved, err := resolveRoot(rootID, h.Root, h.Registry)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]any{"error": err.Error()})
		return
	}

	store := fs.NewViewVersionStore(resolved.ManagedDir, routeID)
	versions, err := store.ListVersionsWithMeta()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]any{"error": err.Error()})
		return
	}

	// Convert to response format
	result := make([]map[string]any, 0, len(versions))
	for _, v := range versions {
		result = append(result, map[string]any{
			"version":    v.Version,
			"prompt":     v.Prompt,
			"agent":      v.Agent,
			"parent":     v.Parent,
			"created_at": v.CreatedAt,
		})
	}

	_ = json.NewEncoder(w).Encode(map[string]any{"versions": result})
}

// handleGenerate triggers view generation
func (h *ViewHandler) handleGenerate(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req struct {
		RootID      string `json:"root_id"`
		RouteID     string `json:"route_id"`
		Prompt      string `json:"prompt"`
		BaseVersion string `json:"base_version"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]any{"error": "invalid json"})
		return
	}

	resolved, err := resolveRoot(req.RootID, h.Root, h.Registry)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]any{"error": err.Error()})
		return
	}

	// Set pending status
	status, _ := router.LoadViewStatus(resolved.ManagedDir)
	status.SetPending(&router.PendingView{
		RouteID: req.RouteID,
	})
	_ = router.SaveViewStatus(resolved.ManagedDir, status)

	// TODO: Actually trigger view generation via session
	// For now, just return a placeholder session key
	_ = json.NewEncoder(w).Encode(map[string]any{
		"status":      "pending",
		"session_key": "view-gen-" + req.RouteID,
	})
}

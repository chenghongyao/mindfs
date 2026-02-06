package api

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"mindfs/server/internal/agent"
)

// AgentHandler handles agent-related API requests.
type AgentHandler struct {
	Prober *agent.Prober
}

// Routes returns the agent API routes.
func (h *AgentHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.handleList)
	r.Post("/{name}/probe", h.handleProbe)
	return r
}

// handleList returns all agent statuses.
// GET /api/agents
func (h *AgentHandler) handleList(w http.ResponseWriter, r *http.Request) {
	if h.Prober == nil {
		writeJSON(w, http.StatusOK, []agent.Status{})
		return
	}
	statuses := h.Prober.GetAllStatuses()
	writeJSON(w, http.StatusOK, statuses)
}

// handleProbe triggers a probe for a specific agent.
// POST /api/agents/:name/probe
func (h *AgentHandler) handleProbe(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	if name == "" {
		writeError(w, http.StatusBadRequest, errInvalidRequest("agent name required"))
		return
	}
	if h.Prober == nil {
		writeError(w, http.StatusServiceUnavailable, errInvalidRequest("prober not available"))
		return
	}
	status := h.Prober.ProbeOne(r.Context(), name)
	writeJSON(w, http.StatusOK, status)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"mindfs/server/internal/skills"
)

func (h *HTTPHandler) handleSkillExecute(w http.ResponseWriter, r *http.Request) {
	rootID := r.URL.Query().Get("root")
	resolved, err := resolveRoot(rootID, h.Root, h.Registry)
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	skillID := chi.URLParam(r, "id")
	if strings.TrimSpace(skillID) == "" {
		writeError(w, http.StatusBadRequest, errInvalidRequest("skill id required"))
		return
	}
	var req struct {
		Params map[string]any `json:"params"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, errInvalidRequest("invalid json"))
		return
	}
	skill, err := skills.LoadSkill(resolved.ManagedDir, skillID)
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	result, err := skills.ExecuteSkill(r.Context(), skill, req.Params)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"result": result})
}

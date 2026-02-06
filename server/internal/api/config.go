package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"mindfs/server/internal/skills"
)

func (h *HTTPHandler) handleDirConfigGet(w http.ResponseWriter, r *http.Request) {
	rootID := chi.URLParam(r, "id")
	resolved, err := resolveRoot(rootID, h.Root, h.Registry)
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	cfg, err := skills.LoadDirConfig(resolved.ManagedDir)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(cfg)
}

func (h *HTTPHandler) handleDirConfigPut(w http.ResponseWriter, r *http.Request) {
	rootID := chi.URLParam(r, "id")
	resolved, err := resolveRoot(rootID, h.Root, h.Registry)
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	var cfg skills.DirConfig
	if err := json.NewDecoder(r.Body).Decode(&cfg); err != nil {
		writeError(w, http.StatusBadRequest, errInvalidRequest("invalid json"))
		return
	}
	cfg.UserDescription = strings.TrimSpace(cfg.UserDescription)
	cfg.DefaultAgent = strings.TrimSpace(cfg.DefaultAgent)
	if err := skills.SaveDirConfig(resolved.ManagedDir, cfg); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(cfg)
}

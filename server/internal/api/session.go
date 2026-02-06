package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"mindfs/server/internal/audit"
	"mindfs/server/internal/session"
)

type SessionService struct {
	Stores *session.StoreManager
}

func (h *HTTPHandler) getSessionManager(rootID string) (*session.Manager, error) {
	if h.Sessions == nil || h.Sessions.Stores == nil {
		return nil, errServiceUnavailable("session store not configured")
	}
	resolved, err := resolveRoot(rootID, h.Root, h.Registry)
	if err != nil {
		return nil, err
	}
	store, err := h.Sessions.Stores.Get(resolved.ManagedDir)
	if err != nil {
		return nil, err
	}
	return session.NewManager(store), nil
}

func (h *HTTPHandler) handleSessions(w http.ResponseWriter, r *http.Request) {
	rootID := r.URL.Query().Get("root")
	manager, err := h.getSessionManager(rootID)
	if err != nil {
		writeError(w, http.StatusServiceUnavailable, err)
		return
	}
	sessions, err := manager.List(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	payload := make([]map[string]any, 0, len(sessions))
	for _, s := range sessions {
		payload = append(payload, sessionResponse(s))
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"sessions": payload})
}

func (h *HTTPHandler) handleSessionGet(w http.ResponseWriter, r *http.Request) {
	rootID := r.URL.Query().Get("root")
	manager, err := h.getSessionManager(rootID)
	if err != nil {
		writeError(w, http.StatusServiceUnavailable, err)
		return
	}
	key := chi.URLParam(r, "key")
	if strings.TrimSpace(key) == "" {
		writeError(w, http.StatusBadRequest, errInvalidRequest("session key required"))
		return
	}
	sessionItem, err := manager.Get(r.Context(), key)
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"session": sessionResponse(sessionItem)})
}

func (h *HTTPHandler) handleSessionCreate(w http.ResponseWriter, r *http.Request) {
	rootID := r.URL.Query().Get("root")
	manager, err := h.getSessionManager(rootID)
	if err != nil {
		writeError(w, http.StatusServiceUnavailable, err)
		return
	}
	var req struct {
		Key   string `json:"key"`
		Type  string `json:"type"`
		Agent string `json:"agent"`
		Name  string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, errInvalidRequest("invalid json"))
		return
	}
	created, err := manager.Create(r.Context(), session.CreateInput{
		Key:   req.Key,
		Type:  req.Type,
		Agent: req.Agent,
		Name:  req.Name,
	})
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	// Audit log
	if logger := h.getAuditLogger(rootID); logger != nil {
		_ = logger.LogSession(audit.ActionSessionCreate, audit.ActorUser, created.Key, created.Agent, map[string]any{
			"type": created.Type,
			"name": created.Name,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"session": sessionResponse(created)})
}

func (h *HTTPHandler) handleSessionMessage(w http.ResponseWriter, r *http.Request) {
	rootID := r.URL.Query().Get("root")
	manager, err := h.getSessionManager(rootID)
	if err != nil {
		writeError(w, http.StatusServiceUnavailable, err)
		return
	}
	key := chi.URLParam(r, "key")
	if strings.TrimSpace(key) == "" {
		writeError(w, http.StatusBadRequest, errInvalidRequest("session key required"))
		return
	}
	var req struct {
		Content string         `json:"content"`
		Context map[string]any `json:"context"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, errInvalidRequest("invalid json"))
		return
	}
	if strings.TrimSpace(req.Content) == "" {
		writeError(w, http.StatusBadRequest, errInvalidRequest("content required"))
		return
	}
	updated, err := manager.AddExchange(r.Context(), key, "user", req.Content)
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}

	// Audit log
	if logger := h.getAuditLogger(rootID); logger != nil {
		_ = logger.LogSession(audit.ActionSessionMessage, audit.ActorUser, key, updated.Agent, map[string]any{
			"content_length": len(req.Content),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"session": sessionResponse(updated)})
}

func sessionResponse(s *session.Session) map[string]any {
	if s == nil {
		return map[string]any{}
	}
	return map[string]any{
		"key":              s.Key,
		"session_key":      s.Key,
		"type":             s.Type,
		"agent":            s.Agent,
		"agent_session_id": s.AgentSessionID,
		"name":             s.Name,
		"status":           s.Status,
		"summary":           s.Summary,
		"exchanges":         s.Exchanges,
		"related_files":     s.RelatedFiles,
		"generated_view":    s.GeneratedView,
		"created_at":        s.CreatedAt,
		"updated_at":        s.UpdatedAt,
		"closed_at":         s.ClosedAt,
	}
}

func writeError(w http.ResponseWriter, status int, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]any{"error": err.Error()})
}

type apiError struct {
	message string
}

func (e apiError) Error() string {
	return e.message
}

func errInvalidRequest(message string) error {
	return apiError{message: message}
}

func errServiceUnavailable(message string) error {
	return apiError{message: message}
}

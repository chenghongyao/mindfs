package api

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"mindfs/server/internal/audit"
	"mindfs/server/internal/fs"
	"mindfs/server/internal/session"
)

// SessionService 提供统一的 session 操作，供 HTTP 和 WebSocket handler 共用
type SessionService struct {
	Stores   *session.StoreManager
	Root     string
	Registry *fs.Registry
	Audit    *audit.WriterPool
}

// ResolvedRoot 包含解析后的 root 路径信息
type ResolvedRoot struct {
	RootID     string
	Path       string
	ManagedDir string
}

// Resolve 解析 rootID 到具体路径
func (s *SessionService) Resolve(rootID string) (*ResolvedRoot, error) {
	resolved, err := resolveRoot(rootID, s.Root, s.Registry)
	if err != nil {
		return nil, err
	}
	return &ResolvedRoot{
		RootID:     rootID,
		Path:       resolved.Path,
		ManagedDir: resolved.ManagedDir,
	}, nil
}

// GetManager 获取指定 root 的 session manager
func (s *SessionService) GetManager(rootID string) (*session.Manager, *ResolvedRoot, error) {
	if s.Stores == nil {
		return nil, nil, errServiceUnavailable("session store not configured")
	}
	resolved, err := s.Resolve(rootID)
	if err != nil {
		return nil, nil, err
	}
	store, err := s.Stores.Get(resolved.ManagedDir)
	if err != nil {
		return nil, nil, err
	}
	return session.NewManager(store), resolved, nil
}

// GetAuditLogger 获取审计日志记录器
func (s *SessionService) GetAuditLogger(rootID string) *audit.Logger {
	if s.Audit == nil {
		return nil
	}
	resolved, err := s.Resolve(rootID)
	if err != nil {
		return nil
	}
	return audit.NewLogger(s.Audit, rootID, resolved.ManagedDir)
}

// CreateSession 创建新 session
func (s *SessionService) CreateSession(ctx context.Context, rootID string, input session.CreateInput) (*session.Session, *ResolvedRoot, error) {
	manager, resolved, err := s.GetManager(rootID)
	if err != nil {
		return nil, nil, err
	}
	created, err := manager.Create(ctx, input)
	if err != nil {
		return nil, resolved, err
	}

	// 审计日志
	if logger := s.GetAuditLogger(rootID); logger != nil {
		_ = logger.LogSession(audit.ActionSessionCreate, audit.ActorUser, created.Key, created.Agent, map[string]any{
			"type": created.Type,
			"name": created.Name,
		})
	}

	return created, resolved, nil
}

// AddMessage 添加消息到 session
func (s *SessionService) AddMessage(ctx context.Context, rootID, sessionKey, role, content string) (*session.Session, *ResolvedRoot, error) {
	manager, resolved, err := s.GetManager(rootID)
	if err != nil {
		return nil, nil, err
	}
	updated, err := manager.AddExchange(ctx, sessionKey, role, content)
	if err != nil {
		return nil, resolved, err
	}

	// 审计日志
	if logger := s.GetAuditLogger(rootID); logger != nil {
		_ = logger.LogSession(audit.ActionSessionMessage, audit.ActorUser, sessionKey, updated.Agent, map[string]any{
			"content_length": len(content),
			"role":           role,
		})
	}

	return updated, resolved, nil
}

// GetSession 获取 session
func (s *SessionService) GetSession(ctx context.Context, rootID, sessionKey string) (*session.Session, error) {
	manager, _, err := s.GetManager(rootID)
	if err != nil {
		return nil, err
	}
	return manager.Get(ctx, sessionKey)
}

// ListSessions 列出所有 session
func (s *SessionService) ListSessions(ctx context.Context, rootID string) ([]*session.Session, error) {
	manager, _, err := s.GetManager(rootID)
	if err != nil {
		return nil, err
	}
	return manager.List(ctx)
}

// CloseSession 关闭 session
func (s *SessionService) CloseSession(ctx context.Context, rootID, sessionKey string) (*session.Session, error) {
	manager, _, err := s.GetManager(rootID)
	if err != nil {
		return nil, err
	}
	closed, err := manager.Close(ctx, sessionKey)
	if err != nil {
		return nil, err
	}

	// 审计日志
	if logger := s.GetAuditLogger(rootID); logger != nil {
		_ = logger.LogSession(audit.ActionSessionClose, audit.ActorUser, sessionKey, closed.Agent, nil)
	}

	return closed, nil
}

// ============ HTTP Handlers ============

func (h *HTTPHandler) handleSessions(w http.ResponseWriter, r *http.Request) {
	rootID := r.URL.Query().Get("root")
	sessions, err := h.Sessions.ListSessions(r.Context(), rootID)
	if err != nil {
		writeError(w, http.StatusServiceUnavailable, err)
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
	key := chi.URLParam(r, "key")
	if strings.TrimSpace(key) == "" {
		writeError(w, http.StatusBadRequest, errInvalidRequest("session key required"))
		return
	}
	sessionItem, err := h.Sessions.GetSession(r.Context(), rootID, key)
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"session": sessionResponse(sessionItem)})
}

func (h *HTTPHandler) handleSessionCreate(w http.ResponseWriter, r *http.Request) {
	rootID := r.URL.Query().Get("root")
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
	created, _, err := h.Sessions.CreateSession(r.Context(), rootID, session.CreateInput{
		Key:   req.Key,
		Type:  req.Type,
		Agent: req.Agent,
		Name:  req.Name,
	})
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"session": sessionResponse(created)})
}

// ============ 辅助函数 ============

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
		"summary":          s.Summary,
		"exchanges":        s.Exchanges,
		"related_files":    s.RelatedFiles,
		"generated_view":   s.GeneratedView,
		"created_at":       s.CreatedAt,
		"updated_at":       s.UpdatedAt,
		"closed_at":        s.ClosedAt,
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

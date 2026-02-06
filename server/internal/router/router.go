package router

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
)

// ActionRequest is the internal action request shape.
type ActionRequest struct {
	Action  string
	Path    string
	Context map[string]any
	Meta    map[string]any
	Version string
	Root    string
}

// ActionResponse is the internal action response shape.
type ActionResponse struct {
	Status  string
	Handled bool
	Data    map[string]any
	View    map[string]any
	Effects []any
	Error   map[string]any
}

// ActionHandler handles action requests.
type ActionHandler func(ctx context.Context, req ActionRequest) (ActionResponse, error)

// Router dispatches actions by action+version.
type Router struct {
	mu             sync.RWMutex
	handlers       map[string]ActionHandler
	fallback       ActionHandler
	allowFallback  bool
}

// New creates a new Router.
func New() *Router {
	return &Router{
		handlers:      make(map[string]ActionHandler),
		allowFallback: true,
	}
}

// Register binds an action+version pair to a handler.
func (r *Router) Register(action, version string, handler ActionHandler) error {
	if action == "" || version == "" {
		return errors.New("action and version are required")
	}
	if handler == nil {
		return errors.New("handler is required")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	key := fmt.Sprintf("%s@%s", action, version)
	r.handlers[key] = handler
	return nil
}

// SetFallback sets the default handler when no action matches.
func (r *Router) SetFallback(handler ActionHandler) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.fallback = handler
}

// Dispatch routes an action to a handler or fallback.
func (r *Router) Dispatch(ctx context.Context, req ActionRequest) (ActionResponse, error) {
	if req.Action == "" || req.Version == "" {
		return ActionResponse{Status: "error", Handled: false, Error: map[string]any{"code": "bad_request", "message": "action and version are required"}}, nil
	}
	key := fmt.Sprintf("%s@%s", req.Action, req.Version)
	r.mu.RLock()
	handler, ok := r.handlers[key]
	fallback := r.fallback
	allowFallback := r.allowFallback
	hasAction := false
	if !ok {
		for k := range r.handlers {
			if strings.HasPrefix(k, req.Action+"@") {
				hasAction = true
				break
			}
		}
	}
	r.mu.RUnlock()
	if ok {
		return handler(ctx, req)
	}
	if hasAction {
		return ActionResponse{Status: "error", Handled: false, Error: map[string]any{"code": "version_mismatch", "message": "no handler for action/version"}}, nil
	}
	if allowFallback && fallback != nil {
		return fallback(ctx, req)
	}
	return ActionResponse{Status: "error", Handled: false, Error: map[string]any{"code": "not_found", "message": "no handler for action/version"}}, nil
}

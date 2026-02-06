package acp

import (
	"context"
	"sync"
	"time"
)

// PermissionHandler manages permission requests to the user.
type PermissionHandler struct {
	pending map[string]chan bool
	mu      sync.Mutex
	timeout time.Duration
}

// NewPermissionHandler creates a new handler.
func NewPermissionHandler(timeout time.Duration) *PermissionHandler {
	return &PermissionHandler{
		pending: make(map[string]chan bool),
		timeout: timeout,
	}
}

// Request sends a permission request and waits for response.
func (h *PermissionHandler) Request(ctx context.Context, req PermissionRequest) (bool, error) {
	h.mu.Lock()
	ch := make(chan bool, 1)
	h.pending[req.RequestID] = ch
	h.mu.Unlock()

	defer func() {
		h.mu.Lock()
		delete(h.pending, req.RequestID)
		h.mu.Unlock()
	}()

	timeoutCtx, cancel := context.WithTimeout(ctx, h.timeout)
	defer cancel()

	select {
	case granted := <-ch:
		return granted, nil
	case <-timeoutCtx.Done():
		return false, timeoutCtx.Err()
	}
}

// Respond handles a permission response from the user.
func (h *PermissionHandler) Respond(resp PermissionResponse) bool {
	h.mu.Lock()
	defer h.mu.Unlock()

	ch, ok := h.pending[resp.RequestID]
	if !ok {
		return false
	}

	select {
	case ch <- resp.Granted:
		return true
	default:
		return false
	}
}

// HasPending returns true if there are pending permission requests.
func (h *PermissionHandler) HasPending() bool {
	h.mu.Lock()
	defer h.mu.Unlock()
	return len(h.pending) > 0
}

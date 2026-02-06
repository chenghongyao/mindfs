package acp

import "time"

// TransportHandler defines agent-specific message handling behavior.
type TransportHandler interface {
	// HandleMessage processes an incoming ACP message.
	HandleMessage(msg ACPMessage) error

	// ShouldFilter returns true if the message should be filtered out.
	ShouldFilter(msg ACPMessage) bool

	// GetIdleTimeout returns the idle timeout for this agent.
	GetIdleTimeout() time.Duration
}

// DefaultTransportHandler provides standard message handling.
type DefaultTransportHandler struct {
	IdleTimeout time.Duration
	OnMessage   func(msg ACPMessage) error
}

// NewDefaultTransportHandler creates a handler with default settings.
func NewDefaultTransportHandler(onMessage func(ACPMessage) error) *DefaultTransportHandler {
	return &DefaultTransportHandler{
		IdleTimeout: 500 * time.Millisecond,
		OnMessage:   onMessage,
	}
}

// HandleMessage processes the message using the callback.
func (h *DefaultTransportHandler) HandleMessage(msg ACPMessage) error {
	if h.OnMessage != nil {
		return h.OnMessage(msg)
	}
	return nil
}

// ShouldFilter returns false - default handler passes all messages.
func (h *DefaultTransportHandler) ShouldFilter(msg ACPMessage) bool {
	return false
}

// GetIdleTimeout returns the configured idle timeout.
func (h *DefaultTransportHandler) GetIdleTimeout() time.Duration {
	return h.IdleTimeout
}

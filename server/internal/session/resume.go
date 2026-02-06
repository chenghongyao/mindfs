package session

import (
	"context"
	"errors"
)

type Resumer interface {
	Resume(ctx context.Context, session *Session) error
}

var ErrResumeFailed = errors.New("resume failed")

// FallbackResumer implements the two-step resume policy.
// PreferResume returns nil on success, otherwise Fallback is attempted.
type FallbackResumer struct {
	PreferResume func(ctx context.Context, session *Session) error
	Fallback     func(ctx context.Context, session *Session) error
}

func (r *FallbackResumer) Resume(ctx context.Context, session *Session) error {
	if session == nil {
		return errors.New("session required")
	}
	if session.AgentSessionID != nil && r.PreferResume != nil {
		if err := r.PreferResume(ctx, session); err == nil {
			return nil
		}
	}
	if r.Fallback != nil {
		if err := r.Fallback(ctx, session); err == nil {
			return nil
		}
	}
	return ErrResumeFailed
}

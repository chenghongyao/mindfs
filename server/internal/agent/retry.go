package agent

import (
	"context"
	"errors"
	"time"
)

// RetryConfig holds retry settings.
type RetryConfig struct {
	MaxAttempts int
	InitialWait time.Duration
	MaxWait     time.Duration
	Multiplier  float64
}

// DefaultRetryConfig returns default retry settings.
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxAttempts: 3,
		InitialWait: 1 * time.Second,
		MaxWait:     4 * time.Second,
		Multiplier:  2.0,
	}
}

// ErrMaxRetries is returned when max retries exceeded.
var ErrMaxRetries = errors.New("max retries exceeded")

// ErrProcessCrashed is returned when the agent process crashes.
var ErrProcessCrashed = errors.New("agent process crashed")

// ErrResponseTimeout is returned when response times out.
var ErrResponseTimeout = errors.New("response timeout")

// Retry executes fn with exponential backoff.
func Retry(ctx context.Context, cfg RetryConfig, fn func() error) error {
	var lastErr error
	wait := cfg.InitialWait

	for attempt := 0; attempt < cfg.MaxAttempts; attempt++ {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		lastErr = fn()
		if lastErr == nil {
			return nil
		}

		// Don't retry on context errors
		if errors.Is(lastErr, context.Canceled) || errors.Is(lastErr, context.DeadlineExceeded) {
			return lastErr
		}

		// Wait before next attempt
		if attempt < cfg.MaxAttempts-1 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(wait):
			}
			wait = time.Duration(float64(wait) * cfg.Multiplier)
			if wait > cfg.MaxWait {
				wait = cfg.MaxWait
			}
		}
	}

	return errors.Join(ErrMaxRetries, lastErr)
}

// StartProcessWithRetry starts a process with retry on failure.
func StartProcessWithRetry(ctx context.Context, def Definition, rootPath string, cfg RetryConfig) (Process, error) {
	var proc Process
	err := Retry(ctx, cfg, func() error {
		var err error
		proc, err = StartProcess(ctx, def, rootPath)
		return err
	})
	return proc, err
}

// RecoverablePool wraps Pool with automatic recovery.
type RecoverablePool struct {
	*Pool
	retryCfg RetryConfig
}

// NewRecoverablePool creates a pool with recovery support.
func NewRecoverablePool(cfg Config, retryCfg RetryConfig) *RecoverablePool {
	return &RecoverablePool{
		Pool:     NewPool(cfg),
		retryCfg: retryCfg,
	}
}

// GetOrCreateWithRetry returns or creates a process with retry.
func (p *RecoverablePool) GetOrCreateWithRetry(ctx context.Context, sessionKey, agentName, rootPath string) (Process, error) {
	// First try to get existing
	p.mu.Lock()
	if proc, ok := p.processes[sessionKey]; ok {
		p.mu.Unlock()
		return proc, nil
	}
	p.mu.Unlock()

	def, ok := p.cfg.Agents[agentName]
	if !ok {
		return nil, errors.New("agent not configured: " + agentName)
	}

	// Start with retry
	proc, err := StartProcessWithRetry(ctx, def, rootPath, p.retryCfg)
	if err != nil {
		return nil, err
	}

	p.mu.Lock()
	p.processes[sessionKey] = proc
	p.mu.Unlock()
	return proc, nil
}

// RecoverProcess attempts to restart a crashed process.
func (p *RecoverablePool) RecoverProcess(ctx context.Context, sessionKey, agentName, rootPath string) (Process, error) {
	// Close existing if any
	p.Close(sessionKey)

	// Start new with retry
	return p.GetOrCreateWithRetry(ctx, sessionKey, agentName, rootPath)
}

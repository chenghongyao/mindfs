package api

import (
	"context"
	"time"

	"mindfs/server/internal/audit"
)

// RecoveryConfig holds configuration for error recovery
type RecoveryConfig struct {
	AgentTimeoutRetries    int           // Number of retries for agent timeout
	AgentRestartDelay      time.Duration // Delay before restarting crashed agent
	SessionResumeRetries   int           // Number of retries for session resume
	EnableAuditLogging     bool          // Whether to log errors to audit
}

// DefaultRecoveryConfig returns the default recovery configuration
func DefaultRecoveryConfig() RecoveryConfig {
	return RecoveryConfig{
		AgentTimeoutRetries:  1,
		AgentRestartDelay:    500 * time.Millisecond,
		SessionResumeRetries: 1,
		EnableAuditLogging:   true,
	}
}

// RecoveryStrategy handles error recovery for various failure scenarios
type RecoveryStrategy struct {
	config RecoveryConfig
	audit  *audit.WriterPool
}

// NewRecoveryStrategy creates a new recovery strategy
func NewRecoveryStrategy(config RecoveryConfig, auditPool *audit.WriterPool) *RecoveryStrategy {
	return &RecoveryStrategy{
		config: config,
		audit:  auditPool,
	}
}

// RecoveryResult represents the result of a recovery attempt
type RecoveryResult struct {
	Recovered bool
	Retries   int
	Error     error
	Action    string
}

// RecoverableFunc is a function that can be retried
type RecoverableFunc func(ctx context.Context) error

// WithAgentTimeoutRecovery wraps a function with agent timeout recovery
func (s *RecoveryStrategy) WithAgentTimeoutRecovery(
	ctx context.Context,
	fn RecoverableFunc,
	logger *audit.Logger,
) RecoveryResult {
	var lastErr error
	retries := 0

	for i := 0; i <= s.config.AgentTimeoutRetries; i++ {
		err := fn(ctx)
		if err == nil {
			return RecoveryResult{
				Recovered: i > 0,
				Retries:   retries,
				Action:    "agent_timeout_retry",
			}
		}

		lastErr = err
		retries++

		// Check if it's a timeout error
		if !isTimeoutError(err) {
			break
		}

		// Log retry attempt
		if logger != nil && s.config.EnableAuditLogging {
			_ = logger.Log(audit.NewEntry(audit.EntryTypeSession, "retry", audit.ActorSystem).
				WithDetails(map[string]any{
					"error":  err.Error(),
					"retry":  i + 1,
					"reason": "agent_timeout",
				}))
		}
	}

	return RecoveryResult{
		Recovered: false,
		Retries:   retries,
		Error:     lastErr,
		Action:    "agent_timeout_retry",
	}
}

// WithAgentCrashRecovery handles agent process crash recovery
func (s *RecoveryStrategy) WithAgentCrashRecovery(
	ctx context.Context,
	restartFn func() error,
	retryFn RecoverableFunc,
	logger *audit.Logger,
) RecoveryResult {
	// Wait before restart
	time.Sleep(s.config.AgentRestartDelay)

	// Attempt to restart the agent
	if err := restartFn(); err != nil {
		if logger != nil && s.config.EnableAuditLogging {
			_ = logger.Log(audit.NewEntry(audit.EntryTypeSession, "restart_failed", audit.ActorSystem).
				WithError(err.Error()))
		}
		return RecoveryResult{
			Recovered: false,
			Error:     err,
			Action:    "agent_restart",
		}
	}

	// Log successful restart
	if logger != nil && s.config.EnableAuditLogging {
		_ = logger.Log(audit.NewEntry(audit.EntryTypeSession, "restart_success", audit.ActorSystem))
	}

	// Retry the operation
	if err := retryFn(ctx); err != nil {
		return RecoveryResult{
			Recovered: false,
			Retries:   1,
			Error:     err,
			Action:    "agent_restart",
		}
	}

	return RecoveryResult{
		Recovered: true,
		Retries:   1,
		Action:    "agent_restart",
	}
}

// WithSessionResumeRecovery handles session resume failure with fallback to exchanges
func (s *RecoveryStrategy) WithSessionResumeRecovery(
	ctx context.Context,
	nativeResumeFn RecoverableFunc,
	exchangesFallbackFn RecoverableFunc,
	logger *audit.Logger,
) RecoveryResult {
	// Try native resume first
	for i := 0; i <= s.config.SessionResumeRetries; i++ {
		if err := nativeResumeFn(ctx); err == nil {
			return RecoveryResult{
				Recovered: i > 0,
				Retries:   i,
				Action:    "native_resume",
			}
		}
	}

	// Log fallback to exchanges
	if logger != nil && s.config.EnableAuditLogging {
		_ = logger.Log(audit.NewEntry(audit.EntryTypeSession, "resume_fallback", audit.ActorSystem).
			WithDetails(map[string]any{
				"fallback": "exchanges",
			}))
	}

	// Fallback to exchanges-based resume
	if err := exchangesFallbackFn(ctx); err != nil {
		return RecoveryResult{
			Recovered: false,
			Error:     err,
			Action:    "exchanges_fallback",
		}
	}

	return RecoveryResult{
		Recovered: true,
		Retries:   s.config.SessionResumeRetries + 1,
		Action:    "exchanges_fallback",
	}
}

// LogError logs an error to the audit log
func (s *RecoveryStrategy) LogError(
	logger *audit.Logger,
	errorCode ErrorCode,
	err error,
	details map[string]any,
) {
	if logger == nil || !s.config.EnableAuditLogging {
		return
	}

	entry := audit.NewEntry(audit.EntryTypeSession, "error", audit.ActorSystem).
		WithError(err.Error())

	if details == nil {
		details = make(map[string]any)
	}
	details["error_code"] = string(errorCode)
	entry.WithDetails(details)

	_ = logger.Log(entry)
}

// isTimeoutError checks if an error is a timeout error
func isTimeoutError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return contains(errStr, "timeout") || contains(errStr, "deadline exceeded")
}

// contains checks if a string contains a substring (case-insensitive)
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || containsSubstring(s, substr))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

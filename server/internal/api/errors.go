package api

import "fmt"

// ErrorCode represents a structured error code
type ErrorCode string

// Session error codes
const (
	ErrSessionNotFound      ErrorCode = "session.not_found"
	ErrSessionAlreadyClosed ErrorCode = "session.already_closed"
	ErrSessionResumeFailed  ErrorCode = "session.resume_failed"
	ErrSessionCreateFailed  ErrorCode = "session.create_failed"
	ErrSessionUnavailable   ErrorCode = "session.unavailable"
)

// Agent error codes
const (
	ErrAgentNotAvailable    ErrorCode = "agent.not_available"
	ErrAgentProcessCrashed  ErrorCode = "agent.process_crashed"
	ErrAgentTimeout         ErrorCode = "agent.timeout"
	ErrAgentInvalidResponse ErrorCode = "agent.invalid_response"
	ErrAgentPermissionDenied ErrorCode = "agent.permission_denied"
)

// View error codes
const (
	ErrViewNotFound         ErrorCode = "view.not_found"
	ErrViewInvalidSchema    ErrorCode = "view.invalid_schema"
	ErrViewGenerationFailed ErrorCode = "view.generation_failed"
	ErrViewVersionNotFound  ErrorCode = "view.version_not_found"
)

// File error codes
const (
	ErrFileNotFound        ErrorCode = "file.not_found"
	ErrFilePermissionDenied ErrorCode = "file.permission_denied"
	ErrFileReadFailed      ErrorCode = "file.read_failed"
	ErrFileWriteFailed     ErrorCode = "file.write_failed"
)

// Skill error codes
const (
	ErrSkillNotFound         ErrorCode = "skill.not_found"
	ErrSkillPermissionDenied ErrorCode = "skill.permission_denied"
	ErrSkillExecutionFailed  ErrorCode = "skill.execution_failed"
)

// Task error codes
const (
	ErrTaskNotFound   ErrorCode = "task.not_found"
	ErrTaskCancelled  ErrorCode = "task.cancelled"
	ErrTaskFailed     ErrorCode = "task.failed"
)

// Auth error codes
const (
	ErrAuthInvalidToken  ErrorCode = "auth.invalid_token"
	ErrAuthTokenExpired  ErrorCode = "auth.token_expired"
	ErrAuthUnauthorized  ErrorCode = "auth.unauthorized"
)

// General error codes
const (
	ErrInvalidRequest    ErrorCode = "invalid_request"
	ErrInternalError     ErrorCode = "internal_error"
	ErrServiceUnavailable ErrorCode = "service_unavailable"
)

// APIError represents a structured API error
type APIError struct {
	Code    ErrorCode `json:"code"`
	Message string    `json:"message"`
	Details any       `json:"details,omitempty"`
}

func (e *APIError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// NewAPIError creates a new API error
func NewAPIError(code ErrorCode, message string) *APIError {
	return &APIError{
		Code:    code,
		Message: message,
	}
}

// WithDetails adds details to the error
func (e *APIError) WithDetails(details any) *APIError {
	e.Details = details
	return e
}

// Error constructors for common errors

func errSessionNotFound(key string) *APIError {
	return NewAPIError(ErrSessionNotFound, fmt.Sprintf("session not found: %s", key))
}

func errSessionAlreadyClosed(key string) *APIError {
	return NewAPIError(ErrSessionAlreadyClosed, fmt.Sprintf("session already closed: %s", key))
}

func errAgentNotAvailable(agent string) *APIError {
	return NewAPIError(ErrAgentNotAvailable, fmt.Sprintf("agent not available: %s", agent))
}

func errAgentTimeout(agent string) *APIError {
	return NewAPIError(ErrAgentTimeout, fmt.Sprintf("agent timeout: %s", agent))
}

func errViewNotFound(path string) *APIError {
	return NewAPIError(ErrViewNotFound, fmt.Sprintf("view not found: %s", path))
}

func errFileNotFound(path string) *APIError {
	return NewAPIError(ErrFileNotFound, fmt.Sprintf("file not found: %s", path))
}

func errSkillNotFound(name string) *APIError {
	return NewAPIError(ErrSkillNotFound, fmt.Sprintf("skill not found: %s", name))
}

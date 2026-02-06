package api

import (
	"context"
	"net/http"
	"strings"
)

type contextKey string

const (
	// TokenContextKey is the context key for the authenticated token
	TokenContextKey contextKey = "auth_token"
	// UserContextKey is the context key for the authenticated user ID
	UserContextKey contextKey = "auth_user"
)

// AuthMiddleware provides HTTP authentication middleware
type AuthMiddleware struct {
	Auth     *AuthService
	Optional bool // If true, unauthenticated requests are allowed
}

// Handler returns an HTTP middleware handler
func (m *AuthMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenValue := extractToken(r)

		if tokenValue == "" {
			if m.Optional {
				next.ServeHTTP(w, r)
				return
			}
			writeAuthError(w, http.StatusUnauthorized, "auth.missing_token", "authentication required")
			return
		}

		token, valid := m.Auth.Validate(tokenValue)
		if !valid {
			if m.Optional {
				next.ServeHTTP(w, r)
				return
			}
			writeAuthError(w, http.StatusUnauthorized, "auth.invalid_token", "invalid or expired token")
			return
		}

		// Add token and user to context
		ctx := context.WithValue(r.Context(), TokenContextKey, token)
		ctx = context.WithValue(ctx, UserContextKey, token.UserID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// extractToken extracts the token from the request
func extractToken(r *http.Request) string {
	// Check Authorization header first
	auth := r.Header.Get("Authorization")
	if auth != "" {
		// Support "Bearer <token>" format
		if strings.HasPrefix(auth, "Bearer ") {
			return strings.TrimPrefix(auth, "Bearer ")
		}
		// Support plain token
		return auth
	}

	// Check query parameter
	if token := r.URL.Query().Get("token"); token != "" {
		return token
	}

	// Check cookie
	if cookie, err := r.Cookie("mindfs_token"); err == nil {
		return cookie.Value
	}

	return ""
}

// writeAuthError writes an authentication error response
func writeAuthError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = w.Write([]byte(`{"error":{"code":"` + code + `","message":"` + message + `"}}`))
}

// GetToken retrieves the token from the request context
func GetToken(ctx context.Context) *Token {
	if token, ok := ctx.Value(TokenContextKey).(*Token); ok {
		return token
	}
	return nil
}

// GetUserID retrieves the user ID from the request context
func GetUserID(ctx context.Context) string {
	if userID, ok := ctx.Value(UserContextKey).(string); ok {
		return userID
	}
	return ""
}

// WSAuthMiddleware provides WebSocket authentication
type WSAuthMiddleware struct {
	Auth *AuthService
}

// ValidateWSConnection validates a WebSocket connection request
func (m *WSAuthMiddleware) ValidateWSConnection(r *http.Request) (*Token, error) {
	tokenValue := extractToken(r)
	if tokenValue == "" {
		return nil, errInvalidRequest("authentication required")
	}

	token, valid := m.Auth.Validate(tokenValue)
	if !valid {
		return nil, errInvalidRequest("invalid or expired token")
	}

	return token, nil
}

// RequireAuth is a convenience function to create an auth middleware
func RequireAuth(auth *AuthService) func(http.Handler) http.Handler {
	m := &AuthMiddleware{Auth: auth, Optional: false}
	return m.Handler
}

// OptionalAuth is a convenience function to create an optional auth middleware
func OptionalAuth(auth *AuthService) func(http.Handler) http.Handler {
	m := &AuthMiddleware{Auth: auth, Optional: true}
	return m.Handler
}

package api

import (
	"crypto/rand"
	"encoding/hex"
	"sync"
	"time"
)

// TokenConfig holds authentication configuration
type TokenConfig struct {
	Secret     string        // Secret key for token generation
	Expiration time.Duration // Token expiration duration
}

// Token represents an authentication token
type Token struct {
	Value     string    `json:"value"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
	UserID    string    `json:"user_id,omitempty"`
}

// IsExpired checks if the token has expired
func (t *Token) IsExpired() bool {
	return time.Now().After(t.ExpiresAt)
}

// TokenStore manages authentication tokens
type TokenStore struct {
	mu         sync.RWMutex
	tokens     map[string]*Token
	config     TokenConfig
	cleanupTTL time.Duration
}

// NewTokenStore creates a new token store
func NewTokenStore(config TokenConfig) *TokenStore {
	if config.Expiration == 0 {
		config.Expiration = 24 * time.Hour
	}
	store := &TokenStore{
		tokens:     make(map[string]*Token),
		config:     config,
		cleanupTTL: 5 * time.Minute,
	}
	go store.cleanupLoop()
	return store
}

// Generate creates a new token
func (s *TokenStore) Generate(userID string) (*Token, error) {
	value, err := generateTokenValue()
	if err != nil {
		return nil, err
	}

	now := time.Now()
	token := &Token{
		Value:     value,
		CreatedAt: now,
		ExpiresAt: now.Add(s.config.Expiration),
		UserID:    userID,
	}

	s.mu.Lock()
	s.tokens[value] = token
	s.mu.Unlock()

	return token, nil
}

// Validate checks if a token is valid
func (s *TokenStore) Validate(value string) (*Token, bool) {
	s.mu.RLock()
	token, ok := s.tokens[value]
	s.mu.RUnlock()

	if !ok {
		return nil, false
	}

	if token.IsExpired() {
		s.Revoke(value)
		return nil, false
	}

	return token, true
}

// Revoke removes a token
func (s *TokenStore) Revoke(value string) {
	s.mu.Lock()
	delete(s.tokens, value)
	s.mu.Unlock()
}

// Refresh extends a token's expiration
func (s *TokenStore) Refresh(value string) (*Token, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	token, ok := s.tokens[value]
	if !ok || token.IsExpired() {
		return nil, false
	}

	token.ExpiresAt = time.Now().Add(s.config.Expiration)
	return token, true
}

// cleanupLoop periodically removes expired tokens
func (s *TokenStore) cleanupLoop() {
	ticker := time.NewTicker(s.cleanupTTL)
	defer ticker.Stop()

	for range ticker.C {
		s.cleanup()
	}
}

func (s *TokenStore) cleanup() {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	for value, token := range s.tokens {
		if now.After(token.ExpiresAt) {
			delete(s.tokens, value)
		}
	}
}

// generateTokenValue creates a cryptographically secure random token
func generateTokenValue() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// AuthService provides authentication functionality
type AuthService struct {
	store  *TokenStore
	config TokenConfig
}

// NewAuthService creates a new authentication service
func NewAuthService(config TokenConfig) *AuthService {
	return &AuthService{
		store:  NewTokenStore(config),
		config: config,
	}
}

// Login generates a new token for the user
func (s *AuthService) Login(userID string) (*Token, error) {
	return s.store.Generate(userID)
}

// Logout revokes a token
func (s *AuthService) Logout(tokenValue string) {
	s.store.Revoke(tokenValue)
}

// Validate checks if a token is valid
func (s *AuthService) Validate(tokenValue string) (*Token, bool) {
	return s.store.Validate(tokenValue)
}

// Refresh extends a token's expiration
func (s *AuthService) Refresh(tokenValue string) (*Token, bool) {
	return s.store.Refresh(tokenValue)
}

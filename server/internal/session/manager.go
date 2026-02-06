package session

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"
)

type SummaryGenerator func(context.Context, *Session) (*SessionSummary, error)

type Manager struct {
	store           *Store
	now             func() time.Time
	summaryGenerate SummaryGenerator
	resume          Resumer
}

type CreateInput struct {
	Key   string
	Type  string
	Agent string
	Name  string
}

func NewManager(store *Store, opts ...Option) *Manager {
	m := &Manager{store: store, now: time.Now}
	for _, opt := range opts {
		opt(m)
	}
	return m
}

type Option func(*Manager)

func WithClock(now func() time.Time) Option {
	return func(m *Manager) {
		m.now = now
	}
}

func WithSummaryGenerator(gen SummaryGenerator) Option {
	return func(m *Manager) {
		m.summaryGenerate = gen
	}
}

func WithResumer(resumer Resumer) Option {
	return func(m *Manager) {
		m.resume = resumer
	}
}

func (m *Manager) Create(ctx context.Context, input CreateInput) (*Session, error) {
	if m.store == nil {
		return nil, errors.New("store not configured")
	}
	if strings.TrimSpace(input.Type) == "" {
		return nil, errors.New("session type required")
	}
	if strings.TrimSpace(input.Agent) == "" {
		return nil, errors.New("agent required")
	}
	key := input.Key
	if key == "" {
		key = generateKey()
	}
	now := m.now().UTC()
	name := strings.TrimSpace(input.Name)
	if name == "" {
		name = "New Session"
	}
	session := &Session{
		Key:       key,
		Type:      input.Type,
		Agent:     input.Agent,
		Name:      name,
		Status:    StatusActive,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := m.store.Create(session); err != nil {
		return nil, err
	}
	return session, nil
}

func (m *Manager) Get(_ context.Context, key string) (*Session, error) {
	if m.store == nil {
		return nil, errors.New("store not configured")
	}
	return m.store.Get(key)
}

func (m *Manager) List(_ context.Context) ([]*Session, error) {
	if m.store == nil {
		return nil, errors.New("store not configured")
	}
	return m.store.List()
}

func (m *Manager) AddExchange(_ context.Context, key, role, content string) (*Session, error) {
	if m.store == nil {
		return nil, errors.New("store not configured")
	}
	session, err := m.store.Get(key)
	if err != nil {
		return nil, err
	}
	session.Exchanges = append(session.Exchanges, Exchange{
		Role:      role,
		Content:   content,
		Timestamp: m.now().UTC(),
	})
	session.Status = StatusActive
	session.UpdatedAt = m.now().UTC()
	if err := m.store.Save(session); err != nil {
		return nil, err
	}
	return session, nil
}

func (m *Manager) AddRelatedFile(_ context.Context, key string, file RelatedFile) (*Session, error) {
	if m.store == nil {
		return nil, errors.New("store not configured")
	}
	if strings.TrimSpace(file.Path) == "" {
		return nil, errors.New("file path required")
	}
	session, err := m.store.Get(key)
	if err != nil {
		return nil, err
	}
	for _, existing := range session.RelatedFiles {
		if existing.Path == file.Path {
			return session, nil
		}
	}
	session.RelatedFiles = append(session.RelatedFiles, file)
	session.UpdatedAt = m.now().UTC()
	if err := m.store.Save(session); err != nil {
		return nil, err
	}
	return session, nil
}

func (m *Manager) UpdateAgentSessionID(_ context.Context, key string, agentSessionID string) (*Session, error) {
	if m.store == nil {
		return nil, errors.New("store not configured")
	}
	if strings.TrimSpace(agentSessionID) == "" {
		return nil, errors.New("agent session id required")
	}
	session, err := m.store.Get(key)
	if err != nil {
		return nil, err
	}
	if session.AgentSessionID != nil && *session.AgentSessionID == agentSessionID {
		return session, nil
	}
	session.AgentSessionID = &agentSessionID
	session.UpdatedAt = m.now().UTC()
	if err := m.store.Save(session); err != nil {
		return nil, err
	}
	return session, nil
}

func (m *Manager) Close(ctx context.Context, key string) (*Session, error) {
	if m.store == nil {
		return nil, errors.New("store not configured")
	}
	session, err := m.store.Get(key)
	if err != nil {
		return nil, err
	}
	if session.Status == StatusClosed {
		return session, nil
	}
	now := m.now().UTC()
	if session.Summary == nil {
		if m.summaryGenerate != nil {
			if summary, err := m.summaryGenerate(ctx, session); err == nil {
				session.Summary = summary
			}
		}
		if session.Summary == nil {
			session.Summary = &SessionSummary{
				Title:       session.Name,
				Description: "",
				KeyActions:  []string{},
				Outputs:     []string{},
				GeneratedAt: now,
			}
		}
	}
	session.Status = StatusClosed
	session.ClosedAt = &now
	session.UpdatedAt = now
	if err := m.store.Save(session); err != nil {
		return nil, err
	}
	return session, nil
}

func (m *Manager) MarkIdle(_ context.Context, key string) (*Session, error) {
	if m.store == nil {
		return nil, errors.New("store not configured")
	}
	session, err := m.store.Get(key)
	if err != nil {
		return nil, err
	}
	if session.Status != StatusActive {
		return session, nil
	}
	session.Status = StatusIdle
	session.UpdatedAt = m.now().UTC()
	if err := m.store.Save(session); err != nil {
		return nil, err
	}
	return session, nil
}

func (m *Manager) CheckIdle(ctx context.Context, idleAfter, closeAfter time.Duration) ([]*Session, []*Session, error) {
	if m.store == nil {
		return nil, nil, errors.New("store not configured")
	}
	if idleAfter <= 0 || closeAfter <= 0 {
		return nil, nil, errors.New("idle and close thresholds required")
	}
	sessions, err := m.store.List()
	if err != nil {
		return nil, nil, err
	}
	now := m.now().UTC()
	markedIdle := []*Session{}
	closed := []*Session{}
	for _, s := range sessions {
		last := s.UpdatedAt
		idleFor := now.Sub(last)
		switch s.Status {
		case StatusActive:
			if idleFor >= idleAfter {
				updated, err := m.MarkIdle(ctx, s.Key)
				if err == nil {
					markedIdle = append(markedIdle, updated)
				}
			}
		case StatusIdle:
			if idleFor >= closeAfter {
				updated, err := m.Close(ctx, s.Key)
				if err == nil {
					closed = append(closed, updated)
				}
			}
		}
	}
	return markedIdle, closed, nil
}

func (m *Manager) Resume(ctx context.Context, key string) (*Session, error) {
	if m.store == nil {
		return nil, errors.New("store not configured")
	}
	session, err := m.store.Get(key)
	if err != nil {
		return nil, err
	}
	if m.resume != nil {
		if err := m.resume.Resume(ctx, session); err != nil {
			return nil, err
		}
	}
	session.Status = StatusActive
	session.UpdatedAt = m.now().UTC()
	if err := m.store.Save(session); err != nil {
		return nil, err
	}
	return session, nil
}

func generateKey() string {
	buf := make([]byte, 6)
	_, err := rand.Read(buf)
	if err != nil {
		return fmt.Sprintf("s-%d", time.Now().UTC().UnixNano())
	}
	return fmt.Sprintf("s-%d-%s", time.Now().UTC().UnixNano(), hex.EncodeToString(buf))
}

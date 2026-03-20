package relay

import (
	"context"
	"log"
	"strings"
	"sync"
)

type Manager struct {
	service         *Service
	initialBindCode string

	mu      sync.Mutex
	ctx     context.Context
	cancel  context.CancelFunc
	started bool
}

func NewManager(localAddr, bindCode string) (*Manager, error) {
	service, err := NewService(localAddr, bindCode)
	if err != nil {
		return nil, err
	}
	return &Manager{
		service:         service,
		initialBindCode: strings.TrimSpace(bindCode),
	}, nil
}

func (m *Manager) Start(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.started {
		return nil
	}
	m.ctx = ctx
	m.started = true

	if m.initialBindCode != "" {
		if _, err := m.service.Bind(ctx, m.initialBindCode); err != nil {
			m.started = false
			m.ctx = nil
			return err
		}
	}

	creds, err := m.service.store.Load()
	if err != nil {
		m.started = false
		m.ctx = nil
		return err
	}
	if creds.Relay.DeviceToken == "" || creds.Relay.Endpoint == "" {
		return nil
	}

	m.startLocked(ctx)
	return nil
}

func (m *Manager) Bind(ctx context.Context, bindCode string) (Credentials, error) {
	creds, err := m.service.Bind(ctx, bindCode)
	if err != nil {
		return Credentials{}, err
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	if m.started {
		m.restartLocked()
	}
	return creds, nil
}

func (m *Manager) startLocked(parent context.Context) {
	runCtx, cancel := context.WithCancel(parent)
	m.ctx = parent
	m.cancel = cancel
	go func() {
		if err := m.service.Run(runCtx); err != nil && runCtx.Err() == nil {
			log.Printf("[relay] stopped: %v", err)
		}
	}()
}

func (m *Manager) restartLocked() {
	if m.cancel != nil {
		m.cancel()
		m.cancel = nil
	}
	if m.ctx != nil {
		m.startLocked(m.ctx)
	}
}

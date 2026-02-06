package agent

import (
	"context"
	"errors"
	"sync"
)

// Pool manages agent processes by session key.
type Pool struct {
	cfg       Config
	mu        sync.Mutex
	processes map[string]Process
}

// NewPool creates a new agent pool.
func NewPool(cfg Config) *Pool {
	return &Pool{cfg: cfg, processes: make(map[string]Process)}
}

// GetOrCreate returns an existing process or creates a new one.
func (p *Pool) GetOrCreate(ctx context.Context, sessionKey, agentName, rootPath string) (Process, error) {
	if sessionKey == "" {
		return nil, errors.New("session key required")
	}
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

	proc, err := StartProcess(ctx, def, rootPath)
	if err != nil {
		return nil, err
	}

	p.mu.Lock()
	p.processes[sessionKey] = proc
	p.mu.Unlock()
	return proc, nil
}

// Close closes and removes a process by session key.
func (p *Pool) Close(sessionKey string) {
	p.mu.Lock()
	proc, ok := p.processes[sessionKey]
	if ok {
		delete(p.processes, sessionKey)
	}
	p.mu.Unlock()
	if ok {
		_ = proc.Close()
	}
}

// Config returns the pool configuration.
func (p *Pool) Config() Config {
	return p.cfg
}

// CloseAll closes all processes.
func (p *Pool) CloseAll() {
	p.mu.Lock()
	procs := make([]Process, 0, len(p.processes))
	for _, proc := range p.processes {
		procs = append(procs, proc)
	}
	p.processes = make(map[string]Process)
	p.mu.Unlock()
	for _, proc := range procs {
		_ = proc.Close()
	}
}

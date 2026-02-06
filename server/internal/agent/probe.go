package agent

import (
	"context"
	"sync"
	"time"
)

type Status struct {
	Name      string    `json:"name"`
	Available bool      `json:"available"`
	Version   string    `json:"version,omitempty"`
	Error     string    `json:"error,omitempty"`
	LastProbe time.Time `json:"last_probe"`
}

// Prober 管理 Agent 可用性探测
type Prober struct {
	cfg           *Config
	statuses      map[string]Status
	mu            sync.RWMutex
	probeInterval time.Duration
	stopCh        chan struct{}
}

func NewProber(cfg *Config, probeInterval time.Duration) *Prober {
	if probeInterval <= 0 {
		probeInterval = 5 * time.Minute
	}
	return &Prober{
		cfg:           cfg,
		statuses:      make(map[string]Status),
		probeInterval: probeInterval,
		stopCh:        make(chan struct{}),
	}
}

// Start 启动定期探测
func (p *Prober) Start(ctx context.Context) {
	// 立即执行一次探测
	p.ProbeAll(ctx)

	// 启动定期探测
	ticker := time.NewTicker(p.probeInterval)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				p.ProbeAll(ctx)
			case <-p.stopCh:
				return
			case <-ctx.Done():
				return
			}
		}
	}()
}

// Stop 停止定期探测
func (p *Prober) Stop() {
	select {
	case <-p.stopCh:
		return
	default:
		close(p.stopCh)
	}
}

// ProbeAll 探测所有配置的 Agent
func (p *Prober) ProbeAll(ctx context.Context) []Status {
	if p.cfg == nil {
		return nil
	}

	statuses := make([]Status, 0, len(p.cfg.Agents))
	for name, def := range p.cfg.Agents {
		status := ProbeAgent(ctx, name, def)
		statuses = append(statuses, status)

		p.mu.Lock()
		p.statuses[name] = status
		p.mu.Unlock()
	}
	return statuses
}

// ProbeOne 探测单个 Agent 并更新缓存
func (p *Prober) ProbeOne(ctx context.Context, name string) Status {
	if p.cfg == nil {
		return Status{Name: name, Available: false, Error: "config not loaded", LastProbe: time.Now().UTC()}
	}

	def, ok := p.cfg.Agents[name]
	if !ok {
		return Status{Name: name, Available: false, Error: "agent not configured", LastProbe: time.Now().UTC()}
	}

	status := ProbeAgent(ctx, name, def)

	p.mu.Lock()
	p.statuses[name] = status
	p.mu.Unlock()

	return status
}

// GetStatus 获取缓存的 Agent 状态
func (p *Prober) GetStatus(name string) (Status, bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	status, ok := p.statuses[name]
	return status, ok
}

// GetAllStatuses 获取所有缓存的 Agent 状态
func (p *Prober) GetAllStatuses() []Status {
	p.mu.RLock()
	defer p.mu.RUnlock()

	statuses := make([]Status, 0, len(p.statuses))
	for _, status := range p.statuses {
		statuses = append(statuses, status)
	}
	return statuses
}

// IsAvailable 检查 Agent 是否可用
func (p *Prober) IsAvailable(name string) bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	status, ok := p.statuses[name]
	return ok && status.Available
}

// ProbeAgent 探测单个 Agent
func ProbeAgent(ctx context.Context, name string, def Definition) Status {
	status := Status{Name: name, Available: false, LastProbe: time.Now().UTC()}
	if def.Command == "" {
		status.Error = "command required"
		return status
	}
	cmd := def.Command
	args := def.ProbeArgs
	if len(args) == 0 {
		args = []string{"--version"}
	}
	probeCtx, cancel := context.WithTimeout(ctx, 4*time.Second)
	defer cancel()
	out, err := runCommand(probeCtx, cmd, args...)
	if err != nil {
		status.Error = err.Error()
		return status
	}
	status.Available = true
	status.Version = parseVersion(out)
	return status
}

// Pool 的 ProbeAll 方法保持兼容
func (p *Pool) ProbeAll(ctx context.Context) []Status {
	statuses := make([]Status, 0, len(p.cfg.Agents))
	for name, def := range p.cfg.Agents {
		statuses = append(statuses, ProbeAgent(ctx, name, def))
	}
	return statuses
}

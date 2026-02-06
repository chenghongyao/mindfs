package session

import (
	"context"
	"sort"
	"time"
)

// StatusChangeCallback 状态变更回调，用于通知前端
type StatusChangeCallback func(session *Session, oldStatus, newStatus string)

type IdleChecker struct {
	stores          *StoreManager
	interval        time.Duration
	idleFor         time.Duration
	closeFor        time.Duration
	maxIdleSessions int
	onStatusChange  StatusChangeCallback
	stopCh          chan struct{}
}

type IdleCheckerOption func(*IdleChecker)

func WithMaxIdleSessions(max int) IdleCheckerOption {
	return func(c *IdleChecker) {
		c.maxIdleSessions = max
	}
}

func WithStatusChangeCallback(cb StatusChangeCallback) IdleCheckerOption {
	return func(c *IdleChecker) {
		c.onStatusChange = cb
	}
}

func NewIdleChecker(stores *StoreManager, interval, idleFor, closeFor time.Duration, opts ...IdleCheckerOption) *IdleChecker {
	if interval <= 0 {
		interval = 1 * time.Minute
	}
	if idleFor <= 0 {
		idleFor = 10 * time.Minute
	}
	if closeFor <= 0 {
		closeFor = 30 * time.Minute
	}
	c := &IdleChecker{
		stores:          stores,
		interval:        interval,
		idleFor:         idleFor,
		closeFor:        closeFor,
		maxIdleSessions: 3, // 默认最多 3 个 idle session
		stopCh:          make(chan struct{}),
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

func (c *IdleChecker) Start(ctx context.Context) {
	if c == nil || c.stores == nil {
		return
	}
	ticker := time.NewTicker(c.interval)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				for _, managedDir := range c.stores.List() {
					store, err := c.stores.Get(managedDir)
					if err != nil {
						continue
					}
					manager := NewManager(store)
					markedIdle, closed, _ := manager.CheckIdle(ctx, c.idleFor, c.closeFor)

					// 通知状态变更
					if c.onStatusChange != nil {
						for _, s := range markedIdle {
							c.onStatusChange(s, StatusActive, StatusIdle)
						}
						for _, s := range closed {
							c.onStatusChange(s, StatusIdle, StatusClosed)
						}
					}

					// 检查并关闭超出限制的 idle session
					c.enforceMaxIdleSessions(ctx, store)
				}
			case <-c.stopCh:
				return
			case <-ctx.Done():
				return
			}
		}
	}()
}

// enforceMaxIdleSessions 关闭超出限制的最老的 idle session
func (c *IdleChecker) enforceMaxIdleSessions(ctx context.Context, store *Store) {
	if c.maxIdleSessions <= 0 {
		return
	}

	sessions, err := store.List()
	if err != nil {
		return
	}

	// 筛选 idle session
	idleSessions := []*Session{}
	for _, s := range sessions {
		if s.Status == StatusIdle {
			idleSessions = append(idleSessions, s)
		}
	}

	if len(idleSessions) <= c.maxIdleSessions {
		return
	}

	// 按 UpdatedAt 排序，最老的在前
	sort.Slice(idleSessions, func(i, j int) bool {
		return idleSessions[i].UpdatedAt.Before(idleSessions[j].UpdatedAt)
	})

	// 关闭超出限制的 session
	manager := NewManager(store)
	toClose := len(idleSessions) - c.maxIdleSessions
	for i := 0; i < toClose; i++ {
		s := idleSessions[i]
		closed, err := manager.Close(ctx, s.Key)
		if err == nil && c.onStatusChange != nil {
			c.onStatusChange(closed, StatusIdle, StatusClosed)
		}
	}
}

func (c *IdleChecker) Stop() {
	select {
	case <-c.stopCh:
		return
	default:
		close(c.stopCh)
	}
}

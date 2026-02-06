package acp

import (
	"sync"
	"time"
)

// IdleDetector signals when no data has been received for a duration.
type IdleDetector struct {
	timeout time.Duration
	timer   *time.Timer
	done    chan struct{}
	mu      sync.Mutex
	stopped bool
}

// NewIdleDetector creates a detector with the given timeout.
func NewIdleDetector(timeout time.Duration) *IdleDetector {
	return &IdleDetector{
		timeout: timeout,
		done:    make(chan struct{}),
	}
}

// Start begins the idle detection timer.
func (d *IdleDetector) Start() {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.stopped {
		return
	}
	d.timer = time.AfterFunc(d.timeout, func() {
		d.mu.Lock()
		defer d.mu.Unlock()
		if !d.stopped {
			close(d.done)
			d.stopped = true
		}
	})
}

// Reset restarts the idle timer. Call this when data is received.
func (d *IdleDetector) Reset() {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.stopped {
		return
	}
	if d.timer != nil {
		d.timer.Stop()
		d.timer.Reset(d.timeout)
	}
}

// Stop cancels the idle detector.
func (d *IdleDetector) Stop() {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.stopped {
		return
	}
	d.stopped = true
	if d.timer != nil {
		d.timer.Stop()
	}
	select {
	case <-d.done:
	default:
		close(d.done)
	}
}

// Done returns a channel that closes when idle timeout is reached.
func (d *IdleDetector) Done() <-chan struct{} {
	return d.done
}

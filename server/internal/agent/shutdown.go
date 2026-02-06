package agent

import (
	"os"
	"syscall"
	"time"
)

// ShutdownConfig holds shutdown timing settings.
type ShutdownConfig struct {
	GracePeriod   time.Duration // Time to wait for graceful exit
	TermWait      time.Duration // Time to wait after SIGTERM
	ForceKillWait time.Duration // Time to wait before SIGKILL
}

// DefaultShutdownConfig returns default shutdown settings.
func DefaultShutdownConfig() ShutdownConfig {
	return ShutdownConfig{
		GracePeriod:   2 * time.Second,
		TermWait:      1 * time.Second,
		ForceKillWait: 500 * time.Millisecond,
	}
}

// GracefulShutdown performs a graceful shutdown of a process.
// Steps:
// 1. Close stdin (signal end of input)
// 2. Wait for graceful exit (GracePeriod)
// 3. Send SIGTERM
// 4. Wait (TermWait)
// 5. Send SIGKILL
func GracefulShutdown(p *os.Process, cfg ShutdownConfig) error {
	if p == nil {
		return nil
	}

	// Create a channel to track process exit
	done := make(chan error, 1)
	go func() {
		_, err := p.Wait()
		done <- err
	}()

	// Step 1-2: Wait for graceful exit
	select {
	case <-done:
		return nil
	case <-time.After(cfg.GracePeriod):
	}

	// Step 3: Send SIGTERM
	if err := p.Signal(syscall.SIGTERM); err != nil {
		// Process might already be dead
		if err == os.ErrProcessDone {
			return nil
		}
	}

	// Step 4: Wait after SIGTERM
	select {
	case <-done:
		return nil
	case <-time.After(cfg.TermWait):
	}

	// Step 5: Send SIGKILL
	if err := p.Signal(syscall.SIGKILL); err != nil {
		if err == os.ErrProcessDone {
			return nil
		}
		return err
	}

	// Wait for SIGKILL to take effect
	select {
	case <-done:
		return nil
	case <-time.After(cfg.ForceKillWait):
		// Force kill should always work, but just in case
		return nil
	}
}

// ProcessCloser wraps a process with graceful shutdown.
type ProcessCloser struct {
	process *os.Process
	stdin   interface{ Close() error }
	cfg     ShutdownConfig
}

// NewProcessCloser creates a new process closer.
func NewProcessCloser(process *os.Process, stdin interface{ Close() error }, cfg ShutdownConfig) *ProcessCloser {
	return &ProcessCloser{
		process: process,
		stdin:   stdin,
		cfg:     cfg,
	}
}

// Close performs graceful shutdown.
func (c *ProcessCloser) Close() error {
	// Close stdin first to signal end of input
	if c.stdin != nil {
		_ = c.stdin.Close()
	}

	// Perform graceful shutdown
	return GracefulShutdown(c.process, c.cfg)
}

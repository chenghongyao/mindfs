package agent

import (
	"context"
	"errors"

	"mindfs/server/internal/agent/acp"
	"mindfs/server/internal/agent/claude"
	"mindfs/server/internal/agent/codex"
	"mindfs/server/internal/agent/gemini"
)

// Process is the unified interface for all agent processes.
type Process interface {
	// SendMessage sends a message and streams responses via callback.
	SendMessage(ctx context.Context, content string, onUpdate func(acp.SessionUpdate)) error

	// SessionID returns the current session ID.
	SessionID() string

	// Close terminates the process.
	Close() error
}

// StartProcess spawns an agent process based on the definition.
func StartProcess(ctx context.Context, def Definition, rootPath string) (Process, error) {
	args := def.BuildArgs(rootPath)
	cwd := def.ResolveCwd(rootPath)

	switch def.Protocol {
	case ProtocolStreamJSON:
		return claude.Start(ctx, args, cwd, def.Env)

	case ProtocolACP:
		proc, err := gemini.Start(ctx, args, cwd, def.Env)
		if err != nil {
			return nil, err
		}
		// Initialize ACP handshake
		if err := proc.Initialize(ctx); err != nil {
			_ = proc.Close()
			return nil, err
		}
		if err := proc.NewSession(ctx, cwd); err != nil {
			_ = proc.Close()
			return nil, err
		}
		return proc, nil

	case ProtocolMCP:
		proc, err := codex.Start(ctx, args, cwd, def.Env)
		if err != nil {
			return nil, err
		}
		// Initialize MCP handshake
		if err := proc.Initialize(ctx); err != nil {
			_ = proc.Close()
			return nil, err
		}
		return proc, nil

	default:
		return nil, errors.New("unsupported protocol: " + string(def.Protocol))
	}
}

// StreamChunk is the legacy streaming chunk type for backward compatibility.
type StreamChunk struct {
	Type    string `json:"type"`
	Content string `json:"content,omitempty"`
	Tool    string `json:"tool,omitempty"`
	Path    string `json:"path,omitempty"`
	Size    int64  `json:"size,omitempty"`
	Percent int    `json:"percent,omitempty"`
}

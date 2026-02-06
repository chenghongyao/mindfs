package codex

import (
	"context"
	"encoding/json"
	"io"
	"os/exec"
	"sync"
	"sync/atomic"

	"mindfs/server/internal/agent/acp"
)

// MCP JSON-RPC message types
type mcpRequest struct {
	JSONRPC string `json:"jsonrpc"`
	ID      int64  `json:"id"`
	Method  string `json:"method"`
	Params  any    `json:"params,omitempty"`
}

type mcpResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      int64           `json:"id,omitempty"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *mcpError       `json:"error,omitempty"`
}

type mcpError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type mcpNotification struct {
	JSONRPC string          `json:"jsonrpc"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

// Process manages a Codex CLI process using MCP protocol.
type Process struct {
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	stdout io.ReadCloser
	mu     sync.Mutex

	sessionID      string
	conversationID string
	nextID         atomic.Int64

	// Channel for receiving parsed messages
	msgCh chan any
}

// Start spawns a Codex process with MCP server mode.
func Start(ctx context.Context, command string, args []string, cwd string, env map[string]string) (*Process, error) {
	if command == "" {
		command = "codex"
	}
	cmd := exec.CommandContext(ctx, command, args...)
	if cwd != "" {
		cmd.Dir = cwd
	}
	if len(env) > 0 {
		cmd.Env = cmd.Environ()
		for k, v := range env {
			cmd.Env = append(cmd.Env, k+"="+v)
		}
	}

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	cmd.Stderr = nil

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	p := &Process{
		cmd:    cmd,
		stdin:  stdin,
		stdout: stdout,
		msgCh:  make(chan any, 100),
	}

	// Start reading messages
	go p.readLoop()

	return p, nil
}

func (p *Process) readLoop() {
	decoder := json.NewDecoder(p.stdout)
	for {
		var raw json.RawMessage
		if err := decoder.Decode(&raw); err != nil {
			close(p.msgCh)
			return
		}

		// Try to determine message type
		var peek struct {
			Method string `json:"method"`
			ID     *int64 `json:"id"`
		}
		_ = json.Unmarshal(raw, &peek)

		if peek.Method == "codex/event" {
			var notif mcpNotification
			_ = json.Unmarshal(raw, &notif)
			p.msgCh <- notif
		} else if peek.ID != nil {
			var resp mcpResponse
			_ = json.Unmarshal(raw, &resp)
			p.msgCh <- resp
		}
	}
}

// Initialize performs MCP handshake.
func (p *Process) Initialize(ctx context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	req := mcpRequest{
		JSONRPC: "2.0",
		ID:      p.nextID.Add(1),
		Method:  "initialize",
		Params: map[string]any{
			"protocolVersion": "2024-11-05",
			"capabilities": map[string]any{
				"elicitation": map[string]any{},
			},
			"clientInfo": map[string]string{
				"name":    "mindfs",
				"version": "1.0.0",
			},
		},
	}
	if err := p.writeJSON(req); err != nil {
		return err
	}

	// Wait for response
	for msg := range p.msgCh {
		if resp, ok := msg.(mcpResponse); ok && resp.ID == req.ID {
			break
		}
	}

	// Send initialized notification
	notif := mcpRequest{
		JSONRPC: "2.0",
		Method:  "notifications/initialized",
	}
	return p.writeJSON(notif)
}

// SendMessage sends a prompt via MCP tool call.
func (p *Process) SendMessage(ctx context.Context, content string, onUpdate func(acp.SessionUpdate)) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Call codex tool
	params := map[string]any{
		"prompt":            content,
		"sandbox":           "workspace-write",
		"approval-policy":   "on-failure",
	}
	if p.sessionID != "" {
		params["experimental_resume"] = p.sessionID
	}

	req := mcpRequest{
		JSONRPC: "2.0",
		ID:      p.nextID.Add(1),
		Method:  "tools/call",
		Params: map[string]any{
			"name":      "codex",
			"arguments": params,
		},
	}
	if err := p.writeJSON(req); err != nil {
		return err
	}

	// Process events until tool call completes
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case msg, ok := <-p.msgCh:
			if !ok {
				return io.EOF
			}

			switch m := msg.(type) {
			case mcpNotification:
				if m.Method == "codex/event" {
					update := p.eventToUpdate(m.Params)
					if update.Type != "" && onUpdate != nil {
						update.SessionID = p.sessionID
						onUpdate(update)
					}
				}
			case mcpResponse:
				if m.ID == req.ID {
					// Tool call completed
					if onUpdate != nil {
						onUpdate(acp.SessionUpdate{
							Type:      acp.UpdateTypeMessageDone,
							SessionID: p.sessionID,
						})
					}
					return nil
				}
			}
		}
	}
}

func (p *Process) eventToUpdate(params json.RawMessage) acp.SessionUpdate {
	var event struct {
		Msg struct {
			Type      string `json:"type"`
			SessionID string `json:"session_id"`
			Data      struct {
				Delta string `json:"delta"`
			} `json:"data"`
		} `json:"msg"`
	}
	_ = json.Unmarshal(params, &event)

	// Extract session ID
	if event.Msg.SessionID != "" {
		p.sessionID = event.Msg.SessionID
	}

	switch event.Msg.Type {
	case "message_delta":
		data, _ := json.Marshal(acp.MessageChunk{Content: event.Msg.Data.Delta})
		return acp.SessionUpdate{Type: acp.UpdateTypeMessageChunk, Data: data}
	}
	return acp.SessionUpdate{}
}

func (p *Process) writeJSON(v any) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	_, err = p.stdin.Write(append(data, '\n'))
	return err
}

// SessionID returns the current session ID.
func (p *Process) SessionID() string {
	return p.sessionID
}

// Close terminates the process.
func (p *Process) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.stdin != nil {
		_ = p.stdin.Close()
	}
	if p.cmd != nil && p.cmd.Process != nil {
		_ = p.cmd.Process.Kill()
	}
	return nil
}

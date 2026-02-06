package gemini

import (
	"context"
	"encoding/json"
	"io"
	"os/exec"
	"sync"
	"sync/atomic"

	"mindfs/server/internal/agent/acp"
)

// Process manages a Gemini CLI process using ACP protocol.
type Process struct {
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	stdout io.ReadCloser
	parser *acp.Parser
	idle   *acp.IdleDetector
	mu     sync.Mutex

	sessionID string
	nextID    atomic.Int64
}

// Start spawns a Gemini process with ACP mode.
func Start(ctx context.Context, command string, args []string, cwd string, env map[string]string) (*Process, error) {
	if command == "" {
		command = "gemini"
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

	return &Process{
		cmd:    cmd,
		stdin:  stdin,
		stdout: stdout,
		parser: acp.NewParser(stdout),
	}, nil
}

// Initialize performs ACP handshake.
func (p *Process) Initialize(ctx context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Send initialize request
	initReq := map[string]any{
		"jsonrpc": "2.0",
		"id":      p.nextID.Add(1),
		"method":  "initialize",
		"params": map[string]any{
			"protocolVersion": 1,
			"clientInfo": map[string]string{
				"name":    "mindfs",
				"version": "1.0.0",
			},
		},
	}
	if err := p.writeJSON(initReq); err != nil {
		return err
	}

	// Wait for initialize response
	for {
		msg, err := p.parser.ReadMessage()
		if err != nil {
			return err
		}
		if msg.Result != nil {
			break
		}
	}
	return nil
}

// NewSession creates a new ACP session.
func (p *Process) NewSession(ctx context.Context, cwd string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	req := map[string]any{
		"jsonrpc": "2.0",
		"id":      p.nextID.Add(1),
		"method":  "newSession",
		"params": map[string]any{
			"cwd": cwd,
		},
	}
	if err := p.writeJSON(req); err != nil {
		return err
	}

	// Wait for session response
	for {
		msg, err := p.parser.ReadMessage()
		if err != nil {
			return err
		}
		if msg.Result != nil {
			var result struct {
				SessionID string `json:"sessionId"`
			}
			_ = json.Unmarshal(msg.Result, &result)
			p.sessionID = result.SessionID
			break
		}
	}
	return nil
}

// SendMessage sends a prompt and streams responses.
func (p *Process) SendMessage(ctx context.Context, content string, onUpdate func(acp.SessionUpdate)) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	req := map[string]any{
		"jsonrpc": "2.0",
		"id":      p.nextID.Add(1),
		"method":  "prompt",
		"params": map[string]any{
			"sessionId": p.sessionID,
			"prompt": []map[string]string{
				{"type": "text", "text": content},
			},
		},
	}
	if err := p.writeJSON(req); err != nil {
		return err
	}

	// Read responses
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		msg, err := p.parser.ReadMessage()
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}

		// Check for turn complete
		if msg.Method == acp.MethodAgentTurnComplete {
			if onUpdate != nil {
				onUpdate(acp.SessionUpdate{
					Type:      acp.UpdateTypeMessageDone,
					SessionID: p.sessionID,
				})
			}
			return nil
		}

		// Convert to session update
		update := p.toSessionUpdate(msg)
		if update.Type != "" && onUpdate != nil {
			update.SessionID = p.sessionID
			onUpdate(update)
		}
	}
}

func (p *Process) toSessionUpdate(msg acp.ACPMessage) acp.SessionUpdate {
	switch msg.Method {
	case acp.MethodAssistantMessage:
		var params acp.AssistantMessageParams
		_ = json.Unmarshal(msg.Params, &params)
		data, _ := json.Marshal(acp.MessageChunk{Content: params.Content})
		return acp.SessionUpdate{Type: acp.UpdateTypeMessageChunk, Data: data}

	case acp.MethodAssistantThinking:
		var params acp.AssistantThinkingParams
		_ = json.Unmarshal(msg.Params, &params)
		data, _ := json.Marshal(acp.ThoughtChunk{Content: params.Content})
		return acp.SessionUpdate{Type: acp.UpdateTypeThoughtChunk, Data: data}

	case acp.MethodToolUseStart:
		var params acp.ToolUseStartParams
		_ = json.Unmarshal(msg.Params, &params)
		data, _ := json.Marshal(acp.ToolCall{
			CallID:    params.ToolUseID,
			Name:      params.Name,
			Arguments: params.Input,
			Status:    "running",
		})
		return acp.SessionUpdate{Type: acp.UpdateTypeToolCall, Data: data}

	case acp.MethodToolUseResult:
		var params acp.ToolUseResultParams
		_ = json.Unmarshal(msg.Params, &params)
		status := "complete"
		if params.IsError {
			status = "failed"
		}
		data, _ := json.Marshal(acp.ToolCallUpdate{
			CallID: params.ToolUseID,
			Status: status,
			Result: params.Output,
		})
		return acp.SessionUpdate{Type: acp.UpdateTypeToolUpdate, Data: data}

	case acp.MethodRequestPermission:
		var params acp.RequestPermissionParams
		_ = json.Unmarshal(msg.Params, &params)
		data, _ := json.Marshal(acp.PermissionRequest{
			Permission: params.Permission,
			Resource:   params.Resource,
			Action:     params.Action,
		})
		return acp.SessionUpdate{Type: acp.UpdateTypePermissionReq, Data: data}
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

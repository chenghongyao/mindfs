package claude

import (
	"context"
	"encoding/json"
	"io"
	"os/exec"
	"sync"

	"mindfs/server/internal/agent/acp"
)

// Process manages a Claude CLI process using stream-json protocol.
type Process struct {
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	stdout io.ReadCloser
	parser *Parser
	mu     sync.Mutex

	sessionID string
	model     string
}

// Start spawns a Claude process with stream-json format.
func Start(ctx context.Context, command string, args []string, cwd string, env map[string]string) (*Process, error) {
	if command == "" {
		command = "claude"
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
	// Ignore stderr for now
	cmd.Stderr = nil

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	return &Process{
		cmd:    cmd,
		stdin:  stdin,
		stdout: stdout,
		parser: NewParser(stdout),
	}, nil
}

// SendMessage sends a user message and streams responses.
func (p *Process) SendMessage(ctx context.Context, content string, onUpdate func(acp.SessionUpdate)) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Send user message
	msg := NewUserMessage(content)
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	if _, err := p.stdin.Write(append(data, '\n')); err != nil {
		return err
	}

	// Read responses until result
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

		// Handle system init
		if msg.Type == TypeSystem && msg.Subtype == "init" {
			p.sessionID = msg.SessionID
			p.model = msg.Model
			continue
		}

		// Handle result (end of response)
		if msg.Type == TypeResult {
			// Send completion update
			if onUpdate != nil {
				onUpdate(acp.SessionUpdate{
					Type:      acp.UpdateTypeMessageDone,
					SessionID: p.sessionID,
				})
			}
			return nil
		}

		// Handle assistant message
		if msg.Type == TypeAssistant && msg.Message != nil {
			for _, part := range msg.Message.Content {
				update := p.partToUpdate(part)
				if update.Type != "" && onUpdate != nil {
					update.SessionID = p.sessionID
					onUpdate(update)
				}
			}
		}
	}
}

// partToUpdate converts a content part to a session update.
func (p *Process) partToUpdate(part ContentPart) acp.SessionUpdate {
	switch part.Type {
	case "text":
		data, _ := json.Marshal(acp.MessageChunk{Content: part.Text})
		return acp.SessionUpdate{Type: acp.UpdateTypeMessageChunk, Data: data}

	case "tool_use":
		data, _ := json.Marshal(acp.ToolCall{
			CallID:    part.ID,
			Name:      part.Name,
			Arguments: part.Input,
			Status:    "running",
		})
		return acp.SessionUpdate{Type: acp.UpdateTypeToolCall, Data: data}

	case "tool_result":
		data, _ := json.Marshal(acp.ToolCallUpdate{
			CallID: part.ToolUseID,
			Status: "complete",
			Result: part.Content,
		})
		return acp.SessionUpdate{Type: acp.UpdateTypeToolUpdate, Data: data}
	}
	return acp.SessionUpdate{}
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

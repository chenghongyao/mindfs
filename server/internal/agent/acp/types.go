package acp

import "encoding/json"

// ACPMessage represents a JSON-RPC 2.0 style message from the agent.
type ACPMessage struct {
	JSONRPC string          `json:"jsonrpc,omitempty"`
	Method  string          `json:"method,omitempty"`
	ID      json.RawMessage `json:"id,omitempty"`
	Params  json.RawMessage `json:"params,omitempty"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *RPCError       `json:"error,omitempty"`
}

// RPCError represents a JSON-RPC error.
type RPCError struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data,omitempty"`
}

// SessionUpdate types sent to frontend.
const (
	UpdateTypeMessageChunk  = "agent_message_chunk"
	UpdateTypeThoughtChunk  = "agent_thought_chunk"
	UpdateTypeToolCall      = "tool_call"
	UpdateTypeToolUpdate    = "tool_call_update"
	UpdateTypeMessageDone   = "agent_message_complete"
	UpdateTypePermissionReq = "permission_request"
)

// SessionUpdate is sent to the frontend via WebSocket.
type SessionUpdate struct {
	Type      string          `json:"type"`
	SessionID string          `json:"sessionId"`
	Data      json.RawMessage `json:"data,omitempty"`
}

// MessageChunk contains a text chunk from the agent.
type MessageChunk struct {
	Content string `json:"content"`
}

// ThoughtChunk contains agent thinking/reasoning.
type ThoughtChunk struct {
	Content string `json:"content"`
}

// ToolCall represents a tool invocation.
type ToolCall struct {
	CallID    string          `json:"callId"`
	Name      string          `json:"name"`
	Arguments json.RawMessage `json:"arguments,omitempty"`
	Status    string          `json:"status"` // pending, running, complete, failed
}

// ToolCallUpdate updates tool call status.
type ToolCallUpdate struct {
	CallID string `json:"callId"`
	Status string `json:"status"`
	Result string `json:"result,omitempty"`
	Error  string `json:"error,omitempty"`
}

// PermissionRequest asks user for permission.
type PermissionRequest struct {
	RequestID  string `json:"requestId"`
	Permission string `json:"permission"`
	Resource   string `json:"resource,omitempty"`
	Action     string `json:"action,omitempty"`
}

// PermissionResponse is the user's answer.
type PermissionResponse struct {
	RequestID string `json:"requestId"`
	Granted   bool   `json:"granted"`
}

// AgentOutput methods from Happy project.
const (
	MethodAssistantMessage       = "assistant_message"
	MethodAssistantThinking      = "assistant_thinking"
	MethodToolUseStart           = "tool_use_start"
	MethodToolUseResult          = "tool_use_result"
	MethodRequestPermission      = "request_permission"
	MethodSystemMessage          = "system_message"
	MethodAgentTurnComplete      = "agent_turn_complete"
)

// AssistantMessageParams for assistant_message method.
type AssistantMessageParams struct {
	Content string `json:"content"`
}

// AssistantThinkingParams for assistant_thinking method.
type AssistantThinkingParams struct {
	Content string `json:"content"`
}

// ToolUseStartParams for tool_use_start method.
type ToolUseStartParams struct {
	ToolUseID string          `json:"tool_use_id"`
	Name      string          `json:"name"`
	Input     json.RawMessage `json:"input,omitempty"`
}

// ToolUseResultParams for tool_use_result method.
type ToolUseResultParams struct {
	ToolUseID string `json:"tool_use_id"`
	Output    string `json:"output,omitempty"`
	IsError   bool   `json:"is_error,omitempty"`
}

// RequestPermissionParams for request_permission method.
type RequestPermissionParams struct {
	Permission string `json:"permission"`
	Resource   string `json:"resource,omitempty"`
	Action     string `json:"action,omitempty"`
}

// SystemMessageParams for system_message method.
type SystemMessageParams struct {
	Message string `json:"message"`
	Level   string `json:"level,omitempty"` // info, warning, error
}

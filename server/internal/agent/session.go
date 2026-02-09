package agent

import "context"

// Session is the interface for all agent sessions.
type Session interface {
	// SendMessage sends a message to the current session.
	SendMessage(ctx context.Context, content string) error

	// OnUpdate registers a callback for streaming updates.
	OnUpdate(onUpdate func(Event))

	// SessionID returns the current session ID.
	SessionID() string

	// Close terminates the session (not the process).
	Close() error
}

// EventType defines the type of a session event.
type EventType string

const (
	EventTypeMessageChunk EventType = "message_chunk"
	EventTypeThoughtChunk EventType = "thought_chunk"
	EventTypeToolCall     EventType = "tool_call"
	EventTypeToolUpdate   EventType = "tool_update"
	EventTypeMessageDone  EventType = "message_done"
)

// Event is a normalized session update emitted by any agent backend.
type Event struct {
	Type      EventType
	SessionID string
	Data      any
}

type MessageChunk struct {
	Content string
}

type ThoughtChunk struct {
	Content string
}

type ToolKind string

const (
	ToolKindRead       ToolKind = "read"
	ToolKindEdit       ToolKind = "edit"
	ToolKindDelete     ToolKind = "delete"
	ToolKindMove       ToolKind = "move"
	ToolKindSearch     ToolKind = "search"
	ToolKindExecute    ToolKind = "execute"
	ToolKindThink      ToolKind = "think"
	ToolKindFetch      ToolKind = "fetch"
	ToolKindSwitchMode ToolKind = "switch_mode"
	ToolKindOther      ToolKind = "other"
)

type ToolCallLocation struct {
	Path string
	Line *int
}

type ToolCall struct {
	CallID    string
	Name      string
	Status    string
	Kind      ToolKind
	Locations []ToolCallLocation
}

func (tc ToolCall) IsWriteOperation() bool {
	switch tc.Kind {
	case ToolKindEdit, ToolKindDelete, ToolKindMove:
		return true
	default:
		return false
	}
}

func (tc ToolCall) GetAffectedPaths() []string {
	paths := make([]string, 0, len(tc.Locations))
	for _, loc := range tc.Locations {
		if loc.Path != "" {
			paths = append(paths, loc.Path)
		}
	}
	return paths
}

type ToolCallUpdate struct {
	CallID string
	Status string
	Result string
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

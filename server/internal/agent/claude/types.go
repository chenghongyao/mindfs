package claude

import (
	"bufio"
	"encoding/json"
	"io"
)

// Message types from Claude stream-json output.
const (
	TypeSystem    = "system"
	TypeUser      = "user"
	TypeAssistant = "assistant"
	TypeResult    = "result"
)

// Message is a Claude stream-json message.
type Message struct {
	Type    string          `json:"type"`
	Subtype string          `json:"subtype,omitempty"`
	Message *ContentMessage `json:"message,omitempty"`

	// System message fields
	SessionID string   `json:"session_id,omitempty"`
	Model     string   `json:"model,omitempty"`
	Tools     []string `json:"tools,omitempty"`

	// Result fields
	NumTurns     int     `json:"num_turns,omitempty"`
	DurationMs   int     `json:"duration_ms,omitempty"`
	TotalCostUSD float64 `json:"total_cost_usd,omitempty"`
}

// ContentMessage wraps role and content.
type ContentMessage struct {
	Role    string        `json:"role"`
	Content []ContentPart `json:"content"`
}

// ContentPart is a single content element.
type ContentPart struct {
	Type string `json:"type"`

	// Text content
	Text string `json:"text,omitempty"`

	// Tool use
	ID    string          `json:"id,omitempty"`
	Name  string          `json:"name,omitempty"`
	Input json.RawMessage `json:"input,omitempty"`

	// Tool result
	ToolUseID string `json:"tool_use_id,omitempty"`
	Content   string `json:"content,omitempty"`
}

// UserMessage creates a user message for stdin.
type UserMessage struct {
	Type    string             `json:"type"`
	Message UserMessageContent `json:"message"`
}

// UserMessageContent is the content of a user message.
type UserMessageContent struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// NewUserMessage creates a user message to send to Claude.
func NewUserMessage(content string) UserMessage {
	return UserMessage{
		Type: "user",
		Message: UserMessageContent{
			Role:    "user",
			Content: content,
		},
	}
}

// ControlRequest is sent to Claude for permission responses.
type ControlRequest struct {
	RequestID string         `json:"request_id"`
	Type      string         `json:"type"` // "control_request"
	Request   ControlPayload `json:"request"`
}

// ControlPayload is the payload of a control request.
type ControlPayload struct {
	Subtype string `json:"subtype"` // "interrupt", etc.
}

// ControlResponse is received from Claude for control requests.
type ControlResponse struct {
	Type     string                 `json:"type"` // "control_response"
	Response ControlResponsePayload `json:"response"`
}

// ControlResponsePayload is the payload of a control response.
type ControlResponsePayload struct {
	Subtype   string `json:"subtype"` // "success" or "error"
	RequestID string `json:"request_id"`
	Error     string `json:"error,omitempty"`
}

// Parser reads Claude stream-json messages.
type Parser struct {
	reader *bufio.Reader
}

// NewParser creates a new Claude message parser.
func NewParser(r io.Reader) *Parser {
	return &Parser{reader: bufio.NewReader(r)}
}

// ReadMessage reads the next message from stdout.
func (p *Parser) ReadMessage() (Message, error) {
	line, err := p.reader.ReadBytes('\n')
	if err != nil {
		return Message{}, err
	}
	if len(line) == 0 {
		return p.ReadMessage()
	}
	var msg Message
	if err := json.Unmarshal(line, &msg); err != nil {
		return Message{}, err
	}
	return msg, nil
}

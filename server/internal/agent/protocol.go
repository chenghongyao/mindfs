package agent

// Protocol defines the communication protocol for an agent.
type Protocol string

const (
	// ProtocolStreamJSON is Claude's stream-json format.
	// stdin/stdout: one JSON per line
	// Command: claude --output-format stream-json --input-format stream-json
	ProtocolStreamJSON Protocol = "stream-json"

	// ProtocolACP is the Agent Client Protocol (JSON-RPC 2.0 over ndJSON).
	// Used by Gemini with --experimental-acp flag.
	ProtocolACP Protocol = "acp"

	// ProtocolMCP is the Model Context Protocol.
	// Used by Codex with mcp-server subcommand.
	ProtocolMCP Protocol = "mcp"
)

// DefaultProtocol returns the default protocol for known agents.
func DefaultProtocol(agentName string) Protocol {
	switch agentName {
	case "claude":
		return ProtocolStreamJSON
	case "gemini":
		return ProtocolACP
	case "codex":
		return ProtocolMCP
	default:
		return ProtocolStreamJSON
	}
}

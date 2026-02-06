package api

// JSONRPCRequest defines a generic JSON-RPC request.
type JSONRPCRequest struct {
	JSONRPC string         `json:"jsonrpc"`
	ID      string         `json:"id"`
	Method  string         `json:"method"`
	Params  map[string]any `json:"params,omitempty"`
}

// JSONRPCResponse defines a generic JSON-RPC response.
type JSONRPCResponse struct {
	JSONRPC string         `json:"jsonrpc"`
	ID      string         `json:"id"`
	Result  map[string]any `json:"result,omitempty"`
	Error   *JSONRPCError  `json:"error,omitempty"`
}

// JSONRPCError captures an error response.
type JSONRPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

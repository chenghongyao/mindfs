package agent

import (
	"encoding/json"
	"path/filepath"
	"strings"

	"mindfs/server/internal/agent/acp"
	"mindfs/server/internal/fs"
)

// FileWriteTracker tracks file writes from agent tool calls
type FileWriteTracker struct {
	managedDir  string
	sessionKey  string
	sessionName string
	agentName   string
}

// NewFileWriteTracker creates a new file write tracker
func NewFileWriteTracker(managedDir, sessionKey, sessionName, agentName string) *FileWriteTracker {
	return &FileWriteTracker{
		managedDir:  managedDir,
		sessionKey:  sessionKey,
		sessionName: sessionName,
		agentName:   agentName,
	}
}

// TrackToolCall checks if a tool call is a file write and records metadata
func (t *FileWriteTracker) TrackToolCall(toolName string, input json.RawMessage) {
	// Check for common file write tool names
	writeTools := map[string]bool{
		"write":       true,
		"Write":       true,
		"write_file":  true,
		"WriteFile":   true,
		"create_file": true,
		"CreateFile":  true,
		"edit":        true,
		"Edit":        true,
	}

	if !writeTools[toolName] {
		return
	}

	// Try to extract file path from input
	var params struct {
		Path     string `json:"path"`
		FilePath string `json:"file_path"`
		Filename string `json:"filename"`
	}

	if err := json.Unmarshal(input, &params); err != nil {
		return
	}

	filePath := params.Path
	if filePath == "" {
		filePath = params.FilePath
	}
	if filePath == "" {
		filePath = params.Filename
	}
	if filePath == "" {
		return
	}

	// Convert to relative path if absolute
	if filepath.IsAbs(filePath) {
		// Try to make it relative to the root
		rootPath := filepath.Dir(t.managedDir) // .mindfs parent is root
		if rel, err := filepath.Rel(rootPath, filePath); err == nil && !strings.HasPrefix(rel, "..") {
			filePath = rel
		}
	}

	// Record file metadata
	_ = fs.UpdateFileMetaFull(
		t.managedDir,
		filePath,
		t.sessionKey,
		t.sessionName,
		t.agentName,
		"agent",
	)
}

// TrackToolResult checks tool result for file creation confirmation
func (t *FileWriteTracker) TrackToolResult(toolUseID, output string, isError bool) {
	if isError {
		return
	}

	// Some agents report file creation in the result
	// Try to extract file path from common patterns
	patterns := []string{
		"Created file:",
		"Wrote to:",
		"File written:",
		"Successfully created",
	}

	for _, pattern := range patterns {
		if idx := strings.Index(output, pattern); idx >= 0 {
			// Extract path after pattern
			rest := strings.TrimSpace(output[idx+len(pattern):])
			// Take first word/path
			parts := strings.Fields(rest)
			if len(parts) > 0 {
				filePath := strings.Trim(parts[0], "\"'`")
				if filePath != "" && !strings.HasPrefix(filePath, "/") {
					_ = fs.UpdateFileMetaFull(
						t.managedDir,
						filePath,
						t.sessionKey,
						t.sessionName,
						t.agentName,
						"agent",
					)
				}
			}
			break
		}
	}
}

// ProcessUpdate processes a session update and tracks file writes
func (t *FileWriteTracker) ProcessUpdate(update acp.SessionUpdate) {
	switch update.Type {
	case acp.UpdateTypeToolCall:
		var toolCall acp.ToolCall
		if err := json.Unmarshal(update.Data, &toolCall); err == nil {
			t.TrackToolCall(toolCall.Name, toolCall.Arguments)
		}

	case acp.UpdateTypeToolUpdate:
		var toolUpdate acp.ToolCallUpdate
		if err := json.Unmarshal(update.Data, &toolUpdate); err == nil {
			t.TrackToolResult(toolUpdate.CallID, toolUpdate.Result, toolUpdate.Error != "")
		}
	}
}

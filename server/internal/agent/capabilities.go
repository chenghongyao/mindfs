package agent

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

// Capability represents an agent capability
type Capability struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Agent       string `json:"agent"`
}

// CapabilitiesConfig stores agent capabilities
type CapabilitiesConfig struct {
	Capabilities map[string][]Capability `json:"capabilities"` // agent -> capabilities
}

// CapabilitiesStore manages agent capabilities
type CapabilitiesStore struct {
	mu       sync.RWMutex
	config   *CapabilitiesConfig
	filePath string
}

// NewCapabilitiesStore creates a new capabilities store
func NewCapabilitiesStore() (*CapabilitiesStore, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	configDir := filepath.Join(homeDir, ".config", "mindfs")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return nil, err
	}

	filePath := filepath.Join(configDir, "capabilities.json")
	store := &CapabilitiesStore{
		filePath: filePath,
		config:   defaultCapabilities(),
	}

	// Load existing config
	if err := store.load(); err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	return store, nil
}

// defaultCapabilities returns built-in agent capabilities
func defaultCapabilities() *CapabilitiesConfig {
	return &CapabilitiesConfig{
		Capabilities: map[string][]Capability{
			"claude": {
				{ID: "claude:read_file", Name: "读取文件", Description: "读取指定文件的内容", Agent: "claude"},
				{ID: "claude:write_file", Name: "写入文件", Description: "创建或修改文件内容", Agent: "claude"},
				{ID: "claude:search", Name: "搜索代码", Description: "在代码库中搜索内容", Agent: "claude"},
				{ID: "claude:run_command", Name: "执行命令", Description: "在终端执行 shell 命令", Agent: "claude"},
				{ID: "claude:web_search", Name: "网络搜索", Description: "搜索网络获取信息", Agent: "claude"},
			},
			"gemini": {
				{ID: "gemini:read_file", Name: "读取文件", Description: "读取指定文件的内容", Agent: "gemini"},
				{ID: "gemini:write_file", Name: "写入文件", Description: "创建或修改文件内容", Agent: "gemini"},
				{ID: "gemini:search", Name: "搜索代码", Description: "在代码库中搜索内容", Agent: "gemini"},
				{ID: "gemini:run_command", Name: "执行命令", Description: "在终端执行 shell 命令", Agent: "gemini"},
			},
			"codex": {
				{ID: "codex:read_file", Name: "读取文件", Description: "读取指定文件的内容", Agent: "codex"},
				{ID: "codex:write_file", Name: "写入文件", Description: "创建或修改文件内容", Agent: "codex"},
				{ID: "codex:run_command", Name: "执行命令", Description: "在终端执行 shell 命令", Agent: "codex"},
			},
		},
	}
}

// load reads capabilities from file
func (s *CapabilitiesStore) load() error {
	data, err := os.ReadFile(s.filePath)
	if err != nil {
		return err
	}

	var config CapabilitiesConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Merge with defaults
	if config.Capabilities == nil {
		config.Capabilities = make(map[string][]Capability)
	}

	// Add any missing default capabilities
	defaults := defaultCapabilities()
	for agent, caps := range defaults.Capabilities {
		if _, ok := config.Capabilities[agent]; !ok {
			config.Capabilities[agent] = caps
		}
	}

	s.config = &config
	return nil
}

// save writes capabilities to file
func (s *CapabilitiesStore) save() error {
	s.mu.RLock()
	data, err := json.MarshalIndent(s.config, "", "  ")
	s.mu.RUnlock()

	if err != nil {
		return err
	}

	return os.WriteFile(s.filePath, data, 0644)
}

// GetCapabilities returns capabilities for an agent
func (s *CapabilitiesStore) GetCapabilities(agent string) []Capability {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if caps, ok := s.config.Capabilities[agent]; ok {
		// Return a copy
		result := make([]Capability, len(caps))
		copy(result, caps)
		return result
	}
	return nil
}

// GetAllCapabilities returns all agent capabilities
func (s *CapabilitiesStore) GetAllCapabilities() map[string][]Capability {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Return a copy
	result := make(map[string][]Capability)
	for agent, caps := range s.config.Capabilities {
		capsCopy := make([]Capability, len(caps))
		copy(capsCopy, caps)
		result[agent] = capsCopy
	}
	return result
}

// AddCapability adds a custom capability for an agent
func (s *CapabilitiesStore) AddCapability(agent string, cap Capability) error {
	s.mu.Lock()
	if s.config.Capabilities == nil {
		s.config.Capabilities = make(map[string][]Capability)
	}
	s.config.Capabilities[agent] = append(s.config.Capabilities[agent], cap)
	s.mu.Unlock()

	return s.save()
}

// RemoveCapability removes a capability by ID
func (s *CapabilitiesStore) RemoveCapability(agent, capID string) error {
	s.mu.Lock()
	if caps, ok := s.config.Capabilities[agent]; ok {
		filtered := make([]Capability, 0, len(caps))
		for _, c := range caps {
			if c.ID != capID {
				filtered = append(filtered, c)
			}
		}
		s.config.Capabilities[agent] = filtered
	}
	s.mu.Unlock()

	return s.save()
}

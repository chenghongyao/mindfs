package audit

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const historyFileName = "history.jsonl"

// Record captures a minimal audit record for history.jsonl.
type Record struct {
	TS        time.Time `json:"ts"`
	Actor     string    `json:"actor"`
	Origin    string    `json:"origin"`
	Dir       string    `json:"dir"`
	Action    string    `json:"action"`
	Path      string    `json:"path,omitempty"`
	Status    string    `json:"status"`
	Handled   bool      `json:"handled"`
	LatencyMS int64     `json:"latency_ms,omitempty"`
	Summary   string    `json:"summary,omitempty"`
	Effects   []any     `json:"effects,omitempty"`
	Error     any       `json:"error,omitempty"`
	Meta      any       `json:"meta,omitempty"`
}

// Append writes a record to the managed directory's history.jsonl.
func Append(managedDir string, record Record) error {
	if managedDir == "" {
		return errors.New("managed dir required")
	}
	if record.TS.IsZero() {
		record.TS = time.Now().UTC()
	}
	payload, err := json.Marshal(record)
	if err != nil {
		return err
	}
	path := filepath.Join(managedDir, historyFileName)
	file, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer file.Close()
	if _, err := file.Write(append(payload, '\n')); err != nil {
		return err
	}
	return nil
}

// Writer writes audit entries to a JSONL file with buffering
type Writer struct {
	mu       sync.Mutex
	basePath string
	file     *os.File
}

// NewWriter creates a new audit writer for the given managed directory
func NewWriter(managedDir string) (*Writer, error) {
	historyPath := filepath.Join(managedDir, historyFileName)

	// Ensure directory exists
	if err := os.MkdirAll(managedDir, 0755); err != nil {
		return nil, err
	}

	// Open file in append mode
	file, err := os.OpenFile(historyPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}

	return &Writer{
		basePath: historyPath,
		file:     file,
	}, nil
}

// Write writes an audit entry to the log
func (w *Writer) Write(entry *Entry) error {
	if entry == nil {
		return nil
	}

	data, err := json.Marshal(entry)
	if err != nil {
		return err
	}

	w.mu.Lock()
	defer w.mu.Unlock()

	if _, err := w.file.Write(append(data, '\n')); err != nil {
		return err
	}

	return nil
}

// Close closes the audit writer
func (w *Writer) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.file != nil {
		return w.file.Close()
	}
	return nil
}

// WriterPool manages audit writers for multiple directories
type WriterPool struct {
	mu      sync.RWMutex
	writers map[string]*Writer
}

// NewWriterPool creates a new writer pool
func NewWriterPool() *WriterPool {
	return &WriterPool{
		writers: make(map[string]*Writer),
	}
}

// Get returns or creates a writer for the given managed directory
func (p *WriterPool) Get(managedDir string) (*Writer, error) {
	p.mu.RLock()
	if w, ok := p.writers[managedDir]; ok {
		p.mu.RUnlock()
		return w, nil
	}
	p.mu.RUnlock()

	p.mu.Lock()
	defer p.mu.Unlock()

	// Double-check after acquiring write lock
	if w, ok := p.writers[managedDir]; ok {
		return w, nil
	}

	w, err := NewWriter(managedDir)
	if err != nil {
		return nil, err
	}

	p.writers[managedDir] = w
	return w, nil
}

// Close closes all writers
func (p *WriterPool) Close() {
	p.mu.Lock()
	defer p.mu.Unlock()

	for _, w := range p.writers {
		_ = w.Close()
	}
	p.writers = make(map[string]*Writer)
}

// Logger provides a convenient interface for logging audit entries
type Logger struct {
	pool       *WriterPool
	rootID     string
	managedDir string
}

// NewLogger creates a new logger for a specific root
func NewLogger(pool *WriterPool, rootID, managedDir string) *Logger {
	return &Logger{
		pool:       pool,
		rootID:     rootID,
		managedDir: managedDir,
	}
}

// Log writes an audit entry
func (l *Logger) Log(entry *Entry) error {
	if l.pool == nil || l.managedDir == "" {
		return nil
	}

	entry.RootID = l.rootID

	w, err := l.pool.Get(l.managedDir)
	if err != nil {
		return err
	}

	return w.Write(entry)
}

// LogSession logs a session-related event
func (l *Logger) LogSession(action Action, actor Actor, sessionKey, agentName string, details map[string]any) error {
	entry := NewEntry(EntryTypeSession, action, actor).
		WithSession(sessionKey).
		WithAgent(agentName).
		WithDetails(details)
	return l.Log(entry)
}

// LogFile logs a file-related event
func (l *Logger) LogFile(action Action, actor Actor, path, sessionKey string, details map[string]any) error {
	entry := NewEntry(EntryTypeFile, action, actor).
		WithPath(path).
		WithSession(sessionKey).
		WithDetails(details)
	return l.Log(entry)
}

// LogView logs a view-related event
func (l *Logger) LogView(action Action, actor Actor, path, sessionKey string, details map[string]any) error {
	entry := NewEntry(EntryTypeView, action, actor).
		WithPath(path).
		WithSession(sessionKey).
		WithDetails(details)
	return l.Log(entry)
}

// LogSkill logs a skill-related event
func (l *Logger) LogSkill(action Action, actor Actor, sessionKey, skillName string, details map[string]any) error {
	entry := NewEntry(EntryTypeSkill, action, actor).
		WithSession(sessionKey).
		WithDetails(details)
	if entry.Details == nil {
		entry.Details = map[string]any{"skill": skillName}
	} else {
		entry.Details["skill"] = skillName
	}
	return l.Log(entry)
}

// LogDir logs a directory-related event
func (l *Logger) LogDir(action Action, actor Actor, path string, details map[string]any) error {
	entry := NewEntry(EntryTypeDir, action, actor).
		WithPath(path).
		WithDetails(details)
	return l.Log(entry)
}

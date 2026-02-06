package audit

import "time"

// EntryType defines the category of audit entry
type EntryType string

const (
	EntryTypeSession EntryType = "session"
	EntryTypeFile    EntryType = "file"
	EntryTypeView    EntryType = "view"
	EntryTypeSkill   EntryType = "skill"
	EntryTypeDir     EntryType = "dir"
	EntryTypeAuth    EntryType = "auth"
)

// Action defines the specific action within a type
type Action string

// Session actions
const (
	ActionSessionCreate  Action = "create"
	ActionSessionMessage Action = "message"
	ActionSessionClose   Action = "close"
	ActionSessionResume  Action = "resume"
)

// File actions
const (
	ActionFileOpen   Action = "open"
	ActionFileCreate Action = "create"
	ActionFileDelete Action = "delete"
	ActionFileRename Action = "rename"
	ActionFileUpdate Action = "update"
)

// View actions
const (
	ActionViewGenerate Action = "generate"
	ActionViewSwitch   Action = "switch"
	ActionViewRevert   Action = "revert"
)

// Skill actions
const (
	ActionSkillExecute Action = "execute"
	ActionSkillCancel  Action = "cancel"
)

// Dir actions
const (
	ActionDirAdd    Action = "add"
	ActionDirRemove Action = "remove"
)

// Auth actions
const (
	ActionAuthLogin  Action = "login"
	ActionAuthLogout Action = "logout"
)

// Actor defines who performed the action
type Actor string

const (
	ActorUser   Actor = "user"
	ActorAgent  Actor = "agent"
	ActorSystem Actor = "system"
)

// Entry represents a single audit log entry
type Entry struct {
	Ts        int64             `json:"ts"`
	Type      EntryType         `json:"type"`
	Action    Action            `json:"action"`
	Actor     Actor             `json:"actor"`
	Session   string            `json:"session,omitempty"`
	Path      string            `json:"path,omitempty"`
	Agent     string            `json:"agent,omitempty"`
	RootID    string            `json:"root_id,omitempty"`
	Details   map[string]any    `json:"details,omitempty"`
	Error     string            `json:"error,omitempty"`
	Duration  int64             `json:"duration_ms,omitempty"`
}

// NewEntry creates a new audit entry with current timestamp
func NewEntry(entryType EntryType, action Action, actor Actor) *Entry {
	return &Entry{
		Ts:     time.Now().UnixMilli(),
		Type:   entryType,
		Action: action,
		Actor:  actor,
	}
}

// WithSession sets the session key
func (e *Entry) WithSession(session string) *Entry {
	e.Session = session
	return e
}

// WithPath sets the file path
func (e *Entry) WithPath(path string) *Entry {
	e.Path = path
	return e
}

// WithAgent sets the agent name
func (e *Entry) WithAgent(agent string) *Entry {
	e.Agent = agent
	return e
}

// WithRootID sets the root ID
func (e *Entry) WithRootID(rootID string) *Entry {
	e.RootID = rootID
	return e
}

// WithDetails sets additional details
func (e *Entry) WithDetails(details map[string]any) *Entry {
	e.Details = details
	return e
}

// WithError sets the error message
func (e *Entry) WithError(err string) *Entry {
	e.Error = err
	return e
}

// WithDuration sets the duration in milliseconds
func (e *Entry) WithDuration(ms int64) *Entry {
	e.Duration = ms
	return e
}

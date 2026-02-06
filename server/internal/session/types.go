package session

import "time"

const (
	TypeChat  = "chat"
	TypeView  = "view"
	TypeSkill = "skill"

	StatusActive = "active"
	StatusIdle   = "idle"
	StatusClosed = "closed"
)

type Session struct {
	Key            string          `json:"key"`
	Type           string          `json:"type"`
	Agent          string          `json:"agent"`
	AgentSessionID *string         `json:"agent_session_id,omitempty"`
	Name           string          `json:"name"`
	Status         string          `json:"status"`
	Summary        *SessionSummary `json:"summary,omitempty"`
	Exchanges      []Exchange      `json:"exchanges"`
	RelatedFiles   []RelatedFile   `json:"related_files"`
	GeneratedView  string          `json:"generated_view,omitempty"`
	CreatedAt      time.Time       `json:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at"`
	ClosedAt       *time.Time      `json:"closed_at,omitempty"`
}

type SessionSummary struct {
	Title       string    `json:"title"`
	Description string    `json:"description"`
	KeyActions  []string  `json:"key_actions"`
	Outputs     []string  `json:"outputs"`
	GeneratedAt time.Time `json:"generated_at"`
}

type Exchange struct {
	Role      string    `json:"role"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
}

type RelatedFile struct {
	Path             string `json:"path"`
	Relation         string `json:"relation"`
	CreatedBySession bool   `json:"created_by_session"`
}

package domain

import (
	"time"
)

// Event событие
type Event struct {
	Timestamp time.Time `json:"timestamp"`
	Hostname  string    `json:"hostname"`
	Source    string    `json:"source"`
	RawLog    string    `json:"raw_log"`
	EventType string    `json:"event_type,omitempty"`
	Severity  string    `json:"severity,omitempty"`
	User      string    `json:"user,omitempty"`
	Process   string    `json:"process,omitempty"`
	Command   string    `json:"command,omitempty"`
}

// Batch набор событий
type Batch struct {
	AgentID   string    `json:"agent_id"`
	Timestamp time.Time `json:"timestamp"`
	Events    []Event   `json:"events"`
}

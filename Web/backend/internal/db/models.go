package db

import "time"

type DBRequest struct {
	Database string           `json:"database"`
	Command  string           `json:"operation"`
	Data     []map[string]any `json:"data,omitempty"`
	Query    map[string]any   `json:"query,omitempty"`
}

type DBResponse struct {
	Status  string           `json:"status"`
	Message string           `json:"message,omitempty"`
	Data    []map[string]any `json:"data,omitempty"`
	Count   int              `json:"count,omitempty"`
}

type DashboardStats struct {
	ActiveAgents  map[string]time.Time `json:"active_agents"`
	EventsByType  map[string]int       `json:"events_by_type"`
	SeverityDist  map[string]int       `json:"severity_distribution"`
	TopUsers      map[string]int       `json:"top_users"`
	TopProcesses  map[string]int       `json:"top_processes"`
	EventsPerHour map[int]int          `json:"events_per_hour"`
	LastLogins    []map[string]any     `json:"last_logins"`
}

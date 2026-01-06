package sender

import (
	"time"

	"github.com/Narotan/SIEM-Agent/internal/domain"
	"github.com/Narotan/SIEM-Agent/internal/filter"
	"github.com/Narotan/SIEM-Agent/internal/storage"
)

// DBRequest запрос к NoSQLdb
type DBRequest struct {
	Database string           `json:"database"`
	Command  string           `json:"operation"`
	Data     []map[string]any `json:"data,omitempty"`
}

// DBResponse ответ NoSQLdb
type DBResponse struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
	Count   int    `json:"count,omitempty"`
}

type Sender interface {
	Send(batch domain.Batch) error
	Close() error
}

type Config struct {
	AgentID      string
	BatchSize    int
	FlushTimeout time.Duration
	DiskBuffer   *storage.DiskBuffer
	Filter       filter.Filter
	RetryConfig  RetryConfig
}

type RetryConfig struct {
	MaxAttempts  int
	InitialDelay time.Duration
	MaxDelay     time.Duration
}

type Pipeline interface {
	Start(input <-chan domain.Event)
	Stop()
}

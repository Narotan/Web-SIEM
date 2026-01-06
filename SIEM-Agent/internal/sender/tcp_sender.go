package sender

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net"
	"sync"
	"time"

	"github.com/Narotan/SIEM-Agent/internal/domain"
)

type TCPSender struct {
	host       string
	port       int
	conn       net.Conn
	mu         sync.Mutex
	collection string
}

func NewTCPSender(host string, port int) *TCPSender {
	return &TCPSender{
		host:       host,
		port:       port,
		collection: "security_events",
	}
}

func (s *TCPSender) SetCollection(name string) {
	s.collection = name
}

func (s *TCPSender) connect() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.conn != nil {
		return nil
	}

	addr := fmt.Sprintf("%s:%d", s.host, s.port)
	conn, err := net.DialTimeout("tcp", addr, 5*time.Second)
	if err != nil {
		return fmt.Errorf("failed to connect to %s: %w", addr, err)
	}

	s.conn = conn
	log.Printf("Connected to NoSQLdb server at %s", addr)
	return nil
}

func (s *TCPSender) batchToDBRequest(batch domain.Batch) DBRequest {
	data := make([]map[string]any, len(batch.Events))

	for i, event := range batch.Events {
		data[i] = map[string]any{
			"agent_id":   batch.AgentID,
			"batch_time": batch.Timestamp.Format(time.RFC3339),
			"timestamp":  event.Timestamp.Format(time.RFC3339),
			"hostname":   event.Hostname,
			"source":     event.Source,
			"event_type": event.EventType,
			"severity":   event.Severity,
			"user":       event.User,
			"process":    event.Process,
			"command":    event.Command,
			"raw_log":    event.RawLog,
		}
	}

	return DBRequest{
		Database: s.collection,
		Command:  "insert",
		Data:     data,
	}
}

func (s *TCPSender) Send(batch domain.Batch) error {
	return s.SendWithRetry(batch, 3, 2*time.Second, 30*time.Second)
}

func (s *TCPSender) SendWithRetry(batch domain.Batch, maxAttempts int, initialDelay, maxDelay time.Duration) error {
	var lastErr error

	for attempt := 0; attempt < maxAttempts; attempt++ {
		if s.conn == nil {
			if err := s.connect(); err != nil {
				lastErr = err
				if attempt < maxAttempts-1 {
					delay := s.calculateBackoff(attempt, initialDelay, maxDelay)
					log.Printf("Connection failed (attempt %d/%d), retrying in %v: %v",
						attempt+1, maxAttempts, delay, err)
					time.Sleep(delay)
					continue
				}
				return fmt.Errorf("failed to connect after %d attempts: %w", maxAttempts, err)
			}
		}

		dbReq := s.batchToDBRequest(batch)

		data, err := json.Marshal(dbReq)
		if err != nil {
			return fmt.Errorf("failed to marshal DB request: %w", err)
		}

		data = append(data, '\n')

		s.mu.Lock()

		_, err = s.conn.Write(data)
		if err != nil {
			s.conn = nil
			s.mu.Unlock()
			lastErr = err

			if attempt < maxAttempts-1 {
				delay := s.calculateBackoff(attempt, initialDelay, maxDelay)
				log.Printf("Send failed (attempt %d/%d), retrying in %v: %v",
					attempt+1, maxAttempts, delay, err)
				time.Sleep(delay)
				continue
			}
			return fmt.Errorf("failed to send after %d attempts: %w", maxAttempts, err)
		}

		decoder := json.NewDecoder(s.conn)
		var resp DBResponse
		if err := decoder.Decode(&resp); err != nil {
			s.mu.Unlock()
			log.Printf("Warning: failed to read response from DB: %v", err)
			return nil
		}

		s.mu.Unlock()

		if resp.Status != "success" {
			log.Printf("Warning: DB returned error: %s", resp.Message)
			return fmt.Errorf("database error: %s", resp.Message)
		}

		log.Printf("Successfully sent batch with %d events to NoSQLdb (inserted: %d)",
			len(batch.Events), resp.Count)
		return nil
	}

	return fmt.Errorf("failed after %d attempts: %w", maxAttempts, lastErr)
}

func (s *TCPSender) calculateBackoff(attempt int, initialDelay, maxDelay time.Duration) time.Duration {
	delay := initialDelay * time.Duration(math.Pow(2, float64(attempt)))
	if delay > maxDelay {
		delay = maxDelay
	}
	return delay
}

func (s *TCPSender) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.conn != nil {
		err := s.conn.Close()
		s.conn = nil
		return err
	}
	return nil
}

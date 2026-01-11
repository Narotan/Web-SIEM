package service

import (
	"sort"
	"sync"
	"time"

	"github.com/Narotan/Web-SIEM/Web/backend/internal/repository"
	"github.com/Narotan/Web-SIEM/Web/backend/internal/service/domain"
)

type Service interface {
	GetEvents(page, limit int) (*domain.EventsPage, error)
	GetStats() (*domain.DashboardStats, error)
	ExportAllEvents() ([]map[string]any, error)
}

type siemService struct {
	repo            repository.Repository
	dbName          string
	statsCache      *domain.DashboardStats
	statsCacheTime  time.Time
	statsCacheMutex sync.RWMutex
	statsCacheTTL   time.Duration
}

func NewSiemService(repo repository.Repository, dbName string) Service {
	return &siemService{
		repo:          repo,
		dbName:        dbName,
		statsCacheTTL: 10 * time.Second,
	}
}

func (s *siemService) GetEvents(page, limit int) (*domain.EventsPage, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}

	data, err := s.repo.FindAll(s.dbName, map[string]any{})
	if err != nil {
		return nil, err
	}

	sort.Slice(data, func(i, j int) bool {
		t1, _ := data[i]["timestamp"].(string)
		t2, _ := data[j]["timestamp"].(string)
		if t1 != t2 {
			return t1 > t2
		}
		id1, _ := data[i]["_id"].(string)
		id2, _ := data[j]["_id"].(string)
		return id1 > id2
	})

	totalCount := len(data)
	startIndex := (page - 1) * limit
	endIndex := startIndex + limit

	if startIndex >= totalCount {
		return &domain.EventsPage{
			Data:       []map[string]any{},
			Count:      0,
			Total:      totalCount,
			Page:       page,
			Limit:      limit,
			TotalPages: (totalCount + limit - 1) / limit,
		}, nil
	}

	if endIndex > totalCount {
		endIndex = totalCount
	}

	paginatedData := data[startIndex:endIndex]

	return &domain.EventsPage{
		Data:       paginatedData,
		Count:      len(paginatedData),
		Total:      totalCount,
		Page:       page,
		Limit:      limit,
		TotalPages: (totalCount + limit - 1) / limit,
	}, nil
}

func (s *siemService) GetStats() (*domain.DashboardStats, error) {
	s.statsCacheMutex.RLock()
	if s.statsCache != nil && time.Since(s.statsCacheTime) < s.statsCacheTTL {
		cached := *s.statsCache
		s.statsCacheMutex.RUnlock()
		return &cached, nil
	}
	s.statsCacheMutex.RUnlock()

	data, err := s.repo.FindAll(s.dbName, map[string]any{})
	if err != nil {
		return nil, err
	}

	stats := domain.DashboardStats{
		ActiveAgents:  make(map[string]time.Time),
		EventsByType:  make(map[string]int),
		SeverityDist:  make(map[string]int),
		TopUsers:      make(map[string]int),
		TopProcesses:  make(map[string]int),
		EventsPerHour: make(map[int]int),
		LastLogins:    []map[string]any{},
	}

	for _, event := range data {
		tsStr, _ := event["timestamp"].(string)
		agent, ok := event["agent_id"].(string)
		if !ok {
			continue
		}

		eType, _ := event["event_type"].(string)
		sev, _ := event["severity"].(string)
		user, _ := event["user"].(string)
		proc, _ := event["process"].(string)

		parsedTime, err := time.Parse(time.RFC3339, tsStr)
		if err != nil {
			continue
		}

		if lastActive, exists := stats.ActiveAgents[agent]; !exists || parsedTime.After(lastActive) {
			stats.ActiveAgents[agent] = parsedTime
		}

		if time.Since(parsedTime) <= 24*time.Hour {
			stats.EventsByType[eType]++
			stats.SeverityDist[sev]++

			if user != "" {
				stats.TopUsers[user]++
			}
			if proc != "" {
				stats.TopProcesses[proc]++
			}

			stats.EventsPerHour[parsedTime.Hour()]++
		}

		if eType == "user_login" || eType == "auth_failure" {
			stats.LastLogins = append(stats.LastLogins, event)
		}
	}

	sort.Slice(stats.LastLogins, func(i, j int) bool {
		t1, _ := time.Parse(time.RFC3339, stats.LastLogins[i]["timestamp"].(string))
		t2, _ := time.Parse(time.RFC3339, stats.LastLogins[j]["timestamp"].(string))
		return t1.After(t2)
	})

	if len(stats.LastLogins) > 10 {
		stats.LastLogins = stats.LastLogins[:10]
	}

	s.statsCacheMutex.Lock()
	s.statsCache = &stats
	s.statsCacheTime = time.Now()
	s.statsCacheMutex.Unlock()

	return &stats, nil
}

func (s *siemService) ExportAllEvents() ([]map[string]any, error) {
	data, err := s.repo.FindAll(s.dbName, map[string]any{})
	if err != nil {
		return nil, err
	}

	sort.Slice(data, func(i, j int) bool {
		t1, _ := data[i]["timestamp"].(string)
		t2, _ := data[j]["timestamp"].(string)
		return t1 > t2
	})

	return data, nil
}

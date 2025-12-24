package api

import (
	"net/http"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/Narotan/Web-SIEM/internal/config"
	"github.com/Narotan/Web-SIEM/internal/db"
	"github.com/gin-gonic/gin"
)

// Cache for stats to avoid recomputing on every request
var (
	statsCache      *db.DashboardStats
	statsCacheTime  time.Time
	statsCacheMutex sync.RWMutex
	statsCacheTTL   = 10 * time.Second // Cache stats for 10 seconds
)

// get /api/health
func HealthHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"message": "SIEM Web API работает",
	})
}

// get /api/events with pagination support
// Query params: page (default 1), limit (default 50, max 200)
func GetEventsHandler(c *gin.Context) {
	cfg := config.GetConfig()

	// Parse pagination params
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))

	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}

	dbReq := db.DBRequest{
		Database: cfg.DBName,
		Command:  "find",
		Query:    map[string]any{},
	}

	resp, err := db.SendQuery(dbReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": "error",
			"error":  "Ошибка связи с базой данных: " + err.Error(),
		})
		return
	}

	if resp.Status == "error" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "error",
			"error":  resp.Message,
		})
		return
	}

	// Sort by timestamp descending
	sort.Slice(resp.Data, func(i, j int) bool {
		t1, _ := resp.Data[i]["timestamp"].(string)
		t2, _ := resp.Data[j]["timestamp"].(string)
		return t1 > t2
	})

	// Apply pagination
	totalCount := len(resp.Data)
	startIndex := (page - 1) * limit
	endIndex := startIndex + limit

	if startIndex >= totalCount {
		c.JSON(http.StatusOK, gin.H{
			"status":     "success",
			"count":      0,
			"total":      totalCount,
			"page":       page,
			"limit":      limit,
			"totalPages": (totalCount + limit - 1) / limit,
			"data":       []map[string]any{},
		})
		return
	}

	if endIndex > totalCount {
		endIndex = totalCount
	}

	paginatedData := resp.Data[startIndex:endIndex]

	c.JSON(http.StatusOK, gin.H{
		"status":     "success",
		"count":      len(paginatedData),
		"total":      totalCount,
		"page":       page,
		"limit":      limit,
		"totalPages": (totalCount + limit - 1) / limit,
		"data":       paginatedData,
	})
}

// get /api/stats with caching
func GetStatsHandler(c *gin.Context) {
	// Check cache first
	statsCacheMutex.RLock()
	if statsCache != nil && time.Since(statsCacheTime) < statsCacheTTL {
		cached := *statsCache
		statsCacheMutex.RUnlock()
		c.JSON(http.StatusOK, cached)
		return
	}
	statsCacheMutex.RUnlock()

	cfg := config.GetConfig()

	dbReq := db.DBRequest{
		Database: cfg.DBName,
		Command:  "find",
		Query:    map[string]any{},
	}

	resp, err := db.SendQuery(dbReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": "error",
			"error":  err.Error(),
		})
		return
	}

	if resp.Status == "error" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "error",
			"error":  resp.Message,
		})
		return
	}

	stats := db.DashboardStats{
		ActiveAgents:  make(map[string]time.Time),
		EventsByType:  make(map[string]int),
		SeverityDist:  make(map[string]int),
		TopUsers:      make(map[string]int),
		TopProcesses:  make(map[string]int),
		EventsPerHour: make(map[int]int),
		LastLogins:    []map[string]any{},
	}

	for _, event := range resp.Data {
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

	// Update cache
	statsCacheMutex.Lock()
	statsCache = &stats
	statsCacheTime = time.Now()
	statsCacheMutex.Unlock()

	c.JSON(http.StatusOK, stats)
}

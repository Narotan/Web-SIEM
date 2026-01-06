package parser

import (
	"regexp"
	"strings"
	"time"

	"github.com/Narotan/SIEM-Agent/internal/domain"
	"github.com/Narotan/SIEM-Agent/internal/reader"
)

type SyslogParser struct{}

func NewSyslogParser() *SyslogParser {
	return &SyslogParser{}
}

func (p *SyslogParser) Match(event reader.RawEvent) bool {
	return strings.Contains(event.Source, "syslog") ||
		strings.Contains(event.Source, "auth.log") ||
		strings.Contains(event.Source, "messages")
}

func (p *SyslogParser) Parse(event reader.RawEvent) (domain.Event, error) {
	result := domain.Event{
		Timestamp: event.Timestamp,
		Source:    "syslog",
		RawLog:    event.Data,
		Severity:  "info",
	}

	if strings.Contains(event.Source, "auth.log") {
		result.Source = "auth"
		result.Severity = "medium"
	}

	syslogPattern := regexp.MustCompile(`^(\w+\s+\d+\s+\d+:\d+:\d+)\s+(\S+)\s+(\S+?)(?:\[(\d+)\])?:\s*(.*)$`)

	if matches := syslogPattern.FindStringSubmatch(event.Data); len(matches) > 5 {
		timestamp, err := time.Parse("Jan 2 15:04:05", matches[1])
		if err == nil {
			now := time.Now()
			timestamp = time.Date(now.Year(), timestamp.Month(), timestamp.Day(),
				timestamp.Hour(), timestamp.Minute(), timestamp.Second(), 0, time.Local)
			result.Timestamp = timestamp
		}

		result.Process = matches[3]
		result.Command = matches[5]
	} else {
		result.Command = event.Data
	}

	if userMatch := regexp.MustCompile(`user[=\s]+(\w+)`).FindStringSubmatch(strings.ToLower(result.Command)); len(userMatch) > 1 {
		result.User = userMatch[1]
	}
	if forMatch := regexp.MustCompile(`for\s+(\w+)`).FindStringSubmatch(result.Command); len(forMatch) > 1 && result.User == "" {
		result.User = forMatch[1]
	}

	lowerCmd := strings.ToLower(result.Command)
	switch {
	case strings.Contains(lowerCmd, "failed") || strings.Contains(lowerCmd, "error"):
		result.EventType = "error"
		result.Severity = "high"
	case strings.Contains(lowerCmd, "authentication failure") || strings.Contains(lowerCmd, "invalid user"):
		result.EventType = "auth_failure"
		result.Severity = "high"
	case strings.Contains(lowerCmd, "accepted") || strings.Contains(lowerCmd, "session opened"):
		result.EventType = "user_login"
		result.Severity = "medium"
	case strings.Contains(lowerCmd, "sudo"):
		result.EventType = "sudo_command"
		result.Severity = "medium"
	case strings.Contains(lowerCmd, "session closed"):
		result.EventType = "user_logout"
		result.Severity = "low"
	default:
		result.EventType = "system_event"
		result.Severity = "info"
	}

	return result, nil
}

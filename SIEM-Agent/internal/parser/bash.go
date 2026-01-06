package parser

import (
	"path/filepath"
	"regexp"
	"strings"

	"github.com/Narotan/SIEM-Agent/internal/domain"
	"github.com/Narotan/SIEM-Agent/internal/reader"
)

type BashParser struct{}

func NewBashParser() *BashParser {
	return &BashParser{}
}

func (p *BashParser) Match(event reader.RawEvent) bool {
	return strings.Contains(event.Source, ".bash_history") ||
		strings.Contains(event.Source, "bash_history")
}

func (p *BashParser) Parse(event reader.RawEvent) (domain.Event, error) {
	result := domain.Event{
		Timestamp: event.Timestamp,
		Source:    "bash_history",
		RawLog:    event.Data,
		EventType: "command_executed",
		Severity:  "low",
	}

	command := event.Data

	if strings.HasPrefix(command, "#") && len(command) > 1 {
		
		return result, nil
	}

	result.Command = command

	if pathParts := strings.Split(event.Source, "/"); len(pathParts) > 2 {
		for i, part := range pathParts {
			if part == "home" && i+1 < len(pathParts) {
				result.User = pathParts[i+1]
				break
			}
		}
	}

	if result.User == "" {
		dir := filepath.Dir(event.Source)
		result.User = filepath.Base(dir)
	}

	if parts := strings.Fields(command); len(parts) > 0 {
		result.Process = parts[0]
	}

	lowerCmd := strings.ToLower(command)
	switch {
	case strings.Contains(lowerCmd, "sudo") || strings.Contains(lowerCmd, "su "):
		result.Severity = "high"
		result.EventType = "privileged_command"
	case containsAny(lowerCmd, []string{"rm -rf", "chmod", "chown", "iptables", "systemctl"}):
		result.Severity = "medium"
		result.EventType = "system_command"
	case containsAny(lowerCmd, []string{"ssh", "scp", "wget", "curl", "nc ", "netcat"}):
		result.Severity = "medium"
		result.EventType = "network_command"
	case containsAny(lowerCmd, []string{"passwd", "useradd", "usermod", "userdel"}):
		result.Severity = "high"
		result.EventType = "user_management"
	default:
		result.Severity = "low"
		result.EventType = "command_executed"
	}

	return result, nil
}

func containsAny(s string, substrings []string) bool {
	for _, substr := range substrings {
		if strings.Contains(s, substr) {
			return true
		}
	}
	return false
}

func extractTimestamp(line string) (int64, bool) {
	if !strings.HasPrefix(line, "#") {
		return 0, false
	}

	tsPattern := regexp.MustCompile(`^#(\d+)$`)
	if matches := tsPattern.FindStringSubmatch(line); len(matches) > 1 {
		return parseInt64(matches[1]), true
	}

	return 0, false
}

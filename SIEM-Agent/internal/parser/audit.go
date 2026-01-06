package parser

import (
	"regexp"
	"strings"
	"time"

	"github.com/Narotan/SIEM-Agent/internal/domain"
	"github.com/Narotan/SIEM-Agent/internal/reader"
)

type AuditParser struct{}

func NewAuditParser() *AuditParser {
	return &AuditParser{}
}

func (p *AuditParser) Match(event reader.RawEvent) bool {
	return strings.Contains(event.Source, "audit.log") ||
		strings.HasPrefix(event.Data, "type=")
}

func (p *AuditParser) Parse(event reader.RawEvent) (domain.Event, error) {
	result := domain.Event{
		Timestamp: event.Timestamp,
		Source:    "auditd",
		RawLog:    event.Data,
		Severity:  "medium",
	}

	if typeMatch := regexp.MustCompile(`type=(\w+)`).FindStringSubmatch(event.Data); len(typeMatch) > 1 {
		result.EventType = strings.ToLower(typeMatch[1])
	} else {
		result.EventType = "audit_event"
	}

	if msgMatch := regexp.MustCompile(`msg=audit\((\d+)\.(\d+):\d+\)`).FindStringSubmatch(event.Data); len(msgMatch) > 2 {
		sec := parseInt64(msgMatch[1])
		nsec := parseInt64(msgMatch[2]) * 1000000
		result.Timestamp = time.Unix(sec, nsec)
	}

	if uidMatch := regexp.MustCompile(`uid=(\d+)`).FindStringSubmatch(event.Data); len(uidMatch) > 1 {
		result.User = uidMatch[1]
	}
	if nameMatch := regexp.MustCompile(`name="([^"]+)"`).FindStringSubmatch(event.Data); len(nameMatch) > 1 {
		if result.User == "" {
			result.User = nameMatch[1]
		}
	}
	if exeMatch := regexp.MustCompile(`exe="([^"]+)"`).FindStringSubmatch(event.Data); len(exeMatch) > 1 {
		result.Process = exeMatch[1]
	}
	if commMatch := regexp.MustCompile(`comm="([^"]+)"`).FindStringSubmatch(event.Data); len(commMatch) > 1 {
		result.Command = commMatch[1]
	}

	switch result.EventType {
	case "user_auth", "user_login", "user_acct", "cred_acq":
		result.Severity = "high"
	case "syscall", "execve":
		result.Severity = "medium"
	default:
		result.Severity = "low"
	}

	return result, nil
}

func parseInt64(s string) int64 {
	var n int64
	for _, c := range s {
		if c >= '0' && c <= '9' {
			n = n*10 + int64(c-'0')
		}
	}
	return n
}

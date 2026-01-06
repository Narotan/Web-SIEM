package filter

import (
	"regexp"
	"strings"

	"github.com/Narotan/SIEM-Agent/internal/domain"
)

type Filter interface {
	Match(event domain.Event) bool
}

type Config struct {
	ExcludePatterns   []string `yaml:"exclude_patterns"`
	IncludePatterns   []string `yaml:"include_patterns"`
	SeverityThreshold string   `yaml:"severity_threshold"`
	ExcludeSources    []string `yaml:"exclude_sources"`
}

// EventFilter комбинированный фильтр
type EventFilter struct {
	excludeRegexps []*regexp.Regexp
	includeRegexps []*regexp.Regexp
	minSeverity    int
	excludeSources map[string]bool
}

func NewFilter(cfg Config) (*EventFilter, error) {
	f := &EventFilter{
		excludeSources: make(map[string]bool),
	}

	for _, pattern := range cfg.ExcludePatterns {
		re, err := regexp.Compile(pattern)
		if err != nil {
			return nil, err
		}
		f.excludeRegexps = append(f.excludeRegexps, re)
	}

	for _, pattern := range cfg.IncludePatterns {
		re, err := regexp.Compile(pattern)
		if err != nil {
			return nil, err
		}
		f.includeRegexps = append(f.includeRegexps, re)
	}

	f.minSeverity = severityToInt(cfg.SeverityThreshold)

	for _, source := range cfg.ExcludeSources {
		f.excludeSources[source] = true
	}

	return f, nil
}

// Match проверяет событие
func (f *EventFilter) Match(event domain.Event) bool {
	if f.excludeSources[event.Source] {
		return false
	}

	if severityToInt(event.Severity) < f.minSeverity {
		return false
	}

	for _, re := range f.excludeRegexps {
		if re.MatchString(event.RawLog) ||
			re.MatchString(event.Command) ||
			re.MatchString(event.EventType) {
			return false
		}
	}

	if len(f.includeRegexps) > 0 {
		matched := false
		for _, re := range f.includeRegexps {
			if re.MatchString(event.RawLog) ||
				re.MatchString(event.Command) ||
				re.MatchString(event.EventType) {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}

	return true
}

func severityToInt(severity string) int {
	switch strings.ToLower(severity) {
	case "low", "info":
		return 1
	case "medium", "warning":
		return 2
	case "high", "error", "critical":
		return 3
	default:
		return 0
	}
}

// NoOpFilter пропускает все события
type NoOpFilter struct{}

func (f *NoOpFilter) Match(event domain.Event) bool {
	return true
}

func NewNoOpFilter() *NoOpFilter {
	return &NoOpFilter{}
}

package config

import (
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server       ServerConfig   `yaml:"server"`
	Logging      LoggingConfig  `yaml:"logging"`
	AgentLogging AgentLogConfig `yaml:"agent_logging"`
	Buffer       BufferConfig   `yaml:"buffer"`
	Filters      FilterConfig   `yaml:"filters"`
	Retry        RetryConfig    `yaml:"retry"`
}

type ServerConfig struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

type LoggingConfig struct {
	AgentID      string        `yaml:"agent_id"`
	Sources      []string      `yaml:"sources"`
	SendInterval time.Duration `yaml:"send_interval"`
	BatchSize    int           `yaml:"batch_size"`
}

type AgentLogConfig struct {
	File       string `yaml:"file"`
	Level      string `yaml:"level"`
	MaxSize    int64  `yaml:"max_size"`
	MaxBackups int    `yaml:"max_backups"`
}

type BufferConfig struct {
	Directory string `yaml:"directory"`
	MaxSize   int64  `yaml:"max_size"`
}

type FilterConfig struct {
	ExcludePatterns   []string `yaml:"exclude_patterns"`
	IncludePatterns   []string `yaml:"include_patterns"`
	SeverityThreshold string   `yaml:"severity_threshold"`
	ExcludeSources    []string `yaml:"exclude_sources"`
}

type RetryConfig struct {
	MaxAttempts  int           `yaml:"max_attempts"`
	InitialDelay time.Duration `yaml:"initial_delay"`
	MaxDelay     time.Duration `yaml:"max_delay"`
}

func Load(configPath string) (*Config, error) {
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, err
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var cfg Config
	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}

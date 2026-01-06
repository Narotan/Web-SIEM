package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/Narotan/SIEM-Agent/internal/config"
	"github.com/Narotan/SIEM-Agent/internal/domain"
	"github.com/Narotan/SIEM-Agent/internal/filter"
	"github.com/Narotan/SIEM-Agent/internal/logger"
	"github.com/Narotan/SIEM-Agent/internal/parser"
	"github.com/Narotan/SIEM-Agent/internal/reader"
	"github.com/Narotan/SIEM-Agent/internal/sender"
	"github.com/Narotan/SIEM-Agent/internal/storage"
)

func main() {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "configs/config.yaml"
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	logLevel := parseLogLevel(cfg.AgentLogging.Level)
	err = logger.Init(logger.Config{
		FilePath:   cfg.AgentLogging.File,
		Level:      logLevel,
		MaxSize:    cfg.AgentLogging.MaxSize * 1024 * 1024,
		MaxBackups: cfg.AgentLogging.MaxBackups,
	})
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Close()

	logger.Info("SIEM Agent starting...")
	logger.Info("Agent ID: %s", cfg.Logging.AgentID)
	logger.Info("Target server: %s:%d", cfg.Server.Host, cfg.Server.Port)

	diskBuffer, err := storage.NewDiskBuffer(
		cfg.Buffer.Directory,
		cfg.Buffer.MaxSize*1024*1024,
	)
	if err != nil {
		logger.Error("Failed to create disk buffer: %v", err)
		log.Fatalf("Failed to create disk buffer: %v", err)
	}
	logger.Info("Disk buffer initialized: %s (max: %dMB)", cfg.Buffer.Directory, cfg.Buffer.MaxSize)

	eventFilter, err := filter.NewFilter(filter.Config{
		ExcludePatterns:   cfg.Filters.ExcludePatterns,
		IncludePatterns:   cfg.Filters.IncludePatterns,
		SeverityThreshold: cfg.Filters.SeverityThreshold,
		ExcludeSources:    cfg.Filters.ExcludeSources,
	})
	if err != nil {
		logger.Error("Failed to create filter: %v", err)
		log.Fatalf("Failed to create filter: %v", err)
	}
	logger.Info("Event filter initialized (threshold: %s)", cfg.Filters.SeverityThreshold)

	parsedEvents := make(chan domain.Event, 100)

	tcpSender := sender.NewTCPSender(cfg.Server.Host, cfg.Server.Port)
	tcpSender.SetCollection("security_events")
	defer tcpSender.Close()

	pipeline := sender.NewPipeline(tcpSender, sender.Config{
		AgentID:      cfg.Logging.AgentID,
		BatchSize:    cfg.Logging.BatchSize,
		FlushTimeout: cfg.Logging.SendInterval,
		DiskBuffer:   diskBuffer,
		Filter:       eventFilter,
		RetryConfig: sender.RetryConfig{
			MaxAttempts:  cfg.Retry.MaxAttempts,
			InitialDelay: cfg.Retry.InitialDelay,
			MaxDelay:     cfg.Retry.MaxDelay,
		},
	})

	pipeline.Start(parsedEvents)
	defer pipeline.Stop()

	logger.Info("Pipeline started (batch_size: %d, interval: %s)",
		cfg.Logging.BatchSize, cfg.Logging.SendInterval)

	auditParser := parser.NewAuditParser()
	syslogParser := parser.NewSyslogParser()
	bashParser := parser.NewBashParser()

	router := parser.NewRouter(auditParser, syslogParser, bashParser)

	readers := make([]*reader.Reader, 0)

	predefinedSources := map[string]string{
		"auditd":       "/var/log/audit/audit.log",
		"syslog":       "/var/log/syslog",
		"bash_history": os.Getenv("HOME") + "/.bash_history",
	}

	for _, source := range cfg.Logging.Sources {
		path, ok := predefinedSources[source]
		if !ok {
			path = source
		}

		r, err := reader.New(path)
		if err != nil {
			logger.Warn("Failed to create reader for %s: %v", path, err)
			log.Printf("Warning: failed to create reader for %s: %v", path, err)
			continue
		}

		readers = append(readers, r)

		if err := r.Start(); err != nil {
			logger.Warn("Failed to start reader for %s: %v", path, err)
			log.Printf("Warning: failed to start reader for %s: %v", path, err)
			continue
		}

		go router.Start(r.Events(), r.Errors(), parsedEvents)

		logger.Info("Monitoring: %s", path)
		log.Printf("Started monitoring: %s", path)
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	log.Printf("SIEM Agent started with agent_id=%s", cfg.Logging.AgentID)
	log.Printf("Sending batches to %s:%d every %d seconds or %d events",
		cfg.Server.Host, cfg.Server.Port, cfg.Logging.SendInterval, cfg.Logging.BatchSize)

	<-sigChan
	log.Println("Shutting down...")

	for _, r := range readers {
		r.Stop()
	}

	router.Stop()
	close(parsedEvents)

	log.Println("SIEM Agent stopped")
}

func parseLogLevel(level string) logger.Level {
	switch level {
	case "debug":
		return logger.DEBUG
	case "info":
		return logger.INFO
	case "warn":
		return logger.WARN
	case "error":
		return logger.ERROR
	default:
		return logger.INFO
	}
}

package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type Level int

const (
	DEBUG Level = iota
	INFO
	WARN
	ERROR
)

func (l Level) String() string {
	switch l {
	case DEBUG:
		return "DEBUG"
	case INFO:
		return "INFO"
	case WARN:
		return "WARN"
	case ERROR:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// Logger пишет логи коротко
type Logger struct {
	level      Level
	file       *os.File
	logger     *log.Logger
	mu         sync.Mutex
	maxSize    int64
	maxBackups int
}

// Config параметры логгера кратко
type Config struct {
	FilePath   string
	Level      Level
	MaxSize    int64
	MaxBackups int
}

var (
	globalLogger *Logger
	once         sync.Once
)

// Init глобальный логгер
func Init(cfg Config) error {
	var err error
	once.Do(func() {
		globalLogger, err = New(cfg)
	})
	return err
}

// New создаёт логгер
func New(cfg Config) (*Logger, error) {
	dir := filepath.Dir(cfg.FilePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}

	file, err := os.OpenFile(cfg.FilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}

	multiWriter := io.MultiWriter(file, os.Stdout)

	l := &Logger{
		level:      cfg.Level,
		file:       file,
		logger:     log.New(multiWriter, "", 0),
		maxSize:    cfg.MaxSize,
		maxBackups: cfg.MaxBackups,
	}

	return l, nil
}

// Close закрывает файл
func (l *Logger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.file != nil {
		return l.file.Close()
	}
	return nil
}

// log внутренняя запись
func (l *Logger) log(level Level, format string, args ...interface{}) {
	if level < l.level {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	if l.maxSize > 0 {
		if err := l.checkRotation(); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to rotate log: %v\n", err)
		}
	}

	timestamp := time.Now().Format("2006/01/02 15:04:05")
	message := fmt.Sprintf(format, args...)
	l.logger.Printf("[%s] %s: %s", timestamp, level.String(), message)
}

// Debug пишет отладку
func (l *Logger) Debug(format string, args ...interface{}) {
	l.log(DEBUG, format, args...)
}

// Info пишет инфо
func (l *Logger) Info(format string, args ...interface{}) {
	l.log(INFO, format, args...)
}

// Warn пишет предупреждение
func (l *Logger) Warn(format string, args ...interface{}) {
	l.log(WARN, format, args...)
}

// Error пишет ошибку
func (l *Logger) Error(format string, args ...interface{}) {
	l.log(ERROR, format, args...)
}

// checkRotation смотрит размер
func (l *Logger) checkRotation() error {
	info, err := l.file.Stat()
	if err != nil {
		return err
	}

	if info.Size() < l.maxSize {
		return nil
	}

	return l.rotate()
}

// rotate обновляет файлы
func (l *Logger) rotate() error {
	if err := l.file.Close(); err != nil {
		return err
	}

	for i := l.maxBackups - 1; i > 0; i-- {
		oldPath := fmt.Sprintf("%s.%d", l.file.Name(), i)
		newPath := fmt.Sprintf("%s.%d", l.file.Name(), i+1)

		if _, err := os.Stat(oldPath); err == nil {
			os.Rename(oldPath, newPath)
		}
	}

	if err := os.Rename(l.file.Name(), l.file.Name()+".1"); err != nil {
		return err
	}

	file, err := os.OpenFile(l.file.Name(), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	l.file = file
	multiWriter := io.MultiWriter(file, os.Stdout)
	l.logger.SetOutput(multiWriter)

	return nil
}

func Debug(format string, args ...interface{}) {
	if globalLogger != nil {
		globalLogger.Debug(format, args...)
	}
}

func Info(format string, args ...interface{}) {
	if globalLogger != nil {
		globalLogger.Info(format, args...)
	}
}

func Warn(format string, args ...interface{}) {
	if globalLogger != nil {
		globalLogger.Warn(format, args...)
	}
}

func Error(format string, args ...interface{}) {
	if globalLogger != nil {
		globalLogger.Error(format, args...)
	}
}

func Close() error {
	if globalLogger != nil {
		return globalLogger.Close()
	}
	return nil
}

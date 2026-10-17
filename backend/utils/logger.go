package utils

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"
	"sync"
	"time"
)

type Level int

const (
	// Log levels
	DEBUG Level = iota
	INFO
	WARN
	ERROR
	FATAL
)

// String representation of the log level
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
	case FATAL:
		return "FATAL"
	default:
		return "UNKNOWN"
	}
}

type Logger struct {
	mu       sync.Mutex
	file     *os.File
	output   io.Writer
	level    Level
	closed   bool
}

func NewLogger(logDir string, level Level) (*Logger, error) {
	// Check if the log directory exists and if it does not create it
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, fmt.Errorf("create log directory: %w", err)
	}

	latestPath := filepath.Join(logDir, "latest.log")

	// Save the previous log file
	if _, err := os.Stat(latestPath); err == nil {
		archiveName := "latest-" + time.Now().Format("20060102-150405.000000000") + ".log"

		archivePath := filepath.Join(logDir, archiveName)

		if err := os.Rename(latestPath, archivePath); err != nil {
			return nil, fmt.Errorf("archive previous log: %w", err)
		}
	} else if !os.IsNotExist(err) {
		return nil, fmt.Errorf("inspect latest log: %w", err)
	}

	file, err := os.OpenFile(
		latestPath,
		os.O_CREATE|os.O_WRONLY|os.O_APPEND,
		0644,
	)
	if err != nil {
		return nil, fmt.Errorf("open latest log: %w", err)
	}

	return &Logger{
		file:  file,
		output: io.MultiWriter(os.Stdout, file),
		level: level,
	}, nil
}

func (l *Logger) log(level Level, format string, args ...any) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.closed || level < l.level {
		return
	}

	message := fmt.Sprintf(format, args...)
	message = strings.TrimRight(message, "\r\n")

	timestamp := time.Now().Format("2006-01-02 15:04:05.000")
	entry := fmt.Sprintf(
		"[%s] [%s] %s\n",
		timestamp,
		level.String(),
		message,
	)

	if _, err := io.WriteString(l.output, entry); err != nil {
		//
		fmt.Fprintf(os.Stderr, "logger: writing failed: %v\n", err)
	}
}

func (l *Logger) Debug(format string, args ...any) {
	l.log(DEBUG, format, args...)
}

func (l *Logger) Info(format string, args ...any) {
	l.log(INFO, format, args...)
}

func (l *Logger) Warn(format string, args ...any) {
	l.log(WARN, format, args...)
}

func (l *Logger) Error(format string, args ...any) {
	l.log(ERROR, format, args...)
}

//
//
func (l *Logger) Fatal(format string, args ...any) {
	l.log(FATAL, format, args...)
}

// Panic
func (l *Logger) Panic(format string, args ...any) {
	message := fmt.Sprintf(format, args...)
	l.log(ERROR, "PANIC: %s", message)
	l.log(ERROR, "Stack trace:\n%s", debug.Stack())
	panic(message)
}

func (l *Logger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.closed {
		return nil
	}

	l.closed = true
	return l.file.Close()
}
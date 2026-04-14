package logger

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type LogLevel string

const (
	LogLevelInfo  LogLevel = "info"
	LogLevelWarn  LogLevel = "warn"
	LogLevelError LogLevel = "error"
)

type LogEntry struct {
	ServiceName string    `json:"serviceName"`
	Level       LogLevel  `json:"level"`
	Message     string    `json:"message"`
	Timestamp   time.Time `json:"timestamp"`
}

type Logger struct {
	logDir  string
	mu      sync.RWMutex
	writers map[string]*os.File
}

func New(logDir string) (*Logger, error) {
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, err
	}

	return &Logger{
		logDir:  logDir,
		writers: make(map[string]*os.File),
	}, nil
}

func (l *Logger) GetLogFile(serviceName string) string {
	date := time.Now().Format("2006-01-02")
	return filepath.Join(l.logDir, fmt.Sprintf("%s_%s.log", serviceName, date))
}

func (l *Logger) Write(serviceName, message string, level LogLevel) {
	entry := LogEntry{
		ServiceName: serviceName,
		Level:       level,
		Message:     message,
		Timestamp:   time.Now(),
	}

	data, _ := json.Marshal(entry)
	line := string(data) + "\n"

	// 写入文件
	logFile := l.GetLogFile(serviceName)
	f, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err == nil {
		f.WriteString(line)
		f.Close()
	}
}

func (l *Logger) Read(serviceName string, lines int) ([]string, error) {
	logFile := l.GetLogFile(serviceName)

	data, err := os.ReadFile(logFile)
	if err != nil {
		return nil, err
	}

	allLines := strings.Split(string(data), "\n")
	if lines > 0 && len(allLines) > lines {
		allLines = allLines[len(allLines)-lines:]
	}

	var result []string
	for _, line := range allLines {
		if line != "" {
			result = append(result, line)
		}
	}

	return result, nil
}

func (l *Logger) ReadSince(serviceName string, since time.Time) ([]string, error) {
	logFile := l.GetLogFile(serviceName)

	data, err := os.ReadFile(logFile)
	if err != nil {
		return nil, err
	}

	var result []string
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}

		var entry LogEntry
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			continue
		}

		if entry.Timestamp.After(since) {
			result = append(result, line)
		}
	}

	return result, nil
}
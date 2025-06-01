package internal

import (
	"fmt"
	"sync"
)

// AppLogger handles application logging with thread-safe operations
type AppLogger struct {
	logMutex sync.RWMutex
	logLines []string
	maxLines int
}

// NewAppLogger creates a new application logger
func NewAppLogger() *AppLogger {
	return &AppLogger{
		logLines: make([]string, 0, 1000),
		maxLines: 1000,
	}
}

// Log adds a message to the internal log buffer
func (l *AppLogger) Log(message string) {
	l.logMutex.Lock()
	defer l.logMutex.Unlock()

	// Keep only last maxLines
	if len(l.logLines) >= l.maxLines {
		l.logLines = l.logLines[1:]
	}

	l.logLines = append(l.logLines, message)
	fmt.Println(message) // Also print to console
}

// GetLogs returns a copy of the current log lines
func (l *AppLogger) GetLogs() []string {
	l.logMutex.RLock()
	defer l.logMutex.RUnlock()

	logs := make([]string, len(l.logLines))
	copy(logs, l.logLines)
	return logs
}

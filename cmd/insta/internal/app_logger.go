package internal

import (
	"fmt"
	"sync"
	"time"
)

// LogEntry represents a single log entry with timestamp
type LogEntry struct {
	Message   string `json:"message"`
	Timestamp string `json:"timestamp"`
}

// AppLogger handles application logging with thread-safe operations
type AppLogger struct {
	logMutex   sync.RWMutex
	logLines   []string
	logEntries []LogEntry
	maxLines   int
}

// NewAppLogger creates a new application logger
func NewAppLogger() *AppLogger {
	return &AppLogger{
		logLines:   make([]string, 0, 1000),
		logEntries: make([]LogEntry, 0, 1000),
		maxLines:   1000,
	}
}

// Log adds a message to the internal log buffer
func (l *AppLogger) Log(message string) {
	l.logMutex.Lock()
	defer l.logMutex.Unlock()

	now := time.Now()
	entry := LogEntry{
		Message:   message,
		Timestamp: now.Format(time.RFC3339),
	}

	// Keep only last maxLines for both logLines and logEntries
	if len(l.logLines) >= l.maxLines {
		l.logLines = l.logLines[1:]
		l.logEntries = l.logEntries[1:]
	}

	l.logLines = append(l.logLines, message)
	l.logEntries = append(l.logEntries, entry)
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

// GetLogEntries returns a copy of the current log entries with timestamps
func (l *AppLogger) GetLogEntries() []LogEntry {
	l.logMutex.RLock()
	defer l.logMutex.RUnlock()

	entries := make([]LogEntry, len(l.logEntries))
	copy(entries, l.logEntries)
	return entries
}

// GetLogsSince returns log entries since the given timestamp
func (l *AppLogger) GetLogsSince(since time.Time) []LogEntry {
	l.logMutex.RLock()
	defer l.logMutex.RUnlock()

	var newEntries []LogEntry
	for _, entry := range l.logEntries {
		entryTime, err := time.Parse(time.RFC3339, entry.Timestamp)
		if err == nil && entryTime.After(since) {
			newEntries = append(newEntries, entry)
		}
	}
	return newEntries
}

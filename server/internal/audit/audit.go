package audit

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type Event struct {
	Time    time.Time      `json:"time"`
	Actor   string         `json:"actor"`
	Action  string         `json:"action"`
	Details map[string]any `json:"details,omitempty"`
}

type Logger struct {
	mu      sync.Mutex
	file    *os.File
	records []Event
	limit   int
}

func NewLogger(path string) (*Logger, error) {
	if path == "" {
		path = "./audit.log"
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, err
	}
	file, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return nil, err
	}

	return &Logger{
		file:    file,
		records: make([]Event, 0, 512),
		limit:   1000,
	}, nil
}

func (l *Logger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.file == nil {
		return nil
	}
	return l.file.Close()
}

func (l *Logger) Log(actor, action string, details map[string]any) {
	l.mu.Lock()
	defer l.mu.Unlock()

	event := Event{
		Time:    time.Now(),
		Actor:   actor,
		Action:  action,
		Details: details,
	}

	b, _ := json.Marshal(event)
	if l.file != nil {
		_, _ = l.file.Write(append(b, '\n'))
	}

	l.records = append(l.records, event)
	if len(l.records) > l.limit {
		l.records = l.records[len(l.records)-l.limit:]
	}
}

func (l *Logger) List(limit int) []Event {
	l.mu.Lock()
	defer l.mu.Unlock()

	if limit <= 0 || limit > len(l.records) {
		limit = len(l.records)
	}

	start := len(l.records) - limit
	out := make([]Event, 0, limit)
	out = append(out, l.records[start:]...)
	return out
}

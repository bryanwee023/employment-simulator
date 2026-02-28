package main

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"
)

const logBufSize = 4096

type LogEntry struct {
	Timestamp time.Time       `json:"timestamp"`
	Event     string          `json:"event"`
	Data      json.RawMessage `json:"data"`
}

func (e LogEntry) String() string {
	return fmt.Sprintf("[%s] %s: %s", e.Timestamp.Format(time.DateTime), e.Event, e.Data)
}

type Logger struct {
	mu   sync.RWMutex
	buf  [logBufSize]LogEntry
	head int
	size int
}

func NewLogger() *Logger {
	return &Logger{}
}

func (l *Logger) Log(event string, data any) {
	l.mu.Lock()
	defer l.mu.Unlock()

	raw, _ := json.Marshal(data)

	l.buf[l.head] = LogEntry{
		Timestamp: time.Now(),
		Event:     event,
		Data:      raw,
	}

	l.head = (l.head + 1) % logBufSize
	if l.size < logBufSize {
		l.size++
	}

	log.Printf("%s: %s\n", event, raw)
}

func (l *Logger) Entries() []LogEntry {
	l.mu.RLock()
	defer l.mu.RUnlock()

	out := make([]LogEntry, l.size)
	start := (l.head - l.size + logBufSize) % logBufSize
	for i := 0; i < l.size; i++ {
		out[i] = l.buf[(start+i)%logBufSize]
	}

	return out
}

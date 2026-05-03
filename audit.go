package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
)

type AuditEntry struct {
	ID string `json:"id"`
	Timestamp string `json:"timestamp"`
	AgentName string `json:"agent_name"`
	Method string `json:"method"`
	Path string `json:"path"`
	Status int `json:"status"`
	Duration string `json:"duration"`
	IPAddress string `json:"ip_address"`
}

type AuditLogger struct {
	mu sync.Mutex
	file *os.File
}

func newAuditLogger(logPath string) (*AuditLogger, error) {
	file, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("could not open audit log: %w", err)
	}
	return &AuditLogger{file: file}, nil
}

func (a *AuditLogger) Log(entry AuditEntry) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	data, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("could not marshal audit entry: %w", err)
	}
	_, err = fmt.Fprintf(a.file, "%s\n", data)
	return err
}

func (a *AuditLogger) Close() {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.file.Close()
}
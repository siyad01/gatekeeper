package main

import (
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

type ServerStatus string

const (
	StatusHealthy ServerStatus = "healthy"
	StatusUnhealthy ServerStatus = "unhealthy"
	StatusUnknown ServerStatus = "unknown"
)

type ServerHealth struct {
	Name        string       `json:"name"`
	URL         string       `json:"url"`
	Status      ServerStatus `json:"status"`
	LastChecked time.Time    `json:"last_checked"`
	LastError   string       `json:"last_error,omitempty"`
	Latency     string       `json:"latency,omitempty"`
}

type HealthChecker struct {
	mu       sync.RWMutex
	statuses map[string]*ServerHealth
	config   *Config
	client   *http.Client
}

func newHealthChecker(config *Config) *HealthChecker {
	hc := &HealthChecker{
		statuses: make(map[string]*ServerHealth),
		config:   config,
		client:   &http.Client{Timeout: 5 * time.Second},
	}

	for _, server := range config.Servers {
		hc.statuses[server.Name] = &ServerHealth{
			Name:   server.Name,
			URL:    server.URL,
			Status: StatusUnknown,
		}
	}
	return hc
}

func (hc *HealthChecker) start() {
	hc.checkAll()
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		for range ticker.C {
			hc.checkAll()
		}
	}()
}

func (hc *HealthChecker) checkAll() {
	for _, server := range hc.config.Servers {
		go hc.checkServer(server)
	}
}

func (hc *HealthChecker) checkServer(server MCPServer) {
	start := time.Now()
	healthURL := server.URL + "/health"

	resp, err := hc.client.Get(healthURL)
	latency := time.Since(start)

	hc.mu.Lock()
	defer hc.mu.Unlock()

	status := hc.statuses[server.Name]
	status.LastChecked = time.Now()
	status.Latency = latency.String()

	if err != nil {
		status.Status = StatusUnhealthy
		status.LastError = err.Error()
		log.Printf("HEALTH | ❌ %s → unhealthy | %v", server.Name, err)
		return
	}

	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		status.Status = StatusHealthy
		status.LastError = ""
		log.Printf("HEALTH | ✅ %s → healthy | latency=%s", server.Name, latency)
	} else {
		status.Status = StatusUnhealthy
		status.LastError = fmt.Sprintf("unexpected status: %d", resp.StatusCode)
		log.Printf("HEALTH | ❌ %s → unhealthy | status=%d", server.Name, resp.StatusCode)
	}
}

func (hc *HealthChecker) isHealthy(serverName string) bool {
	hc.mu.RLock()
	defer hc.mu.RUnlock()

	status, exists := hc.statuses[serverName]
	if !exists {
		return false
	}
	return status.Status == StatusHealthy
}

func (hc *HealthChecker) getStatuses() map[string]*ServerHealth {
	hc.mu.RLock()
	defer hc.mu.RUnlock()

	copy := make(map[string]*ServerHealth)
	for k,v := range hc.statuses {
		statusCopy := *v
		copy[k] = &statusCopy
	}
	return copy
}
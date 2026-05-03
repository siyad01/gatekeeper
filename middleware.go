package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
)

func requestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		log.Printf("-> %s %s", r.Method, r.URL.Path)

		next.ServeHTTP(w, r)

		duration := time.Since(start)
		log.Printf("<- %s %s completed in %v", r.Method, r.URL.Path, duration)

	})

}

func auditMiddleware(audit *AuditLogger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			wrapped := newResponseWriter(w)

			next.ServeHTTP(wrapped, r)

			entry := AuditEntry{
				ID: newID(),
				Timestamp: start.UTC().Format(time.RFC3339),
				AgentName: r.Header.Get("X-Agent-Name"),
				Method: r.Method,
				Path: r.URL.Path,
				Status: wrapped.status,
				Duration: time.Since(start).String(),
				IPAddress: r.RemoteAddr,
			}

			if err := audit.Log(entry); err != nil {
				log.Printf("audit log error: %v", err)
			}

			fmt.Printf("AUDIT | %s | agent=%-20s | %s %s | status=%d | %s\n",  entry.Timestamp, entry.AgentName, entry.Method, entry.Path, entry.Status, entry.Duration)

		})
	}
}

func newID() string {
	return uuid.New().String()
}
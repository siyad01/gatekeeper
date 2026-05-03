package main

import (
	"net/http"
	"sync"
	"time"
	"fmt"
)

type requestRecord struct {
	timestamps []time.Time
	mu 	   sync.Mutex
}

type RateLimiter struct {
	mu sync.Mutex
	agents map[string]*requestRecord
	limit int
	window time.Duration
}

func newRateLimiter(limit int, window time.Duration) *RateLimiter {
	r1 := &RateLimiter{
		agents: make(map[string]*requestRecord),
		limit: limit,
		window: window,
	}
	go r1.cleanup()
	return r1
}

func (r1 *RateLimiter) isAllowed(agentName string) (bool, int) {
	r1.mu.Lock()
	record, exists := r1.agents[agentName]
	if !exists {
		record = &requestRecord{}
		r1.agents[agentName] = record
	}
	r1.mu.Unlock()

	record.mu.Lock()
	defer record.mu.Unlock()

	now := time.Now()
	windowStart := now.Add(-r1.window)

	valid := []time.Time{}
	for _,t := range record.timestamps {
		if t.After(windowStart) {
			valid = append(valid, t)
		}
	}
	record.timestamps = valid

	if len(record.timestamps) >= r1.limit {
		return false, 0
	}

	record.timestamps = append(record.timestamps, now)
	remaining := r1.limit - len(record.timestamps)
	return true, remaining

}

func (r1 *RateLimiter) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	for range ticker.C {
		r1.mu.Lock()
		now := time.Now()
		for agent, record := range r1.agents {
			record.mu.Lock()
			windowStart := now.Add(-r1.window)
			hasRecent := false
			for _,t := range record.timestamps {
				if t.After(windowStart) {
					hasRecent = true
					break
				}
			}
			if !hasRecent {
				delete(r1.agents, agent)
			}
			record.mu.Unlock()
		}
		r1.mu.Unlock()
	}
}

func rateLimitMiddleware(rl *RateLimiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			agentName := r.Header.Get("X-Agent-Name")
			if agentName == "" {
				next.ServeHTTP(w,r)
				return
			}

			allowed, remaining := rl.isAllowed(agentName)

			w.Header().Set("X-RateLimit-Limit", "10")
			w.Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))
			w.Header().Set("X-RateLimit-Window", "60s")

			if !allowed {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusTooManyRequests)
				fmt.Fprintf(w, `{"error":"rate limit exceeded","agent":"%s","limit":10,"window":"60s"}`, agentName)
				return
			}
			next.ServeHTTP(w,r)

		})
	}
}
package main

import (
	"net/http"
	"strings"
)

var publicRoutes = []string{
	"/health",
	"/dashboard",
}

func isPublicRoute(path string) bool {
	for _, route := range publicRoutes {
		if path == route {
			return true
		}
	}
	return false
}

func buildKeyMap(config *Config) map[string]string {
	keys := make(map[string]string)

	for _, k := range config.Auth.APIKeys {
		keys[k.Key] = k.Agent
	}
	return keys
}

func authMiddleware(config *Config) func(http.Handler) http.Handler {
	keyMap := buildKeyMap(config)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			if isPublicRoute(r.URL.Path) {
				next.ServeHTTP(w, r)
				return
			}
			authHeader := r.Header.Get("Authorization")

			if authHeader == "" {
				http.Error(w, `{"error": "missing API Key"}`, http.StatusUnauthorized)
				return
			}

			if !strings.HasPrefix(authHeader, "Bearer ") {
				http.Error(w, `{"error": "invalid auth format. use: Bearer <key>"}`, http.StatusUnauthorized)
				return
			}

			apiKey := strings.TrimPrefix(authHeader, "Bearer ")
			agentName, valid := keyMap[apiKey]

			if !valid {
				http.Error(w, `{"error": "invalid api key"}`, http.StatusUnauthorized)
				return
			}

			r.Header.Set("X-Agent-Name", agentName)
			next.ServeHTTP(w, r)

		})
	}
}

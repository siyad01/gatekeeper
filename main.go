package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

func main(){
	fmt.Println("Gatekeeper Starting....")

	config, err := loadConfig("config.yaml")
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	fmt.Printf("Loaded %d API keys\n", len(config.Auth.APIKeys))
	fmt.Printf("Loaded %d MCP servers\n", len(config.Servers))

	audit, err := newAuditLogger(config.Audit.LogPath)
	if err != nil {
		log.Fatal("Failed to start audit logger: ", err)
	}
	defer audit.Close()
	fmt.Printf("Audit log → %s\n", config.Audit.LogPath)

	r1 := newRateLimiter(
		config.RateLimit.RequestsPerWindow,
		time.Duration(config.RateLimit.WindowSeconds)*time.Second,
	)

	fmt.Printf("Rate limit → %d requests per %ds\n",
		config.RateLimit.RequestsPerWindow,
		config.RateLimit.WindowSeconds,
	)

	hc := newHealthChecker(config)
	hc.start()
	fmt.Println("Health checker started → checking every 30s")



	mux:=http.NewServeMux()
	mux.HandleFunc("/health", healthHandler)
	mux.HandleFunc("/status", statusHandler(hc))
	mux.HandleFunc("/dashboard", dashboardPageHandler())
	mux.HandleFunc("/api/dashboard", dashboardAPIHandler(hc, config))
	mux.HandleFunc("/mcp/", proxyHandler(config, hc))

	stack := requestLogger(authMiddleware(config)(rateLimitMiddleware(r1)(auditMiddleware(audit)(mux))))
	server:=&http.Server{
		Addr: ":" + config.Server.Port,
		Handler: stack,

	}

	fmt.Printf("Gatekeeper running on http://localhost:%s\n", config.Server.Port)
	fmt.Printf("Dashboard → http://localhost:%s/dashboard\n", config.Server.Port)
	log.Fatal(server.ListenAndServe())

}

func healthHandler(w http.ResponseWriter, r *http.Request){
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, `{"status": "ok", "service": "gatekeeper"}`)
}

func statusHandler(hc *HealthChecker) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		statuses := hc.getStatuses()
		json.NewEncoder(w).Encode(map[string]interface{}{
			"service": "gatekeeper",
			"servers": statuses,
			"time": time.Now().UTC().Format(time.RFC3339),
		})
	}
}

func protectedHandler(w http.ResponseWriter, r *http.Request){
	agentName := r.Header.Get("X-Agent-Name")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"message": "Welcome to Gatekeeper", "agent": "%s"}`, agentName)
}
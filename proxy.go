package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

type MCPServer struct {
	Name string `yaml:"name"`
	URL string `yaml:"url"`
	Prefix string `yaml:"prefix"`
}

type JSONRPCRequest struct {
	JSONRPC string `json:"jsonrpc"`
	ID interface{} `json:"id"`
	Method string `json:"method"`
	Params json.RawMessage `json:"params,omitempty"`
}

type JSONRPCError struct {
	Code int `json:"code"`
	Message string `json:"message"`
}

type JSONRPCResponse struct {
	JSONRPC string `json:"jsonrpc"`
	ID interface{} `json:"id"`
	Result json.RawMessage `json:"result,omitempty"`
	Error *JSONRPCError `json:"error,omitempty"`
}

func proxyHandler(config *Config, hc *HealthChecker) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if r.Method != http.MethodPost {
			writeJSONRPCError(w, nil, -32600, "only Post method allowed")
			return
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			writeJSONRPCError(w, nil, -32700, "could not read request body")
			return
		}
		defer r.Body.Close()

		var rpcReq JSONRPCRequest
		if err := json.Unmarshal(body, &rpcReq); err != nil {
			writeJSONRPCError(w, nil, -32700, "invalid JSON_RPC request")
			return
		}

		server := findServer(config, r.URL.Path)
		if server == nil {
			writeJSONRPCError(w, rpcReq.ID, -32601, fmt.Sprintf("no MCP Server found for path: %s", r.URL.Path))
			return
		}

		if !hc.isHealthy(server.Name) {
			writeJSONRPCError(w, rpcReq.ID, -32603, fmt.Sprintf("MCP server unavailable: %s — currently unhealthy", server.Name))
			return
		}
		agentName := r.Header.Get("X-Agent-Name")
		log.Printf("PROXY | agent=%s | server=%s | method=%s", agentName, server.Name, rpcReq.Method)

		targetURL := server.URL
		proxyReq, err := http.NewRequest(http.MethodPost, targetURL, bytes.NewReader(body))
		if err != nil {
			writeJSONRPCError(w, rpcReq.ID, -32603, "could not create proxy request")
			return
		}

		proxyReq.Header.Set("Content-Type", "application/json")
		proxyReq.Header.Set("X-Forwarded-Agent", agentName)
		proxyReq.Header.Set("X_Gatekeeper-Version", "0.1.0")

		client := &http.Client{Timeout: 30 * time.Second}
		resp, err := client.Do(proxyReq)
		if err != nil {
			writeJSONRPCError(w, rpcReq.ID, -32603, fmt.Sprintf("MCP server unreachable: %s", server.Name))
			return
		}
		defer resp.Body.Close()

		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			writeJSONRPCError(w, rpcReq.ID, -32603, "could not read MCP server response")
			return
		}

		w.WriteHeader(resp.StatusCode)
		w.Write(respBody)

	}
}

func findServer(config *Config, path string)*MCPServer {
	for i := range config.Servers {
		if strings.HasPrefix(path, config.Servers[i].Prefix) {
			return &config.Servers[i]
		}
	}
	return nil
}

func writeJSONRPCError(w http.ResponseWriter, id interface{}, code int, message string) {
	resp := JSONRPCResponse{
		JSONRPC: "2.0",
		ID: id,
		Error: &JSONRPCError{
			Code: code,
			Message: message,
		},
	}
	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(resp)
}
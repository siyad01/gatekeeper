package main

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type APIKey struct {
	Key string `yaml:"key"`
	Agent string `yaml:"agent"`
}

type AuthConfig struct {
	APIKeys []APIKey `yaml:"api_keys"`
}

type ServerConfig struct {
	Port string `yaml:"port"`
	Name string `yaml:"name"`
}

type AuditConfig struct {
	LogPath string `yaml:"log_path"`
}

type RateLimitConfig struct {
	RequestsPerWindow int `yaml:"requests_per_window"`
	WindowSeconds     int `yaml:"window_seconds"`
}

type Config struct {
	Server ServerConfig `yaml:"server"`
	Auth AuthConfig `yaml:"auth"`
	Audit AuditConfig `yaml:"audit"`
	RateLimit RateLimitConfig `yaml:"rate_limit"`
	Servers []MCPServer `yaml:"mcp_servers"`
}

func loadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("could not read config file: %w", err)
	}

	var config Config
	err = yaml.Unmarshal(data, &config)
	if err!= nil {
		return nil, fmt.Errorf("could not parse config file: %w", err)
	}
	return &config, nil
}

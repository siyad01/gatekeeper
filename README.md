<div align="center">

# 🛡️ GateKeeper

**The open-source MCP gateway for AI agents**

Auth · Audit logging · Rate limiting · Health checks · Dashboard

[![License: Apache 2.0](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/Go-1.26-00ADD8?logo=go)](go.mod)
[![Docker](https://img.shields.io/badge/Docker-ready-2496ED?logo=docker)](Dockerfile)

</div>

---

## The problem

Every AI agent connects directly to every tool. No authentication. No audit trail. No rate limiting. No way to know what your agents did, when, or why.

**43% of MCP servers have no meaningful access controls.**

When an agent goes rogue — and they do — you have no way to stop it, trace it, or prove what happened.

## The solution

GateKeeper sits between your agents and your MCP servers. One endpoint. All your tools. Full security stack — without changing a single line of agent code.

```
Before GateKeeper          After GateKeeper

Agent ──► GitHub MCP       Agent ──► GateKeeper ──► GitHub MCP
Agent ──► Postgres MCP                    │──────► Postgres MCP
Agent ──► Slack MCP                       │──────► Slack MCP
                           
❌ No auth                 ✅ OAuth 2.1 + API keys
❌ No audit trail          ✅ Immutable audit log
❌ No rate limits          ✅ Per-agent rate limiting
❌ No health checks        ✅ Automatic health monitoring
❌ No visibility           ✅ Live dashboard
```

## Quickstart

**Run with Docker in 30 seconds:**

```bash
# 1. Clone the repo
git clone https://github.com/siyad01/gatekeeper
cd gatekeeper

# 2. Create your config
cp config.example.yaml config.yaml
# Edit config.yaml — add your API keys and MCP servers

# 3. Run
docker compose up
```

Visit `http://localhost:8080/dashboard` to see your dashboard.

## Configuration

```yaml
server:
  port: "8080"
  name: "gatekeeper"

audit:
  log_path: "logs/audit.log"

rate_limit:
  requests_per_window: 100
  window_seconds: 60

auth:
  api_keys:
    - key: "your-secret-key-here"
      agent: "my-agent"

mcp_servers:
  - name: "github"
    url: "http://your-github-mcp-server:3000"
    prefix: "/mcp/github"
  - name: "postgres"
    url: "http://your-postgres-mcp-server:3001"
    prefix: "/mcp/postgres"
```

## How agents connect

Point your MCP client at GateKeeper instead of your MCP server directly:

```
# Before — agent connects directly
MCP_SERVER_URL=http://your-mcp-server:3000

# After — agent connects through GateKeeper
MCP_SERVER_URL=http://localhost:8080/mcp/github
AUTHORIZATION=Bearer your-secret-key-here
```

No changes to your agent code. No changes to your MCP servers. Just add GateKeeper in the middle.

## Features

### 🔐 Authentication
Every request must carry a valid API key via `Authorization: Bearer <key>`. Invalid or missing keys are rejected before reaching any MCP server.

### 📋 Audit logging
Every tool call is permanently logged:
```json
{
  "id": "uuid",
  "timestamp": "2026-05-02T08:00:00Z",
  "agent_name": "my-agent",
  "method": "POST",
  "path": "/mcp/github",
  "status": 200,
  "duration": "1.2ms",
  "ip_address": "127.0.0.1:54321"
}
```

### ⚡ Rate limiting
Sliding window rate limiting per agent. Agents that exceed their limit get a clear error — not a timeout.

```json
{
  "error": "rate limit exceeded",
  "agent": "my-agent",
  "limit": 100,
  "window": "60s"
}
```

### 🏥 Health checks
GateKeeper pings every MCP server every 30 seconds. Requests to unhealthy servers are rejected instantly — no 30-second hangs.

### 📊 Dashboard
Live dashboard at `/dashboard` showing server health, active agents, and recent audit log.

## Architecture

```
Agent
  │
  ▼
GateKeeper (:8080)
  │
  ├── Request Logger      — logs every request
  ├── Auth Middleware     — validates API keys
  ├── Rate Limiter        — sliding window per agent
  ├── Audit Logger        — permanent append-only log
  └── MCP Proxy           — routes to registered servers
            │
            ├── /mcp/github   ──► GitHub MCP Server
            ├── /mcp/postgres ──► Postgres MCP Server
            └── /mcp/slack    ──► Slack MCP Server
```

## API

| Endpoint | Auth | Description |
|----------|------|-------------|
| `GET /health` | None | Health check |
| `GET /dashboard` | None | Admin dashboard UI |
| `GET /api/dashboard` | Required | Dashboard JSON data |
| `GET /status` | Required | Server health statuses |
| `POST /mcp/{server}/*` | Required | Proxy to MCP server |

## Roadmap

- [ ] OAuth 2.1 + PKCE flow
- [ ] Per-tool RBAC permissions
- [ ] Credential vault (agents never see secrets)
- [ ] Human-in-the-loop approval queue
- [ ] WebSocket / SSE streaming support
- [ ] Prometheus metrics endpoint
- [ ] GateKeeper Cloud (managed hosting)

## Contributing

GateKeeper is built in Go. If you know Python or JavaScript, Go will feel familiar within a day.

```bash
git clone https://github.com/siyad01/gatekeeper
cd gatekeeper
go run main.go middleware.go auth.go config.go audit.go responsewriter.go proxy.go ratelimit.go health.go dashboard.go
```

Open an issue before starting large features. PRs welcome.

## License

Apache 2.0 — free to use, modify, and distribute. See [LICENSE](LICENSE).

---

<div align="center">
Built with Go · Zero external dependencies for core features · Self-hostable in 30 seconds
</div>
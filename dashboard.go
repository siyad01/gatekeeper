package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)

type DashboardStats struct {
	TotalRequests  int                       `json:"total_requests"`
	BlockedRequests int                      `json:"blocked_requests"`
	ActiveAgents   []string                  `json:"active_agents"`
	ServerStatuses map[string]*ServerHealth  `json:"server_statuses"`
	RecentLogs     []AuditEntry              `json:"recent_logs"`
	Uptime         string                    `json:"uptime"`
}

var startTime = time.Now()

func dashboardAPIHandler(hc *HealthChecker, config *Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")

		recentLogs := readRecentLogs(config.Audit.LogPath, 20)

		totalRequests := len(recentLogs)
		blockedRequests := 0
		agentSet := make(map[string]bool)

		for _, entry := range recentLogs {
			if entry.Status == 401 || entry.Status == 429 {
				blockedRequests++
			}
			if entry.AgentName != "" {
				agentSet[entry.AgentName] = true
			}
		}

		activeAgents := []string{}
		for agent := range agentSet {
			activeAgents = append(activeAgents, agent)
		}

		stats := DashboardStats{
			TotalRequests:   totalRequests,
			BlockedRequests: blockedRequests,
			ActiveAgents:    activeAgents,
			ServerStatuses:  hc.getStatuses(),
			RecentLogs:      recentLogs,
			Uptime:          time.Since(startTime).Round(time.Second).String(),
		}

		json.NewEncoder(w).Encode(stats)
	}
}

func readRecentLogs(logPath string, limit int) []AuditEntry {
	entries := []AuditEntry{}

	file, err := os.Open(logPath)
	if err != nil {
		return entries
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	start := len(lines) - limit
	if start < 0 {
		start = 0
	}
	recentLines := lines[start:]

	for _, line := range recentLines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		var entry AuditEntry
		if err := json.Unmarshal([]byte(line), &entry); err == nil {
			entries = append(entries, entry)
		}
	}

	return entries
}

func dashboardPageHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, dashboardHTML())
	}
}

func dashboardHTML() string {
	return `<!DOCTYPE html>
<html class="dark" lang="en">
<head>
    <meta charset="utf-8"/>
    <meta content="width=device-width, initial-scale=1.0" name="viewport"/>
    <title>GateKeeper Dashboard</title>
    <script src="https://cdn.tailwindcss.com?plugins=forms,container-queries"></script>
    <link href="https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600;700;800&amp;family=Material+Symbols+Outlined:wght,FILL@100..700,0..1&amp;display=swap" rel="stylesheet"/>
    <style>
        .material-symbols-outlined {
            font-variation-settings: 'FILL' 0, 'wght' 400, 'GRAD' 0, 'opsz' 24;
        }
        body {
            background-color: #0b1326;
            margin: 0;
            padding: 0;
        }
    </style>
    <script id="tailwind-config">
        tailwind.config = {
            darkMode: "class",
            theme: {
                extend: {
                    "colors": {
                        "on-secondary-container": "#9af7ca",
                        "on-background": "#dae2fd",
                        "on-secondary": "#003825",
                        "surface-container-highest": "#2d3449",
                        "primary": "#b7c8e1",
                        "surface-container-lowest": "#060e20",
                        "primary-container": "#64748b",
                        "outline-variant": "#44474c",
                        "surface-tint": "#b7c8e1",
                        "inverse-surface": "#dae2fd",
                        "surface-container-low": "#131b2e",
                        "surface": "#0b1326",
                        "outline": "#8e9197",
                        "secondary": "#7cd8ad",
                        "background": "#0b1326",
                        "tertiary-fixed-dim": "#ffb3b1",
                        "on-primary-fixed-variant": "#38485d",
                        "on-primary-fixed": "#0b1c30",
                        "tertiary-container": "#ba5253",
                        "on-primary-container": "#f9f9ff",
                        "error-container": "#93000a",
                        "secondary-fixed-dim": "#7cd8ad",
                        "secondary-container": "#007451",
                        "tertiary-fixed": "#ffdad8",
                        "surface-bright": "#31394d",
                        "on-surface-variant": "#c4c6cd",
                        "inverse-primary": "#505f76",
                        "on-tertiary-fixed-variant": "#80272b",
                        "surface-dim": "#0b1326",
                        "primary-fixed-dim": "#b7c8e1",
                        "on-tertiary-fixed": "#410007",
                        "surface-container-high": "#222a3d",
                        "on-tertiary": "#620f17",
                        "on-secondary-fixed-variant": "#005137",
                        "primary-fixed": "#d3e4fe",
                        "inverse-on-surface": "#283044",
                        "secondary-fixed": "#98f5c8",
                        "on-tertiary-container": "#fff8f8",
                        "surface-container": "#171f33",
                        "on-surface": "#dae2fd",
                        "tertiary": "#ffb3b1",
                        "on-secondary-fixed": "#002114",
                        "on-primary": "#213145",
                        "on-error-container": "#ffdad6",
                        "on-error": "#690005",
                        "error": "#ffb4ab",
                        "surface-variant": "#2d3449"
                    },
                    "borderRadius": {
                        "DEFAULT": "0.125rem",
                        "lg": "0.25rem",
                        "xl": "0.5rem",
                        "full": "0.75rem"
                    },
                    "spacing": {
                        "unit": "8px",
                        "table-cell-padding": "16px 24px",
                        "container-margin": "32px",
                        "gutter": "24px",
                        "card-padding": "24px"
                    },
                    "fontFamily": {
                        "label-caps": ["Inter"],
                        "label-md": ["Inter"],
                        "h3": ["Inter"],
                        "body-md": ["Inter"],
                        "h2": ["Inter"],
                        "code": ["monospace"],
                        "body-lg": ["Inter"],
                        "h1": ["Inter"],
                        "body-sm": ["Inter"],
                        "display": ["Inter"]
                    },
                    "fontSize": {
                        "label-caps": ["12px", {"lineHeight": "1", "letterSpacing": "0.1em", "fontWeight": "600"}],
                        "label-md": ["12px", {"lineHeight": "1", "letterSpacing": "0.02em", "fontWeight": "500"}],
                        "h3": ["20px", {"lineHeight": "1.4", "letterSpacing": "0", "fontWeight": "500"}],
                        "body-md": ["16px", {"lineHeight": "1.6", "letterSpacing": "0", "fontWeight": "400"}],
                        "h2": ["24px", {"lineHeight": "1.3", "letterSpacing": "-0.01em", "fontWeight": "500"}],
                        "code": ["13px", {"lineHeight": "1.5", "fontWeight": "400"}],
                        "body-lg": ["18px", {"lineHeight": "1.6", "letterSpacing": "0", "fontWeight": "400"}],
                        "h1": ["32px", {"lineHeight": "1.2", "letterSpacing": "-0.02em", "fontWeight": "600"}],
                        "body-sm": ["14px", {"lineHeight": "1.5", "letterSpacing": "0", "fontWeight": "400"}],
                        "display": ["48px", {"lineHeight": "1.1", "letterSpacing": "-0.02em", "fontWeight": "600"}]
                    }
                }
            }
        }
    </script>
</head>
<body class="font-body-md text-on-surface selection:bg-primary-container/30">
    <header class="bg-slate-950/80 backdrop-blur-md text-slate-400 dark:text-slate-500 font-inter antialiased tracking-tight fixed top-0 w-full border-b border-slate-800 flex justify-between items-center h-16 px-6 z-50">
        <div class="flex items-center gap-unit">
            <span class="text-lg font-bold tracking-tighter text-slate-100">GateKeeper</span>
            <span class="ml-4 px-2 py-0.5 bg-surface-container-high text-[10px] rounded border border-outline-variant/30 text-on-surface-variant font-code">v1.2.4-stable</span>
        </div>
        <div class="flex items-center gap-6">
            <div class="flex items-center gap-4">
                <div class="text-[10px] uppercase tracking-widest text-secondary font-bold hidden md:block" id="uptime">Uptime: --:--:--</div>
                <div class="h-8 w-8 rounded-full border border-outline-variant/40 overflow-hidden bg-surface-variant">
                    <img alt="Administrator Profile" class="w-full h-full object-cover" src="https://lh3.googleusercontent.com/aida-public/AB6AXuAf3qHP37cb_tL0rChyHrkTZHdvAWSr28Cu1ZZRjFSQgloEpteTss8B3GgBbMfJlnt-qS6NmCWLECs8GoWQ7JZWBgG1sIU6D2YQglPR8bqbSc-PWGTHEP3tjhxj07Kiwwk4HLWsieDXjMwQkRQWj7W1-EjBxNijf9APVrX1qXs3yJpt0e6ClAttcuiD7SPP_21qP0CyGXAkE_BusuoJch4rFKgkjfLc2jQTLnwBxEtayBq7binvr9RDwUWKFEDhltXBQ-X7lZYplZo"/>
                </div>
            </div>
        </div>
    </header>

    <main class="pt-16 min-h-screen bg-background">
        <div class="p-8 max-w-[1600px] mx-auto space-y-gutter">
            <div class="flex justify-between items-end mb-4">
                <div>
                    <h1 class="font-h1 text-h1 text-on-background">Operations Overview</h1>
                    <p class="text-body-sm text-on-surface-variant mt-1">Real-time gateway status and security heuristics</p>
                </div>
                <div class="flex items-center gap-3">
                    <button onclick="fetchData()" class="flex items-center gap-2 px-4 py-2 border border-outline-variant/30 rounded text-label-md font-label-md text-on-surface hover:bg-surface-variant/40 transition-all">
                        <span class="material-symbols-outlined text-sm">refresh</span>
                        Manual Refresh
                    </button>
                    <button class="flex items-center gap-2 px-4 py-2 bg-secondary-container text-on-secondary-container rounded text-label-md font-label-md hover:opacity-90 transition-all">
                        <span class="material-symbols-outlined text-sm">add</span>
                        New Policy
                    </button>
                </div>
            </div>

            <div class="grid grid-cols-1 md:grid-cols-4 gap-gutter">
                <div class="bg-surface-container border border-outline-variant/30 p-card-padding rounded-xl">
                    <div class="flex items-center justify-between mb-4">
                        <span class="text-label-caps font-label-caps text-on-surface-variant">Total Requests</span>
                        <span class="material-symbols-outlined text-primary">database</span>
                    </div>
                    <div class="text-display font-display leading-none" id="total-requests">--</div>
                </div>
                <div class="bg-surface-container border border-outline-variant/30 p-card-padding rounded-xl">
                    <div class="flex items-center justify-between mb-4">
                        <span class="text-label-caps font-label-caps text-on-surface-variant">Blocked</span>
                        <span class="material-symbols-outlined text-tertiary" style="font-variation-settings: 'FILL' 1;">gpp_bad</span>
                    </div>
                    <div class="text-display font-display leading-none text-tertiary" id="blocked-requests">--</div>
                </div>
                <div class="bg-surface-container border border-outline-variant/30 p-card-padding rounded-xl">
                    <div class="flex items-center justify-between mb-4">
                        <span class="text-label-caps font-label-caps text-on-surface-variant">Active Agents</span>
                        <span class="material-symbols-outlined text-primary">smart_toy</span>
                    </div>
                    <div class="text-display font-display leading-none" id="active-agents-count">--</div>
                </div>
                <div class="bg-surface-container border border-outline-variant/30 p-card-padding rounded-xl">
                    <div class="flex items-center justify-between mb-4">
                        <span class="text-label-caps font-label-caps text-on-surface-variant">System Health</span>
                        <span class="material-symbols-outlined text-secondary">vital_signs</span>
                    </div>
                    <div class="text-display font-display leading-none" id="healthy-servers">--</div>
                    <div class="mt-4 w-full bg-surface-container-highest h-1 rounded-full overflow-hidden">
                        <div id="health-bar" class="bg-secondary h-full" style="width: 0%"></div>
                    </div>
                </div>
            </div>

            <div class="grid grid-cols-1 lg:grid-cols-3 gap-gutter">
                <div class="lg:col-span-2 bg-surface-container border border-outline-variant/30 rounded-xl overflow-hidden">
                    <div class="p-6 border-b border-outline-variant/10 flex items-center justify-between">
                        <h3 class="font-h3 text-h3">MCP Servers</h3>
                    </div>
                    <div class="p-6 space-y-4" id="servers-list">
                        <div class="text-center text-on-surface-variant py-8">Loading server data...</div>
                    </div>
                </div>

                <div class="bg-surface-container border border-outline-variant/30 rounded-xl overflow-hidden flex flex-col">
                    <div class="p-6 border-b border-outline-variant/10">
                        <h3 class="font-h3 text-h3">Connected Agents</h3>
                    </div>
                    <div class="p-6 flex-1" id="agents-list">
                        <div class="text-center text-on-surface-variant py-8">Loading agents...</div>
                    </div>
                    <div class="mt-auto p-4 bg-surface-container-high/30">
                        <div class="flex items-center justify-between text-[11px] text-on-surface-variant uppercase tracking-wider font-semibold">
                            <span>Pool Capacity</span>
                            <span id="agent-capacity">--/50 Agents</span>
                        </div>
                        <div class="mt-2 w-full bg-surface-container-highest h-1 rounded-full overflow-hidden">
                            <div id="agent-bar" class="bg-primary h-full" style="width: 0%"></div>
                        </div>
                    </div>
                </div>
            </div>

            <div class="bg-surface-container border border-outline-variant/30 rounded-xl overflow-hidden">
                <div class="p-6 border-b border-outline-variant/10 flex items-center justify-between bg-surface-container-high/20">
                    <div class="flex items-center gap-4">
                        <h3 class="font-h3 text-h3">Security Audit Log</h3>
                    </div>
                </div>
                <div class="overflow-x-auto">
                    <table class="w-full text-left border-collapse">
                        <thead>
                            <tr class="bg-surface-container-high/10 text-label-caps font-label-caps text-on-surface-variant border-b border-outline-variant/10">
                                <th class="p-table-cell-padding">Time</th>
                                <th class="p-table-cell-padding">Identity</th>
                                <th class="p-table-cell-padding">Action</th>
                                <th class="p-table-cell-padding">Status</th>
                                <th class="p-table-cell-padding">Performance</th>
                            </tr>
                        </thead>
                        <tbody class="divide-y divide-outline-variant/5" id="log-container"></tbody>
                    </table>
                </div>
            </div>
        </div>
    </main>

    <script>
        function getKey() {
            var key = localStorage.getItem('gk-api-key');
            if (!key) {
                key = prompt('Enter your GateKeeper API key:');
                if (key) localStorage.setItem('gk-api-key', key);
            }
            return key || '';
        }

        async function fetchData() {
            try {
                const res = await fetch('/api/dashboard', {
                    headers: { 'Authorization': 'Bearer ' + getKey() }
                });
                const data = await res.json();
                render(data);
            } catch(e) {
                console.error('Failed to fetch:', e);
            }
        }

        function render(data) {
            document.getElementById('uptime').textContent = 'Uptime: ' + data.uptime;
            document.getElementById('total-requests').textContent = data.total_requests;
            document.getElementById('blocked-requests').textContent = data.blocked_requests;
            
            var agents = data.active_agents || [];
            document.getElementById('active-agents-count').textContent = agents.length;
            document.getElementById('agent-capacity').textContent = agents.length + '/50 Agents';
            document.getElementById('agent-bar').style.width = (agents.length / 50 * 100) + '%';

            var statuses = data.server_statuses || {};
            var serversArray = Object.values(statuses);
            var healthyCount = serversArray.filter(function(s) { return s.status === 'healthy'; }).length;
            var healthPercent = serversArray.length > 0 ? (healthyCount / serversArray.length * 100).toFixed(1) : 0;
            
            document.getElementById('healthy-servers').textContent = healthPercent + '%';
            document.getElementById('health-bar').style.width = healthPercent + '%';

            var serversList = document.getElementById('servers-list');
            if (serversArray.length === 0) {
                serversList.innerHTML = '<div class="text-center py-8 opacity-50">No servers configured</div>';
            } else {
                serversList.innerHTML = serversArray.map(function(s) {
                    var statusColor = s.status === 'healthy' ? 'bg-secondary' : 
                                     s.status === 'unhealthy' ? 'bg-tertiary' : 'bg-outline';
                    return '<div class="flex items-center justify-between p-4 bg-surface-container-low rounded-lg border border-outline-variant/5">' +
                        '<div class="flex items-center gap-4">' +
                            '<div class="w-2 h-2 rounded-full ' + statusColor + '"></div>' +
                            '<div>' +
                                '<div class="text-body-md font-medium">' + s.name + '</div>' +
                                '<div class="text-xs text-on-surface-variant font-code opacity-70">' + s.url + '</div>' +
                            '</div>' +
                        '</div>' +
                        '<div class="flex items-center gap-4">' +
                            '<span class="px-2 py-1 bg-surface-container-highest rounded text-[10px] text-on-surface font-code">' + (s.latency || '—') + '</span>' +
                        '</div>' +
                    '</div>';
                }).join('');
            }

            var agentsList = document.getElementById('agents-list');
            if (agents.length === 0) {
                agentsList.innerHTML = '<div class="text-center py-8 opacity-50">No active agents</div>';
            } else {
                agentsList.innerHTML = '<div class="flex flex-wrap gap-2">' + agents.map(function(a) {
                    return '<div class="px-3 py-2 bg-surface-container-low border border-outline-variant/20 rounded flex items-center gap-2">' +
                        '<span class="material-symbols-outlined text-xs text-secondary" style="font-variation-settings: \'FILL\' 1;">circle</span>' +
                        '<span class="text-label-md font-label-md">' + a + '</span>' +
                    '</div>';
                }).join('') + '</div>';
            }

            var logs = (data.recent_logs || []).reverse();
            var logContainer = document.getElementById('log-container');
            if (logs.length === 0) {
                logContainer.innerHTML = '<tr><td colspan="5" class="p-8 text-center opacity-50">No audit entries yet</td></tr>';
            } else {
                logContainer.innerHTML = logs.map(function(e) {
                    var isSuccess = e.status === 200 || e.status === '200 OK';
                    var statusClass = isSuccess ? 'text-secondary' : 'text-tertiary';
                    var dotClass = isSuccess ? 'bg-secondary' : 'bg-tertiary';
                    var initials = (e.agent_name || 'AN').substring(0,2).toUpperCase();
                    
                    return '<tr class="hover:bg-surface-variant/10 transition-colors">' +
                        '<td class="p-table-cell-padding font-code text-[13px] text-on-surface-variant">' + new Date(e.timestamp).toLocaleTimeString() + '</td>' +
                        '<td class="p-table-cell-padding">' +
                            '<div class="flex items-center gap-2">' +
                                '<div class="w-6 h-6 rounded-full bg-primary/20 flex items-center justify-center text-[10px] font-bold text-primary">' + initials + '</div>' +
                                '<span class="text-body-sm">' + (e.agent_name || 'anonymous') + '</span>' +
                            '</div>' +
                        '</td>' +
                        '<td class="p-table-cell-padding">' +
                            '<span class="font-code text-primary-container text-xs px-2 py-0.5 bg-primary-container/10 rounded mr-2">' + e.method + '</span>' +
                            '<span class="text-body-sm font-code">' + e.path + '</span>' +
                        '</td>' +
                        '<td class="p-table-cell-padding">' +
                            '<span class="flex items-center gap-2 ' + statusClass + ' text-xs font-semibold">' +
                                '<span class="w-1.5 h-1.5 rounded-full ' + dotClass + '"></span>' + e.status + 
                            '</span>' +
                        '</td>' +
                        '<td class="p-table-cell-padding font-code text-xs text-on-surface-variant">' + e.duration + '</td>' +
                    '</tr>';
                }).join('');
            }
        }

        fetchData();
        setInterval(fetchData, 5000);
    </script>
</body>
</html>`
}
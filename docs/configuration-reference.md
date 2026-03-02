# AgentTeams Configuration Reference

This document describes all configuration options for the AgentTeams Agent.

## Configuration File Location

- Windows: `C:\ProgramData\AgentTeams\agent.yaml`
- Environment variable override: `AGENT_CONFIG_PATH`

## Configuration Structure

```yaml
agent:
  # Agent identification
  id: ""
  token: ""
  server_url: ""

connection:
  # Connection settings
  retry_interval: 5s
  max_retry_interval: 60s
  ping_interval: 10s
  pong_timeout: 5s

heartbeat:
  # Heartbeat settings
  interval: 1s

metrics:
  # Metrics collection
  interval: 1m

task:
  # Task execution
  max_concurrent: 4
  queue_size: 100
  default_timeout: 5m

update:
  # Auto-update
  check_interval: 1h
  idle_required: true

logging:
  # Logging
  level: info
  file: ""
  max_size: 100MB
  max_files: 5
```

## Agent Section

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `id` | string | No | auto-generated | Unique agent identifier |
| `token` | string | Yes | - | Registration token from server |
| `server_url` | string | Yes | - | WebSocket server URL (wss://) |

### Example

```yaml
agent:
  id: "agent-001"
  token: "abc123def456"
  server_url: "wss://agentteams.example.com:443/api/v1/agent/ws"
```

## Connection Section

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `retry_interval` | duration | 5s | Initial reconnect delay |
| `max_retry_interval` | duration | 60s | Maximum reconnect delay |
| `ping_interval` | duration | 10s | WebSocket ping interval |
| `pong_timeout` | duration | 5s | Time to wait for pong response |

### Duration Format

Duration values use Go duration format:
- `s` - seconds (e.g., `30s`)
- `m` - minutes (e.g., `5m`)
- `h` - hours (e.g., `1h`)

## Heartbeat Section

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `interval` | duration | 1s | Heartbeat message interval |

## Metrics Section

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `interval` | duration | 1m | Metrics collection interval |

### Collected Metrics

- CPU usage percentage
- Memory usage (total, used, available, percent)
- Disk usage (total, used, free, percent)
- System uptime

## Task Section

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `max_concurrent` | int | 4 | Maximum concurrent tasks |
| `queue_size` | int | 100 | Maximum queued tasks |
| `default_timeout` | duration | 5m | Default task timeout |

## Update Section

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `check_interval` | duration | 1h | Update check interval |
| `idle_required` | bool | true | Only update when idle |

## Logging Section

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `level` | string | info | Log level |
| `file` | string | - | Log file path |
| `max_size` | size | 100MB | Max log file size |
| `max_files` | int | 5 | Max log files to retain |

### Log Levels

- `trace` - Very detailed debugging
- `debug` - Debugging information
- `info` - General information
- `warn` - Warning messages
- `error` - Error messages

## Environment Variables

Configuration can be overridden with environment variables:

| Variable | Description |
|----------|-------------|
| `AGENT_ID` | Agent ID |
| `AGENT_TOKEN` | Registration token |
| `AGENT_SERVER_URL` | Server URL |
| `AGENT_LOG_LEVEL` | Log level |
| `AGENT_CONFIG_PATH` | Config file path |

## Complete Example

```yaml
# AgentTeams Agent Configuration

agent:
  id: ""
  token: "your-registration-token-here"
  server_url: "wss://agentteams.example.com:443/api/v1/agent/ws"

connection:
  retry_interval: 5s
  max_retry_interval: 60s
  ping_interval: 10s
  pong_timeout: 5s

heartbeat:
  interval: 1s

metrics:
  interval: 1m

task:
  max_concurrent: 4
  queue_size: 100
  default_timeout: 5m

update:
  check_interval: 1h
  idle_required: true

logging:
  level: info
  file: "C:/ProgramData/AgentTeams/logs/agent.log"
  max_size: 100MB
  max_files: 5
```

# Heartbeat Output Plugin Spec

## Objective

Provide an output plugin that allows a Telegraf agent to send a "heartbeat" request to an external endpoint. This plugin enables centralized monitoring of agent health and configuration.

## Overview

This plugin is designed to provide visibility into Telegraf agent deployments, particularly for users managing many agents through the Telegraf UI/Configuration Manager. It is also flexible enough to be used in custom applications.

Key goals:

- Centralized monitoring of agent health
- Visibility into agent configurations
- Optional inclusion of recent logs
- Future support for user-defined status conditions

## Plugin Configuration

Example configuration in TOML:

```toml
[[outputs.heartbeat]]
  # Phase 1 (MVP)
  ## URLs of heartbeat endpoints
  urls = ["http://monitoring.example.com/heartbeat"]
  ## Unique identifier for the agent (required)
  agent_id = "agent-prod-web-01"
  ## HTTP headers to include with the heartbeat request
  headers = {
    Authorization = "Token ..."
  }
  ## Omit the agent hostname from the heartbeat request. Default is global.omit_hostname.
  omit_hostname = false
  ## Omit the agent IP from the heartbeat request.
  omit_ip = false

  # Phase 2
  ## Include logs in the heartbeat request
  send_logs = true
  ## Maximum number of log entries to send with each heartbeat request
  log_limit = 5
  ## Minimum log level to send (debug|info|warning|error)
  log_level_threshold = "warning"

  # Phase 3
  ## Logical conditions that determine the agent status
  status_conditions = # ...
```

## Output Format

The plugin sends a POST request with a JSON body that includes data about the agent.

### Always Included Fields

- `id`: Unique identifier for the agent (from `agent_id`)
- `status`: Current agent status (default: "OK")
- `configs`: List of configuration sources (paths or URLs)
- `telegraf_version`: Version of the running Telegraf agent

### Optional Fields

- `hostname`: Hostname of the machine
- `ip`: IP address of the machine
- `logs`: List of recent log entries (if `send_logs = true`)
  - `level`: Log severity level
  - `message`: Log content
  - `timestamp`: RFC3339-formatted timestamp (respects `global.log_with_timezone`)

### Example Request Body

```json
{
  "id": "agent-prod-web-01",
  "status": "OK",
  "configs": [
    "/etc/telegraf/config.toml",
    "https://localhost:8181/api/configs/xAohlapd"
  ],
  "telegraf_version": "1.25.0",
  "hostname": "web-server-01",
  "ip": "192.168.1.100",
  "logs": [
    {
      "level": "info",
      "message": "Configuration loaded successfully",
      "timestamp": "2025-06-16T18:03:00Z"
    }
  ]
}
```

## Development Phases

### Phase 1 (MVP)

- Send basic agent data (`id`, `status`, `configs`, `telegraf_version`)
- Send optional data (`hostname` and `ip`)

### Phase 2

- Add support for logs
- Allow filtering logs by severity
- Limit number of logs sent

### Phase 3

- Support user-defined logic to compute `status`

## Keywords

agent, outputs, status, heartbeat, monitoring, logs

## Open Questions

- What is the source of logs to include? (Current flush vs historical)
- How will users define `status_conditions`?
- Will Telegraf include any default status rules?

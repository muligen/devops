## ADDED Requirements

### Requirement: Server provides dashboard statistics

The Server SHALL provide dashboard statistics API.

#### Scenario: Overall statistics
- **WHEN** user requests dashboard stats
- **THEN** Server returns:
  - total Agents count
  - online Agents count
  - offline Agents count
  - Agents in maintenance
  - total tasks count (24h)
  - successful tasks count (24h)
  - failed tasks count (24h)

#### Scenario: Real-time statistics
- **WHEN** user subscribes to dashboard updates
- **THEN** Server pushes updates via WebSocket
- **AND** updates include latest statistics

### Requirement: Server provides Agent metrics history

The Server SHALL provide Agent metrics history API.

#### Scenario: Query metrics history
- **WHEN** user requests metrics for Agent
- **THEN** Server returns metrics for specified time range
- **AND** supports resolution: raw (1min), hourly, daily
- **AND** default time range is last 24 hours

#### Scenario: Metrics aggregation
- **WHEN** user requests aggregated metrics
- **THEN** Server returns aggregated values:
  - average CPU usage
  - peak CPU usage
  - average memory usage
  - peak memory usage
  - average disk usage

### Requirement: Server provides task statistics

The Server SHALL provide task statistics API.

#### Scenario: Task statistics by Agent
- **WHEN** user requests task stats by Agent
- **THEN** Server returns:
  - tasks per Agent (24h)
  - success rate per Agent
  - average execution time per Agent

#### Scenario: Task statistics by type
- **WHEN** user requests task stats by type
- **THEN** Server returns:
  - tasks per command type
  - success rate per type
  - average execution time per type

### Requirement: Server supports alerting rules

The Server SHALL support alerting rules configuration.

#### Scenario: Create alert rule
- **WHEN** admin creates alert rule
- **THEN** Server stores rule with conditions:
  - metric type (cpu, memory, disk, agent_offline)
  - threshold value
  - duration
  - notification target

#### Scenario: Alert evaluation
- **WHEN** metric crosses threshold for specified duration
- **THEN** Server triggers alert
- **AND** Server sends notification
- **AND** Server logs alert event

#### Scenario: Alert types
- **WHEN** configuring alert rules
- **THEN** supported alert types are:
  - CPU usage exceeds threshold
  - Memory usage exceeds threshold
  - Disk usage exceeds threshold
  - Agent offline for extended period

### Requirement: Server provides alert history

The Server SHALL provide alert history API.

#### Scenario: Query alerts
- **WHEN** user requests alert history
- **THEN** Server returns alerts filtered by:
  - Agent ID
  - Alert type
  - Time range
  - Status (active, resolved)

#### Scenario: Alert acknowledgment
- **WHEN** user acknowledges alert
- **THEN** Server marks alert as acknowledged
- **AND** Server records acknowledged_by user

### Requirement: Server supports notification channels

The Server SHALL support multiple notification channels.

#### Scenario: Webhook notification
- **WHEN** alert triggers with webhook configured
- **THEN** Server sends POST request to webhook URL
- **AND** request contains alert details (JSON)

#### Scenario: Email notification
- **WHEN** alert triggers with email configured
- **THEN** Server sends email to configured addresses
- **AND** email contains alert details

### Requirement: Server provides system health API

The Server SHALL provide system health API.

#### Scenario: Health check
- **WHEN** health check endpoint is called
- **THEN** Server returns status of:
  - database connection
  - Redis connection
  - RabbitMQ connection
  - object storage connection
- **AND** returns 200 if all healthy
- **AND** returns 503 if any unhealthy

#### Scenario: Metrics endpoint
- **WHEN** Prometheus scrapes metrics endpoint
- **THEN** Server returns metrics in Prometheus format
- **AND** includes:
  - connected Agents gauge
  - task execution histogram
  - request latency histogram

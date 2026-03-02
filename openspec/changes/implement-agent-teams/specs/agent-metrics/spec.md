## ADDED Requirements

### Requirement: Agent collects machine metrics

The Agent SHALL collect machine metrics including CPU, memory, and disk usage.

#### Scenario: CPU metrics collection
- **WHEN** Agent collects metrics
- **THEN** Agent records CPU usage percentage
- **AND** CPU usage is accurate within 5% tolerance

#### Scenario: Memory metrics collection
- **WHEN** Agent collects metrics
- **THEN** Agent records total memory in bytes
- **AND** Agent records used memory in bytes
- **AND** Agent calculates memory usage percentage

#### Scenario: Disk metrics collection
- **WHEN** Agent collects metrics
- **THEN** Agent records total disk space in bytes
- **AND** Agent records used disk space in bytes
- **AND** Agent calculates disk usage percentage

#### Scenario: Uptime collection
- **WHEN** Agent collects metrics
- **THEN** Agent records system uptime in seconds

### Requirement: Agent sends metrics at regular intervals

The Agent SHALL send collected metrics to Server every 1 minute.

#### Scenario: Scheduled metrics send
- **WHEN** 1 minute interval elapses
- **THEN** Agent collects current metrics
- **AND** Agent sends metrics message to Server

#### Scenario: Metrics send failure
- **WHEN** metrics send fails
- **THEN** Agent logs the error
- **AND** Agent buffers metrics locally (up to 5 minutes)
- **AND** Agent retries on next interval

### Requirement: Server stores metrics history

The Server SHALL store metrics for historical analysis.

#### Scenario: Metrics storage
- **WHEN** Server receives metrics from Agent
- **THEN** Server stores metrics in database
- **AND** Server associates metrics with Agent ID
- **AND** Server records collection timestamp

#### Scenario: Metrics retention
- **WHEN** metrics are stored
- **THEN** raw metrics are retained for 7 days
- **AND** aggregated metrics (hourly avg) are retained for 30 days

### Requirement: Server provides metrics query API

The Server SHALL provide API to query Agent metrics.

#### Scenario: Query recent metrics
- **WHEN** user requests metrics for an Agent
- **THEN** Server returns metrics for specified time range
- **AND** Server supports pagination for large result sets

#### Scenario: Query aggregated metrics
- **WHEN** user requests metrics with aggregation
- **THEN** Server returns aggregated data (avg, min, max)
- **AND** Server supports grouping by hour, day, week

### Requirement: Metrics message format

The Agent SHALL use the standard metrics message format.

#### Scenario: Metrics message structure
- **WHEN** Agent sends metrics
- **THEN** message contains `type: "metrics"`
- **AND** message contains `agent_id`
- **AND** message contains `timestamp`
- **AND** message contains `data` object with:
  - `cpu_usage`: float (percentage)
  - `memory`: object with `total`, `used`, `percent`
  - `disk`: object with `total`, `used`, `percent`
  - `uptime`: integer (seconds)

## ADDED Requirements

### Requirement: Agent sends heartbeat at regular intervals

The Agent SHALL send heartbeat messages to Server every 1 second.

#### Scenario: Normal heartbeat
- **WHEN** Agent is connected and running normally
- **THEN** Agent sends heartbeat message every 1 second
- **AND** heartbeat contains AgentID and timestamp

#### Scenario: Heartbeat failure
- **WHEN** heartbeat send fails
- **THEN** Agent logs the error
- **AND** Agent continues to attempt heartbeat
- **AND** Agent triggers reconnection if multiple failures

### Requirement: Server monitors heartbeat

The Server SHALL monitor Agent heartbeat to determine online status.

#### Scenario: Heartbeat received
- **WHEN** Server receives heartbeat from Agent
- **THEN** Server updates `last_seen_at` timestamp
- **AND** Server keeps Agent status as `online`

#### Scenario: Heartbeat timeout
- **WHEN** Server does not receive heartbeat for 30 seconds
- **THEN** Server marks Agent as `offline`
- **AND** Server publishes `agent.offline` event

### Requirement: Heartbeat message format

The Agent SHALL use the standard heartbeat message format.

#### Scenario: Heartbeat message structure
- **WHEN** Agent sends heartbeat
- **THEN** message contains `type: "heartbeat"`
- **AND** message contains `agent_id`
- **AND** message contains `timestamp` in ISO 8601 format

### Requirement: Server handles heartbeat efficiently

The Server SHALL handle high-volume heartbeat messages efficiently.

#### Scenario: Bulk heartbeat processing
- **WHEN** Server receives multiple heartbeats concurrently
- **THEN** Server processes them asynchronously
- **AND** Server does not block other operations
- **AND** Server uses Redis for session state

### Requirement: Connection keepalive

The Agent SHALL use WebSocket ping/pong for connection keepalive.

#### Scenario: Ping interval
- **WHEN** Agent is idle (no messages for 10 seconds)
- **THEN** Agent sends WebSocket ping frame
- **AND** expects pong response within 5 seconds

#### Scenario: Pong timeout
- **WHEN** Agent does not receive pong within 5 seconds
- **THEN** Agent considers connection stale
- **AND** Agent triggers reconnection

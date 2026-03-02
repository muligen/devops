## ADDED Requirements

### Requirement: Agent establishes WebSocket connection to Server

The Agent SHALL establish a WebSocket over TLS (WSS) connection to the Server on startup.

#### Scenario: Successful connection
- **WHEN** Agent starts with valid configuration
- **THEN** Agent establishes WSS connection to Server
- **AND** connection uses TLS 1.2 or higher

#### Scenario: Connection failure with retry
- **WHEN** Agent fails to connect to Server
- **THEN** Agent retries with exponential backoff (5s, 10s, 30s, 60s max)
- **AND** Agent logs the connection error

### Requirement: Agent authenticates with Challenge-Response

The Agent SHALL authenticate using AgentID and Token with Challenge-Response mechanism.

#### Scenario: Successful authentication
- **WHEN** Agent connects and receives challenge nonce
- **THEN** Agent computes HMAC(token, nonce) and sends response
- **AND** Server validates and returns session ID

#### Scenario: Authentication failure
- **WHEN** Agent provides invalid credentials
- **THEN** Server rejects the connection
- **AND** Agent logs authentication failure
- **AND** Agent retries after backoff period

### Requirement: Agent maintains session state

The Agent SHALL maintain session state and reconnect automatically on disconnection.

#### Scenario: Automatic reconnection
- **WHEN** WebSocket connection is lost
- **THEN** Agent attempts to reconnect with exponential backoff
- **AND** Agent re-authenticates on successful reconnection
- **AND** Agent syncs any pending tasks after reconnection

#### Scenario: Session timeout
- **WHEN** Server session expires
- **THEN** Agent re-authenticates
- **AND** Agent continues normal operation

### Requirement: Server validates Agent identity

The Server SHALL validate Agent identity before allowing any operations.

#### Scenario: Valid Agent connection
- **WHEN** Agent presents valid AgentID and Token
- **THEN** Server creates session and stores in Redis
- **AND** Server publishes `agent.online` event

#### Scenario: Unknown Agent connection
- **WHEN** Agent presents unknown AgentID
- **THEN** Server rejects connection with appropriate error code
- **AND** Server logs the attempt

### Requirement: Server manages connection state

The Server SHALL track Agent connection state in real-time.

#### Scenario: Agent online
- **WHEN** Agent successfully authenticates
- **THEN** Server marks Agent as `online` in database
- **AND** Server updates `last_seen_at` timestamp

#### Scenario: Agent offline
- **WHEN** Agent disconnects or heartbeat times out
- **THEN** Server marks Agent as `offline`
- **AND** Server publishes `agent.offline` event

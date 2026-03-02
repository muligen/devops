## ADDED Requirements

### Requirement: Server manages Agent registration

The Server SHALL support Agent registration.

#### Scenario: New Agent registration
- **WHEN** admin registers a new Agent
- **THEN** Server generates unique Agent ID
- **AND** Server generates secret Token
- **AND** Server stores Agent in database
- **AND** Server returns credentials to admin

#### Scenario: Agent name uniqueness
- **WHEN** admin registers Agent with duplicate name
- **THEN** Server rejects registration
- **AND** Server returns appropriate error

### Requirement: Server provides Agent listing

The Server SHALL provide API to list all Agents.

#### Scenario: List all Agents
- **WHEN** user requests Agent list
- **THEN** Server returns all Agents with:
  - ID, name, status
  - last_seen_at timestamp
  - current version
  - metadata

#### Scenario: Filter Agents by status
- **WHEN** user requests Agents with status filter
- **THEN** Server returns only Agents matching status
- **AND** supported statuses: `online`, `offline`, `maintenance`

### Requirement: Server provides Agent details

The Server SHALL provide API to get Agent details.

#### Scenario: Get Agent by ID
- **WHEN** user requests Agent by ID
- **THEN** Server returns Agent details
- **AND** includes current status and version
- **AND** includes last seen timestamp

#### Scenario: Agent not found
- **WHEN** user requests non-existent Agent
- **THEN** Server returns 404 error

### Requirement: Server supports Agent deletion

The Server SHALL support deleting Agents.

#### Scenario: Delete Agent
- **WHEN** admin deletes an Agent
- **THEN** Server marks Agent as deleted (soft delete)
- **AND** Server revokes Agent's Token
- **AND** Agent cannot reconnect with old credentials

#### Scenario: Delete online Agent
- **WHEN** admin deletes online Agent
- **THEN** Server disconnects Agent's WebSocket connection
- **AND** Server marks Agent as deleted

### Requirement: Server supports Agent status update

The Server SHALL support updating Agent status.

#### Scenario: Set maintenance mode
- **WHEN** admin sets Agent to `maintenance` status
- **THEN** Server updates status
- **AND** Agent receives maintenance notification
- **AND** Agent pauses task execution

#### Scenario: Set back to active
- **WHEN** admin sets Agent back to `active`
- **THEN** Server updates status
- **AND** Agent resumes normal operation

### Requirement: Server supports Agent metadata

The Server SHALL support custom Agent metadata.

#### Scenario: Update metadata
- **WHEN** admin updates Agent metadata
- **THEN** Server stores metadata as JSON
- **AND** metadata can include arbitrary key-value pairs

#### Scenario: Query by metadata
- **WHEN** user queries Agents by metadata filter
- **THEN** Server returns matching Agents
- **AND** supports JSONB query syntax

### Requirement: Server tracks Agent version

The Server SHALL track Agent version history.

#### Scenario: Version tracking
- **WHEN** Agent reports new version after update
- **THEN** Server updates current version
- **AND** Server records version history
- **AND** Server records update timestamp

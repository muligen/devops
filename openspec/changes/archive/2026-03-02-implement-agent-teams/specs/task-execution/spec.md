## ADDED Requirements

### Requirement: Server sends commands to Agent

The Server SHALL send commands to connected Agents via WebSocket.

#### Scenario: Single command send
- **WHEN** user creates a task for an Agent
- **THEN** Server sends command message to Agent
- **AND** command contains unique `id`
- **AND** command contains `type` (e.g., `exec_shell`, `init_machine`)
- **AND** command contains `params` object
- **AND** command contains `timeout` in seconds

#### Scenario: Agent offline
- **WHEN** user creates task for offline Agent
- **THEN** Server queues task with status `pending`
- **AND** Server delivers task when Agent reconnects

### Requirement: Agent executes commands concurrently

The Agent SHALL support concurrent command execution.

#### Scenario: Concurrent execution
- **WHEN** Agent receives multiple commands
- **THEN** Agent executes them concurrently up to `max_concurrent` limit
- **AND** each command runs in isolated process
- **AND** command failures do not affect other commands

#### Scenario: Queue overflow
- **WHEN** command queue exceeds `queue_size` limit
- **THEN** Agent rejects new commands with `busy` status
- **AND** Server retries later or returns error to user

### Requirement: Agent enforces command timeout

The Agent SHALL enforce timeout for each command execution.

#### Scenario: Command timeout
- **WHEN** command execution exceeds timeout duration
- **THEN** Agent terminates the command process
- **AND** Agent returns result with status `timeout`
- **AND** Agent includes partial output if available

#### Scenario: Default timeout
- **WHEN** command does not specify timeout
- **THEN** Agent uses default timeout of 300 seconds (5 minutes)

### Requirement: Agent returns command results

The Agent SHALL return execution results to Server.

#### Scenario: Successful execution
- **WHEN** command completes successfully
- **THEN** Agent sends result with `status: "success"`
- **AND** result contains `exit_code: 0`
- **AND** result contains `output` (stdout + stderr)
- **AND** result contains `duration` in seconds

#### Scenario: Failed execution
- **WHEN** command fails
- **THEN** Agent sends result with `status: "failed"`
- **AND** result contains non-zero `exit_code`
- **AND** result contains `output` with error messages

### Requirement: Agent supports streaming output

The Agent SHALL support streaming command output for long-running commands.

#### Scenario: Output streaming
- **WHEN** command produces output during execution
- **THEN** Agent sends `output_chunk` messages periodically
- **AND** each chunk contains `id` (command ID)
- **AND** each chunk contains partial `output`

### Requirement: Agent supports built-in commands

The Agent SHALL support the following built-in commands:

#### Scenario: exec_shell command
- **WHEN** Agent receives `exec_shell` command
- **THEN** Agent executes the shell command in subprocess
- **AND** Agent captures stdout and stderr
- **AND** Agent returns exit code and output

#### Scenario: init_machine command
- **WHEN** Agent receives `init_machine` command
- **THEN** Agent downloads configuration from `config_url`
- **AND** Agent applies configuration (users, services, etc.)
- **AND** Agent returns initialization result

#### Scenario: clean_disk command
- **WHEN** Agent receives `clean_disk` command
- **THEN** Agent cleans specified categories (temp, cache, logs)
- **AND** Agent returns cleanup statistics

### Requirement: Server stores command results

The Server SHALL store command results in database.

#### Scenario: Result storage
- **WHEN** Server receives command result from Agent
- **THEN** Server stores result in `tasks` table
- **AND** Server updates task status
- **AND** Server stores output and exit code
- **AND** Server publishes `task.completed` or `task.failed` event

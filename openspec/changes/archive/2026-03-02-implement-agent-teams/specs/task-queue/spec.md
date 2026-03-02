## ADDED Requirements

### Requirement: Server manages task queue

The Server SHALL maintain a task queue for each Agent.

#### Scenario: Task creation
- **WHEN** user creates a task
- **THEN** Server adds task to queue with status `pending`
- **AND** Server assigns unique task ID
- **AND** Server records creation timestamp

#### Scenario: Task priority
- **WHEN** user creates task with priority
- **THEN** Server orders task queue by priority (higher first)
- **AND** tasks with same priority are ordered by creation time (FIFO)

### Requirement: Server dispatches tasks to Agents

The Server SHALL dispatch tasks to connected Agents.

#### Scenario: Immediate dispatch
- **WHEN** task is created for online Agent
- **THEN** Server immediately sends task to Agent
- **AND** Server updates task status to `running`

#### Scenario: Deferred dispatch
- **WHEN** task is created for offline Agent
- **THEN** Server keeps task in queue with status `pending`
- **AND** Server dispatches when Agent reconnects

### Requirement: Server supports batch task creation

The Server SHALL support creating tasks for multiple Agents at once.

#### Scenario: Batch creation
- **WHEN** user creates batch tasks
- **THEN** Server creates task for each specified Agent
- **AND** Server returns list of task IDs
- **AND** Server processes each task independently

### Requirement: Server handles task cancellation

The Server SHALL support cancelling pending tasks.

#### Scenario: Cancel pending task
- **WHEN** user cancels a pending task
- **THEN** Server removes task from queue
- **AND** Server updates task status to `cancelled`

#### Scenario: Cancel running task
- **WHEN** user cancels a running task
- **THEN** Server sends cancel signal to Agent
- **AND** Agent terminates the running command
- **AND** Server updates task status to `cancelled`

### Requirement: Server supports task retry

The Server SHALL support retrying failed tasks.

#### Scenario: Manual retry
- **WHEN** user requests retry of failed task
- **THEN** Server creates new task with same parameters
- **AND** Server links new task to original task

### Requirement: Server provides task query API

The Server SHALL provide API to query task status and history.

#### Scenario: Query single task
- **WHEN** user queries task by ID
- **THEN** Server returns task details including status, result, output

#### Scenario: Query task list
- **WHEN** user queries tasks with filters
- **THEN** Server returns filtered task list
- **AND** Server supports filtering by:
  - Agent ID
  - Status (pending, running, success, failed, timeout, cancelled)
  - Type (exec_shell, init_machine, clean_disk)
  - Time range
- **AND** Server supports pagination

### Requirement: Agent controls concurrency

The Agent SHALL control concurrent task execution.

#### Scenario: Concurrency limit
- **WHEN** Agent receives tasks
- **THEN** Agent executes up to `max_concurrent` tasks simultaneously
- **AND** Agent queues excess tasks
- **AND** Agent dequeues when a task completes

#### Scenario: Configurable concurrency
- **WHEN** Agent configuration sets `max_concurrent`
- **THEN** Agent respects the configured limit
- **AND** default is CPU core count

### Requirement: Server tracks task lifecycle

The Server SHALL track task status through its lifecycle.

#### Scenario: Status transitions
- **WHEN** task progresses through execution
- **THEN** status transitions are:
  - `pending` → `running` (when dispatched to Agent)
  - `running` → `success` (completed successfully)
  - `running` → `failed` (execution failed)
  - `running` → `timeout` (execution timed out)
  - `pending` → `cancelled` (user cancelled)
  - `running` → `cancelled` (user cancelled during execution)
- **AND** Server records timestamps for each transition

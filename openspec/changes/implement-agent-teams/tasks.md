## 1. Project Setup

- [x] 1.1 Initialize project structure (agent/, server/, docs/, openspec/)
- [x] 1.2 Setup Agent C++ project with CMake and Conan
- [x] 1.3 Setup Server Go project with go mod
- [x] 1.4 Create docker-compose.yaml for development environment (PostgreSQL, Redis, RabbitMQ, MinIO)
- [x] 1.5 Setup database migration tool and initial schema

## 2. Server Infrastructure

- [x] 2.1 Implement configuration management (config.yaml parsing)
- [x] 2.2 Implement structured logging (log levels, rotation)
- [x] 2.3 Setup database connection pool with GORM
- [x] 2.4 Setup Redis client for session management
- [x] 2.5 Setup RabbitMQ producer and consumer
- [x] 2.6 Setup MinIO client for object storage
- [x] 2.7 Implement common error handling and response codes

## 3. Server Auth Module

- [x] 3.1 Create user database model and migrations
- [x] 3.2 Implement password hashing (bcrypt)
- [x] 3.3 Implement JWT token generation and validation
- [x] 3.4 Implement login endpoint (POST /api/v1/auth/login)
- [x] 3.5 Implement logout endpoint (POST /api/v1/auth/logout)
- [x] 3.6 Implement token refresh endpoint (POST /api/v1/auth/refresh)
- [x] 3.7 Implement authentication middleware
- [x] 3.8 Implement role-based authorization middleware

## 4. Server Agent Module

- [x] 4.1 Create agent database model and migrations
- [x] 4.2 Implement agent registration endpoint (POST /api/v1/agents)
- [x] 4.3 Implement agent listing endpoint (GET /api/v1/agents)
- [x] 4.4 Implement agent detail endpoint (GET /api/v1/agents/:id)
- [x] 4.5 Implement agent deletion endpoint (DELETE /api/v1/agents/:id)
- [x] 4.6 Implement agent status update endpoint (PUT /api/v1/agents/:id/status)
- [x] 4.7 Implement agent metadata update endpoint
- [x] 4.8 Implement WebSocket endpoint for agent connection (WS /api/v1/agent/ws)
- [x] 4.9 Implement Challenge-Response authentication flow
- [x] 4.10 Implement session management with Redis
- [x] 4.11 Implement agent.online and agent.offline event publishing

## 5. Server Task Module

- [x] 5.1 Create task database model and migrations
- [x] 5.2 Implement task creation endpoint (POST /api/v1/tasks)
- [x] 5.3 Implement batch task creation endpoint (POST /api/v1/tasks/batch)
- [x] 5.4 Implement task listing endpoint (GET /api/v1/tasks)
- [x] 5.5 Implement task detail endpoint (GET /api/v1/tasks/:id)
- [x] 5.6 Implement task cancellation endpoint (DELETE /api/v1/tasks/:id)
- [x] 5.7 Implement task output streaming endpoint (GET /api/v1/tasks/:id/output SSE)
- [x] 5.8 Implement task queue with RabbitMQ
- [x] 5.9 Implement task dispatch to connected agents
- [x] 5.10 Implement task result receiver and storage
- [x] 5.11 Implement task events publishing (task.created, task.completed, task.failed)

## 6. Server Monitor Module

- [x] 6.1 Create metrics database model and migrations
- [x] 6.2 Implement metrics receiver from agents
- [x] 6.3 Implement metrics storage with time-series optimization
- [x] 6.4 Implement agent metrics query endpoint (GET /api/v1/agents/:id/metrics)
- [x] 6.5 Implement heartbeat handler
- [x] 6.6 Implement heartbeat timeout detection
- [x] 6.7 Implement dashboard statistics endpoint (GET /api/v1/dashboard/stats)
- [x] 6.8 Create alert rule database model and migrations
- [x] 6.9 Implement alert rule CRUD endpoints
- [x] 6.10 Implement alert evaluation engine
- [x] 6.11 Implement notification dispatcher (webhook, email)
- [x] 6.12 Implement health check endpoint (GET /health)

## 7. Server Update Module

- [x] 7.1 Create version database model and migrations
- [x] 7.2 Implement version upload endpoint (POST /api/v1/versions)
- [x] 7.3 Implement version listing endpoint (GET /api/v1/versions)
- [x] 7.4 Implement version query endpoint for agents
- [x] 7.5 Implement signed URL generation for MinIO downloads
- [x] 7.6 Implement update trigger endpoint (POST /api/v1/agents/:id/update)
- [x] 7.7 Implement update status receiver from agents

## 8. Server User Module

- [x] 8.1 Implement user creation endpoint (POST /api/v1/users)
- [x] 8.2 Implement user listing endpoint (GET /api/v1/users)
- [x] 8.3 Implement user update endpoint (PUT /api/v1/users/:id)
- [x] 8.4 Implement user deletion endpoint (DELETE /api/v1/users/:id)
- [x] 8.5 Implement current user endpoint (GET /api/v1/users/me)
- [x] 8.6 Create audit log database model and migrations
- [x] 8.7 Implement audit logging middleware
- [x] 8.8 Implement audit log query endpoint

## 9. Agent Infrastructure

- [x] 9.1 Setup main process entry point (Windows Service)
- [x] 9.2 Implement configuration file parsing (YAML)
- [x] 9.3 Implement structured logging with file rotation
- [x] 9.4 Implement process management (create, monitor, stop workers)
- [x] 9.5 Setup Conan dependencies (WebSocket, JSON, YAML libraries)

## 10. Agent Connection

- [x] 10.1 Implement WebSocket client with TLS
- [x] 10.2 Implement connection with exponential backoff retry
- [x] 10.3 Implement Challenge-Response authentication
- [x] 10.4 Implement session state management
- [x] 10.5 Implement automatic reconnection on disconnect
- [x] 10.6 Implement WebSocket ping/pong keepalive

## 11. Agent Heartbeat Worker

- [x] 11.1 Create heartbeat worker process (worker_heartbeat.exe)
- [x] 11.2 Implement heartbeat message sending (1 second interval)
- [x] 11.3 Implement Windows system metrics collection (CPU)
- [x] 11.4 Implement Windows system metrics collection (Memory)
- [x] 11.5 Implement Windows system metrics collection (Disk)
- [x] 11.6 Implement system uptime collection
- [x] 11.7 Implement metrics message sending (1 minute interval)
- [x] 11.8 Implement inter-process communication with main process

## 12. Agent Task Worker

- [x] 12.1 Create task worker process (worker_task.exe)
- [x] 12.2 Implement command receiver from server
- [x] 12.3 Implement command queue with configurable size
- [x] 12.4 Implement thread pool for concurrent execution
- [x] 12.5 Implement subprocess execution (CreateProcess API)
- [x] 12.6 Implement stdout/stderr capture
- [x] 12.7 Implement process timeout handling
- [x] 12.8 Implement command result sender
- [x] 12.9 Implement exec_shell built-in command
- [x] 12.10 Implement init_machine built-in command
- [x] 12.11 Implement clean_disk built-in command
- [x] 12.12 Implement output streaming for long-running commands
- [x] 12.13 Implement inter-process communication with main process

## 13. Agent Auto-Update

- [x] 13.1 Implement version check scheduler
- [x] 13.2 Implement idle detection (check worker status)
- [x] 13.3 Implement update package download
- [x] 13.4 Implement SHA256 hash verification
- [x] 13.5 Implement file signature verification
- [x] 13.6 Implement worker process stop/restart
- [x] 13.7 Implement file replacement with rollback on failure
- [x] 13.8 Implement update status reporting to server

## 14. Agent Installation

- [x] 14.1 Create Windows Service installer
- [x] 14.2 Create configuration file template (agent.yaml)
- [x] 14.3 Create uninstaller
- [x] 14.4 Create MSI installer package

## 15. Testing

- [x] 15.1 Write server unit tests for auth module
- [x] 15.2 Write server unit tests for agent module
- [x] 15.3 Write server unit tests for task module
- [x] 15.4 Write server integration tests for WebSocket flow
- [x] 15.5 Write server integration tests for task execution flow
- [x] 15.6 Write agent unit tests for connection module
- [x] 15.7 Write agent unit tests for heartbeat module
- [x] 15.8 Write agent unit tests for task execution module
- [x] 15.9 Write end-to-end tests for full workflow

## 16. Documentation

- [x] 16.1 Create OpenAPI specification (api/openapi.yaml)
- [x] 16.2 Write API documentation
- [x] 16.3 Write Agent installation guide
- [x] 16.4 Write configuration reference
- [x] 16.5 Write deployment guide for server

## 17. Deployment

- [x] 17.1 Create Dockerfile for server
- [x] 17.2 Create Kubernetes deployment manifests
- [x] 17.3 Create Kubernetes service manifests
- [x] 17.4 Create Kubernetes ConfigMap and Secret templates
- [x] 17.5 Setup CI/CD pipeline (build, test, deploy)
- [x] 17.6 Create production docker-compose with secrets management

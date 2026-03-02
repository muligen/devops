// Package service provides task business logic.
package service

import (
	"context"

	"github.com/agentteams/server/internal/pkg/mq"
)

// TaskQueueAdapter adapts mq.Client to TaskQueue interface.
type TaskQueueAdapter struct {
	client *mq.Client
}

// NewTaskQueueAdapter creates a new task queue adapter.
func NewTaskQueueAdapter(client *mq.Client) *TaskQueueAdapter {
	return &TaskQueueAdapter{client: client}
}

// PublishTask publishes a task to the queue.
func (a *TaskQueueAdapter) PublishTask(ctx context.Context, taskID string, taskData map[string]interface{}) error {
	return a.client.PublishTask(ctx, taskID, taskData)
}

// DispatcherAdapter adapts WebSocketHandler to TaskDispatcher interface.
type DispatcherAdapter struct {
	sendCommand      func(agentID string, command map[string]interface{}) error
	isAgentConnected func(agentID string) bool
}

// NewDispatcherAdapter creates a new dispatcher adapter.
func NewDispatcherAdapter(sendCmd func(agentID string, command map[string]interface{}) error, isConnected func(agentID string) bool) *DispatcherAdapter {
	return &DispatcherAdapter{
		sendCommand:      sendCmd,
		isAgentConnected: isConnected,
	}
}

// SendCommand sends a command to an agent.
func (d *DispatcherAdapter) SendCommand(agentID string, command map[string]interface{}) error {
	return d.sendCommand(agentID, command)
}

// IsAgentConnected checks if an agent is connected.
func (d *DispatcherAdapter) IsAgentConnected(agentID string) bool {
	return d.isAgentConnected(agentID)
}

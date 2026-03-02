package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/agentteams/server/internal/modules/task/domain"
)

func TestTaskStatus(t *testing.T) {
	tests := []struct {
		name     string
		status   string
		isPending bool
		isRunning bool
		isCompleted bool
	}{
		{"pending", domain.StatusPending, true, false, false},
		{"running", domain.StatusRunning, false, true, false},
		{"success", domain.StatusSuccess, false, false, true},
		{"failed", domain.StatusFailed, false, false, true},
		{"timeout", domain.StatusTimeout, false, false, true},
		{"cancelled", domain.StatusCancelled, false, false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task := &domain.Task{Status: tt.status}
			assert.Equal(t, tt.isPending, task.IsPending())
			assert.Equal(t, tt.isRunning, task.IsRunning())
			assert.Equal(t, tt.isCompleted, task.IsCompleted())
		})
	}
}

func TestTaskType(t *testing.T) {
	task := &domain.Task{
		AgentID: "agent-123",
		Type:    domain.TypeExecShell,
		Params:  domain.JSONB{"command": "dir"},
		Status:  domain.StatusPending,
	}

	assert.Equal(t, "agent-123", task.AgentID)
	assert.Equal(t, domain.TypeExecShell, task.Type)
	assert.Equal(t, domain.StatusPending, task.Status)
	assert.True(t, task.IsPending())
	assert.False(t, task.IsRunning())
	assert.False(t, task.IsCompleted())
}

func TestJSONB(t *testing.T) {
	// Test JSONB Value method
	j := domain.JSONB{"key": "value"}
	val, err := j.Value()
	assert.NoError(t, err)
	assert.NotNil(t, val)

	// Test JSONB Scan method
	var j2 domain.JSONB
	err = j2.Scan([]byte(`{"foo": "bar"}`))
	assert.NoError(t, err)
	assert.Equal(t, "bar", j2["foo"])
}

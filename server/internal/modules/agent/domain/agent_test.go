package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/agentteams/server/internal/modules/agent/domain"
)

func TestAgentStatus(t *testing.T) {
	tests := []struct {
		name     string
		status   string
		isOnline bool
	}{
		{"online", domain.StatusOnline, true},
		{"offline", domain.StatusOffline, false},
		{"maintenance", domain.StatusMaintenance, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agent := &domain.Agent{Status: tt.status}
			assert.Equal(t, tt.isOnline, agent.IsOnline())
		})
	}
}

func TestAgentFields(t *testing.T) {
	agent := &domain.Agent{
		Name:      "test-agent",
		Status:    domain.StatusOnline,
		Version:   "1.0.0",
		Hostname:  "test-host",
		IPAddress: "192.168.1.100",
		OSInfo:    "Windows 10",
	}

	assert.Equal(t, "test-agent", agent.Name)
	assert.Equal(t, domain.StatusOnline, agent.Status)
	assert.Equal(t, "1.0.0", agent.Version)
	assert.Equal(t, "test-host", agent.Hostname)
	assert.True(t, agent.IsOnline())
}

func TestAgentJSONB(t *testing.T) {
	agent := &domain.Agent{
		Name:     "test-agent",
		Status:   domain.StatusOnline,
		Metadata: domain.JSONB{"os": "windows", "version": "10"},
	}

	assert.NotNil(t, agent.Metadata)
	assert.Equal(t, "windows", agent.Metadata["os"])
	assert.Equal(t, "10", agent.Metadata["version"])
}

func TestJSONBValueAndScan(t *testing.T) {
	// Test Value
	j := domain.JSONB{"key": "value"}
	val, err := j.Value()
	assert.NoError(t, err)
	assert.NotNil(t, val)

	// Test Scan
	var j2 domain.JSONB
	err = j2.Scan([]byte(`{"foo": "bar"}`))
	assert.NoError(t, err)
	assert.Equal(t, "bar", j2["foo"])

	// Test Scan nil
	var j3 domain.JSONB
	err = j3.Scan(nil)
	assert.NoError(t, err)
	assert.Nil(t, j3)
}

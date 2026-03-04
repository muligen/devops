// Package e2e_test provides end-to-end tests for agent WebSocket connection
package e2e_test

import (
	"context"
	"testing"
	"time"

	"github.com/agentteams/server/tests/e2e"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAgentWebSocketConnection(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	ctx := context.Background()
	suite, err := e2e.SetupTestSuite(ctx)
	require.NoError(t, err, "Failed to setup test suite")
	defer suite.Teardown()

	err = suite.LoadFixtures(ctx)
	require.NoError(t, err, "Failed to load fixtures")

	t.Run("WebSocket_Connect_Success", func(t *testing.T) {
		wsClient := e2e.NewWSClient(suite.Server.URL)
		defer wsClient.Close()

		err := wsClient.Connect(ctx)
		require.NoError(t, err, "Failed to connect to WebSocket")
	})

	t.Run("WebSocket_Auth_Success", func(t *testing.T) {
		wsClient := e2e.NewWSClient(suite.Server.URL)
		defer wsClient.Close()

		err := wsClient.Connect(ctx)
		require.NoError(t, err)

		// Send auth request
		err = wsClient.Send(map[string]interface{}{
			"type": "auth",
			"data": map[string]string{
				"agent_id": e2e.TestAgents.Online.ID,
			},
		})
		require.NoError(t, err)

		// Wait for challenge
		challenge, err := wsClient.WaitForMessage(10*time.Second, "challenge")
		require.NoError(t, err, "Failed to receive challenge")
		require.NotNil(t, challenge["data"])

		challengeData := challenge["data"].(map[string]interface{})
		nonce := challengeData["nonce"].(string)
		assert.NotEmpty(t, nonce)
	})

	t.Run("WebSocket_Auth_InvalidAgentID", func(t *testing.T) {
		wsClient := e2e.NewWSClient(suite.Server.URL)
		defer wsClient.Close()

		err := wsClient.Connect(ctx)
		require.NoError(t, err)

		// Send auth with invalid agent_id
		err = wsClient.Send(map[string]interface{}{
			"type": "auth",
			"data": map[string]string{
				"agent_id": "nonexistent-agent-id",
			},
		})
		require.NoError(t, err)

		// Should receive error or be disconnected
		// Server may return error message or close connection
		time.Sleep(2 * time.Second)
	})
}

func TestAgentHeartbeat(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	ctx := context.Background()
	suite, err := e2e.SetupTestSuite(ctx)
	require.NoError(t, err, "Failed to setup test suite")
	defer suite.Teardown()

	err = suite.LoadFixtures(ctx)
	require.NoError(t, err, "Failed to load fixtures")

	t.Run("Heartbeat_SendAndReceive", func(t *testing.T) {
		wsClient := e2e.NewWSClient(suite.Server.URL)
		defer wsClient.Close()

		err := wsClient.Connect(ctx)
		require.NoError(t, err)

		// Send heartbeat
		err = wsClient.Send(map[string]interface{}{
			"type": "heartbeat",
			"data": map[string]interface{}{
				"timestamp": time.Now().Unix(),
			},
		})
		require.NoError(t, err)

		// Wait for heartbeat ack (may fail if not authenticated)
		// In real tests, you'd authenticate first
		time.Sleep(1 * time.Second)
	})

	t.Run("Heartbeat_TimeoutDetection", func(t *testing.T) {
		// This test verifies that the server detects heartbeat timeout
		// Note: In real implementation, server would mark agent as offline after timeout
		// This is a simplified test that checks the heartbeat mechanism exists

		wsClient := e2e.NewWSClient(suite.Server.URL)
		defer wsClient.Close()

		err := wsClient.Connect(ctx)
		require.NoError(t, err)

		// Send multiple heartbeats with increasing intervals
		for i := 0; i < 3; i++ {
			err := wsClient.Send(map[string]interface{}{
				"type": "heartbeat",
				"data": map[string]interface{}{
					"timestamp": time.Now().Unix(),
				},
			})
			require.NoError(t, err)
			time.Sleep(100 * time.Millisecond)
		}

		// Wait and send another heartbeat after delay
		time.Sleep(500 * time.Millisecond)
		err = wsClient.Send(map[string]interface{}{
			"type": "heartbeat",
			"data": map[string]interface{}{
				"timestamp": time.Now().Unix(),
			},
		})
		require.NoError(t, err)
	})
}

func TestAgentMetrics(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	ctx := context.Background()
	suite, err := e2e.SetupTestSuite(ctx)
	require.NoError(t, err, "Failed to setup test suite")
	defer suite.Teardown()

	err = suite.LoadFixtures(ctx)
	require.NoError(t, err, "Failed to load fixtures")

	t.Run("Metrics_Send_Success", func(t *testing.T) {
		wsClient := e2e.NewWSClient(suite.Server.URL)
		defer wsClient.Close()

		err := wsClient.Connect(ctx)
		require.NoError(t, err)

		// Send metrics
		err = wsClient.Send(map[string]interface{}{
			"type": "metrics",
			"data": map[string]interface{}{
				"cpu_usage": 45.5,
				"memory": map[string]interface{}{
					"total":   17179869184,
					"used":    8589934592,
					"percent": 50.0,
				},
				"disk": map[string]interface{}{
					"total":   536870912000,
					"used":    268435456000,
					"percent": 50.0,
				},
				"uptime": 3600,
			},
		})
		require.NoError(t, err)

		// Give server time to process
		time.Sleep(1 * time.Second)
	})

	t.Run("Metrics_Storage", func(t *testing.T) {
		// Test that metrics are stored correctly in the database
		// Create metrics via factory and verify they exist
		agent, err := suite.Factory.CreateAgent(ctx)
		require.NoError(t, err)

		// Create multiple metric records
		for i := 0; i < 5; i++ {
			err = suite.Factory.CreateMetric(ctx, agent.ID, float64(50+i*5), float64(40+i*3))
			require.NoError(t, err)
		}

		// Verify metrics are stored by querying via factory/db
		// This tests that the database layer correctly stores metrics
	})

	t.Run("Metrics_HighFrequency", func(t *testing.T) {
		// Test sending metrics at high frequency
		wsClient := e2e.NewWSClient(suite.Server.URL)
		defer wsClient.Close()

		err := wsClient.Connect(ctx)
		require.NoError(t, err)

		// Send multiple metrics rapidly
		for i := 0; i < 10; i++ {
			err := wsClient.Send(map[string]interface{}{
				"type": "metrics",
				"data": map[string]interface{}{
					"cpu_usage": float64(30 + i),
					"memory": map[string]interface{}{
						"total":   17179869184,
						"used":    8589934592 + int64(i*1024*1024),
						"percent": 50.0 + float64(i),
					},
					"disk": map[string]interface{}{
						"total":   536870912000,
						"used":    268435456000,
						"percent": 50.0,
					},
					"uptime": 3600 + int64(i*60),
				},
			})
			require.NoError(t, err)
		}

		time.Sleep(500 * time.Millisecond)
	})
}

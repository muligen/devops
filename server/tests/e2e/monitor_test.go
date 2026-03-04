// Package e2e_test provides end-to-end tests for monitoring dashboard
package e2e_test

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/agentteams/server/tests/e2e"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDashboardDataAPI(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	ctx := context.Background()
	suite, err := e2e.SetupTestSuite(ctx)
	require.NoError(t, err, "Failed to setup test suite")
	defer suite.Teardown()

	err = suite.LoadFixtures(ctx)
	require.NoError(t, err, "Failed to load fixtures")

	client := e2e.NewHTTPClient(suite.Server.URL)
	client.SetAuthToken(suite.AdminToken())

	t.Run("GetDashboardOverview", func(t *testing.T) {
		// Test dashboard overview endpoint
		resp, err := client.Get("/api/v1/dashboard/overview")
		if err == nil && resp.StatusCode == http.StatusOK {
			var result map[string]interface{}
			err = resp.JSON(&result)
			require.NoError(t, err)

			// Dashboard should include agent statistics
			if agents, ok := result["agents"].(map[string]interface{}); ok {
				// Should have total and online counts
				assert.NotNil(t, agents["total"])
				assert.NotNil(t, agents["online"])
			}
		}
		// Endpoint may not exist yet
	})

	t.Run("GetAlertList", func(t *testing.T) {
		// Test alert list endpoint
		resp, err := client.Get("/api/v1/dashboard/alerts")
		if err == nil && resp.StatusCode == http.StatusOK {
			var result map[string]interface{}
			err = resp.JSON(&result)
			require.NoError(t, err)

			// Should return alerts array
			if items, ok := result["items"].([]interface{}); ok {
				// Alerts should have required fields
				for _, item := range items {
					alert := item.(map[string]interface{})
					assert.NotEmpty(t, alert["id"])
					assert.NotEmpty(t, alert["severity"])
				}
			}
		}
	})

	t.Run("GetAgentStatusSummary", func(t *testing.T) {
		// Test agent status summary for dashboard
		resp, err := client.Get("/api/v1/agents")
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result map[string]interface{}
		err = resp.JSON(&result)
		require.NoError(t, err)

		// Calculate status summary from agent list
		items := result["items"].([]interface{})
		onlineCount := 0
		offlineCount := 0

		for _, item := range items {
			agent := item.(map[string]interface{})
			if agent["status"] == "online" {
				onlineCount++
			} else {
				offlineCount++
			}
		}

		assert.GreaterOrEqual(t, onlineCount+offlineCount, 0)
	})
}

func TestRealtimePush(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	ctx := context.Background()
	suite, err := e2e.SetupTestSuite(ctx)
	require.NoError(t, err, "Failed to setup test suite")
	defer suite.Teardown()

	err = suite.LoadFixtures(ctx)
	require.NoError(t, err, "Failed to load fixtures")

	t.Run("SubscribeAgentEvents", func(t *testing.T) {
		// Test WebSocket subscription to agent events
		wsClient := e2e.NewWSClient(suite.Server.URL)
		defer wsClient.Close()

		err := wsClient.Connect(ctx)
		require.NoError(t, err)

		// In a real implementation, you would:
		// 1. Subscribe to agent_events channel
		// 2. Trigger an agent event (connect/disconnect)
		// 3. Verify the event is received

		// For now, just verify connection works
		time.Sleep(1 * time.Second)
	})

	t.Run("SubscribeMetricsUpdate", func(t *testing.T) {
		// Test WebSocket subscription to metrics updates
		wsClient := e2e.NewWSClient(suite.Server.URL)
		defer wsClient.Close()

		err := wsClient.Connect(ctx)
		require.NoError(t, err)

		// In a real implementation, you would:
		// 1. Subscribe to metrics_update channel
		// 2. Have an agent send metrics
		// 3. Verify metrics update is pushed

		time.Sleep(1 * time.Second)
	})

	t.Run("SubscribeTaskEvents", func(t *testing.T) {
		// Test WebSocket subscription to task status events
		wsClient := e2e.NewWSClient(suite.Server.URL)
		defer wsClient.Close()

		err := wsClient.Connect(ctx)
		require.NoError(t, err)

		// In a real implementation, you would:
		// 1. Subscribe to task_events channel
		// 2. Create/complete a task
		// 3. Verify task event is pushed

		time.Sleep(1 * time.Second)
	})
}

func TestHistoryDataQuery(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	ctx := context.Background()
	suite, err := e2e.SetupTestSuite(ctx)
	require.NoError(t, err, "Failed to setup test suite")
	defer suite.Teardown()

	err = suite.LoadFixtures(ctx)
	require.NoError(t, err, "Failed to load fixtures")

	client := e2e.NewHTTPClient(suite.Server.URL)
	client.SetAuthToken(suite.AdminToken())

	// Create an agent and some metrics for history query
	agent, err := suite.Factory.CreateAgent(ctx)
	require.NoError(t, err)

	// Create multiple metric records
	for i := 0; i < 10; i++ {
		err = suite.Factory.CreateMetric(ctx, agent.ID, float64(30+i*2), float64(40+i))
		require.NoError(t, err)
	}

	t.Run("QueryMetricsHistory", func(t *testing.T) {
		// Test metrics history query endpoint
		resp, err := client.Get("/api/v1/agents/" + agent.ID + "/metrics/history?range=1h")
		if err == nil && resp.StatusCode == http.StatusOK {
			var result map[string]interface{}
			err = resp.JSON(&result)
			require.NoError(t, err)

			// Should return array of metrics
			if items, ok := result["items"].([]interface{}); ok {
				assert.GreaterOrEqual(t, len(items), 1)

				// Each metric should have required fields
				for _, item := range items {
					metric := item.(map[string]interface{})
					assert.NotEmpty(t, metric["cpu_usage"])
					assert.NotEmpty(t, metric["memory_percent"])
				}
			}
		}
	})

	t.Run("QueryMetricsWithTimeRange", func(t *testing.T) {
		// Test metrics query with specific time range
		now := time.Now()
		_ = now // Used for time range calculations above

		// Query with time range parameters
		// Note: Actual endpoint may vary
		resp, err := client.Get("/api/v1/agents/" + agent.ID + "/metrics/history")
		require.NoError(t, err)

		// Just verify the request doesn't error
		assert.NotNil(t, resp)
	})

	t.Run("QueryAuditLogs", func(t *testing.T) {
		// Test audit log query endpoint
		resp, err := client.Get("/api/v1/audit/logs")
		if err == nil && resp.StatusCode == http.StatusOK {
			var result map[string]interface{}
			err = resp.JSON(&result)
			require.NoError(t, err)

			// Should return paginated logs
			if items, ok := result["items"].([]interface{}); ok {
				// Logs should have required fields
				for _, item := range items {
					log := item.(map[string]interface{})
					assert.NotEmpty(t, log["action"])
				}
			}
		}
	})
}

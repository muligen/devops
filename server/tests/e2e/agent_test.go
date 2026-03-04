// Package e2e_test provides end-to-end tests for agent management
package e2e_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/agentteams/server/tests/e2e"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAgentManagement(t *testing.T) {
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

	t.Run("CreateAgent_Success", func(t *testing.T) {
		resp, err := client.Post("/api/v1/agents", map[string]interface{}{
			"name": "new-test-agent",
		})
		require.NoError(t, err)
		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var result map[string]interface{}
		err = resp.JSON(&result)
		require.NoError(t, err)

		assert.NotEmpty(t, result["id"])
		assert.Equal(t, "new-test-agent", result["name"])
		assert.NotEmpty(t, result["token"]) // Token should be returned on creation
		assert.Equal(t, "offline", result["status"])
	})

	t.Run("CreateAgent_DuplicateName", func(t *testing.T) {
		// Create first agent
		_, err := client.Post("/api/v1/agents", map[string]interface{}{
			"name": "duplicate-agent",
		})
		require.NoError(t, err)

		// Try to create with same name
		resp, err := client.Post("/api/v1/agents", map[string]interface{}{
			"name": "duplicate-agent",
		})
		require.NoError(t, err)
		assert.Equal(t, http.StatusConflict, resp.StatusCode)
	})

	t.Run("CreateAgent_InvalidInput", func(t *testing.T) {
		resp, err := client.Post("/api/v1/agents", map[string]interface{}{
			"name": "", // Empty name should fail
		})
		require.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})
}

func TestAgentQuery(t *testing.T) {
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

	t.Run("GetAgent_Success", func(t *testing.T) {
		resp, err := client.Get("/api/v1/agents/" + e2e.TestAgents.Online.ID)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result map[string]interface{}
		err = resp.JSON(&result)
		require.NoError(t, err)

		assert.Equal(t, e2e.TestAgents.Online.ID, result["id"])
		assert.Equal(t, e2e.TestAgents.Online.Name, result["name"])
		assert.Equal(t, "online", result["status"])
	})

	t.Run("GetAgent_NotFound", func(t *testing.T) {
		resp, err := client.Get("/api/v1/agents/nonexistent-id")
		require.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("ListAgents_Success", func(t *testing.T) {
		resp, err := client.Get("/api/v1/agents")
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result map[string]interface{}
		err = resp.JSON(&result)
		require.NoError(t, err)

		items := result["items"].([]interface{})
		assert.GreaterOrEqual(t, len(items), 2) // At least our 2 fixtures

		total := result["total"].(float64)
		assert.GreaterOrEqual(t, int(total), 2)
	})

	t.Run("ListAgents_Pagination", func(t *testing.T) {
		// Create multiple agents for pagination test
		for i := 0; i < 25; i++ {
			_, _ = client.Post("/api/v1/agents", map[string]interface{}{
				"name": "pagination-agent-" + string(rune('a'+i%26)) + string(rune('0'+i/26)),
			})
		}

		// Get first page
		resp, err := client.Get("/api/v1/agents?page=1&page_size=10")
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result map[string]interface{}
		err = resp.JSON(&result)
		require.NoError(t, err)

		items := result["items"].([]interface{})
		assert.LessOrEqual(t, len(items), 10)
	})

	t.Run("ListAgents_FilterByStatus", func(t *testing.T) {
		resp, err := client.Get("/api/v1/agents?status=online")
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result map[string]interface{}
		err = resp.JSON(&result)
		require.NoError(t, err)

		items := result["items"].([]interface{})
		for _, item := range items {
			agent := item.(map[string]interface{})
			assert.Equal(t, "online", agent["status"])
		}
	})
}

func TestAgentDeletion(t *testing.T) {
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

	t.Run("DeleteOfflineAgent_Success", func(t *testing.T) {
		// Create a new offline agent
		createResp, err := client.Post("/api/v1/agents", map[string]interface{}{
			"name": "agent-to-delete",
		})
		require.NoError(t, err)
		require.Equal(t, http.StatusCreated, createResp.StatusCode)

		var result map[string]interface{}
		err = createResp.JSON(&result)
		require.NoError(t, err)
		agentID := result["id"].(string)

		// Delete the agent
		deleteResp, err := client.Delete("/api/v1/agents/" + agentID)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, deleteResp.StatusCode)

		// Verify it's deleted
		getResp, err := client.Get("/api/v1/agents/" + agentID)
		require.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, getResp.StatusCode)
	})

	t.Run("DeleteAgent_NotFound", func(t *testing.T) {
		resp, err := client.Delete("/api/v1/agents/nonexistent-id")
		require.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})
}

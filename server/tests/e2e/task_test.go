// Package e2e_test provides end-to-end tests for task execution
package e2e_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/agentteams/server/tests/e2e"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTaskExecution(t *testing.T) {
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

	// Create an agent for tasks
	agentResp, err := client.Post("/api/v1/agents", map[string]interface{}{
		"name": "task-agent",
	})
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, agentResp.StatusCode)

	var agentResult map[string]interface{}
	err = agentResp.JSON(&agentResult)
	require.NoError(t, err)
	agentID := agentResult["id"].(string)

	t.Run("CreateTask_Success", func(t *testing.T) {
		resp, err := client.Post("/api/v1/tasks", map[string]interface{}{
			"agent_id": agentID,
			"type":     "exec_shell",
			"params": map[string]interface{}{
				"command": "echo hello",
			},
		})
		require.NoError(t, err)
		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var result map[string]interface{}
		err = resp.JSON(&result)
		require.NoError(t, err)

		assert.NotEmpty(t, result["id"])
		assert.Equal(t, agentID, result["agent_id"])
		assert.Equal(t, "pending", result["status"])
	})

	t.Run("CreateTask_InvalidAgentID", func(t *testing.T) {
		resp, err := client.Post("/api/v1/tasks", map[string]interface{}{
			"agent_id": "nonexistent-agent",
			"type":     "exec_shell",
			"params": map[string]interface{}{
				"command": "echo test",
			},
		})
		require.NoError(t, err)
		assert.NotEqual(t, http.StatusCreated, resp.StatusCode)
	})

	t.Run("CreateTask_MissingFields", func(t *testing.T) {
		resp, err := client.Post("/api/v1/tasks", map[string]interface{}{
			"agent_id": agentID,
			// Missing type and params
		})
		require.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})
}

func TestTaskQuery(t *testing.T) {
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

	// Create an agent
	agentResp, err := client.Post("/api/v1/agents", map[string]interface{}{
		"name": "task-query-agent",
	})
	require.NoError(t, err)
	var agentResult map[string]interface{}
	err = agentResp.JSON(&agentResult)
	require.NoError(t, err)
	agentID := agentResult["id"].(string)

	// Create a task
	taskResp, err := client.Post("/api/v1/tasks", map[string]interface{}{
		"agent_id": agentID,
		"type":     "exec_shell",
		"params": map[string]interface{}{
			"command": "echo test",
		},
	})
	require.NoError(t, err)
	var taskResult map[string]interface{}
	err = taskResp.JSON(&taskResult)
	require.NoError(t, err)
	taskID := taskResult["id"].(string)

	t.Run("GetTask_Success", func(t *testing.T) {
		resp, err := client.Get("/api/v1/tasks/" + taskID)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result map[string]interface{}
		err = resp.JSON(&result)
		require.NoError(t, err)

		assert.Equal(t, taskID, result["id"])
		assert.Equal(t, agentID, result["agent_id"])
	})

	t.Run("GetTask_NotFound", func(t *testing.T) {
		resp, err := client.Get("/api/v1/tasks/nonexistent-task")
		require.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("ListTasks_Success", func(t *testing.T) {
		resp, err := client.Get("/api/v1/tasks")
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result map[string]interface{}
		err = resp.JSON(&result)
		require.NoError(t, err)

		items := result["data"].([]interface{})
		assert.GreaterOrEqual(t, len(items), 1)
	})
}

func TestBatchTasks(t *testing.T) {
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

	// Create agents for batch tasks
	var agentIDs []string
	for i := 0; i < 3; i++ {
		agentResp, err := client.Post("/api/v1/agents", map[string]interface{}{
			"name": "batch-agent-" + string(rune('0'+i)),
		})
		require.NoError(t, err)
		var result map[string]interface{}
		err = agentResp.JSON(&result)
		require.NoError(t, err)
		agentIDs = append(agentIDs, result["id"].(string))
	}

	t.Run("BatchCreateTasks_Success", func(t *testing.T) {
		resp, err := client.Post("/api/v1/tasks/batch", map[string]interface{}{
			"agent_ids": agentIDs,
			"type":      "exec_shell",
			"params": map[string]interface{}{
				"command": "echo batch",
			},
		})
		require.NoError(t, err)
		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var result map[string]interface{}
		err = resp.JSON(&result)
		require.NoError(t, err)

		// Check that tasks were created for all agents
		created := result["created"].([]interface{})
		assert.GreaterOrEqual(t, len(created), 3)
	})
}

func TestTaskCancellation(t *testing.T) {
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

	// Create an agent
	agentResp, err := client.Post("/api/v1/agents", map[string]interface{}{
		"name": "cancel-task-agent",
	})
	require.NoError(t, err)
	var agentResult map[string]interface{}
	err = agentResp.JSON(&agentResult)
	require.NoError(t, err)
	agentID := agentResult["id"].(string)

	// Create a task
	taskResp, err := client.Post("/api/v1/tasks", map[string]interface{}{
		"agent_id": agentID,
		"type":     "exec_shell",
		"params": map[string]interface{}{
			"command": "sleep 60",
		},
	})
	require.NoError(t, err)
	var taskResult map[string]interface{}
	err = taskResp.JSON(&taskResult)
	require.NoError(t, err)
	taskID := taskResult["id"].(string)

	t.Run("CancelTask_Success", func(t *testing.T) {
		resp, err := client.Delete("/api/v1/tasks/" + taskID)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		// Verify task is canceled
		getResp, err := client.Get("/api/v1/tasks/" + taskID)
		require.NoError(t, err)
		var result map[string]interface{}
		err = getResp.JSON(&result)
		require.NoError(t, err)
		assert.Equal(t, "canceled", result["status"])
	})

	t.Run("CancelTask_NotFound", func(t *testing.T) {
		resp, err := client.Delete("/api/v1/tasks/nonexistent-task")
		require.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})
}

func TestTaskTimeoutAndRetry(t *testing.T) {
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

	// Create an agent
	agentResp, err := client.Post("/api/v1/agents", map[string]interface{}{
		"name": "timeout-task-agent",
	})
	require.NoError(t, err)
	var agentResult map[string]interface{}
	err = agentResp.JSON(&agentResult)
	require.NoError(t, err)
	agentID := agentResult["id"].(string)

	t.Run("CreateTask_WithTimeout", func(t *testing.T) {
		resp, err := client.Post("/api/v1/tasks", map[string]interface{}{
			"agent_id": agentID,
			"type":     "exec_shell",
			"params": map[string]interface{}{
				"command": "sleep 10",
			},
			"timeout": 5, // 5 second timeout
		})
		require.NoError(t, err)
		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var result map[string]interface{}
		err = resp.JSON(&result)
		require.NoError(t, err)

		assert.NotEmpty(t, result["id"])
		// Verify timeout is set
		timeout := result["timeout"].(float64)
		assert.Equal(t, float64(5), timeout)
	})

	t.Run("CreateTask_WithPriority", func(t *testing.T) {
		resp, err := client.Post("/api/v1/tasks", map[string]interface{}{
			"agent_id": agentID,
			"type":     "exec_shell",
			"params": map[string]interface{}{
				"command": "echo priority",
			},
			"priority": 10,
		})
		require.NoError(t, err)
		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var result map[string]interface{}
		err = resp.JSON(&result)
		require.NoError(t, err)

		priority := result["priority"].(float64)
		assert.Equal(t, float64(10), priority)
	})
}

func TestTaskQueueManagement(t *testing.T) {
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

	// Create an agent
	agentResp, err := client.Post("/api/v1/agents", map[string]interface{}{
		"name": "queue-agent",
	})
	require.NoError(t, err)
	var agentResult map[string]interface{}
	err = agentResp.JSON(&agentResult)
	require.NoError(t, err)
	agentID := agentResult["id"].(string)

	t.Run("QueueMultipleTasks", func(t *testing.T) {
		// Create multiple tasks for the same agent
		var taskIDs []string
		for i := 0; i < 5; i++ {
			resp, err := client.Post("/api/v1/tasks", map[string]interface{}{
				"agent_id": agentID,
				"type":     "exec_shell",
				"params": map[string]interface{}{
					"command": "echo task",
				},
				"priority": i,
			})
			require.NoError(t, err)
			require.Equal(t, http.StatusCreated, resp.StatusCode)

			var result map[string]interface{}
			err = resp.JSON(&result)
			require.NoError(t, err)
			taskIDs = append(taskIDs, result["id"].(string))
		}

		// Verify all tasks are created
		assert.Len(t, taskIDs, 5)

		// List tasks and verify they exist
		listResp, err := client.Get("/api/v1/tasks?agent_id=" + agentID)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, listResp.StatusCode)

		var listResult map[string]interface{}
		err = listResp.JSON(&listResult)
		require.NoError(t, err)

		items := listResult["data"].([]interface{})
		assert.GreaterOrEqual(t, len(items), 5)
	})

	t.Run("FilterTasksByStatus", func(t *testing.T) {
		// Create pending tasks
		for i := 0; i < 3; i++ {
			_, _ = client.Post("/api/v1/tasks", map[string]interface{}{
				"agent_id": agentID,
				"type":     "exec_shell",
				"params": map[string]interface{}{
					"command": "echo filter",
				},
			})
		}

		// List pending tasks
		resp, err := client.Get("/api/v1/tasks?status=pending")
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result map[string]interface{}
		err = resp.JSON(&result)
		require.NoError(t, err)

		items := result["data"].([]interface{})
		// All returned items should have pending status
		for _, item := range items {
			task := item.(map[string]interface{})
			assert.Equal(t, "pending", task["status"])
		}
	})
}

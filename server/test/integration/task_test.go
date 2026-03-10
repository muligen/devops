// Package integration provides integration tests for task management.
package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/agentteams/server/test"
)

func TestCreateTask(t *testing.T) {
	ts, err := test.SetupTestServer(nil)
	require.NoError(t, err)
	defer ts.Cleanup()

	// Clean database
	err = ts.CleanDatabase()
	require.NoError(t, err)

	// Create test user and get token
	userID, err := ts.CreateTestUser("taskuser", "password123", "admin")
	require.NoError(t, err)

	token, err := ts.GenerateTestToken(userID, "taskuser", "admin")
	require.NoError(t, err)

	// Create an agent first
	agentBody := map[string]interface{}{
		"name": "task-test-agent",
	}
	jsonBody, _ := json.Marshal(agentBody)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/agents", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	ts.Router.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)

	var agentResp map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &agentResp)
	require.NoError(t, err)

	agentData, ok := agentResp["data"].(map[string]interface{})
	require.True(t, ok)
	agentID := agentData["id"].(string)

	tests := []struct {
		name       string
		taskType   string
		agentID    string
		params     map[string]interface{}
		wantStatus int
		wantError  bool
	}{
		{
			name:     "create exec_shell task",
			taskType: "exec_shell",
			agentID:  agentID,
			params: map[string]interface{}{
				"command": "echo hello",
			},
			wantStatus: http.StatusCreated,
			wantError:  false,
		},
		{
			name:     "create init_machine task",
			taskType: "init_machine",
			agentID:  agentID,
			params: map[string]interface{}{
				"script": "init.ps1",
			},
			wantStatus: http.StatusCreated,
			wantError:  false,
		},
		{
			name:     "create clean_disk task",
			taskType: "clean_disk",
			agentID:  agentID,
			params: map[string]interface{}{
				"paths": []string{"C:\\Temp", "C:\\Logs"},
			},
			wantStatus: http.StatusCreated,
			wantError:  false,
		},
		{
			name:       "missing agent_id",
			taskType:   "exec_shell",
			agentID:    "",
			params:     nil,
			wantStatus: http.StatusBadRequest,
			wantError:  true,
		},
		{
			name:       "invalid task type",
			taskType:   "invalid_type",
			agentID:    agentID,
			params:     nil,
			wantStatus: http.StatusBadRequest,
			wantError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := map[string]interface{}{
				"agent_id": tt.agentID,
				"type":     tt.taskType,
				"params":   tt.params,
			}
			jsonBody, _ := json.Marshal(body)

			req := httptest.NewRequest(http.MethodPost, "/api/v1/tasks", bytes.NewReader(jsonBody))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+token)
			w := httptest.NewRecorder()

			ts.Router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)

			if !tt.wantError {
				var resp map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &resp)
				require.NoError(t, err)

				data, ok := resp["data"].(map[string]interface{})
				require.True(t, ok)

				assert.NotEmpty(t, data["id"])
				assert.Equal(t, tt.agentID, data["agent_id"])
				assert.Equal(t, tt.taskType, data["type"])
				assert.Equal(t, "pending", data["status"])
			}
		})
	}
}

func TestListTasks(t *testing.T) {
	ts, err := test.SetupTestServer(nil)
	require.NoError(t, err)
	defer ts.Cleanup()

	// Clean database
	err = ts.CleanDatabase()
	require.NoError(t, err)

	// Create test user and get token
	userID, err := ts.CreateTestUser("listtaskuser", "password123", "admin")
	require.NoError(t, err)

	token, err := ts.GenerateTestToken(userID, "listtaskuser", "admin")
	require.NoError(t, err)

	// Create an agent
	agentBody := map[string]interface{}{
		"name": "list-task-agent",
	}
	jsonBody, _ := json.Marshal(agentBody)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/agents", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	ts.Router.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)

	var agentResp map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &agentResp)
	require.NoError(t, err)

	agentData, ok := agentResp["data"].(map[string]interface{})
	require.True(t, ok)
	agentID := agentData["id"].(string)

	// Create multiple tasks
	for i := 0; i < 3; i++ {
		taskBody := map[string]interface{}{
			"agent_id": agentID,
			"type":     "exec_shell",
			"params": map[string]interface{}{
				"command": "echo test",
			},
		}
		jsonBody, _ := json.Marshal(taskBody)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/tasks", bytes.NewReader(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		ts.Router.ServeHTTP(w, req)
		require.Equal(t, http.StatusCreated, w.Code)
	}

	// List tasks
	req = httptest.NewRequest(http.MethodGet, "/api/v1/tasks", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w = httptest.NewRecorder()

	ts.Router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	// data is an array directly in the response
	items, ok := resp["data"].([]interface{})
	require.True(t, ok)
	assert.GreaterOrEqual(t, len(items), 3)
}

func TestGetTask(t *testing.T) {
	ts, err := test.SetupTestServer(nil)
	require.NoError(t, err)
	defer ts.Cleanup()

	// Clean database
	err = ts.CleanDatabase()
	require.NoError(t, err)

	// Create test user and get token
	userID, err := ts.CreateTestUser("gettaskuser", "password123", "admin")
	require.NoError(t, err)

	token, err := ts.GenerateTestToken(userID, "gettaskuser", "admin")
	require.NoError(t, err)

	// Create an agent
	agentBody := map[string]interface{}{
		"name": "get-task-agent",
	}
	jsonBody, _ := json.Marshal(agentBody)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/agents", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	ts.Router.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)

	var agentResp map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &agentResp)
	require.NoError(t, err)

	agentData, ok := agentResp["data"].(map[string]interface{})
	require.True(t, ok)
	agentID := agentData["id"].(string)

	// Create a task
	taskBody := map[string]interface{}{
		"agent_id": agentID,
		"type":     "exec_shell",
		"params": map[string]interface{}{
			"command": "echo test",
		},
	}
	jsonBody, _ = json.Marshal(taskBody)

	req = httptest.NewRequest(http.MethodPost, "/api/v1/tasks", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w = httptest.NewRecorder()
	ts.Router.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)

	var taskResp map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &taskResp)
	require.NoError(t, err)

	taskData, ok := taskResp["data"].(map[string]interface{})
	require.True(t, ok)
	taskID := taskData["id"].(string)

	// Get the task
	req = httptest.NewRequest(http.MethodGet, "/api/v1/tasks/"+taskID, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w = httptest.NewRecorder()

	ts.Router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var getResp map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &getResp)
	require.NoError(t, err)

	data, ok := getResp["data"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, taskID, data["id"])
	assert.Equal(t, agentID, data["agent_id"])
	assert.Equal(t, "exec_shell", data["type"])
}

func TestGetTaskNotFound(t *testing.T) {
	ts, err := test.SetupTestServer(nil)
	require.NoError(t, err)
	defer ts.Cleanup()

	// Clean database
	err = ts.CleanDatabase()
	require.NoError(t, err)

	// Create test user and get token
	userID, err := ts.CreateTestUser("notfoundtaskuser", "password123", "admin")
	require.NoError(t, err)

	token, err := ts.GenerateTestToken(userID, "notfoundtaskuser", "admin")
	require.NoError(t, err)

	// Use a valid UUID format that doesn't exist
	req := httptest.NewRequest(http.MethodGet, "/api/v1/tasks/00000000-0000-0000-0000-000000000000", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	ts.Router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestCancelTask(t *testing.T) {
	ts, err := test.SetupTestServer(nil)
	require.NoError(t, err)
	defer ts.Cleanup()

	// Clean database
	err = ts.CleanDatabase()
	require.NoError(t, err)

	// Create test user and get token
	userID, err := ts.CreateTestUser("canceltaskuser", "password123", "admin")
	require.NoError(t, err)

	token, err := ts.GenerateTestToken(userID, "canceltaskuser", "admin")
	require.NoError(t, err)

	// Create an agent
	agentBody := map[string]interface{}{
		"name": "cancel-task-agent",
	}
	jsonBody, _ := json.Marshal(agentBody)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/agents", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	ts.Router.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)

	var agentResp map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &agentResp)
	require.NoError(t, err)

	agentData, ok := agentResp["data"].(map[string]interface{})
	require.True(t, ok)
	agentID := agentData["id"].(string)

	// Create a task
	taskBody := map[string]interface{}{
		"agent_id": agentID,
		"type":     "exec_shell",
		"params": map[string]interface{}{
			"command": "echo test",
		},
	}
	jsonBody, _ = json.Marshal(taskBody)

	req = httptest.NewRequest(http.MethodPost, "/api/v1/tasks", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w = httptest.NewRecorder()
	ts.Router.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)

	var taskResp map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &taskResp)
	require.NoError(t, err)

	taskData, ok := taskResp["data"].(map[string]interface{})
	require.True(t, ok)
	taskID := taskData["id"].(string)

	// Cancel the task
	req = httptest.NewRequest(http.MethodDelete, "/api/v1/tasks/"+taskID, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w = httptest.NewRecorder()

	ts.Router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Verify task is canceled
	req = httptest.NewRequest(http.MethodGet, "/api/v1/tasks/"+taskID, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w = httptest.NewRecorder()

	ts.Router.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	var getResp map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &getResp)
	require.NoError(t, err)

	data, ok := getResp["data"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "canceled", data["status"])
}

func TestBatchCreateTasks(t *testing.T) {
	ts, err := test.SetupTestServer(nil)
	require.NoError(t, err)
	defer ts.Cleanup()

	// Clean database
	err = ts.CleanDatabase()
	require.NoError(t, err)

	// Create test user and get token
	userID, err := ts.CreateTestUser("batchtaskuser", "password123", "admin")
	require.NoError(t, err)

	token, err := ts.GenerateTestToken(userID, "batchtaskuser", "admin")
	require.NoError(t, err)

	// Create agents
	agentIDs := make([]string, 2)
	for i := 0; i < 2; i++ {
		agentBody := map[string]interface{}{
			"name": "batch-agent-" + string(rune('0'+i)),
		}
		jsonBody, _ := json.Marshal(agentBody)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/agents", bytes.NewReader(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		ts.Router.ServeHTTP(w, req)
		require.Equal(t, http.StatusCreated, w.Code)

		var agentResp map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &agentResp)
		require.NoError(t, err)

		agentData, ok := agentResp["data"].(map[string]interface{})
		require.True(t, ok)
		agentIDs[i] = agentData["id"].(string)
	}

	// Batch create tasks
	batchBody := map[string]interface{}{
		"tasks": []map[string]interface{}{
			{
				"agent_id": agentIDs[0],
				"type":     "exec_shell",
				"params": map[string]interface{}{
					"command": "echo task1",
				},
			},
			{
				"agent_id": agentIDs[1],
				"type":     "exec_shell",
				"params": map[string]interface{}{
					"command": "echo task2",
				},
			},
		},
	}
	jsonBody, _ := json.Marshal(batchBody)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/tasks/batch", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	ts.Router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	data, ok := resp["data"].(map[string]interface{})
	require.True(t, ok)

	tasks, ok := data["tasks"].([]interface{})
	require.True(t, ok)
	assert.Len(t, tasks, 2)

	count, ok := data["count"].(float64)
	require.True(t, ok)
	assert.Equal(t, float64(2), count)
}

func TestFilterTasksByStatus(t *testing.T) {
	ts, err := test.SetupTestServer(nil)
	require.NoError(t, err)
	defer ts.Cleanup()

	// Clean database
	err = ts.CleanDatabase()
	require.NoError(t, err)

	// Create test user and get token
	userID, err := ts.CreateTestUser("filtertaskuser", "password123", "admin")
	require.NoError(t, err)

	token, err := ts.GenerateTestToken(userID, "filtertaskuser", "admin")
	require.NoError(t, err)

	// Create an agent
	agentBody := map[string]interface{}{
		"name": "filter-task-agent",
	}
	jsonBody, _ := json.Marshal(agentBody)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/agents", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	ts.Router.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)

	var agentResp map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &agentResp)
	require.NoError(t, err)

	agentData, ok := agentResp["data"].(map[string]interface{})
	require.True(t, ok)
	agentID := agentData["id"].(string)

	// Create a task
	taskBody := map[string]interface{}{
		"agent_id": agentID,
		"type":     "exec_shell",
		"params": map[string]interface{}{
			"command": "echo test",
		},
	}
	jsonBody, _ = json.Marshal(taskBody)

	req = httptest.NewRequest(http.MethodPost, "/api/v1/tasks", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w = httptest.NewRecorder()
	ts.Router.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)

	// List tasks with status filter
	req = httptest.NewRequest(http.MethodGet, "/api/v1/tasks?status=pending", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w = httptest.NewRecorder()

	ts.Router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

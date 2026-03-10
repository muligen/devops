// Package integration provides integration tests for agent management.
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

func TestCreateAgent(t *testing.T) {
	ts, err := test.SetupTestServer(nil)
	require.NoError(t, err)
	defer ts.Cleanup()

	// Clean database
	err = ts.CleanDatabase()
	require.NoError(t, err)

	// Create test user and get token
	_, err = ts.CreateTestUser("agentadmin", "password123", "admin")
	require.NoError(t, err)

	token, err := ts.GenerateTestToken("agentadmin-id", "agentadmin", "admin")
	require.NoError(t, err)

	tests := []struct {
		name       string
		agentName  string
		metadata   map[string]interface{}
		wantStatus int
		wantError  bool
	}{
		{
			name:       "create agent with name only",
			agentName:  "test-agent-1",
			metadata:   nil,
			wantStatus: http.StatusCreated,
			wantError:  false,
		},
		{
			name:      "create agent with metadata",
			agentName: "test-agent-2",
			metadata: map[string]interface{}{
				"location": "datacenter-1",
				"env":      "production",
			},
			wantStatus: http.StatusCreated,
			wantError:  false,
		},
		{
			name:       "empty name",
			agentName:  "",
			metadata:   nil,
			wantStatus: http.StatusBadRequest,
			wantError:  true,
		},
		{
			name:       "name too long",
			agentName:  "this-is-a-very-long-agent-name-that-exceeds-the-maximum-allowed-length-of-100-characters-but-we-keep-going",
			metadata:   nil,
			wantStatus: http.StatusBadRequest,
			wantError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := map[string]interface{}{
				"name": tt.agentName,
			}
			if tt.metadata != nil {
				body["metadata"] = tt.metadata
			}
			jsonBody, _ := json.Marshal(body)

			req := httptest.NewRequest(http.MethodPost, "/api/v1/agents", bytes.NewReader(jsonBody))
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
				assert.Equal(t, tt.agentName, data["name"])
				assert.NotEmpty(t, data["token"]) // Token should be returned on creation
				assert.Equal(t, "offline", data["status"])
			}
		})
	}
}

func TestListAgents(t *testing.T) {
	ts, err := test.SetupTestServer(nil)
	require.NoError(t, err)
	defer ts.Cleanup()

	// Clean database
	err = ts.CleanDatabase()
	require.NoError(t, err)

	// Create test user and get token
	_, err = ts.CreateTestUser("listuser", "password123", "admin")
	require.NoError(t, err)

	token, err := ts.GenerateTestToken("listuser-id", "listuser", "admin")
	require.NoError(t, err)

	// Create test agents
	for i := 1; i <= 3; i++ {
		body := map[string]interface{}{
			"name": "list-agent-" + string(rune('0'+i)),
		}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/agents", bytes.NewReader(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		ts.Router.ServeHTTP(w, req)
		require.Equal(t, http.StatusCreated, w.Code)
	}

	// List agents
	req := httptest.NewRequest(http.MethodGet, "/api/v1/agents", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	ts.Router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	// data is an array directly in the response
	items, ok := resp["data"].([]interface{})
	require.True(t, ok)
	assert.GreaterOrEqual(t, len(items), 3)

	// Test pagination
	req = httptest.NewRequest(http.MethodGet, "/api/v1/agents?page=1&page_size=2", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w = httptest.NewRecorder()

	ts.Router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	err = json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	items, ok = resp["data"].([]interface{})
	require.True(t, ok)
	assert.LessOrEqual(t, len(items), 2)

	// Check pagination info
	pagination, ok := resp["pagination"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, float64(1), pagination["page"])
	assert.Equal(t, float64(2), pagination["page_size"])
}

func TestGetAgent(t *testing.T) {
	ts, err := test.SetupTestServer(nil)
	require.NoError(t, err)
	defer ts.Cleanup()

	// Clean database
	err = ts.CleanDatabase()
	require.NoError(t, err)

	// Create test user and get token
	_, err = ts.CreateTestUser("getuser", "password123", "admin")
	require.NoError(t, err)

	token, err := ts.GenerateTestToken("getuser-id", "getuser", "admin")
	require.NoError(t, err)

	// Create an agent
	body := map[string]interface{}{
		"name": "get-test-agent",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/agents", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	ts.Router.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)

	var createResp map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &createResp)
	require.NoError(t, err)

	data, ok := createResp["data"].(map[string]interface{})
	require.True(t, ok)
	agentID := data["id"].(string)

	// Get the agent
	req = httptest.NewRequest(http.MethodGet, "/api/v1/agents/"+agentID, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w = httptest.NewRecorder()

	ts.Router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var getResp map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &getResp)
	require.NoError(t, err)

	data, ok = getResp["data"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, agentID, data["id"])
	assert.Equal(t, "get-test-agent", data["name"])
}

func TestGetAgentNotFound(t *testing.T) {
	ts, err := test.SetupTestServer(nil)
	require.NoError(t, err)
	defer ts.Cleanup()

	// Clean database
	err = ts.CleanDatabase()
	require.NoError(t, err)

	// Create test user and get token
	_, err = ts.CreateTestUser("notfounduser", "password123", "admin")
	require.NoError(t, err)

	token, err := ts.GenerateTestToken("notfounduser-id", "notfounduser", "admin")
	require.NoError(t, err)

	// Use a valid UUID format that doesn't exist
	req := httptest.NewRequest(http.MethodGet, "/api/v1/agents/00000000-0000-0000-0000-000000000000", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	ts.Router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestDeleteAgent(t *testing.T) {
	ts, err := test.SetupTestServer(nil)
	require.NoError(t, err)
	defer ts.Cleanup()

	// Clean database
	err = ts.CleanDatabase()
	require.NoError(t, err)

	// Create test user and get token
	_, err = ts.CreateTestUser("deleteuser", "password123", "admin")
	require.NoError(t, err)

	token, err := ts.GenerateTestToken("deleteuser-id", "deleteuser", "admin")
	require.NoError(t, err)

	// Create an agent
	body := map[string]interface{}{
		"name": "delete-test-agent",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/agents", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	ts.Router.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)

	var createResp map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &createResp)
	require.NoError(t, err)

	data, ok := createResp["data"].(map[string]interface{})
	require.True(t, ok)
	agentID := data["id"].(string)

	// Delete the agent
	req = httptest.NewRequest(http.MethodDelete, "/api/v1/agents/"+agentID, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w = httptest.NewRecorder()

	ts.Router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Verify agent is deleted
	req = httptest.NewRequest(http.MethodGet, "/api/v1/agents/"+agentID, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w = httptest.NewRecorder()

	ts.Router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestUpdateAgentStatus(t *testing.T) {
	ts, err := test.SetupTestServer(nil)
	require.NoError(t, err)
	defer ts.Cleanup()

	// Clean database
	err = ts.CleanDatabase()
	require.NoError(t, err)

	// Create test user and get token
	_, err = ts.CreateTestUser("statususer", "password123", "admin")
	require.NoError(t, err)

	token, err := ts.GenerateTestToken("statususer-id", "statususer", "admin")
	require.NoError(t, err)

	// Create an agent
	body := map[string]interface{}{
		"name": "status-test-agent",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/agents", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	ts.Router.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)

	var createResp map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &createResp)
	require.NoError(t, err)

	data, ok := createResp["data"].(map[string]interface{})
	require.True(t, ok)
	agentID := data["id"].(string)

	// Update status
	updateBody := map[string]string{
		"status": "maintenance",
	}
	jsonBody, _ = json.Marshal(updateBody)

	req = httptest.NewRequest(http.MethodPut, "/api/v1/agents/"+agentID+"/status", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w = httptest.NewRecorder()

	ts.Router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Verify status updated
	req = httptest.NewRequest(http.MethodGet, "/api/v1/agents/"+agentID, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w = httptest.NewRecorder()

	ts.Router.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	var getResp map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &getResp)
	require.NoError(t, err)

	data, ok = getResp["data"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "maintenance", data["status"])
}

func TestUnauthorizedAccess(t *testing.T) {
	ts, err := test.SetupTestServer(nil)
	require.NoError(t, err)
	defer ts.Cleanup()

	tests := []struct {
		name   string
		method string
		path   string
	}{
		{
			name:   "create agent without token",
			method: http.MethodPost,
			path:   "/api/v1/agents",
		},
		{
			name:   "list agents without token",
			method: http.MethodGet,
			path:   "/api/v1/agents",
		},
		{
			name:   "get agent without token",
			method: http.MethodGet,
			path:   "/api/v1/agents/some-id",
		},
		{
			name:   "delete agent without token",
			method: http.MethodDelete,
			path:   "/api/v1/agents/some-id",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			w := httptest.NewRecorder()

			ts.Router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusUnauthorized, w.Code)
		})
	}
}

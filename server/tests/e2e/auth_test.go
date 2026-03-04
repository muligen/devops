// Package e2e_test provides end-to-end tests for user authentication
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

func TestUserAuth(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	ctx := context.Background()
	suite, err := e2e.SetupTestSuite(ctx)
	require.NoError(t, err, "Failed to setup test suite")
	defer suite.Teardown()

	// Load test fixtures
	err = suite.LoadFixtures(ctx)
	require.NoError(t, err, "Failed to load fixtures")

	client := e2e.NewHTTPClient(suite.Server.URL)

	t.Run("Login_Success", func(t *testing.T) {
		// Test admin login
		resp, err := client.Post("/api/v1/auth/login", map[string]string{
			"username": e2e.TestUsers.Admin.Username,
			"password": "test-password",
		})
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result map[string]interface{}
		err = resp.JSON(&result)
		require.NoError(t, err)

		assert.NotEmpty(t, result["access_token"])
		assert.NotEmpty(t, result["refresh_token"])
		assert.NotEmpty(t, result["expires_in"])

		user := result["user"].(map[string]interface{})
		assert.Equal(t, e2e.TestUsers.Admin.ID, user["id"])
		assert.Equal(t, e2e.TestUsers.Admin.Username, user["username"])
		assert.Equal(t, e2e.TestUsers.Admin.Role, user["role"])
	})

	t.Run("Login_InvalidCredentials", func(t *testing.T) {
		resp, err := client.Post("/api/v1/auth/login", map[string]string{
			"username": e2e.TestUsers.Admin.Username,
			"password": "wrong-password",
		})
		require.NoError(t, err)
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("Login_UserNotFound", func(t *testing.T) {
		resp, err := client.Post("/api/v1/auth/login", map[string]string{
			"username": "nonexistent-user",
			"password": "any-password",
		})
		require.NoError(t, err)
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("Login_MissingFields", func(t *testing.T) {
		resp, err := client.Post("/api/v1/auth/login", map[string]string{
			"username": "",
			"password": "",
		})
		require.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})
}

func TestTokenValidation(t *testing.T) {
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

	t.Run("ValidToken_AccessProtectedEndpoint", func(t *testing.T) {
		token := suite.AdminToken()
		client.SetAuthToken(token)

		resp, err := client.Get("/api/v1/agents")
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("MissingToken_Rejected", func(t *testing.T) {
		client.SetAuthToken("")

		resp, err := client.Get("/api/v1/agents")
		require.NoError(t, err)
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("InvalidToken_Rejected", func(t *testing.T) {
		client.SetAuthToken("invalid-token-12345")

		resp, err := client.Get("/api/v1/agents")
		require.NoError(t, err)
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("TokenRefresh_Success", func(t *testing.T) {
		// First, login to get refresh token
		resp, err := client.Post("/api/v1/auth/login", map[string]string{
			"username": e2e.TestUsers.Admin.Username,
			"password": "test-password",
		})
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)

		var result map[string]interface{}
		err = resp.JSON(&result)
		require.NoError(t, err)

		refreshToken := result["refresh_token"].(string)
		require.NotEmpty(t, refreshToken)

		// Refresh token
		refreshResp, err := client.Post("/api/v1/auth/refresh", map[string]string{
			"refresh_token": refreshToken,
		})
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, refreshResp.StatusCode)

		var refreshResult map[string]interface{}
		err = refreshResp.JSON(&refreshResult)
		require.NoError(t, err)

		assert.NotEmpty(t, refreshResult["access_token"])
		assert.NotEmpty(t, refreshResult["refresh_token"])

		// New tokens should be different from old ones
		assert.NotEqual(t, result["access_token"], refreshResult["access_token"])
	})
}

func TestRoleBasedAccess(t *testing.T) {
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

	t.Run("Admin_CanCreateAgent", func(t *testing.T) {
		client.SetAuthToken(suite.AdminToken())

		resp, err := client.Post("/api/v1/agents", map[string]interface{}{
			"name": "test-agent-new",
		})
		require.NoError(t, err)
		assert.Equal(t, http.StatusCreated, resp.StatusCode)
	})

	t.Run("Admin_CanDeleteAgent", func(t *testing.T) {
		// Create agent first
		client.SetAuthToken(suite.AdminToken())
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
	})

	t.Run("Operator_CanCreateTask", func(t *testing.T) {
		client.SetAuthToken(suite.OperatorToken())

		// First create an agent
		createResp, err := client.Post("/api/v1/agents", map[string]interface{}{
			"name": "agent-for-task",
		})
		require.NoError(t, err)

		var result map[string]interface{}
		err = createResp.JSON(&result)
		require.NoError(t, err)
		agentID := result["id"].(string)

		// Create task
		taskResp, err := client.Post("/api/v1/tasks", map[string]interface{}{
			"agent_id": agentID,
			"type":     "exec_shell",
			"params":   map[string]string{"command": "echo test"},
		})
		require.NoError(t, err)
		assert.Equal(t, http.StatusCreated, taskResp.StatusCode)
	})

	t.Run("Viewer_CanListAgents", func(t *testing.T) {
		client.SetAuthToken(suite.ViewerToken())

		resp, err := client.Get("/api/v1/agents")
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("Viewer_CannotCreateAgent", func(t *testing.T) {
		client.SetAuthToken(suite.ViewerToken())

		resp, err := client.Post("/api/v1/agents", map[string]interface{}{
			"name": "should-not-work",
		})
		require.NoError(t, err)
		// Viewer should not have permission (403 Forbidden or 401 Unauthorized depending on middleware)
		assert.NotEqual(t, http.StatusCreated, resp.StatusCode)
		assert.NotEqual(t, http.StatusOK, resp.StatusCode)
	})
}

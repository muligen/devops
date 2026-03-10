// Package integration provides integration tests for authentication.
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

func TestAuthLogin(t *testing.T) {
	ts, err := test.SetupTestServer(nil)
	require.NoError(t, err)
	defer ts.Cleanup()

	// Clean database
	err = ts.CleanDatabase()
	require.NoError(t, err)

	// Create test user
	userID, err := ts.CreateTestUser("testuser", "password123", "admin")
	require.NoError(t, err)
	require.NotEmpty(t, userID)

	tests := []struct {
		name       string
		username   string
		password   string
		wantStatus int
		wantError  bool
	}{
		{
			name:       "valid credentials",
			username:   "testuser",
			password:   "password123",
			wantStatus: http.StatusOK,
			wantError:  false,
		},
		{
			name:       "invalid password",
			username:   "testuser",
			password:   "wrongpassword",
			wantStatus: http.StatusUnauthorized,
			wantError:  true,
		},
		{
			name:       "invalid username",
			username:   "nonexistent",
			password:   "password123",
			wantStatus: http.StatusUnauthorized,
			wantError:  true,
		},
		{
			name:       "empty credentials",
			username:   "",
			password:   "",
			wantStatus: http.StatusBadRequest,
			wantError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := map[string]string{
				"username": tt.username,
				"password": tt.password,
			}
			jsonBody, _ := json.Marshal(body)

			req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(jsonBody))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			ts.Router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)

			if !tt.wantError {
				var resp map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &resp)
				require.NoError(t, err)

				data, ok := resp["data"].(map[string]interface{})
				require.True(t, ok)

				assert.NotEmpty(t, data["access_token"])
				assert.NotEmpty(t, data["refresh_token"])
				assert.NotEmpty(t, data["user"])
			}
		})
	}
}

func TestAuthRefreshToken(t *testing.T) {
	ts, err := test.SetupTestServer(nil)
	require.NoError(t, err)
	defer ts.Cleanup()

	// Clean database
	err = ts.CleanDatabase()
	require.NoError(t, err)

	// Create test user
	_, err = ts.CreateTestUser("refreshuser", "password123", "admin")
	require.NoError(t, err)

	// First, login to get refresh token
	loginBody := map[string]string{
		"username": "refreshuser",
		"password": "password123",
	}
	jsonBody, _ := json.Marshal(loginBody)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	ts.Router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var loginResp map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &loginResp)
	require.NoError(t, err)

	data, ok := loginResp["data"].(map[string]interface{})
	require.True(t, ok)
	refreshToken := data["refresh_token"].(string)
	require.NotEmpty(t, refreshToken)

	// Now test refresh token
	refreshReq := map[string]string{
		"refresh_token": refreshToken,
	}
	jsonBody, _ = json.Marshal(refreshReq)

	req = httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	ts.Router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var refreshResp map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &refreshResp)
	require.NoError(t, err)

	data, ok = refreshResp["data"].(map[string]interface{})
	require.True(t, ok)
	assert.NotEmpty(t, data["access_token"])
	assert.NotEmpty(t, data["refresh_token"])
}

func TestAuthInvalidRefreshToken(t *testing.T) {
	ts, err := test.SetupTestServer(nil)
	require.NoError(t, err)
	defer ts.Cleanup()

	refreshReq := map[string]string{
		"refresh_token": "invalid-token",
	}
	jsonBody, _ := json.Marshal(refreshReq)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	ts.Router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthLogout(t *testing.T) {
	ts, err := test.SetupTestServer(nil)
	require.NoError(t, err)
	defer ts.Cleanup()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", nil)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	ts.Router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	data, ok := resp["data"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "logged out successfully", data["message"])
}

func TestCreateUser(t *testing.T) {
	ts, err := test.SetupTestServer(nil)
	require.NoError(t, err)
	defer ts.Cleanup()

	// Clean database
	err = ts.CleanDatabase()
	require.NoError(t, err)

	// Create admin user and get token
	_, err = ts.CreateTestUser("admin", "admin123", "admin")
	require.NoError(t, err)

	token, err := ts.GenerateTestToken("admin-id", "admin", "admin")
	require.NoError(t, err)

	tests := []struct {
		name       string
		username   string
		password   string
		email      string
		role       string
		wantStatus int
	}{
		{
			name:       "create operator user",
			username:   "operator1",
			password:   "password123",
			email:      "operator1@test.com",
			role:       "operator",
			wantStatus: http.StatusCreated,
		},
		{
			name:       "create viewer user",
			username:   "viewer1",
			password:   "password123",
			email:      "viewer1@test.com",
			role:       "viewer",
			wantStatus: http.StatusCreated,
		},
		{
			name:       "duplicate username",
			username:   "operator1",
			password:   "password123",
			email:      "operator1-dup@test.com",
			role:       "operator",
			wantStatus: http.StatusConflict,
		},
		{
			name:       "invalid role",
			username:   "invalid-role-user",
			password:   "password123",
			email:      "invalid@test.com",
			role:       "invalid",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "short password",
			username:   "shortpass",
			password:   "short",
			email:      "short@test.com",
			role:       "operator",
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := map[string]string{
				"username": tt.username,
				"password": tt.password,
				"email":    tt.email,
				"role":     tt.role,
			}
			jsonBody, _ := json.Marshal(body)

			req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/users", bytes.NewReader(jsonBody))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+token)
			w := httptest.NewRecorder()

			ts.Router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

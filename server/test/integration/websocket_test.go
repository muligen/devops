// Package integration provides integration tests for WebSocket authentication.
package integration

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/require"

	"github.com/agentteams/server/test"
)

// WSMessage represents a WebSocket message.
type WSMessage struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

// Note: WebSocket tests are skipped because the test helper
// doesn't include WebSocket routes. These tests should be run
// with the full server setup or in E2E tests.

func TestWebSocketAuthFlow(t *testing.T) {
	t.Skip("WebSocket tests require full server setup with WebSocket routes")

	ts, err := test.SetupTestServer(nil)
	require.NoError(t, err)
	defer ts.Cleanup()

	// Clean database
	err = ts.CleanDatabase()
	require.NoError(t, err)

	// Create test user and get token
	userID, err := ts.CreateTestUser("wsuser", "password123", "admin")
	require.NoError(t, err)

	token, err := ts.GenerateTestToken(userID, "wsuser", "admin")
	require.NoError(t, err)

	// Create an agent
	agentBody := map[string]interface{}{
		"name": "ws-test-agent",
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
	agentToken := agentData["token"].(string)

	// Connect to WebSocket
	wsURL := "ws" + strings.TrimPrefix(ts.Server.URL, "http") + "/api/v1/agent/ws"

	dialer := websocket.DefaultDialer
	conn, resp, err := dialer.Dial(wsURL, nil)
	require.NoError(t, err)
	defer conn.Close()
	require.Equal(t, http.StatusSwitchingProtocols, resp.StatusCode)

	// Step 1: Send auth request
	authMsg := map[string]interface{}{
		"type": "auth",
		"data": map[string]string{
			"agent_id": agentID,
		},
	}
	err = conn.WriteJSON(authMsg)
	require.NoError(t, err)

	// Step 2: Receive challenge
	var challengeMsg WSMessage
	err = conn.ReadJSON(&challengeMsg)
	require.NoError(t, err)
	require.Equal(t, "challenge", challengeMsg.Type)

	var challengeData struct {
		Nonce string `json:"nonce"`
	}
	err = json.Unmarshal(challengeMsg.Data, &challengeData)
	require.NoError(t, err)
	require.NotEmpty(t, challengeData.Nonce)

	// Step 3: Compute HMAC response
	response := computeHMAC(agentToken, challengeData.Nonce)

	// Step 4: Send challenge response
	responseMsg := map[string]interface{}{
		"type": "challenge",
		"data": map[string]string{
			"response": response,
		},
	}
	err = conn.WriteJSON(responseMsg)
	require.NoError(t, err)

	// Step 5: Receive auth result
	var authResultMsg WSMessage
	err = conn.ReadJSON(&authResultMsg)
	require.NoError(t, err)
	require.Equal(t, "auth_result", authResultMsg.Type)

	var authResultData struct {
		Success bool   `json:"success"`
		Message string `json:"message"`
	}
	err = json.Unmarshal(authResultMsg.Data, &authResultData)
	require.NoError(t, err)
	require.True(t, authResultData.Success)
	require.Equal(t, "authenticated", authResultData.Message)
}

func TestWebSocketInvalidAgentID(t *testing.T) {
	t.Skip("WebSocket tests require full server setup with WebSocket routes")

	ts, err := test.SetupTestServer(nil)
	require.NoError(t, err)
	defer ts.Cleanup()

	// Clean database
	err = ts.CleanDatabase()
	require.NoError(t, err)

	// Connect to WebSocket
	wsURL := "ws" + strings.TrimPrefix(ts.Server.URL, "http") + "/api/v1/agent/ws"

	dialer := websocket.DefaultDialer
	conn, _, err := dialer.Dial(wsURL, nil)
	require.NoError(t, err)
	defer conn.Close()

	// Send auth with invalid agent ID
	authMsg := map[string]interface{}{
		"type": "auth",
		"data": map[string]string{
			"agent_id": "non-existent-agent-id",
		},
	}
	err = conn.WriteJSON(authMsg)
	require.NoError(t, err)

	// Should receive error
	var errorMsg WSMessage
	err = conn.ReadJSON(&errorMsg)
	require.NoError(t, err)
	require.Equal(t, "error", errorMsg.Type)
}

func TestWebSocketInvalidChallengeResponse(t *testing.T) {
	t.Skip("WebSocket tests require full server setup with WebSocket routes")

	ts, err := test.SetupTestServer(nil)
	require.NoError(t, err)
	defer ts.Cleanup()

	// Clean database
	err = ts.CleanDatabase()
	require.NoError(t, err)

	// Create test user and get token
	userID, err := ts.CreateTestUser("wsuser2", "password123", "admin")
	require.NoError(t, err)

	token, err := ts.GenerateTestToken(userID, "wsuser2", "admin")
	require.NoError(t, err)

	// Create an agent
	agentBody := map[string]interface{}{
		"name": "ws-test-agent-2",
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

	// Connect to WebSocket
	wsURL := "ws" + strings.TrimPrefix(ts.Server.URL, "http") + "/api/v1/agent/ws"

	dialer := websocket.DefaultDialer
	conn, _, err := dialer.Dial(wsURL, nil)
	require.NoError(t, err)
	defer conn.Close()

	// Send auth request
	authMsg := map[string]interface{}{
		"type": "auth",
		"data": map[string]string{
			"agent_id": agentID,
		},
	}
	err = conn.WriteJSON(authMsg)
	require.NoError(t, err)

	// Receive challenge
	var challengeMsg WSMessage
	err = conn.ReadJSON(&challengeMsg)
	require.NoError(t, err)

	// Send invalid response
	responseMsg := map[string]interface{}{
		"type": "challenge",
		"data": map[string]string{
			"response": "invalid-response",
		},
	}
	err = conn.WriteJSON(responseMsg)
	require.NoError(t, err)

	// Should receive error
	var errorMsg WSMessage
	err = conn.ReadJSON(&errorMsg)
	require.NoError(t, err)
	require.Equal(t, "auth_result", errorMsg.Type)

	var authResultData struct {
		Success bool   `json:"success"`
		Message string `json:"message"`
	}
	err = json.Unmarshal(errorMsg.Data, &authResultData)
	require.NoError(t, err)
	require.False(t, authResultData.Success)
}

func TestWebSocketHeartbeatWithoutAuth(t *testing.T) {
	t.Skip("WebSocket tests require full server setup with WebSocket routes")

	ts, err := test.SetupTestServer(nil)
	require.NoError(t, err)
	defer ts.Cleanup()

	// Connect to WebSocket
	wsURL := "ws" + strings.TrimPrefix(ts.Server.URL, "http") + "/api/v1/agent/ws"

	dialer := websocket.DefaultDialer
	conn, _, err := dialer.Dial(wsURL, nil)
	require.NoError(t, err)
	defer conn.Close()

	// Try to send heartbeat without auth
	heartbeatMsg := map[string]interface{}{
		"type": "heartbeat",
		"data": map[string]int64{
			"timestamp": time.Now().Unix(),
		},
	}
	err = conn.WriteJSON(heartbeatMsg)
	require.NoError(t, err)

	// Should receive error
	var errorMsg WSMessage
	err = conn.ReadJSON(&errorMsg)
	require.NoError(t, err)
	require.Equal(t, "error", errorMsg.Type)
}

func computeHMAC(key, data string) string {
	h := hmac.New(sha256.New, []byte(key))
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}

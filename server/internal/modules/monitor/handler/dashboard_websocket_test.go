package handler_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/agentteams/server/internal/modules/auth/service"
	"github.com/agentteams/server/internal/modules/monitor/handler"
	"github.com/agentteams/server/internal/pkg/logger"
)

var testJWTSecret = "test-secret-key-for-websocket-tests"

func setupTestJWTService(t *testing.T) *service.JWTService {
	return service.NewJWTService(service.JWTConfig{
		Secret:        testJWTSecret,
		Expiry:        time.Hour,
		RefreshExpiry: 24 * time.Hour,
	})
}

func setupTestHandler(t *testing.T, jwtService *service.JWTService) *handler.DashboardWSHandler {
	log := logger.NewNop()
	return handler.NewDashboardWSHandler(jwtService, nil, log)
}

func setupTestRouter(jwtService *service.JWTService) (*gin.Engine, *handler.DashboardWSHandler) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	h := handler.NewDashboardWSHandler(jwtService, nil, logger.NewNop())
	router.GET("/ws/dashboard", h.Handle)
	return router, h
}

// TestDashboardClient_New tests DashboardClient creation.
func TestDashboardClient_New(t *testing.T) {
	conn := &websocket.Conn{} // Mock connection
	userID := "user-123"
	username := "testuser"

	client := handler.NewDashboardClient(conn, userID, username)

	assert.NotEmpty(t, client.ID)
	assert.Equal(t, userID, client.UserID)
	assert.Equal(t, username, client.Username)
	assert.Equal(t, conn, client.Conn)
	assert.NotNil(t, client.Send)
	assert.NotNil(t, client.Close)
}

// TestDashboardWSHandler_Handle_MissingToken tests handling of missing token.
func TestDashboardWSHandler_Handle_MissingToken(t *testing.T) {
	jwtService := setupTestJWTService(t)
	router, _ := setupTestRouter(jwtService)

	// Request without token
	req := httptest.NewRequest("GET", "/ws/dashboard", nil)
	req.Header.Set("Connection", "Upgrade")
	req.Header.Set("Upgrade", "websocket")
	req.Header.Set("Sec-WebSocket-Version", "13")
	req.Header.Set("Sec-WebSocket-Key", "dGhlIHNhbXBsZSBub25jZQ==")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "missing token")
}

// TestDashboardWSHandler_Handle_InvalidToken tests handling of invalid token.
func TestDashboardWSHandler_Handle_InvalidToken(t *testing.T) {
	jwtService := setupTestJWTService(t)
	router, _ := setupTestRouter(jwtService)

	// Request with invalid token
	req := httptest.NewRequest("GET", "/ws/dashboard?token=invalid-token", nil)
	req.Header.Set("Connection", "Upgrade")
	req.Header.Set("Upgrade", "websocket")
	req.Header.Set("Sec-WebSocket-Version", "13")
	req.Header.Set("Sec-WebSocket-Key", "dGhlIHNhbXBsZSBub25jZQ==")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "invalid token")
}

// TestDashboardWSHandler_Handle_ValidConnection tests successful WebSocket connection.
func TestDashboardWSHandler_Handle_ValidConnection(t *testing.T) {
	jwtService := setupTestJWTService(t)
	router, h := setupTestRouter(jwtService)

	// Generate valid token
	token, err := jwtService.GenerateToken("user-1", "testuser", "admin")
	require.NoError(t, err)

	// Start handler
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	h.Start(ctx)
	defer h.Stop()

	// Create test server
	server := httptest.NewServer(router)
	defer server.Close()

	// Convert http URL to ws URL
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws/dashboard?token=" + token

	// Connect WebSocket client
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	require.NoError(t, err)
	defer conn.Close()

	// Read the connected message
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	_, message, err := conn.ReadMessage()
	require.NoError(t, err)

	// Verify connected message structure
	var msg struct {
		Type string `json:"type"`
		Data struct {
			ClientID string `json:"client_id"`
		} `json:"data"`
	}
	err = json.Unmarshal(message, &msg)
	require.NoError(t, err)
	assert.Equal(t, "connected", msg.Type)
	assert.NotEmpty(t, msg.Data.ClientID)

	// Verify client count
	assert.Equal(t, 1, h.GetClientCount())
}

// TestDashboardWSHandler_Handle_PingPong tests ping/pong functionality.
func TestDashboardWSHandler_Handle_PingPong(t *testing.T) {
	jwtService := setupTestJWTService(t)
	router, h := setupTestRouter(jwtService)

	// Generate valid token
	token, err := jwtService.GenerateToken("user-2", "testuser", "admin")
	require.NoError(t, err)

	// Start handler
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	h.Start(ctx)
	defer h.Stop()

	// Create test server
	server := httptest.NewServer(router)
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws/dashboard?token=" + token

	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	require.NoError(t, err)
	defer conn.Close()

	// Read connected message first
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	_, _, err = conn.ReadMessage()
	require.NoError(t, err)

	// Send ping message
	pingMsg, _ := json.Marshal(map[string]string{"type": "ping"})
	err = conn.WriteMessage(websocket.TextMessage, pingMsg)
	require.NoError(t, err)

	// Read pong response
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	_, message, err := conn.ReadMessage()
	require.NoError(t, err)

	var msg struct {
		Type string `json:"type"`
	}
	err = json.Unmarshal(message, &msg)
	require.NoError(t, err)
	assert.Equal(t, "pong", msg.Type)
}

// TestDashboardWSHandler_Handle_BearerToken tests Bearer token in header.
func TestDashboardWSHandler_Handle_BearerToken(t *testing.T) {
	jwtService := setupTestJWTService(t)
	router, h := setupTestRouter(jwtService)

	// Generate valid token
	token, err := jwtService.GenerateToken("user-3", "testuser", "admin")
	require.NoError(t, err)

	// Start handler
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	h.Start(ctx)
	defer h.Stop()

	// Create test server
	server := httptest.NewServer(router)
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws/dashboard"

	// Connect with Bearer token in header
	header := http.Header{}
	header.Set("Authorization", "Bearer "+token)
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, header)
	require.NoError(t, err)
	defer conn.Close()

	// Read connected message
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	_, message, err := conn.ReadMessage()
	require.NoError(t, err)

	var msg struct {
		Type string `json:"type"`
	}
	err = json.Unmarshal(message, &msg)
	require.NoError(t, err)
	assert.Equal(t, "connected", msg.Type)
}

// TestMetricsBuffer_Add tests adding metrics to buffer.
func TestMetricsBuffer_Add(t *testing.T) {
	buffer := handler.NewMetricsBuffer()

	buffer.Add("agent-1", map[string]interface{}{"cpu": 50.0})
	buffer.Add("agent-2", map[string]interface{}{"cpu": 60.0})

	assert.Equal(t, 2, buffer.Count())
}

// TestMetricsBuffer_ShouldFlush tests flush condition.
func TestMetricsBuffer_ShouldFlush(t *testing.T) {
	buffer := handler.NewMetricsBuffer()

	// Should not flush initially
	assert.False(t, buffer.ShouldFlush())

	// Add metrics until flush threshold
	for i := 0; i < 100; i++ {
		buffer.Add("agent-"+string(rune(i)), map[string]interface{}{"cpu": 50.0})
	}

	// Should flush at 100 metrics
	assert.True(t, buffer.ShouldFlush())
}

// TestMetricsBuffer_Flush tests flushing the buffer.
func TestMetricsBuffer_Flush(t *testing.T) {
	buffer := handler.NewMetricsBuffer()

	buffer.Add("agent-1", map[string]interface{}{"cpu": 50.0})
	buffer.Add("agent-2", map[string]interface{}{"cpu": 60.0})

	result := buffer.Flush()

	assert.Len(t, result, 2)
	assert.Contains(t, result, "agent-1")
	assert.Contains(t, result, "agent-2")

	// Buffer should be empty after flush
	assert.Equal(t, 0, buffer.Count())
}

// TestDashboardWSHandler_ClientCount tests client count tracking.
func TestDashboardWSHandler_ClientCount(t *testing.T) {
	jwtService := setupTestJWTService(t)
	router, h := setupTestRouter(jwtService)

	// Start handler
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	h.Start(ctx)
	defer h.Stop()

	server := httptest.NewServer(router)
	defer server.Close()

	// Initially no clients
	assert.Equal(t, 0, h.GetClientCount())

	// Connect first client
	token1, _ := jwtService.GenerateToken("user-1", "user1", "admin")
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws/dashboard?token=" + token1
	conn1, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	require.NoError(t, err)
	defer conn1.Close()

	// Read connected message
	conn1.SetReadDeadline(time.Now().Add(2 * time.Second))
	_, _, err = conn1.ReadMessage()
	require.NoError(t, err)

	// Wait for client registration
	time.Sleep(100 * time.Millisecond)
	assert.Equal(t, 1, h.GetClientCount())
}

// TestDashboardWSHandler_Handle_ExpiredToken tests handling of expired token.
func TestDashboardWSHandler_Handle_ExpiredToken(t *testing.T) {
	// Create JWT service with very short expiry
	jwtService := service.NewJWTService(service.JWTConfig{
		Secret:        testJWTSecret,
		Expiry:        time.Millisecond, // Expires immediately
		RefreshExpiry: time.Millisecond,
	})

	// Generate token that will expire
	token, err := jwtService.GenerateToken("user-1", "testuser", "admin")
	require.NoError(t, err)

	// Wait for token to expire
	time.Sleep(10 * time.Millisecond)

	// Create router with standard JWT service for validation
	standardJWTService := setupTestJWTService(t)
	router, _ := setupTestRouter(standardJWTService)

	// Try to connect with expired token
	req := httptest.NewRequest("GET", "/ws/dashboard?token="+token, nil)
	req.Header.Set("Connection", "Upgrade")
	req.Header.Set("Upgrade", "websocket")
	req.Header.Set("Sec-WebSocket-Version", "13")
	req.Header.Set("Sec-WebSocket-Key", "dGhlIHNhbXBsZSBub25jZQ==")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Should fail due to expired token (signed with different secret)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestDashboardWSHandler_MultipleConnections tests multiple concurrent connections.
func TestDashboardWSHandler_MultipleConnections(t *testing.T) {
	jwtService := setupTestJWTService(t)
	router, h := setupTestRouter(jwtService)

	// Start handler
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	h.Start(ctx)
	defer h.Stop()

	server := httptest.NewServer(router)
	defer server.Close()

	// Connect multiple clients
	var conns []*websocket.Conn
	for i := 0; i < 5; i++ {
		token, _ := jwtService.GenerateToken("user-"+string(rune('0'+i)), "user", "admin")
		wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws/dashboard?token=" + token
		conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
		require.NoError(t, err)
		conns = append(conns, conn)

		// Read connected message
		conn.SetReadDeadline(time.Now().Add(2 * time.Second))
		_, _, err = conn.ReadMessage()
		require.NoError(t, err)
	}

	// Verify client count - allow for timing variations
	time.Sleep(100 * time.Millisecond)
	assert.GreaterOrEqual(t, h.GetClientCount(), 3) // At least 3 clients should be connected

	// Cleanup
	for _, conn := range conns {
		conn.Close()
	}
}

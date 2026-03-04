// Package handler provides WebSocket handlers for agent connections.
package handler

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/agentteams/server/internal/modules/agent/domain"
	"github.com/agentteams/server/internal/modules/agent/service"
	"github.com/agentteams/server/internal/pkg/cache"
	"github.com/agentteams/server/internal/pkg/logger"
	"github.com/agentteams/server/internal/pkg/mq"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

const (
	// Session expiration time
	sessionExpiration = 24 * time.Hour

	// Nonce expiration time
	nonceExpiration = 5 * time.Minute
)

// WebSocket message types
const (
	TypeAuth        = "auth"
	TypeChallenge   = "challenge"
	TypeAuthResult  = "auth_result"
	TypeHeartbeat   = "heartbeat"
	TypeMetrics     = "metrics"
	TypeCommand     = "command"
	TypeCommandAck  = "command_ack"
	TypeResult      = "result"
	TypeError       = "error"
)

// AuthState represents authentication state.
type AuthState int

const (
	AuthStatePending AuthState = iota
	AuthStateChallenged
	AuthStateAuthenticated
)

// Session represents an agent WebSocket session.
type Session struct {
	AgentID    string
	AgentName  string
	Conn       *websocket.Conn
	State      AuthState
	Nonce      string
	LastSeen   time.Time
	SendChan   chan []byte
	CloseChan  chan struct{}
}

// NewSession creates a new session.
func NewSession(conn *websocket.Conn) *Session {
	return &Session{
		Conn:      conn,
		State:     AuthStatePending,
		SendChan:  make(chan []byte, 256),
		CloseChan: make(chan struct{}),
		LastSeen:  time.Now(),
	}
}

// Send sends a message to the session.
func (s *Session) Send(msg []byte) bool {
	select {
	case s.SendChan <- msg:
		return true
	default:
		return false
	}
}

// Close closes the session.
func (s *Session) Close() {
	close(s.CloseChan)
}

// WebSocketHandler handles WebSocket connections from agents.
type WebSocketHandler struct {
	agentService   *service.Service
	taskService    TaskResultService
	metricsService MetricsService
	cache          *cache.Client
	mq             *mq.Client
	upgrader       websocket.Upgrader
	sessions       sync.Map // map[string]*Session (agentID -> Session)
	logger         *logger.Logger
}

// TaskResultService defines the interface for task result handling.
type TaskResultService interface {
	UpdateTaskResult(ctx context.Context, id string, status string, exitCode int, output string, duration float64) error
	StartTask(ctx context.Context, id string) error
}

// MetricsService defines the interface for metrics handling.
type MetricsService interface {
	StoreMetric(ctx context.Context, agentID string, cpuUsage float64, memTotal, memUsed int64, memPercent float64, diskTotal, diskUsed int64, diskPercent float64, uptime int64) error
}

// NewWebSocketHandler creates a new WebSocket handler.
func NewWebSocketHandler(agentService *service.Service, cache *cache.Client, mq *mq.Client, log *logger.Logger) *WebSocketHandler {
	return &WebSocketHandler{
		agentService: agentService,
		cache:        cache,
		mq:           mq,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // Allow all origins for now
			},
			ReadBufferSize:  4096,
			WriteBufferSize: 4096,
		},
		logger: log,
	}
}

// SetTaskService sets the task result service.
func (h *WebSocketHandler) SetTaskService(svc TaskResultService) {
	h.taskService = svc
}

// SetMetricsService sets the metrics service.
func (h *WebSocketHandler) SetMetricsService(svc MetricsService) {
	h.metricsService = svc
}

// Handle handles WebSocket connection upgrade.
func (h *WebSocketHandler) Handle(c *gin.Context) {
	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		h.logger.Errorw("Failed to upgrade connection", "error", err)
		return
	}

	session := NewSession(conn)
	defer h.cleanup(session)

	// Start write goroutine
	go h.writePump(session)

	// Read messages
	h.readPump(session)
}

// readPump reads messages from the WebSocket connection.
func (h *WebSocketHandler) readPump(session *Session) {
	defer session.Conn.Close()

	for {
		_, message, err := session.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				h.logger.Errorw("WebSocket read error", "error", err)
			}
			break
		}

		session.LastSeen = time.Now()

		if err := h.handleMessage(session, message); err != nil {
			h.logger.Errorw("Failed to handle message", "error", err)
			h.sendError(session, err.Error())
		}
	}
}

// writePump writes messages to the WebSocket connection.
func (h *WebSocketHandler) writePump(session *Session) {
	ticker := time.NewTicker(30 * time.Second)
	defer func() {
		ticker.Stop()
		session.Conn.Close()
	}()

	for {
		select {
		case <-session.CloseChan:
			return
		case message := <-session.SendChan:
			if err := session.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
				h.logger.Errorw("WebSocket write error", "error", err)
				return
			}
		case <-ticker.C:
			// Send ping for keepalive
			if err := session.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// handleMessage handles incoming WebSocket messages.
func (h *WebSocketHandler) handleMessage(session *Session, message []byte) error {
	var msg struct {
		Type string          `json:"type"`
		Data json.RawMessage `json:"data"`
	}

	if err := json.Unmarshal(message, &msg); err != nil {
		return fmt.Errorf("invalid message format: %w", err)
	}

	switch msg.Type {
	case TypeAuth:
		return h.handleAuth(session, msg.Data)
	case TypeChallenge:
		return h.handleChallengeResponse(session, msg.Data)
	case TypeHeartbeat:
		return h.handleHeartbeat(session, msg.Data)
	case TypeMetrics:
		return h.handleMetrics(session, msg.Data)
	case TypeResult:
		return h.handleResult(session, msg.Data)
	default:
		return fmt.Errorf("unknown message type: %s", msg.Type)
	}
}

// handleAuth handles initial authentication request.
func (h *WebSocketHandler) handleAuth(session *Session, data json.RawMessage) error {
	var req struct {
		AgentID string `json:"agent_id"`
	}
	if err := json.Unmarshal(data, &req); err != nil {
		return fmt.Errorf("invalid auth request: %w", err)
	}

	// Get agent by ID
	agent, err := h.agentService.GetAgent(context.Background(), req.AgentID)
	if err != nil {
		return fmt.Errorf("agent not found")
	}

	// Generate nonce
	nonce, err := generateNonce()
	if err != nil {
		return fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Store nonce in Redis
	ctx := context.Background()
	nonceKey := fmt.Sprintf("agent:nonce:%s", agent.ID)
	if err := h.cache.Set(ctx, nonceKey, nonce, nonceExpiration); err != nil {
		return fmt.Errorf("failed to store nonce: %w", err)
	}

	// Update session state
	session.AgentID = agent.ID
	session.AgentName = agent.Name
	session.Nonce = nonce
	session.State = AuthStateChallenged

	// Send challenge
	response := map[string]interface{}{
		"type": TypeChallenge,
		"data": map[string]string{
			"nonce": nonce,
		},
	}
	h.sendMessage(session, response)

	return nil
}

// handleChallengeResponse handles challenge response.
func (h *WebSocketHandler) handleChallengeResponse(session *Session, data json.RawMessage) error {
	if session.State != AuthStateChallenged {
		return fmt.Errorf("invalid auth state")
	}

	var req struct {
		Response string `json:"response"`
	}
	if err := json.Unmarshal(data, &req); err != nil {
		return fmt.Errorf("invalid challenge response: %w", err)
	}

	// Get agent
	agent, err := h.agentService.GetAgent(context.Background(), session.AgentID)
	if err != nil {
		return fmt.Errorf("agent not found")
	}

	// Verify response
	// Response should be HMAC(token_hash, nonce)
	expectedResponse := hmacSHA256(agent.TokenHash, session.Nonce)
	if !hmac.Equal([]byte(req.Response), []byte(expectedResponse)) {
		h.sendAuthResult(session, false, "invalid response")
		return fmt.Errorf("invalid challenge response")
	}

	// Clear nonce
	ctx := context.Background()
	nonceKey := fmt.Sprintf("agent:nonce:%s", agent.ID)
	_ = h.cache.Delete(ctx, nonceKey)

	// Create session in Redis
	sessionKey := fmt.Sprintf("agent:session:%s", agent.ID)
	sessionData := map[string]interface{}{
		"agent_id":   agent.ID,
		"agent_name": agent.Name,
		"connected":  time.Now().Unix(),
	}
	if err := h.cache.SetJSON(ctx, sessionKey, sessionData, sessionExpiration); err != nil {
		h.logger.Errorw("Failed to store session", "error", err)
	}

	// Update agent status to online
	if err := h.agentService.UpdateAgentStatus(ctx, agent.ID, domain.StatusOnline); err != nil {
		h.logger.Errorw("Failed to update agent status", "error", err)
	}

	// Update session info
	session.State = AuthStateAuthenticated

	// Store session
	h.sessions.Store(agent.ID, session)

	// Publish online event
	_ = h.mq.PublishEvent(ctx, mq.EventAgentOnline, map[string]interface{}{
		"agent_id":   agent.ID,
		"agent_name": agent.Name,
		"timestamp":  time.Now().Unix(),
	})

	// Send success response
	h.sendAuthResult(session, true, "authenticated")

	h.logger.Infow("Agent authenticated", "agent_id", agent.ID, "agent_name", agent.Name)

	return nil
}

// handleHeartbeat handles heartbeat messages.
func (h *WebSocketHandler) handleHeartbeat(session *Session, data json.RawMessage) error {
	if session.State != AuthStateAuthenticated {
		return fmt.Errorf("not authenticated")
	}

	// Update last seen
	ctx := context.Background()
	if err := h.agentService.UpdateAgentStatus(ctx, session.AgentID, domain.StatusOnline); err != nil {
		h.logger.Errorw("Failed to update agent last seen", "error", err)
	}

	// Send ack
	response := map[string]interface{}{
		"type": TypeHeartbeat,
		"data": map[string]int64{
			"timestamp": time.Now().Unix(),
		},
	}
	h.sendMessage(session, response)

	return nil
}

// handleMetrics handles metrics messages.
func (h *WebSocketHandler) handleMetrics(session *Session, data json.RawMessage) error {
	if session.State != AuthStateAuthenticated {
		return fmt.Errorf("not authenticated")
	}

	var metrics struct {
		CPU     float64 `json:"cpu_usage"`
		Memory  struct {
			Total   uint64  `json:"total"`
			Used    uint64  `json:"used"`
			Percent float64 `json:"percent"`
		} `json:"memory"`
		Disk struct {
			Total   uint64  `json:"total"`
			Used    uint64  `json:"used"`
			Percent float64 `json:"percent"`
		} `json:"disk"`
		Uptime uint64 `json:"uptime"`
	}
	if err := json.Unmarshal(data, &metrics); err != nil {
		return fmt.Errorf("invalid metrics format: %w", err)
	}

	ctx := context.Background()

	// Store metrics in database
	if h.metricsService != nil {
		if err := h.metricsService.StoreMetric(
			ctx,
			session.AgentID,
			metrics.CPU,
			int64(metrics.Memory.Total),   //nolint:gosec // safe conversion for memory values
			int64(metrics.Memory.Used),    //nolint:gosec // safe conversion for memory values
			metrics.Memory.Percent,
			int64(metrics.Disk.Total),     //nolint:gosec // safe conversion for disk values
			int64(metrics.Disk.Used),      //nolint:gosec // safe conversion for disk values
			metrics.Disk.Percent,
			int64(metrics.Uptime),         //nolint:gosec // safe conversion for uptime
		); err != nil {
			h.logger.Errorw("Failed to store metrics", "error", err, "agent_id", session.AgentID)
		}
	}

	// Publish metrics event
	_ = h.mq.PublishEvent(ctx, mq.EventAgentHeartbeat, map[string]interface{}{
		"agent_id":   session.AgentID,
		"agent_name": session.AgentName,
		"metrics":    metrics,
		"timestamp":  time.Now().Unix(),
	})

	// Send ack
	response := map[string]interface{}{
		"type": TypeMetrics,
		"data": map[string]string{
			"status": "received",
		},
	}
	h.sendMessage(session, response)

	return nil
}

// handleResult handles command result messages.
func (h *WebSocketHandler) handleResult(session *Session, data json.RawMessage) error {
	if session.State != AuthStateAuthenticated {
		return fmt.Errorf("not authenticated")
	}

	var result struct {
		CommandID string  `json:"command_id"`
		Status    string  `json:"status"`
		ExitCode  int     `json:"exit_code"`
		Output    string  `json:"output"`
		Duration  float64 `json:"duration"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return fmt.Errorf("invalid result format: %w", err)
	}

	ctx := context.Background()

	// Store result in database via task service
	if h.taskService != nil {
		if err := h.taskService.UpdateTaskResult(ctx, result.CommandID, result.Status, result.ExitCode, result.Output, result.Duration); err != nil {
			h.logger.Errorw("Failed to update task result", "error", err, "command_id", result.CommandID)
		}
	}

	// Determine event type based on status
	var eventType string
	switch result.Status {
	case "success":
		eventType = mq.EventTaskCompleted
	case "failed", "timeout":
		eventType = mq.EventTaskFailed
	default:
		eventType = mq.EventTaskCompleted
	}

	// Publish result event
	_ = h.mq.PublishEvent(ctx, eventType, map[string]interface{}{
		"agent_id":   session.AgentID,
		"command_id": result.CommandID,
		"status":     result.Status,
		"exit_code":  result.ExitCode,
		"output":     result.Output,
		"duration":   result.Duration,
		"timestamp":  time.Now().Unix(),
	})

	h.logger.Infow("Task result received",
		"agent_id", session.AgentID,
		"command_id", result.CommandID,
		"status", result.Status,
		"exit_code", result.ExitCode,
	)

	return nil
}

// SendCommand sends a command to an agent.
func (h *WebSocketHandler) SendCommand(agentID string, command map[string]interface{}) error {
	value, ok := h.sessions.Load(agentID)
	if !ok {
		return fmt.Errorf("agent not connected")
	}

	session, ok := value.(*Session)
	if !ok {
		return fmt.Errorf("invalid session type")
	}
	if session.State != AuthStateAuthenticated {
		return fmt.Errorf("agent not authenticated")
	}

	msg := map[string]interface{}{
		"type": TypeCommand,
		"data": command,
	}
	h.sendMessage(session, msg)

	return nil
}

// sendMessage sends a message to the session.
func (h *WebSocketHandler) sendMessage(session *Session, msg interface{}) {
	data, err := json.Marshal(msg)
	if err != nil {
		h.logger.Errorw("Failed to marshal message", "error", err)
		return
	}
	session.Send(data)
}

// sendAuthResult sends authentication result.
func (h *WebSocketHandler) sendAuthResult(session *Session, success bool, message string) {
	response := map[string]interface{}{
		"type": TypeAuthResult,
		"data": map[string]interface{}{
			"success": success,
			"message": message,
		},
	}
	h.sendMessage(session, response)
}

// sendError sends an error message.
func (h *WebSocketHandler) sendError(session *Session, message string) {
	response := map[string]interface{}{
		"type": TypeError,
		"data": map[string]string{
			"message": message,
		},
	}
	h.sendMessage(session, response)
}

// cleanup cleans up session resources.
func (h *WebSocketHandler) cleanup(session *Session) {
	if session.AgentID != "" {
		// Remove from sessions map
		h.sessions.Delete(session.AgentID)

		// Delete session from Redis
		ctx := context.Background()
		sessionKey := fmt.Sprintf("agent:session:%s", session.AgentID)
		_ = h.cache.Delete(ctx, sessionKey)

		// Update agent status to offline
		_ = h.agentService.UpdateAgentStatus(ctx, session.AgentID, domain.StatusOffline)

		// Publish offline event
		_ = h.mq.PublishEvent(ctx, mq.EventAgentOffline, map[string]interface{}{
			"agent_id":   session.AgentID,
			"agent_name": session.AgentName,
			"timestamp":  time.Now().Unix(),
		})

		h.logger.Infow("Agent disconnected", "agent_id", session.AgentID)
	}

	session.Conn.Close()
}

// generateNonce generates a random nonce.
func generateNonce() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// hmacSHA256 computes HMAC-SHA256.
func hmacSHA256(key, data string) string {
	h := hmac.New(sha256.New, []byte(key))
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}

// IsAgentConnected checks if an agent is connected.
func (h *WebSocketHandler) IsAgentConnected(agentID string) bool {
	_, ok := h.sessions.Load(agentID)
	return ok
}

// GetConnectedAgents returns list of connected agent IDs.
func (h *WebSocketHandler) GetConnectedAgents() []string {
	var agents []string
	h.sessions.Range(func(key, value interface{}) bool {
		if agentID, ok := key.(string); ok {
			agents = append(agents, agentID)
		}
		return true
	})
	return agents
}

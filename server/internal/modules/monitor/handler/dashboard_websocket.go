// Package handler provides HTTP handlers for monitoring.
package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/agentteams/server/internal/modules/auth/service"
	"github.com/agentteams/server/internal/pkg/logger"
	"github.com/agentteams/server/internal/pkg/mq"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// Dashboard WebSocket message types
const (
	DashboardMsgTypeAgentStatus  = "agent_status"
	DashboardMsgTypeMetrics      = "metrics"
	DashboardMsgTypeAlert        = "alert"
	DashboardMsgTypeDashboard    = "dashboard"
	DashboardMsgTypeConnected    = "connected"
	DashboardMsgTypeError        = "error"
)

// DashboardClient represents a connected dashboard client.
type DashboardClient struct {
	ID       string
	UserID   string
	Username string
	Conn     *websocket.Conn
	Send     chan []byte
	Close    chan struct{}
	mu       sync.RWMutex
}

// NewDashboardClient creates a new dashboard client.
func NewDashboardClient(conn *websocket.Conn, userID, username string) *DashboardClient {
	return &DashboardClient{
		ID:       generateClientID(),
		UserID:   userID,
		Username: username,
		Conn:     conn,
		Send:     make(chan []byte, 256),
		Close:    make(chan struct{}),
	}
}

// DashboardWSHandler handles WebSocket connections from the dashboard frontend.
type DashboardWSHandler struct {
	jwtService      *service.JWTService
	mq              *mq.Client
	logger          *logger.Logger
	upgrader        websocket.Upgrader
	clients         sync.Map // map[string]*DashboardClient
	broadcast       chan []byte
	register        chan *DashboardClient
	unregister      chan *DashboardClient
	metricsBuffer   *MetricsBuffer
	statsProvider   StatsProvider
	alertProvider   AlertProvider
	stopCh          chan struct{}
	wg              sync.WaitGroup
}

// StatsProvider provides dashboard statistics.
type StatsProvider interface {
	GetDashboardStats(ctx context.Context) (interface{}, error)
}

// AlertProvider provides alert information.
type AlertProvider interface {
	GetPendingCount(ctx context.Context) (int64, error)
}

// NewDashboardWSHandler creates a new dashboard WebSocket handler.
func NewDashboardWSHandler(jwtService *service.JWTService, mq *mq.Client, log *logger.Logger) *DashboardWSHandler {
	h := &DashboardWSHandler{
		jwtService: jwtService,
		mq:         mq,
		logger:     log,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // Configure appropriately for production
			},
			ReadBufferSize:  1024,
			WriteBufferSize: 4096,
		},
		broadcast:     make(chan []byte, 100),
		register:      make(chan *DashboardClient, 10),
		unregister:    make(chan *DashboardClient, 10),
		metricsBuffer: NewMetricsBuffer(),
		stopCh:        make(chan struct{}),
	}
	return h
}

// SetStatsProvider sets the stats provider.
func (h *DashboardWSHandler) SetStatsProvider(p StatsProvider) {
	h.statsProvider = p
}

// SetAlertProvider sets the alert provider.
func (h *DashboardWSHandler) SetAlertProvider(p AlertProvider) {
	h.alertProvider = p
}

// Start starts the dashboard WebSocket handler.
func (h *DashboardWSHandler) Start(ctx context.Context) {
	// Start hub goroutine
	h.wg.Add(1)
	go h.runHub(ctx)

	// Start periodic stats push
	h.wg.Add(1)
	go h.pushStats(ctx)

	// Start MQ subscriber
	h.wg.Add(1)
	go h.subscribeEvents(ctx)
}

// Stop stops the dashboard WebSocket handler.
func (h *DashboardWSHandler) Stop() {
	close(h.stopCh)
	h.wg.Wait()
}

// runHub handles client registration and broadcasting.
func (h *DashboardWSHandler) runHub(ctx context.Context) {
	defer h.wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		case <-h.stopCh:
			return
		case client := <-h.register:
			h.clients.Store(client.ID, client)
			h.logger.Infow("Dashboard client connected", "client_id", client.ID, "user", client.Username)

			// Send connected message
			h.sendToClient(client, map[string]interface{}{
				"type": DashboardMsgTypeConnected,
				"data": map[string]interface{}{
					"client_id": client.ID,
					"timestamp": time.Now().Unix(),
				},
			})

		case client := <-h.unregister:
			if _, loaded := h.clients.LoadAndDelete(client.ID); loaded {
				close(client.Send)
				h.logger.Infow("Dashboard client disconnected", "client_id", client.ID)
			}

		case message := <-h.broadcast:
			h.clients.Range(func(key, value interface{}) bool {
				client, ok := value.(*DashboardClient)
				if !ok {
					return true
				}
				select {
				case client.Send <- message:
				default:
					// Client buffer full, close connection
					close(client.Send)
					h.clients.Delete(key)
				}
				return true
			})
		}
	}
}

// pushStats periodically pushes dashboard statistics to all clients.
func (h *DashboardWSHandler) pushStats(ctx context.Context) {
	defer h.wg.Done()

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-h.stopCh:
			return
		case <-ticker.C:
			if h.statsProvider == nil {
				continue
			}

			stats, err := h.statsProvider.GetDashboardStats(ctx)
			if err != nil {
				h.logger.Errorw("Failed to get dashboard stats", "error", err)
				continue
			}

			h.broadcast <- mustMarshal(map[string]interface{}{
				"type": DashboardMsgTypeDashboard,
				"data": stats,
			})
		}
	}
}

// subscribeEvents subscribes to message queue events and broadcasts to clients.
func (h *DashboardWSHandler) subscribeEvents(ctx context.Context) {
	defer h.wg.Done()

	if h.mq == nil {
		return
	}

	// Subscribe to relevant events
	events := []string{
		mq.EventAgentOnline,
		mq.EventAgentOffline,
		mq.EventAgentHeartbeat,
		"alert.triggered",
		"alert.resolved",
		mq.EventTaskCompleted,
		mq.EventTaskFailed,
	}

	for _, event := range events {
		go h.subscribeEvent(ctx, event)
	}

	<-ctx.Done()
}

// subscribeEvent subscribes to a specific event type.
func (h *DashboardWSHandler) subscribeEvent(ctx context.Context, eventType string) {
	// Create a unique queue name for this subscription
	queueName := fmt.Sprintf("dashboard.%s.%s", eventType, randomString(8))

	deliveries, err := h.mq.ConsumeEvents(ctx, queueName, []string{eventType})
	if err != nil {
		h.logger.Errorw("Failed to subscribe to event", "event", eventType, "error", err)
		return
	}

	for {
		select {
		case <-ctx.Done():
			return
		case <-h.stopCh:
			return
		case d, ok := <-deliveries:
			if !ok {
				return
			}

			// Acknowledge message
			_ = h.mq.Ack(d.DeliveryTag)

			// Parse message
			var msg map[string]interface{}
			if err := json.Unmarshal(d.Body, &msg); err != nil {
				continue
			}

			// Extract data from event
			data := msg
			if dataField, ok := msg["data"].(map[string]interface{}); ok {
				data = dataField
			}

			h.handleEvent(eventType, data)
		}
	}
}

// handleEvent handles an event from the message queue.
func (h *DashboardWSHandler) handleEvent(eventType string, msg map[string]interface{}) {
	var msgType string
	var data interface{} = msg

	switch eventType {
	case mq.EventAgentOnline, mq.EventAgentOffline:
		msgType = DashboardMsgTypeAgentStatus
	case mq.EventAgentHeartbeat:
		msgType = DashboardMsgTypeMetrics
		// Buffer metrics and aggregate
		if agentID, ok := msg["agent_id"].(string); ok {
			h.metricsBuffer.Add(agentID, msg)
			// Only broadcast if buffer is ready
			if !h.metricsBuffer.ShouldFlush() {
				return
			}
			data = h.metricsBuffer.Flush()
		}
	case "alert.triggered", "alert.resolved":
		msgType = DashboardMsgTypeAlert
	case mq.EventTaskCompleted, mq.EventTaskFailed:
		// Task events can be broadcast as dashboard updates
		msgType = DashboardMsgTypeDashboard
	default:
		return
	}

	h.broadcast <- mustMarshal(map[string]interface{}{
		"type":      msgType,
		"data":      data,
		"timestamp": time.Now().Unix(),
	})
}

// Handle handles WebSocket connection upgrade.
func (h *DashboardWSHandler) Handle(c *gin.Context) {
	// Get JWT token from query parameter or header
	token := c.Query("token")
	if token == "" {
		token = c.GetHeader("Authorization")
		if len(token) > 7 && token[:7] == "Bearer " {
			token = token[7:]
		}
	}

	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing token"})
		return
	}

	// Validate JWT
	claims, err := h.jwtService.ValidateToken(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
		return
	}

	// Upgrade to WebSocket
	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		h.logger.Errorw("Failed to upgrade connection", "error", err)
		return
	}

	client := NewDashboardClient(conn, claims.UserID, claims.Username)

	// Register client
	h.register <- client

	// Start read and write goroutines
	go h.writePump(client)
	h.readPump(client)

	// Unregister on disconnect
	h.unregister <- client
}

// readPump reads messages from the WebSocket connection.
func (h *DashboardWSHandler) readPump(client *DashboardClient) {
	defer func() {
		client.Conn.Close()
	}()

	for {
		_, message, err := client.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				h.logger.Errorw("WebSocket read error", "error", err, "client_id", client.ID)
			}
			break
		}

		// Handle incoming message (ping/pong, subscriptions, etc.)
		var msg struct {
			Type string `json:"type"`
		}
		if err := json.Unmarshal(message, &msg); err != nil {
			continue
		}

		// Handle ping
		if msg.Type == "ping" {
			h.sendToClient(client, map[string]interface{}{
				"type": "pong",
				"data": map[string]int64{
					"timestamp": time.Now().Unix(),
				},
			})
		}
	}
}

// writePump writes messages to the WebSocket connection.
func (h *DashboardWSHandler) writePump(client *DashboardClient) {
	ticker := time.NewTicker(30 * time.Second)
	defer func() {
		ticker.Stop()
		client.Conn.Close()
	}()

	for {
		select {
		case <-client.Close:
			return
		case message, ok := <-client.Send:
			if !ok {
				_ = client.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := client.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
				h.logger.Errorw("WebSocket write error", "error", err, "client_id", client.ID)
				return
			}
		case <-ticker.C:
			// Send ping for keepalive
			if err := client.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// sendToClient sends a message to a specific client.
func (h *DashboardWSHandler) sendToClient(client *DashboardClient, msg interface{}) {
	data, err := json.Marshal(msg)
	if err != nil {
		return
	}
	select {
	case client.Send <- data:
	default:
		// Buffer full
	}
}

// GetClientCount returns the number of connected clients.
func (h *DashboardWSHandler) GetClientCount() int {
	count := 0
	h.clients.Range(func(key, value interface{}) bool {
		count++
		return true
	})
	return count
}

// MetricsBuffer buffers metrics for aggregation.
type MetricsBuffer struct {
	mu        sync.Mutex
	metrics   map[string]interface{}
	lastFlush time.Time
	count     int
}

// NewMetricsBuffer creates a new metrics buffer.
func NewMetricsBuffer() *MetricsBuffer {
	return &MetricsBuffer{
		metrics:   make(map[string]interface{}),
		lastFlush: time.Now(),
	}
}

// Add adds metrics to the buffer.
func (b *MetricsBuffer) Add(agentID string, metrics interface{}) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.metrics[agentID] = metrics
	b.count++
}

// ShouldFlush returns true if the buffer should be flushed.
func (b *MetricsBuffer) ShouldFlush() bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	// Flush every 5 seconds or 100 metrics
	return time.Since(b.lastFlush) >= 5*time.Second || b.count >= 100
}

// Flush returns all buffered metrics and resets the buffer.
func (b *MetricsBuffer) Flush() map[string]interface{} {
	b.mu.Lock()
	defer b.mu.Unlock()
	result := b.metrics
	b.metrics = make(map[string]interface{})
	b.lastFlush = time.Now()
	b.count = 0
	return result
}

// Count returns the current number of buffered metrics.
func (b *MetricsBuffer) Count() int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.count
}

// generateClientID generates a unique client ID.
func generateClientID() string {
	return time.Now().Format("20060102150405") + "-" + randomString(8)
}

// randomString generates a random string of given length.
func randomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[time.Now().Nanosecond()%len(letters)]
	}
	return string(b)
}

// mustMarshal marshals to JSON or returns empty object.
func mustMarshal(v interface{}) []byte {
	data, err := json.Marshal(v)
	if err != nil {
		return []byte("{}")
	}
	return data
}

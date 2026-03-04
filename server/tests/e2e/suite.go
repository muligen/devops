// Package e2e provides end-to-end testing infrastructure
package e2e

import (
	"context"
	"fmt"
	"net/http/httptest"
	"strings"
	"time"

	agentdomain "github.com/agentteams/server/internal/modules/agent/domain"
	"github.com/agentteams/server/internal/modules/agent/handler"
	agentservice "github.com/agentteams/server/internal/modules/agent/service"
	authdomain "github.com/agentteams/server/internal/modules/auth/domain"
	authhandler "github.com/agentteams/server/internal/modules/auth/handler"
	authservice "github.com/agentteams/server/internal/modules/auth/service"
	monitordomain "github.com/agentteams/server/internal/modules/monitor/domain"
	"github.com/agentteams/server/internal/pkg/cache"
	"github.com/agentteams/server/internal/pkg/database"
	"github.com/agentteams/server/internal/pkg/logger"
	"github.com/agentteams/server/internal/pkg/mq"
	taskhandler "github.com/agentteams/server/internal/modules/task/handler"
	taskservice "github.com/agentteams/server/internal/modules/task/service"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// TestSuite provides a complete test environment
type TestSuite struct {
	DB         *gorm.DB
	Containers *TestContainers
	Config     *TestConfig
	Factory    *Factory
	Fixtures   *FixtureLoader
	Server     *httptest.Server
	Router     *gin.Engine
	JWTService *authservice.JWTService
	Cache      *cache.Client
	MQ         *mq.Client
}

// SetupTestSuite initializes the complete test environment
func SetupTestSuite(ctx context.Context) (*TestSuite, error) {
	// Initialize logger
	logConfig := &logger.Config{Level: "debug", Format: "console", Output: "stdout"}
	if err := logger.Init(logConfig); err != nil {
		return nil, fmt.Errorf("failed to init logger: %w", err)
	}

	// Start containers
	containers, testConfig, err := StartContainers(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to start containers: %w", err)
	}

	// Initialize database using the app's database package
	dbConfig := &database.Config{
		Host:            extractHost(testConfig.DatabaseURL),
		Port:            extractPort(testConfig.DatabaseURL),
		User:            "test",
		Password:        "test",
		Name:            "agentteams_test",
		SSLMode:         "disable",
		MaxOpenConns:    10,
		MaxIdleConns:    5,
		ConnMaxLifetime: time.Hour,
	}

	databaseObj, err := database.New(dbConfig)
	if err != nil {
		_ = StopContainers(containers)
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}
	db := databaseObj.DB

	// Run migrations
	if err := runMigrations(db); err != nil {
		_ = StopContainers(containers)
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	// Initialize Redis cache
	cacheClient, err := cache.New(&cache.Config{
		Host: extractRedisHost(testConfig.RedisAddr),
		Port: extractRedisPort(testConfig.RedisAddr),
	})
	if err != nil {
		_ = StopContainers(containers)
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	// Initialize RabbitMQ
	mqClient, err := mq.New(&mq.Config{
		Host:     extractRabbitHost(testConfig.RabbitMQURL),
		Port:     extractRabbitPort(testConfig.RabbitMQURL),
		User:     "guest",
		Password: "guest",
		VHost:    "/",
	})
	if err != nil {
		_ = StopContainers(containers)
		return nil, fmt.Errorf("failed to connect to rabbitmq: %w", err)
	}

	// Initialize JWT service
	jwtService := authservice.NewJWTService(authservice.JWTConfig{
		Secret:        "test-secret-key-for-e2e-testing-only",
		Expiry:        time.Hour,
		RefreshExpiry: 24 * time.Hour,
	})

	suite := &TestSuite{
		DB:         db,
		Containers: containers,
		Config:     testConfig,
		Factory:    NewFactory(db),
		Fixtures:   NewFixtureLoader(db),
		JWTService: jwtService,
		Cache:      cacheClient,
		MQ:         mqClient,
	}

	// Setup HTTP server
	suite.setupRouter()

	return suite, nil
}

// setupRouter configures the Gin router with all routes
func (s *TestSuite) setupRouter() {
	gin.SetMode(gin.TestMode)
	s.Router = gin.New()

	// Create services
	agentRepo := agentservice.NewRepository(s.DB)
	agentSvc := agentservice.NewService(agentRepo)

	taskRepo := taskservice.NewRepository(s.DB)
	taskSvc := taskservice.NewService(taskRepo)

	authRepo := authservice.NewRepository(s.DB)
	authSvc := authservice.NewService(authRepo)

	// Create handlers
	authHandler := authhandler.NewHandler(authSvc, s.JWTService)
	agentHandler := handler.NewHandler(agentSvc)
	wsHandler := handler.NewWebSocketHandler(agentSvc, s.Cache, s.MQ, logger.NewNop())
	taskHandler := taskhandler.NewHandler(taskSvc)

	wsHandler.SetTaskService(taskSvc)

	// Public routes
	public := s.Router.Group("/api/v1")
	{
		public.POST("/auth/login", authHandler.Login)
		public.POST("/auth/refresh", authHandler.RefreshToken)
	}

	// Protected routes
	protected := s.Router.Group("/api/v1")
	protected.Use(authMiddleware(s.JWTService))
	{
		// Agents
		protected.GET("/agents", agentHandler.ListAgents)
		protected.GET("/agents/:id", agentHandler.GetAgent)
		protected.POST("/agents", agentHandler.CreateAgent)
		protected.DELETE("/agents/:id", agentHandler.DeleteAgent)

		// Tasks
		protected.POST("/tasks", taskHandler.CreateTask)
		protected.POST("/tasks/batch", taskHandler.BatchCreateTasks)
		protected.GET("/tasks", taskHandler.ListTasks)
		protected.GET("/tasks/:id", taskHandler.GetTask)
		protected.DELETE("/tasks/:id", taskHandler.CancelTask)
	}

	// WebSocket endpoint
	s.Router.GET("/ws", wsHandler.Handle)

	// Create test server
	s.Server = httptest.NewServer(s.Router)
}

// authMiddleware creates an authentication middleware for testing
func authMiddleware(jwtSvc *authservice.JWTService) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(401, gin.H{"error": "missing authorization header"})
			c.Abort()
			return
		}

		// Extract Bearer token
		if len(authHeader) < 7 || authHeader[:7] != "Bearer " {
			c.JSON(401, gin.H{"error": "invalid authorization header"})
			c.Abort()
			return
		}

		token := authHeader[7:]
		claims, err := jwtSvc.ValidateToken(token)
		if err != nil {
			c.JSON(401, gin.H{"error": "invalid token"})
			c.Abort()
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("role", claims.Role)
		c.Next()
	}
}

// Teardown cleans up the test environment
func (s *TestSuite) Teardown() error {
	if s.Server != nil {
		s.Server.Close()
	}
	if s.MQ != nil {
		s.MQ.Close()
	}
	if s.Cache != nil {
		s.Cache.Close()
	}
	return StopContainers(s.Containers)
}

// LoadFixtures loads all test fixtures
func (s *TestSuite) LoadFixtures(ctx context.Context) error {
	return s.Fixtures.LoadAll(ctx)
}

// CleanupData cleans all test data
func (s *TestSuite) CleanupData(ctx context.Context) error {
	return s.Fixtures.Cleanup(ctx)
}

// GenerateToken generates a valid JWT token for testing
func (s *TestSuite) GenerateToken(userID, username, role string) (string, error) {
	return s.JWTService.GenerateToken(userID, username, role)
}

// AdminToken returns a valid admin token
func (s *TestSuite) AdminToken() string {
	token, _ := s.GenerateToken(TestUsers.Admin.ID, TestUsers.Admin.Username, TestUsers.Admin.Role)
	return token
}

// OperatorToken returns a valid operator token
func (s *TestSuite) OperatorToken() string {
	token, _ := s.GenerateToken(TestUsers.Operator.ID, TestUsers.Operator.Username, TestUsers.Operator.Role)
	return token
}

// ViewerToken returns a valid viewer token
func (s *TestSuite) ViewerToken() string {
	token, _ := s.GenerateToken(TestUsers.Viewer.ID, TestUsers.Viewer.Username, TestUsers.Viewer.Role)
	return token
}

// runMigrations runs auto migrations
func runMigrations(db *gorm.DB) error {
	return db.AutoMigrate(
		&authdomain.User{},
		&agentdomain.Agent{},
		&monitordomain.AgentMetric{},
	)
}

// extractRabbitHost extracts host from RabbitMQ URL
func extractRabbitHost(url string) string {
	// URL format: amqp://guest:guest@host:port/
	parts := strings.Split(url, "@")
	if len(parts) > 1 {
		hostPort := strings.Split(parts[1], ":")
		if len(hostPort) > 0 {
			return hostPort[0]
		}
	}
	return "localhost"
}

// extractRabbitPort extracts port from RabbitMQ URL
func extractRabbitPort(url string) int {
	// URL format: amqp://guest:guest@host:port/
	parts := strings.Split(url, "@")
	if len(parts) > 1 {
		hostPortVhost := strings.Split(parts[1], "/")
		if len(hostPortVhost) > 0 {
			hostPort := strings.Split(hostPortVhost[0], ":")
			if len(hostPort) > 1 {
				var port int
				fmt.Sscanf(hostPort[1], "%d", &port)
				return port
			}
		}
	}
	return 5672
}

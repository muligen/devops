// Package test provides testing utilities for the server.
package test

import (
	"context"
	"fmt"
	"net/http/httptest"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	agentHandler "github.com/agentteams/server/internal/modules/agent/handler"
	agentService "github.com/agentteams/server/internal/modules/agent/service"
	authHandler "github.com/agentteams/server/internal/modules/auth/handler"
	authService "github.com/agentteams/server/internal/modules/auth/service"
	taskHandler "github.com/agentteams/server/internal/modules/task/handler"
	taskService "github.com/agentteams/server/internal/modules/task/service"
	"github.com/agentteams/server/internal/pkg/cache"
	"github.com/agentteams/server/internal/pkg/middleware"
)

// TestConfig holds test configuration.
type TestConfig struct {
	DatabaseHost     string
	DatabasePort     int
	DatabaseUser     string
	DatabasePassword string
	DatabaseName     string
	RedisHost        string
	RedisPort        int
	JWTSecret        string
}

// DefaultTestConfig returns default test configuration.
func DefaultTestConfig() *TestConfig {
	return &TestConfig{
		DatabaseHost:     "localhost",
		DatabasePort:     5433,
		DatabaseUser:     "test",
		DatabasePassword: "test",
		DatabaseName:     "agentteams_test",
		RedisHost:        "localhost",
		RedisPort:        6380,
		JWTSecret:        "test-secret-key-for-testing-only",
	}
}

// TestServer holds test server dependencies.
type TestServer struct {
	DB         *gorm.DB
	Redis      *cache.Client
	Router     *gin.Engine
	JWTService *authService.JWTService
	Server     *httptest.Server
	authH      *authHandler.Handler
}

// SetupTestServer creates a test server with all dependencies.
func SetupTestServer(cfg *TestConfig) (*TestServer, error) {
	if cfg == nil {
		cfg = DefaultTestConfig()
	}

	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Connect to database
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		cfg.DatabaseHost, cfg.DatabasePort, cfg.DatabaseUser, cfg.DatabasePassword, cfg.DatabaseName,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Connect to Redis
	redis, err := cache.New(&cache.Config{
		Host:     cfg.RedisHost,
		Port:     cfg.RedisPort,
		Password: "",
		DB:       1,
		PoolSize: 5,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	// Create JWT service
	jwtService := authService.NewJWTService(authService.JWTConfig{
		Secret:        cfg.JWTSecret,
		Expiry:        time.Hour,
		RefreshExpiry: 24 * time.Hour,
	})

	// Initialize services
	authRepo := authService.NewRepository(db)
	authSvc := authService.NewService(authRepo)
	authH := authHandler.NewHandler(authSvc, jwtService)

	agentRepo := agentService.NewRepository(db)
	agentSvc := agentService.NewService(agentRepo)
	agentH := agentHandler.NewHandler(agentSvc)

	taskRepo := taskService.NewRepository(db)
	taskSvc := taskService.NewService(taskRepo)
	taskH := taskHandler.NewHandler(taskSvc)

	// Setup router
	router := gin.New()
	router.Use(gin.Recovery())

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		authH.RegisterRoutes(v1)

		// User creation routes (admin only)
		users := v1.Group("/auth")
		users.Use(middleware.AuthMiddleware(jwtService))
		{
			users.POST("/users", authH.CreateUser)
			users.GET("/users", authH.ListUsers)
		}

		protected := v1.Group("")
		protected.Use(middleware.AuthMiddleware(jwtService))
		{
			agentH.RegisterRoutes(protected)
			taskH.RegisterRoutes(protected)
		}
	}

	return &TestServer{
		DB:         db,
		Redis:      redis,
		Router:     router,
		JWTService: jwtService,
		Server:     httptest.NewServer(router),
		authH:      authH,
	}, nil
}

// Cleanup cleans up test resources.
func (s *TestServer) Cleanup() {
	if s.Server != nil {
		s.Server.Close()
	}
	if s.Redis != nil {
		s.Redis.Close()
	}

	// Clean up database tables
	if s.DB != nil {
		sqlDB, _ := s.DB.DB()
		if sqlDB != nil {
			sqlDB.Close()
		}
	}
}

// CleanDatabase truncates all tables.
func (s *TestServer) CleanDatabase() error {
	tables := []string{
		"tasks",
		"agents",
		"users",
		"agent_metrics",
		"alert_events",
	}

	for _, table := range tables {
		if err := s.DB.Exec(fmt.Sprintf("TRUNCATE TABLE %s CASCADE", table)).Error; err != nil {
			// Ignore error if table doesn't exist
			continue
		}
	}

	return nil
}

// CreateTestUser creates a test user and returns the user ID.
func (s *TestServer) CreateTestUser(username, password, role string) (string, error) {
	ctx := context.Background()

	authRepo := authService.NewRepository(s.DB)
	authSvc := authService.NewService(authRepo)

	user, err := authSvc.CreateUser(ctx, username, password, username+"@test.com", role)
	if err != nil {
		return "", err
	}

	return user.ID, nil
}

// GenerateTestToken generates a JWT token for testing.
func (s *TestServer) GenerateTestToken(userID, username, role string) (string, error) {
	return s.JWTService.GenerateToken(userID, username, role)
}

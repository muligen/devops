package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	agentHandler "github.com/agentteams/server/internal/modules/agent/handler"
	agentService "github.com/agentteams/server/internal/modules/agent/service"
	authHandler "github.com/agentteams/server/internal/modules/auth/handler"
	authService "github.com/agentteams/server/internal/modules/auth/service"
	monitorHandler "github.com/agentteams/server/internal/modules/monitor/handler"
	monitorService "github.com/agentteams/server/internal/modules/monitor/service"
	taskHandler "github.com/agentteams/server/internal/modules/task/handler"
	taskService "github.com/agentteams/server/internal/modules/task/service"
	"github.com/agentteams/server/internal/infrastructure/migrate"
	"github.com/agentteams/server/internal/pkg/cache"
	"github.com/agentteams/server/internal/pkg/config"
	"github.com/agentteams/server/internal/pkg/database"
	"github.com/agentteams/server/internal/pkg/logger"
	"github.com/agentteams/server/internal/pkg/middleware"
	"github.com/agentteams/server/internal/pkg/mq"
)

// Version information (set via ldflags)
var (
	Version   = "dev"
	BuildDate = "unknown"
	GitCommit = "unknown"
)

// Application holds all dependencies
type Application struct {
	config     *config.Config
	logger     *logger.Logger
	db         *database.Database
	redis      *cache.Client
	rabbitMQ   *mq.Client
	jwtService *authService.JWTService
}

func main() {
	// Parse command line flags
	configPath := flag.String("config", "configs/config.yaml", "Path to configuration file")
	showVersion := flag.Bool("version", false, "Show version information")
	flag.Parse()

	if *showVersion {
		fmt.Printf("AgentTeams Server\n")
		fmt.Printf("Version:    %s\n", Version)
		fmt.Printf("Build Date: %s\n", BuildDate)
		fmt.Printf("Git Commit: %s\n", GitCommit)
		os.Exit(0)
	}

	// Load configuration
	cfg, err := config.Load(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	log, err := logger.New(logger.Config{
		Level:  cfg.Log.Level,
		Format: cfg.Log.Format,
		Output: cfg.Log.Output,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}

	log.Infow("Starting AgentTeams Server", "version", Version)

	// Initialize database
	db, err := database.New(&database.Config{
		Host:            cfg.Database.Host,
		Port:            cfg.Database.Port,
		User:            cfg.Database.User,
		Password:        cfg.Database.Password,
		Name:            cfg.Database.Name,
		SSLMode:         cfg.Database.SSLMode,
		MaxOpenConns:    cfg.Database.MaxOpenConns,
		MaxIdleConns:    cfg.Database.MaxIdleConns,
		ConnMaxLifetime: cfg.Database.ConnMaxLifetime,
	})
	if err != nil {
		log.Errorw("Failed to initialize database", "error", err)
		os.Exit(1)
	}
	log.Infow("Database connected")

	// Run migrations
	migrator, err := migrate.New(db.DB)
	if err != nil {
		log.Errorw("Failed to create migrator", "error", err)
		os.Exit(1)
	}
	if err := migrator.Up(); err != nil {
		log.Errorw("Failed to run migrations", "error", err)
		os.Exit(1)
	}
	log.Infow("Database migrations completed")

	// Initialize Redis
	redis, err := cache.New(&cache.Config{
		Host:     cfg.Redis.Host,
		Port:     cfg.Redis.Port,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
		PoolSize: cfg.Redis.PoolSize,
	})
	if err != nil {
		log.Errorw("Failed to initialize Redis", "error", err)
		os.Exit(1)
	}
	log.Infow("Redis connected")

	// Initialize RabbitMQ
	rabbitMQ, err := mq.New(&mq.Config{
		Host:     cfg.RabbitMQ.Host,
		Port:     cfg.RabbitMQ.Port,
		User:     cfg.RabbitMQ.User,
		Password: cfg.RabbitMQ.Password,
		VHost:    cfg.RabbitMQ.VHost,
	})
	if err != nil {
		log.Errorw("Failed to initialize RabbitMQ", "error", err)
		os.Exit(1)
	}
	log.Infow("RabbitMQ connected")

	// Initialize JWT service
	jwtService := authService.NewJWTService(authService.JWTConfig{
		Secret:        cfg.JWT.Secret,
		Expiry:        cfg.JWT.Expiry,
		RefreshExpiry: cfg.JWT.RefreshExpiry,
	})

	app := &Application{
		config:     cfg,
		logger:     log,
		db:         db,
		redis:      redis,
		rabbitMQ:   rabbitMQ,
		jwtService: jwtService,
	}

	// Setup Gin router
	router := app.setupRouter()

	// Start HTTP server
	srv := &http.Server{
		Addr:         cfg.Server.Addr,
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	// Start server in goroutine
	go func() {
		log.Infow("Server starting", "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Errorw("Failed to start server", "error", err)
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal for graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Infow("Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Errorw("Server forced to shutdown", "error", err)
	}

	// Close connections
	if rabbitMQ != nil {
		_ = rabbitMQ.Close()
	}
	if redis != nil {
		_ = redis.Close()
	}
	if db != nil {
		_ = db.Close()
	}

	log.Infow("Server exited")
}

func (app *Application) setupRouter() *gin.Engine {
	// Set Gin mode
	if app.config.Log.Level == "debug" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	// Middleware
	router.Use(gin.Recovery())
	router.Use(middleware.RequestLogger(app.logger))
	router.Use(middleware.CORS([]string{"*"}))

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"version": Version,
		})
	})

	// Initialize services
	authRepo := authService.NewRepository(app.db.DB)
	authSvc := authService.NewService(authRepo)
	authH := authHandler.NewHandler(authSvc, app.jwtService)

	agentRepo := agentService.NewRepository(app.db.DB)
	agentSvc := agentService.NewService(agentRepo)
	agentH := agentHandler.NewHandler(agentSvc)

	// Create WebSocket handler first (needed for task dispatching)
	wsHandler := agentHandler.NewWebSocketHandler(agentSvc, app.redis, app.rabbitMQ, app.logger)

	// Create dashboard WebSocket handler for frontend
	dashboardWSHandler := monitorHandler.NewDashboardWSHandler(app.jwtService, app.rabbitMQ, app.logger)

	// Initialize task service with queue and dispatcher
	taskRepo := taskService.NewRepository(app.db.DB)
	taskSvc := taskService.NewService(taskRepo)

	// Set up task queue adapter
	taskQueueAdapter := taskService.NewTaskQueueAdapter(app.rabbitMQ)
	taskSvc.SetQueue(taskQueueAdapter)

	// Set up task dispatcher adapter
	taskDispatcher := taskService.NewDispatcherAdapter(
		wsHandler.SendCommand,
		wsHandler.IsAgentConnected,
	)
	taskSvc.SetDispatcher(taskDispatcher)

	// Wire WebSocket handler to task service for result handling
	wsHandler.SetTaskService(taskSvc)

	// Initialize monitor service
	monitorRepo := monitorService.NewRepository(app.db.DB)
	monitorSvc := monitorService.NewService(monitorRepo)
	monitorSvc.SetMQ(app.rabbitMQ)

	// Initialize alert event repository
	alertEventRepo := monitorService.NewAlertEventRepository(app.db.DB)
	monitorSvc.SetAlertEventRepository(alertEventRepo)

	// Wire WebSocket handler to monitor service for metrics handling
	wsHandler.SetMetricsService(monitorSvc)

	// Wire monitor service to agent handler for metrics display
	agentH.SetLatestMetricsService(monitorSvc)

	monitorH := monitorHandler.NewHandler(monitorSvc)

	taskH := taskHandler.NewHandler(taskSvc)

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Public routes
		authH.RegisterRoutes(v1)

		// Agent WebSocket (public, handles own auth)
		v1.GET("/agent/ws", func(c *gin.Context) {
			wsHandler.Handle(c)
		})

		// Dashboard WebSocket for frontend (protected by JWT)
		v1.GET("/ws/dashboard", func(c *gin.Context) {
			dashboardWSHandler.Handle(c)
		})

		// Protected routes
		protected := v1.Group("")
		protected.Use(middleware.AuthMiddleware(app.jwtService))
		{
			agentH.RegisterRoutes(protected)
			taskH.RegisterRoutes(protected)
			monitorH.RegisterRoutes(protected)
		}
	}

	return router
}

// Package e2e provides end-to-end testing infrastructure
package e2e

import (
	"context"
	"fmt"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/modules/rabbitmq"
	"github.com/testcontainers/testcontainers-go/modules/redis"
	"github.com/testcontainers/testcontainers-go/wait"
)

// TestContainers holds all test container instances
type TestContainers struct {
	Postgres *postgres.PostgresContainer
	Redis    *redis.RedisContainer
	RabbitMQ *rabbitmq.RabbitMQContainer
}

// TestConfig holds test environment configuration
type TestConfig struct {
	DatabaseURL string
	RedisAddr   string
	RabbitMQURL string
	ServerPort  int
}

// StartContainers starts all required test containers
func StartContainers(ctx context.Context) (*TestContainers, *TestConfig, error) {
	// Start PostgreSQL
	pgContainer, err := postgres.Run(ctx, "postgres:15-alpine",
		postgres.WithDatabase("agentteams_test"),
		postgres.WithUsername("test"),
		postgres.WithPassword("test"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(60*time.Second),
		),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to start postgres: %w", err)
	}

	pgConnStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get postgres connection string: %w", err)
	}

	// Start Redis
	redisContainer, err := redis.Run(ctx, "redis:7-alpine",
		testcontainers.WithWaitStrategy(
			wait.ForLog("Ready to accept connections").
				WithStartupTimeout(30*time.Second),
		),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to start redis: %w", err)
	}

	redisHost, err := redisContainer.Host(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get redis host: %w", err)
	}
	redisPort, err := redisContainer.MappedPort(ctx, "6379")
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get redis port: %w", err)
	}

	// Start RabbitMQ
	rabbitContainer, err := rabbitmq.Run(ctx, "rabbitmq:3.12-alpine",
		testcontainers.WithWaitStrategy(
			wait.ForLog("Server startup complete").
				WithStartupTimeout(60*time.Second),
		),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to start rabbitmq: %w", err)
	}

	rabbitHost, err := rabbitContainer.Host(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get rabbitmq host: %w", err)
	}
	rabbitPort, err := rabbitContainer.MappedPort(ctx, "5672")
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get rabbitmq port: %w", err)
	}

	config := &TestConfig{
		DatabaseURL: pgConnStr,
		RedisAddr:   fmt.Sprintf("%s:%s", redisHost, redisPort.Port()),
		RabbitMQURL: fmt.Sprintf("amqp://guest:guest@%s:%s/", rabbitHost, rabbitPort.Port()),
		ServerPort:  8080,
	}

	containers := &TestContainers{
		Postgres: pgContainer,
		Redis:    redisContainer,
		RabbitMQ: rabbitContainer,
	}

	return containers, config, nil
}

// StopContainers stops all test containers
func StopContainers(containers *TestContainers) error {
	if containers == nil {
		return nil
	}

	var errs []error

	if containers.Postgres != nil {
		if err := testcontainers.TerminateContainer(containers.Postgres); err != nil {
			errs = append(errs, fmt.Errorf("failed to stop postgres: %w", err))
		}
	}

	if containers.Redis != nil {
		if err := testcontainers.TerminateContainer(containers.Redis); err != nil {
			errs = append(errs, fmt.Errorf("failed to stop redis: %w", err))
		}
	}

	if containers.RabbitMQ != nil {
		if err := testcontainers.TerminateContainer(containers.RabbitMQ); err != nil {
			errs = append(errs, fmt.Errorf("failed to stop rabbitmq: %w", err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors stopping containers: %v", errs)
	}

	return nil
}

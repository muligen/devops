package main

import (
	"context"
	"fmt"
	"log"

	agentService "github.com/agentteams/server/internal/modules/agent/service"
	"github.com/agentteams/server/internal/pkg/database"
)

func main() {
	// Connect to database
	db, err := database.New(&database.Config{
		Host:     "localhost",
		Port:     5432,
		User:     "postgres",
		Password: "postgres",
		Name:     "agentteams",
		SSLMode:  "disable",
	})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Create agent service
	repo := agentService.NewRepository(db.DB)
	svc := agentService.NewService(repo)

	// Check if agent already exists
	agentID := "b0381bda-fb22-474f-964e-137ce03b9b34"
	_, err = svc.GetAgent(context.Background(), agentID)
	if err == nil {
		fmt.Printf("Agent %s already exists\n", agentID)
		return
	}

	// Create new agent
	agent, token, err := svc.CreateAgent(context.Background(), "test-agent", nil)
	if err != nil {
		log.Fatalf("Failed to create agent: %v", err)
	}

	fmt.Printf("Agent created successfully!\n")
	fmt.Printf("ID: %s\n", agent.ID)
	fmt.Printf("Token: %s\n", token)
}

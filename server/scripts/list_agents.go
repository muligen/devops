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

	// List all agents
	agents, total, err := svc.ListAgents(context.Background(), 1, 10, "")
	if err != nil {
		log.Fatalf("Failed to list agents: %v", err)
	}

	fmt.Printf("Total agents: %d\n", total)
	for _, agent := range agents {
		fmt.Printf("  - ID: %s, Name: %s, Status: %s\n", agent.ID, agent.Name, agent.Status)
	}
}

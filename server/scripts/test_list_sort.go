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

	// Test ListWithSort with cpu_usage sort
	opts := agentService.ListOptions{
		Page:     1,
		PageSize: 12,
		Sort:     "cpu_usage",
		Order:    "desc",
	}

	fmt.Println("Testing ListWithSort with cpu_usage sort...")
	agents, total, err := repo.ListWithSort(context.Background(), opts)
	if err != nil {
		log.Fatalf("ListWithSort failed: %v", err)
	}

	fmt.Printf("Total: %d, Agents: %d\n", total, len(agents))
	for _, a := range agents {
		fmt.Printf("  - ID: %s, Name: %s, Status: %s\n", a.ID, a.Name, a.Status)
	}
}

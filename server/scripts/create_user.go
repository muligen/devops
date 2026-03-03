package main

import (
	"context"
	"fmt"
	"log"

	authService "github.com/agentteams/server/internal/modules/auth/service"
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

	// Create auth service
	repo := authService.NewRepository(db.DB)
	svc := authService.NewService(repo)

	// Create test user
	username := "admin"
	password := "admin123"
	email := "admin@example.com"

	user, err := svc.CreateUser(context.Background(), username, password, email, "admin")
	if err != nil {
		log.Printf("Failed to create user (may already exist): %v", err)
		// Try to list users
		users, _, err := svc.ListUsers(context.Background(), 1, 10)
		if err != nil {
			log.Fatalf("Failed to list users: %v", err)
		}
		fmt.Println("Existing users:")
		for _, u := range users {
			fmt.Printf("  - Username: %s, Email: %s, Role: %s\n", u.Username, u.Email, u.Role)
		}
		return
	}

	fmt.Printf("User created successfully!\n")
	fmt.Printf("Username: %s\n", user.Username)
	fmt.Printf("Email: %s\n", user.Email)
	fmt.Printf("Role: %s\n", user.Role)
}

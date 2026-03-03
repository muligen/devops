package main

import (
	"fmt"
	"log"

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

	// Check tables
	tables := []string{"alert_rules", "alert_events"}
	for _, table := range tables {
		var exists bool
		db.Raw("SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = ?)", table).Scan(&exists)
		fmt.Printf("%s exists: %v\n", table, exists)
	}

	// Create alert_rules table if not exists
	db.Exec(`
		CREATE TABLE IF NOT EXISTS alert_rules (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			name VARCHAR(100) NOT NULL,
			description TEXT,
			metric_type VARCHAR(50) NOT NULL,
			condition VARCHAR(20) NOT NULL,
			threshold FLOAT NOT NULL,
			duration_seconds INT DEFAULT 60,
			severity VARCHAR(20) NOT NULL DEFAULT 'warning',
			enabled BOOLEAN DEFAULT true,
			agent_id UUID REFERENCES agents(id) ON DELETE CASCADE,
			created_by UUID REFERENCES users(id) ON DELETE SET NULL,
			created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
			CONSTRAINT chk_severity CHECK (severity IN ('info', 'warning', 'critical'))
		)
	`)
	fmt.Println("Created alert_rules table (if not exists)")

	// Create alert_events table if not exists
	db.Exec(`
		CREATE TABLE IF NOT EXISTS alert_events (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			rule_id UUID REFERENCES alert_rules(id) ON DELETE SET NULL,
			agent_id UUID REFERENCES agents(id) ON DELETE CASCADE,
			metric_value FLOAT NOT NULL,
			threshold FLOAT NOT NULL,
			status VARCHAR(20) NOT NULL DEFAULT 'pending',
			message TEXT,
			triggered_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
			resolved_at TIMESTAMP WITH TIME ZONE,
			acknowledged_by UUID REFERENCES users(id) ON DELETE SET NULL,
			acknowledged_at TIMESTAMP WITH TIME ZONE,
			created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
			CONSTRAINT chk_status CHECK (status IN ('pending', 'acknowledged', 'resolved'))
		)
	`)
	fmt.Println("Created alert_events table (if not exists)")

	// Verify
	for _, table := range tables {
		var exists bool
		db.Raw("SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = ?)", table).Scan(&exists)
		fmt.Printf("%s exists: %v\n", table, exists)
	}
}

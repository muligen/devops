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

	// Check if agent_metrics table exists
	var tableExists bool
	db.Raw("SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'agent_metrics')").Scan(&tableExists)
	fmt.Printf("agent_metrics table exists: %v\n", tableExists)

	if tableExists {
		// Get count
		var count int64
		db.Raw("SELECT COUNT(*) FROM agent_metrics").Scan(&count)
		fmt.Printf("agent_metrics count: %d\n", count)

		// Get columns
		type ColumnInfo struct {
			ColumnName string
			DataType   string
		}
		var columns []ColumnInfo
		db.Raw("SELECT column_name, data_type FROM information_schema.columns WHERE table_name = 'agent_metrics'").Scan(&columns)
		fmt.Println("Columns:")
		for _, col := range columns {
			fmt.Printf("  - %s (%s)\n", col.ColumnName, col.DataType)
		}
	}

	// Test the query that's failing
	fmt.Println("\nTesting sort query...")
	var result []map[string]interface{}
	err = db.Table("agent_metrics").
		Select("DISTINCT ON (agent_id) agent_id, cpu_usage, memory_percent, disk_percent").
		Order("agent_id, collected_at DESC").
		Find(&result).Error
	if err != nil {
		fmt.Printf("Query error: %v\n", err)
	} else {
		fmt.Printf("Query result count: %d\n", len(result))
		for _, r := range result {
			fmt.Printf("  - agent_id: %v, cpu: %v, mem: %v, disk: %v\n",
				r["agent_id"], r["cpu_usage"], r["memory_percent"], r["disk_percent"])
		}
	}
}

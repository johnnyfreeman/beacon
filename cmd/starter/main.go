package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/beacon/internal/db"
	"github.com/beacon/internal/temporal"
	"github.com/google/uuid"
	"go.temporal.io/sdk/client"
)

func main() {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL environment variable is required")
	}

	temporalHost := os.Getenv("TEMPORAL_HOST")
	if temporalHost == "" {
		temporalHost = "localhost:7233"
	}

	database, err := db.NewDB(databaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close()

	c, err := client.Dial(client.Options{
		HostPort: temporalHost,
	})
	if err != nil {
		log.Fatalf("Failed to create Temporal client: %v", err)
	}
	defer c.Close()

	// Get all enabled endpoints
	endpoints, err := database.ListEnabledEndpoints()
	if err != nil {
		log.Fatalf("Failed to list endpoints: %v", err)
	}

	fmt.Printf("Found %d enabled endpoints\n", len(endpoints))

	// Start monitoring workflow for each endpoint
	for _, endpoint := range endpoints {
		workflowID := fmt.Sprintf("monitor-endpoint-%s", endpoint.ID)
		
		we, err := c.ExecuteWorkflow(context.Background(), client.StartWorkflowOptions{
			ID:        workflowID,
			TaskQueue: "beacon-monitoring",
		}, temporal.MonitorEndpointWorkflow, endpoint.ID, endpoint.IntervalSec)
		
		if err != nil {
			log.Printf("Failed to start workflow for endpoint %s: %v", endpoint.Name, err)
			continue
		}
		
		fmt.Printf("Started monitoring workflow for endpoint %s (ID: %s, WorkflowID: %s, RunID: %s)\n", 
			endpoint.Name, endpoint.ID, we.GetID(), we.GetRunID())
	}

	// Start aggregate metrics workflow
	aggregateWorkflowID := "aggregate-metrics"
	we, err := c.ExecuteWorkflow(context.Background(), client.StartWorkflowOptions{
		ID:        aggregateWorkflowID,
		TaskQueue: "beacon-monitoring",
	}, temporal.AggregateMetricsWorkflow)
	
	if err != nil {
		log.Printf("Failed to start aggregate metrics workflow: %v", err)
	} else {
		fmt.Printf("Started aggregate metrics workflow (WorkflowID: %s, RunID: %s)\n", 
			we.GetID(), we.GetRunID())
	}

	// Start cleanup workflow
	cleanupWorkflowID := "cleanup-old-data"
	retentionDays := 30
	we, err = c.ExecuteWorkflow(context.Background(), client.StartWorkflowOptions{
		ID:        cleanupWorkflowID,
		TaskQueue: "beacon-monitoring",
	}, temporal.CleanupWorkflow, retentionDays)
	
	if err != nil {
		log.Printf("Failed to start cleanup workflow: %v", err)
	} else {
		fmt.Printf("Started cleanup workflow (WorkflowID: %s, RunID: %s, Retention: %d days)\n", 
			we.GetID(), we.GetRunID(), retentionDays)
	}
}
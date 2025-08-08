package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/beacon/internal/db"
	"github.com/beacon/internal/temporal"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
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

	w := worker.New(c, "beacon-monitoring", worker.Options{})

	activities := &temporal.Activities{
		DB: database,
	}

	w.RegisterActivity(activities.PingEndpoint)
	w.RegisterActivity(activities.CheckIncidentStatus)
	w.RegisterActivity(activities.TriggerWebhooks)
	w.RegisterActivity(activities.AggregateMetrics)
	w.RegisterActivity(activities.CleanupOldData)
	w.RegisterActivity(activities.GetEnabledEndpoints)

	w.RegisterWorkflow(temporal.MonitorEndpointWorkflow)
	w.RegisterWorkflow(temporal.AggregateMetricsWorkflow)
	w.RegisterWorkflow(temporal.CleanupWorkflow)

	err = w.Start()
	if err != nil {
		log.Fatalf("Failed to start worker: %v", err)
	}

	fmt.Println("Beacon worker started successfully")

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	fmt.Println("Shutting down worker...")
	w.Stop()
}


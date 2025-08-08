package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/beacon/internal/db"
	"github.com/beacon/internal/temporal"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
	"go.temporal.io/sdk/client"
)

func MonitorCmd(dbURL string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "monitor",
		Short: "Manage monitoring workflows",
	}

	cmd.AddCommand(startMonitoringCmd(dbURL))
	cmd.AddCommand(stopMonitoringCmd(dbURL))

	return cmd
}

func startMonitoringCmd(dbURL string) *cobra.Command {
	var endpointID string
	var all bool

	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start monitoring workflows",
		RunE: func(cmd *cobra.Command, args []string) error {
			database, err := db.NewDB(dbURL)
			if err != nil {
				return err
			}
			defer database.Close()

			temporalHost := os.Getenv("TEMPORAL_HOST")
			if temporalHost == "" {
				temporalHost = "localhost:7233"
			}

			c, err := client.Dial(client.Options{
				HostPort: temporalHost,
			})
			if err != nil {
				return fmt.Errorf("failed to create Temporal client: %w", err)
			}
			defer c.Close()

			if all {
				// Start monitoring for all enabled endpoints
				endpoints, err := database.ListEnabledEndpoints()
				if err != nil {
					return fmt.Errorf("failed to list endpoints: %w", err)
				}

				fmt.Printf("Starting monitoring for %d endpoints\n", len(endpoints))

				for _, endpoint := range endpoints {
					workflowID := fmt.Sprintf("monitor-endpoint-%s", endpoint.ID)
					
					we, err := c.ExecuteWorkflow(context.Background(), client.StartWorkflowOptions{
						ID:        workflowID,
						TaskQueue: "beacon-monitoring",
					}, temporal.MonitorEndpointWorkflow, endpoint.ID, endpoint.IntervalSec)
					
					if err != nil {
						fmt.Printf("Failed to start workflow for %s: %v\n", endpoint.Name, err)
						continue
					}
					
					fmt.Printf("✓ Started monitoring for %s (ID: %s)\n", endpoint.Name, we.GetID())
				}

				// Also start the aggregate metrics workflow
				aggregateWorkflowID := "aggregate-metrics"
				_, err = c.ExecuteWorkflow(context.Background(), client.StartWorkflowOptions{
					ID:        aggregateWorkflowID,
					TaskQueue: "beacon-monitoring",
				}, temporal.AggregateMetricsWorkflow)
				
				if err != nil {
					fmt.Printf("Failed to start aggregate metrics workflow: %v\n", err)
				} else {
					fmt.Printf("✓ Started aggregate metrics workflow\n")
				}

				// Start cleanup workflow
				cleanupWorkflowID := "cleanup-old-data"
				retentionDays := 30
				_, err = c.ExecuteWorkflow(context.Background(), client.StartWorkflowOptions{
					ID:        cleanupWorkflowID,
					TaskQueue: "beacon-monitoring",
				}, temporal.CleanupWorkflow, retentionDays)
				
				if err != nil {
					fmt.Printf("Failed to start cleanup workflow: %v\n", err)
				} else {
					fmt.Printf("✓ Started cleanup workflow (retention: %d days)\n", retentionDays)
				}

			} else if endpointID != "" {
				// Start monitoring for a specific endpoint
				epID, err := uuid.Parse(endpointID)
				if err != nil {
					return fmt.Errorf("invalid endpoint UUID: %w", err)
				}

				endpoint, err := database.GetEndpoint(epID)
				if err != nil {
					return fmt.Errorf("failed to get endpoint: %w", err)
				}

				workflowID := fmt.Sprintf("monitor-endpoint-%s", endpoint.ID)
				
				we, err := c.ExecuteWorkflow(context.Background(), client.StartWorkflowOptions{
					ID:        workflowID,
					TaskQueue: "beacon-monitoring",
				}, temporal.MonitorEndpointWorkflow, endpoint.ID, endpoint.IntervalSec)
				
				if err != nil {
					return fmt.Errorf("failed to start workflow: %w", err)
				}
				
				fmt.Printf("✓ Started monitoring for %s (ID: %s)\n", endpoint.Name, we.GetID())
			} else {
				return fmt.Errorf("specify --endpoint-id or --all")
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&endpointID, "endpoint-id", "", "Endpoint ID to monitor")
	cmd.Flags().BoolVar(&all, "all", false, "Start monitoring for all enabled endpoints")

	return cmd
}

func stopMonitoringCmd(dbURL string) *cobra.Command {
	var endpointID string
	var all bool

	cmd := &cobra.Command{
		Use:   "stop",
		Short: "Stop monitoring workflows",
		RunE: func(cmd *cobra.Command, args []string) error {
			temporalHost := os.Getenv("TEMPORAL_HOST")
			if temporalHost == "" {
				temporalHost = "localhost:7233"
			}

			c, err := client.Dial(client.Options{
				HostPort: temporalHost,
			})
			if err != nil {
				return fmt.Errorf("failed to create Temporal client: %w", err)
			}
			defer c.Close()

			if all {
				database, err := db.NewDB(dbURL)
				if err != nil {
					return err
				}
				defer database.Close()

				endpoints, err := database.ListEnabledEndpoints()
				if err != nil {
					return fmt.Errorf("failed to list endpoints: %w", err)
				}

				for _, endpoint := range endpoints {
					workflowID := fmt.Sprintf("monitor-endpoint-%s", endpoint.ID)
					err := c.CancelWorkflow(context.Background(), workflowID, "")
					if err != nil {
						fmt.Printf("Failed to stop workflow for %s: %v\n", endpoint.Name, err)
					} else {
						fmt.Printf("✓ Stopped monitoring for %s\n", endpoint.Name)
					}
				}

				// Stop aggregate and cleanup workflows
				c.CancelWorkflow(context.Background(), "aggregate-metrics", "")
				c.CancelWorkflow(context.Background(), "cleanup-old-data", "")
				fmt.Printf("✓ Stopped aggregate and cleanup workflows\n")

			} else if endpointID != "" {
				epID, err := uuid.Parse(endpointID)
				if err != nil {
					return fmt.Errorf("invalid endpoint UUID: %w", err)
				}

				workflowID := fmt.Sprintf("monitor-endpoint-%s", epID)
				err = c.CancelWorkflow(context.Background(), workflowID, "")
				if err != nil {
					return fmt.Errorf("failed to stop workflow: %w", err)
				}
				
				fmt.Printf("✓ Stopped monitoring for endpoint %s\n", endpointID)
			} else {
				return fmt.Errorf("specify --endpoint-id or --all")
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&endpointID, "endpoint-id", "", "Endpoint ID to stop monitoring")
	cmd.Flags().BoolVar(&all, "all", false, "Stop monitoring for all endpoints")

	return cmd
}
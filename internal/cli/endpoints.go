package cli

import (
	"encoding/json"
	"fmt"

	"github.com/beacon/internal/db"
	"github.com/beacon/internal/models"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

func EndpointsCmd(dbURL string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "endpoints",
		Short: "Manage service endpoints",
	}

	cmd.AddCommand(createEndpointCmd(dbURL))
	cmd.AddCommand(getEndpointCmd(dbURL))
	cmd.AddCommand(listEndpointsCmd(dbURL))
	cmd.AddCommand(updateEndpointCmd(dbURL))
	cmd.AddCommand(deleteEndpointCmd(dbURL))

	return cmd
}

func createEndpointCmd(dbURL string) *cobra.Command {
	var (
		serviceID    string
		name         string
		url          string
		method       string
		expectedCode int
		timeoutMs    int
		intervalSec  int
		enabled      bool
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new endpoint",
		RunE: func(cmd *cobra.Command, args []string) error {
			database, err := db.NewDB(dbURL)
			if err != nil {
				return err
			}
			defer database.Close()

			svcID, err := uuid.Parse(serviceID)
			if err != nil {
				return fmt.Errorf("invalid service UUID: %w", err)
			}

			endpoint := &models.ServiceEndpoint{
				ServiceID:    svcID,
				Name:         name,
				URL:          url,
				Method:       method,
				ExpectedCode: expectedCode,
				TimeoutMs:    timeoutMs,
				IntervalSec:  intervalSec,
				Enabled:      enabled,
				Headers:      make(models.JSONB),
			}

			if err := database.CreateEndpoint(endpoint); err != nil {
				return fmt.Errorf("failed to create endpoint: %w", err)
			}

			data, _ := json.MarshalIndent(endpoint, "", "  ")
			fmt.Println(string(data))
			return nil
		},
	}

	cmd.Flags().StringVar(&serviceID, "service-id", "", "Service ID (required)")
	cmd.Flags().StringVar(&name, "name", "", "Endpoint name (required)")
	cmd.Flags().StringVar(&url, "url", "", "Endpoint URL (required)")
	cmd.Flags().StringVar(&method, "method", "GET", "HTTP method")
	cmd.Flags().IntVar(&expectedCode, "expected-code", 200, "Expected HTTP status code")
	cmd.Flags().IntVar(&timeoutMs, "timeout", 30000, "Timeout in milliseconds")
	cmd.Flags().IntVar(&intervalSec, "interval", 60, "Check interval in seconds")
	cmd.Flags().BoolVar(&enabled, "enabled", true, "Enable endpoint monitoring")
	
	cmd.MarkFlagRequired("service-id")
	cmd.MarkFlagRequired("name")
	cmd.MarkFlagRequired("url")

	return cmd
}

func getEndpointCmd(dbURL string) *cobra.Command {
	return &cobra.Command{
		Use:   "get [id]",
		Short: "Get an endpoint by ID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			database, err := db.NewDB(dbURL)
			if err != nil {
				return err
			}
			defer database.Close()

			id, err := uuid.Parse(args[0])
			if err != nil {
				return fmt.Errorf("invalid UUID: %w", err)
			}

			endpoint, err := database.GetEndpoint(id)
			if err != nil {
				return fmt.Errorf("failed to get endpoint: %w", err)
			}

			data, _ := json.MarshalIndent(endpoint, "", "  ")
			fmt.Println(string(data))
			return nil
		},
	}
}

func listEndpointsCmd(dbURL string) *cobra.Command {
	var serviceID string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List endpoints",
		RunE: func(cmd *cobra.Command, args []string) error {
			database, err := db.NewDB(dbURL)
			if err != nil {
				return err
			}
			defer database.Close()

			var svcID *uuid.UUID
			if serviceID != "" {
				id, err := uuid.Parse(serviceID)
				if err != nil {
					return fmt.Errorf("invalid service UUID: %w", err)
				}
				svcID = &id
			}

			endpoints, err := database.ListEndpoints(svcID)
			if err != nil {
				return fmt.Errorf("failed to list endpoints: %w", err)
			}

			data, _ := json.MarshalIndent(endpoints, "", "  ")
			fmt.Println(string(data))
			return nil
		},
	}

	cmd.Flags().StringVar(&serviceID, "service-id", "", "Filter by service ID")

	return cmd
}

func updateEndpointCmd(dbURL string) *cobra.Command {
	var (
		name         string
		url          string
		method       string
		expectedCode int
		timeoutMs    int
		intervalSec  int
		enabled      *bool
	)

	cmd := &cobra.Command{
		Use:   "update [id]",
		Short: "Update an endpoint",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			database, err := db.NewDB(dbURL)
			if err != nil {
				return err
			}
			defer database.Close()

			id, err := uuid.Parse(args[0])
			if err != nil {
				return fmt.Errorf("invalid UUID: %w", err)
			}

			endpoint, err := database.GetEndpoint(id)
			if err != nil {
				return fmt.Errorf("failed to get endpoint: %w", err)
			}

			if name != "" {
				endpoint.Name = name
			}
			if url != "" {
				endpoint.URL = url
			}
			if method != "" {
				endpoint.Method = method
			}
			if cmd.Flags().Changed("expected-code") {
				endpoint.ExpectedCode = expectedCode
			}
			if cmd.Flags().Changed("timeout") {
				endpoint.TimeoutMs = timeoutMs
			}
			if cmd.Flags().Changed("interval") {
				endpoint.IntervalSec = intervalSec
			}
			if enabled != nil {
				endpoint.Enabled = *enabled
			}

			if err := database.UpdateEndpoint(endpoint); err != nil {
				return fmt.Errorf("failed to update endpoint: %w", err)
			}

			fmt.Printf("Endpoint %s updated successfully\n", id)
			return nil
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Endpoint name")
	cmd.Flags().StringVar(&url, "url", "", "Endpoint URL")
	cmd.Flags().StringVar(&method, "method", "", "HTTP method")
	cmd.Flags().IntVar(&expectedCode, "expected-code", 0, "Expected HTTP status code")
	cmd.Flags().IntVar(&timeoutMs, "timeout", 0, "Timeout in milliseconds")
	cmd.Flags().IntVar(&intervalSec, "interval", 0, "Check interval in seconds")
	
	enabledFlag := false
	cmd.Flags().BoolVar(&enabledFlag, "enabled", false, "Enable/disable endpoint")
	cmd.PreRunE = func(cmd *cobra.Command, args []string) error {
		if cmd.Flags().Changed("enabled") {
			enabled = &enabledFlag
		}
		return nil
	}

	return cmd
}

func deleteEndpointCmd(dbURL string) *cobra.Command {
	return &cobra.Command{
		Use:   "delete [id]",
		Short: "Delete an endpoint",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			database, err := db.NewDB(dbURL)
			if err != nil {
				return err
			}
			defer database.Close()

			id, err := uuid.Parse(args[0])
			if err != nil {
				return fmt.Errorf("invalid UUID: %w", err)
			}

			if err := database.DeleteEndpoint(id); err != nil {
				return fmt.Errorf("failed to delete endpoint: %w", err)
			}

			fmt.Printf("Endpoint %s deleted successfully\n", id)
			return nil
		},
	}
}
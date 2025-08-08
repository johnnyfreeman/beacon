package cli

import (
	"encoding/json"
	"fmt"

	"github.com/beacon/internal/db"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

func IncidentsCmd(dbURL string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "incidents",
		Short: "Manage incidents",
	}

	cmd.AddCommand(getIncidentCmd(dbURL))
	cmd.AddCommand(listIncidentsCmd(dbURL))
	cmd.AddCommand(resolveIncidentCmd(dbURL))

	return cmd
}

func getIncidentCmd(dbURL string) *cobra.Command {
	return &cobra.Command{
		Use:   "get [id]",
		Short: "Get an incident by ID",
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

			incident, err := database.GetIncident(id)
			if err != nil {
				return fmt.Errorf("failed to get incident: %w", err)
			}

			data, _ := json.MarshalIndent(incident, "", "  ")
			fmt.Println(string(data))
			return nil
		},
	}
}

func listIncidentsCmd(dbURL string) *cobra.Command {
	var (
		endpointID string
		status     string
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List incidents",
		RunE: func(cmd *cobra.Command, args []string) error {
			database, err := db.NewDB(dbURL)
			if err != nil {
				return err
			}
			defer database.Close()

			var epID *uuid.UUID
			if endpointID != "" {
				id, err := uuid.Parse(endpointID)
				if err != nil {
					return fmt.Errorf("invalid endpoint UUID: %w", err)
				}
				epID = &id
			}

			incidents, err := database.ListIncidents(epID, status)
			if err != nil {
				return fmt.Errorf("failed to list incidents: %w", err)
			}

			data, _ := json.MarshalIndent(incidents, "", "  ")
			fmt.Println(string(data))
			return nil
		},
	}

	cmd.Flags().StringVar(&endpointID, "endpoint-id", "", "Filter by endpoint ID")
	cmd.Flags().StringVar(&status, "status", "", "Filter by status (open/resolved)")

	return cmd
}

func resolveIncidentCmd(dbURL string) *cobra.Command {
	return &cobra.Command{
		Use:   "resolve [id]",
		Short: "Resolve an incident",
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

			if err := database.ResolveIncident(id); err != nil {
				return fmt.Errorf("failed to resolve incident: %w", err)
			}

			fmt.Printf("Incident %s resolved successfully\n", id)
			return nil
		},
	}
}
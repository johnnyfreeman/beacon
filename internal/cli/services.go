package cli

import (
	"encoding/json"
	"fmt"

	"github.com/beacon/internal/db"
	"github.com/beacon/internal/models"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

func ServicesCmd(dbURL string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "services",
		Short: "Manage services",
	}

	cmd.AddCommand(createServiceCmd(dbURL))
	cmd.AddCommand(getServiceCmd(dbURL))
	cmd.AddCommand(listServicesCmd(dbURL))
	cmd.AddCommand(updateServiceCmd(dbURL))
	cmd.AddCommand(deleteServiceCmd(dbURL))

	return cmd
}

func createServiceCmd(dbURL string) *cobra.Command {
	var name, description string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new service",
		RunE: func(cmd *cobra.Command, args []string) error {
			database, err := db.NewDB(dbURL)
			if err != nil {
				return err
			}
			defer database.Close()

			service := &models.Service{
				Name:        name,
				Description: description,
			}

			if err := database.CreateService(service); err != nil {
				return fmt.Errorf("failed to create service: %w", err)
			}

			data, _ := json.MarshalIndent(service, "", "  ")
			fmt.Println(string(data))
			return nil
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Service name (required)")
	cmd.Flags().StringVar(&description, "description", "", "Service description")
	cmd.MarkFlagRequired("name")

	return cmd
}

func getServiceCmd(dbURL string) *cobra.Command {
	return &cobra.Command{
		Use:   "get [id]",
		Short: "Get a service by ID",
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

			service, err := database.GetService(id)
			if err != nil {
				return fmt.Errorf("failed to get service: %w", err)
			}

			data, _ := json.MarshalIndent(service, "", "  ")
			fmt.Println(string(data))
			return nil
		},
	}
}

func listServicesCmd(dbURL string) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all services",
		RunE: func(cmd *cobra.Command, args []string) error {
			database, err := db.NewDB(dbURL)
			if err != nil {
				return err
			}
			defer database.Close()

			services, err := database.ListServices()
			if err != nil {
				return fmt.Errorf("failed to list services: %w", err)
			}

			data, _ := json.MarshalIndent(services, "", "  ")
			fmt.Println(string(data))
			return nil
		},
	}
}

func updateServiceCmd(dbURL string) *cobra.Command {
	var name, description string

	cmd := &cobra.Command{
		Use:   "update [id]",
		Short: "Update a service",
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

			service, err := database.GetService(id)
			if err != nil {
				return fmt.Errorf("failed to get service: %w", err)
			}

			if name != "" {
				service.Name = name
			}
			if description != "" {
				service.Description = description
			}

			if err := database.UpdateService(service); err != nil {
				return fmt.Errorf("failed to update service: %w", err)
			}

			fmt.Printf("Service %s updated successfully\n", id)
			return nil
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Service name")
	cmd.Flags().StringVar(&description, "description", "", "Service description")

	return cmd
}

func deleteServiceCmd(dbURL string) *cobra.Command {
	return &cobra.Command{
		Use:   "delete [id]",
		Short: "Delete a service",
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

			if err := database.DeleteService(id); err != nil {
				return fmt.Errorf("failed to delete service: %w", err)
			}

			fmt.Printf("Service %s deleted successfully\n", id)
			return nil
		},
	}
}
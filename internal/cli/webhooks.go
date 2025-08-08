package cli

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/beacon/internal/db"
	"github.com/beacon/internal/models"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

func WebhooksCmd(dbURL string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "webhooks",
		Short: "Manage webhooks",
	}

	cmd.AddCommand(createWebhookCmd(dbURL))
	cmd.AddCommand(getWebhookCmd(dbURL))
	cmd.AddCommand(listWebhooksCmd(dbURL))
	cmd.AddCommand(updateWebhookCmd(dbURL))
	cmd.AddCommand(deleteWebhookCmd(dbURL))

	return cmd
}

func createWebhookCmd(dbURL string) *cobra.Command {
	var (
		serviceID string
		name      string
		url       string
		events    string
		enabled   bool
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new webhook",
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

			eventList := strings.Split(events, ",")
			for i := range eventList {
				eventList[i] = strings.TrimSpace(eventList[i])
			}

			webhook := &models.Webhook{
				ServiceID: svcID,
				Name:      name,
				URL:       url,
				Events:    eventList,
				Enabled:   enabled,
				Headers:   make(models.JSONB),
			}

			if err := database.CreateWebhook(webhook); err != nil {
				return fmt.Errorf("failed to create webhook: %w", err)
			}

			data, _ := json.MarshalIndent(webhook, "", "  ")
			fmt.Println(string(data))
			return nil
		},
	}

	cmd.Flags().StringVar(&serviceID, "service-id", "", "Service ID (required)")
	cmd.Flags().StringVar(&name, "name", "", "Webhook name (required)")
	cmd.Flags().StringVar(&url, "url", "", "Webhook URL (required)")
	cmd.Flags().StringVar(&events, "events", "incident_start,incident_resolved", "Comma-separated list of events")
	cmd.Flags().BoolVar(&enabled, "enabled", true, "Enable webhook")
	
	cmd.MarkFlagRequired("service-id")
	cmd.MarkFlagRequired("name")
	cmd.MarkFlagRequired("url")

	return cmd
}

func getWebhookCmd(dbURL string) *cobra.Command {
	return &cobra.Command{
		Use:   "get [id]",
		Short: "Get a webhook by ID",
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

			webhook, err := database.GetWebhook(id)
			if err != nil {
				return fmt.Errorf("failed to get webhook: %w", err)
			}

			data, _ := json.MarshalIndent(webhook, "", "  ")
			fmt.Println(string(data))
			return nil
		},
	}
}

func listWebhooksCmd(dbURL string) *cobra.Command {
	var serviceID string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List webhooks",
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

			webhooks, err := database.ListWebhooks(svcID)
			if err != nil {
				return fmt.Errorf("failed to list webhooks: %w", err)
			}

			data, _ := json.MarshalIndent(webhooks, "", "  ")
			fmt.Println(string(data))
			return nil
		},
	}

	cmd.Flags().StringVar(&serviceID, "service-id", "", "Filter by service ID")

	return cmd
}

func updateWebhookCmd(dbURL string) *cobra.Command {
	var (
		name    string
		url     string
		events  string
		enabled *bool
	)

	cmd := &cobra.Command{
		Use:   "update [id]",
		Short: "Update a webhook",
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

			webhook, err := database.GetWebhook(id)
			if err != nil {
				return fmt.Errorf("failed to get webhook: %w", err)
			}

			if name != "" {
				webhook.Name = name
			}
			if url != "" {
				webhook.URL = url
			}
			if events != "" {
				eventList := strings.Split(events, ",")
				for i := range eventList {
					eventList[i] = strings.TrimSpace(eventList[i])
				}
				webhook.Events = eventList
			}
			if enabled != nil {
				webhook.Enabled = *enabled
			}

			if err := database.UpdateWebhook(webhook); err != nil {
				return fmt.Errorf("failed to update webhook: %w", err)
			}

			fmt.Printf("Webhook %s updated successfully\n", id)
			return nil
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Webhook name")
	cmd.Flags().StringVar(&url, "url", "", "Webhook URL")
	cmd.Flags().StringVar(&events, "events", "", "Comma-separated list of events")
	
	enabledFlag := false
	cmd.Flags().BoolVar(&enabledFlag, "enabled", false, "Enable/disable webhook")
	cmd.PreRunE = func(cmd *cobra.Command, args []string) error {
		if cmd.Flags().Changed("enabled") {
			enabled = &enabledFlag
		}
		return nil
	}

	return cmd
}

func deleteWebhookCmd(dbURL string) *cobra.Command {
	return &cobra.Command{
		Use:   "delete [id]",
		Short: "Delete a webhook",
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

			if err := database.DeleteWebhook(id); err != nil {
				return fmt.Errorf("failed to delete webhook: %w", err)
			}

			fmt.Printf("Webhook %s deleted successfully\n", id)
			return nil
		},
	}
}
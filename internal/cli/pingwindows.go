package cli

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/beacon/internal/db"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

func PingWindowsCmd(dbURL string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ping-windows",
		Short: "Manage ping windows (aggregated metrics)",
	}

	cmd.AddCommand(getPingWindowCmd(dbURL))
	cmd.AddCommand(listPingWindowsCmd(dbURL))

	return cmd
}

func getPingWindowCmd(dbURL string) *cobra.Command {
	return &cobra.Command{
		Use:   "get [id]",
		Short: "Get a ping window by ID",
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

			window, err := database.GetPingWindow(id)
			if err != nil {
				return fmt.Errorf("failed to get ping window: %w", err)
			}

			data, _ := json.MarshalIndent(window, "", "  ")
			fmt.Println(string(data))
			return nil
		},
	}
}

func listPingWindowsCmd(dbURL string) *cobra.Command {
	var (
		endpointID string
		limit      int
		startTime  string
		endTime    string
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List ping windows",
		RunE: func(cmd *cobra.Command, args []string) error {
			database, err := db.NewDB(dbURL)
			if err != nil {
				return err
			}
			defer database.Close()

			epID, err := uuid.Parse(endpointID)
			if err != nil {
				return fmt.Errorf("invalid endpoint UUID: %w", err)
			}

			if startTime != "" && endTime != "" {
				start, err := time.Parse(time.RFC3339, startTime)
				if err != nil {
					return fmt.Errorf("invalid start time: %w", err)
				}
				end, err := time.Parse(time.RFC3339, endTime)
				if err != nil {
					return fmt.Errorf("invalid end time: %w", err)
				}

				windows, err := database.ListPingWindowsByTimeRange(epID, start, end)
				if err != nil {
					return fmt.Errorf("failed to list ping windows: %w", err)
				}

				data, _ := json.MarshalIndent(windows, "", "  ")
				fmt.Println(string(data))
			} else {
				windows, err := database.ListPingWindows(epID, limit)
				if err != nil {
					return fmt.Errorf("failed to list ping windows: %w", err)
				}

				data, _ := json.MarshalIndent(windows, "", "  ")
				fmt.Println(string(data))
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&endpointID, "endpoint-id", "", "Endpoint ID (required)")
	cmd.Flags().IntVar(&limit, "limit", 100, "Maximum number of windows to return")
	cmd.Flags().StringVar(&startTime, "start", "", "Start time (RFC3339 format)")
	cmd.Flags().StringVar(&endTime, "end", "", "End time (RFC3339 format)")
	cmd.MarkFlagRequired("endpoint-id")

	return cmd
}
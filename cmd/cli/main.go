package main

import (
	"fmt"
	"os"

	"github.com/beacon/internal/cli"
	"github.com/spf13/cobra"
)

var (
	databaseURL string
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "beacon",
		Short: "Beacon API monitoring service CLI",
		Long:  "CLI for managing services, endpoints, and monitoring data in Beacon",
	}

	rootCmd.PersistentFlags().StringVar(&databaseURL, "db", os.Getenv("DATABASE_URL"), "Database connection string")

	rootCmd.AddCommand(cli.ServicesCmd(databaseURL))
	rootCmd.AddCommand(cli.EndpointsCmd(databaseURL))
	rootCmd.AddCommand(cli.PingsCmd(databaseURL))
	rootCmd.AddCommand(cli.PingWindowsCmd(databaseURL))
	rootCmd.AddCommand(cli.IncidentsCmd(databaseURL))
	rootCmd.AddCommand(cli.WebhooksCmd(databaseURL))
	rootCmd.AddCommand(cli.MonitorCmd(databaseURL))

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
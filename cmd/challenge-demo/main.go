// Copyright (c) 2025 AccelByte Inc. All Rights Reserved.
// This is licensed software from AccelByte Inc, for limitations
// and restrictions contact your company contract manager.

package main

import (
	"fmt"
	"os"

	"github.com/AccelByte/extend-challenge/extend-challenge-demo-app/internal/app"
	"github.com/AccelByte/extend-challenge/extend-challenge-demo-app/internal/cli/commands"
	"github.com/AccelByte/extend-challenge/extend-challenge-demo-app/internal/tui"
	"github.com/spf13/cobra"
)

var (
	// Global flags
	backendURL        string
	authMode          string
	eventHandlerURL   string
	userID            string
	namespace         string
	email             string
	password          string
	clientID          string
	clientSecret      string
	iamURL            string
	platformURL       string
	format            string
	adminClientID     string
	adminClientSecret string
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "challenge-demo",
		Short: "Challenge Service Demo CLI",
		Long:  "Interactive TUI and CLI tool for testing AccelByte Challenge Service.",
		// If no subcommand, launch TUI (default behavior)
		Run: func(cmd *cobra.Command, args []string) {
			// Create dependency container
			container := app.NewContainer(
				backendURL,
				authMode,
				eventHandlerURL,
				userID,
				namespace,
				email,
				password,
				clientID,
				clientSecret,
				iamURL,
				platformURL,
				adminClientID,
				adminClientSecret,
			)

			// Create and run TUI application
			application := tui.NewApp(container)
			if err := application.Run(); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
		},
	}

	// Global flags (available to all commands)
	rootCmd.PersistentFlags().StringVar(&backendURL, "backend-url", "http://localhost:8000/challenge", "Challenge service backend URL (gRPC Gateway)")
	rootCmd.PersistentFlags().StringVar(&authMode, "auth-mode", "mock", "Authentication mode (mock|password|client)")
	rootCmd.PersistentFlags().StringVar(&eventHandlerURL, "event-handler-url", "localhost:6566", "Event handler gRPC address (for event simulation)")
	rootCmd.PersistentFlags().StringVar(&userID, "user-id", "test-user-123", "User ID for mock mode")
	rootCmd.PersistentFlags().StringVar(&namespace, "namespace", "test", "AccelByte namespace")
	rootCmd.PersistentFlags().StringVar(&email, "email", "", "User email for password mode")
	rootCmd.PersistentFlags().StringVar(&password, "password", "", "User password for password mode")
	rootCmd.PersistentFlags().StringVar(&clientID, "client-id", "", "OAuth2 client ID (for password or client mode)")
	rootCmd.PersistentFlags().StringVar(&clientSecret, "client-secret", "", "OAuth2 client secret (for password or client mode)")
	rootCmd.PersistentFlags().StringVar(&iamURL, "iam-url", "https://demo.accelbyte.io/iam", "AGS IAM URL (for password or client mode)")
	rootCmd.PersistentFlags().StringVar(&platformURL, "platform-url", "https://demo.accelbyte.io/platform", "AGS Platform URL (for reward verification)")
	rootCmd.PersistentFlags().StringVar(&adminClientID, "admin-client-id", "", "Admin OAuth2 client ID (optional - for AGS Platform verification)")
	rootCmd.PersistentFlags().StringVar(&adminClientSecret, "admin-client-secret", "", "Admin OAuth2 client secret (optional - for AGS Platform verification)")
	rootCmd.PersistentFlags().StringVar(&format, "format", "json", "Output format (json|table|text)")

	// Add subcommands
	rootCmd.AddCommand(commands.NewListCommand())
	rootCmd.AddCommand(commands.NewGetCommand())
	rootCmd.AddCommand(commands.NewTriggerCommand())
	rootCmd.AddCommand(commands.NewClaimCommand())
	rootCmd.AddCommand(commands.NewWatchCommand())

	// M3: Add goal assignment commands
	rootCmd.AddCommand(commands.NewInitializeCommand())
	rootCmd.AddCommand(commands.NewSetGoalActiveCommand())

	// Add reward verification commands
	rootCmd.AddCommand(commands.NewVerifyEntitlementCommand())
	rootCmd.AddCommand(commands.NewVerifyWalletCommand())
	rootCmd.AddCommand(commands.NewListInventoryCommand())
	rootCmd.AddCommand(commands.NewListWalletsCommand())

	// Add explicit TUI command (optional, since it's the default)
	tuiCmd := &cobra.Command{
		Use:   "tui",
		Short: "Launch interactive TUI (default)",
		Long:  "Launch the interactive terminal user interface for the Challenge Service demo app.",
		Run: func(cmd *cobra.Command, args []string) {
			// Same as root command - launch TUI
			container := app.NewContainer(
				backendURL,
				authMode,
				eventHandlerURL,
				userID,
				namespace,
				email,
				password,
				clientID,
				clientSecret,
				iamURL,
				platformURL,
				adminClientID,
				adminClientSecret,
			)

			application := tui.NewApp(container)
			if err := application.Run(); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
		},
	}
	rootCmd.AddCommand(tuiCmd)

	// Execute
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

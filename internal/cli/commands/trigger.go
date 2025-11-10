// Copyright (c) 2025 AccelByte Inc. All Rights Reserved.
// This is licensed software from AccelByte Inc, for limitations
// and restrictions contact your company contract manager.

package commands

import (
	"context"
	"fmt"
	"time"

	"github.com/AccelByte/extend-challenge/extend-challenge-demo-app/internal/cli"
	"github.com/AccelByte/extend-challenge/extend-challenge-demo-app/internal/cli/output"
	"github.com/spf13/cobra"
)

// NewTriggerCommand creates the trigger-event command
func NewTriggerCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "trigger-event",
		Short: "Trigger gameplay events",
		Long:  "Trigger gameplay events for testing (login, stat updates).",
	}

	// Add subcommands
	cmd.AddCommand(newTriggerLoginCommand())
	cmd.AddCommand(newTriggerStatUpdateCommand())

	return cmd
}

func newTriggerLoginCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "login",
		Short: "Trigger user login event",
		Long:  "Trigger a user login event to update login-based challenge progress.",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get format flag
			format, _ := cmd.Flags().GetString("format")

			// Create container
			container := cli.GetContainerFromFlags(cmd)

			// Get user ID and namespace (use container's values)
			userID := container.UserID
			namespace := container.Namespace

			// Trigger event
			ctx := context.Background()
			start := time.Now()
			err := container.EventTrigger.TriggerLogin(ctx, userID, namespace)
			duration := time.Since(start)

			// Format result
			formatter := output.NewFormatter(format)
			result := &output.EventResult{
				Event:      "login",
				UserID:     userID,
				Timestamp:  time.Now(),
				Status:     "success",
				DurationMs: duration.Milliseconds(),
				Error:      err,
			}

			if err != nil {
				result.Status = "error"
			}

			formattedResult, formatErr := formatter.FormatEventResult(result)
			if formatErr != nil {
				return fmt.Errorf("failed to format output: %w", formatErr)
			}

			fmt.Print(formattedResult)

			if err != nil {
				return fmt.Errorf("event trigger failed: %w", err)
			}

			return nil
		},
	}

	return cmd
}

func newTriggerStatUpdateCommand() *cobra.Command {
	var statCode string
	var value int

	cmd := &cobra.Command{
		Use:   "stat-update",
		Short: "Trigger statistic update event",
		Long:  "Trigger a statistic update event with custom stat code and value.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if statCode == "" {
				return fmt.Errorf("--stat-code is required")
			}

			// Get format flag
			format, _ := cmd.Flags().GetString("format")

			// Create container
			container := cli.GetContainerFromFlags(cmd)

			// Get user ID and namespace (use container's values)
			userID := container.UserID
			namespace := container.Namespace

			// Trigger event
			ctx := context.Background()
			start := time.Now()
			err := container.EventTrigger.TriggerStatUpdate(ctx, userID, namespace, statCode, value)
			duration := time.Since(start)

			// Format result
			formatter := output.NewFormatter(format)
			result := &output.EventResult{
				Event:      "stat-update",
				UserID:     userID,
				StatCode:   statCode,
				Value:      value,
				Timestamp:  time.Now(),
				Status:     "success",
				DurationMs: duration.Milliseconds(),
				Error:      err,
			}

			if err != nil {
				result.Status = "error"
			}

			formattedResult, formatErr := formatter.FormatEventResult(result)
			if formatErr != nil {
				return fmt.Errorf("failed to format output: %w", formatErr)
			}

			fmt.Print(formattedResult)

			if err != nil {
				return fmt.Errorf("event trigger failed: %w", err)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&statCode, "stat-code", "", "Statistic code (required)")
	cmd.Flags().IntVar(&value, "value", 0, "New statistic value (required)")
	_ = cmd.MarkFlagRequired("stat-code")
	_ = cmd.MarkFlagRequired("value")

	return cmd
}

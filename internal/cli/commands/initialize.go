// Copyright (c) 2025 AccelByte Inc. All Rights Reserved.
// This is licensed software from AccelByte Inc, for limitations
// and restrictions contact your company contract manager.

package commands

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/AccelByte/extend-challenge/extend-challenge-demo-app/internal/cli"
	"github.com/spf13/cobra"
)

// NewInitializeCommand creates the initialize-player command
func NewInitializeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "initialize-player",
		Short: "Initialize player goals with default assignments",
		Long: `Initialize player goals by assigning default goals based on challenge configuration.
This should be called on first login or when config is updated.
Safe to call multiple times (idempotent).`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get format flag
			format, _ := cmd.Flags().GetString("format")

			// Create container
			container := cli.GetContainerFromFlags(cmd)

			// Call API
			ctx := context.Background()
			result, err := container.APIClient.InitializePlayer(ctx)
			if err != nil {
				return fmt.Errorf("failed to initialize player: %w", err)
			}

			// Format output
			switch format {
			case "json":
				output, err := json.MarshalIndent(result, "", "  ")
				if err != nil {
					return fmt.Errorf("failed to format JSON: %w", err)
				}
				fmt.Println(string(output))

			case "table":
				// Table output for assigned goals
				fmt.Printf("Player Initialized Successfully\n")
				fmt.Printf("New Assignments: %d\n", result.NewAssignments)
				fmt.Printf("Total Active: %d\n\n", result.TotalActive)

				if len(result.AssignedGoals) > 0 {
					fmt.Println("Assigned Goals:")
					fmt.Println("─────────────────────────────────────────────────────────────────")
					fmt.Printf("%-20s %-20s %-12s %-10s\n", "Challenge ID", "Goal ID", "Status", "Progress")
					fmt.Println("─────────────────────────────────────────────────────────────────")

					for _, goal := range result.AssignedGoals {
						active := "inactive"
						if goal.IsActive {
							active = "active"
						}
						fmt.Printf("%-20s %-20s %-12s %d/%d\n",
							truncate(goal.ChallengeID, 20),
							truncate(goal.GoalID, 20),
							active,
							goal.Progress,
							goal.Target)
					}
					fmt.Println("─────────────────────────────────────────────────────────────────")
				}

			default: // text
				fmt.Printf("✅ Player initialized successfully\n")
				fmt.Printf("   New assignments: %d\n", result.NewAssignments)
				fmt.Printf("   Total active goals: %d\n", result.TotalActive)

				if len(result.AssignedGoals) > 0 {
					fmt.Printf("\nAssigned goals:\n")
					for _, goal := range result.AssignedGoals {
						status := "inactive"
						if goal.IsActive {
							status = "active"
						}
						fmt.Printf("  - %s / %s (%s) - %d/%d\n",
							goal.ChallengeID,
							goal.GoalID,
							status,
							goal.Progress,
							goal.Target)
					}
				}
			}

			return nil
		},
	}

	return cmd
}

// truncate truncates a string to maxLen with ellipsis
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}

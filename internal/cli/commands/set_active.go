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

// NewSetGoalActiveCommand creates the set-goal-active command
func NewSetGoalActiveCommand() *cobra.Command {
	var isActive bool

	cmd := &cobra.Command{
		Use:   "set-goal-active <challenge-id> <goal-id>",
		Short: "Activate or deactivate a goal",
		Long: `Activate or deactivate a goal for the current player.
Active goals receive event updates and can be claimed.
Inactive goals do not receive event updates.`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			challengeID := args[0]
			goalID := args[1]

			// Get format flag
			format, _ := cmd.Flags().GetString("format")

			// Create container
			container := cli.GetContainerFromFlags(cmd)

			// Call API
			ctx := context.Background()
			result, err := container.APIClient.SetGoalActive(ctx, challengeID, goalID, isActive)
			if err != nil {
				return fmt.Errorf("failed to set goal active status: %w", err)
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
				fmt.Printf("Goal Active Status Updated\n")
				fmt.Println("─────────────────────────────────────────")
				fmt.Printf("Challenge ID: %s\n", result.ChallengeID)
				fmt.Printf("Goal ID:      %s\n", result.GoalID)
				fmt.Printf("Active:       %v\n", result.IsActive)
				fmt.Printf("Assigned At:  %s\n", result.AssignedAt)
				fmt.Println("─────────────────────────────────────────")
				if result.Message != "" {
					fmt.Printf("Message: %s\n", result.Message)
				}

			default: // text
				action := "deactivated"
				if result.IsActive {
					action = "activated"
				}
				fmt.Printf("✅ Goal %s successfully\n", action)
				fmt.Printf("   Challenge: %s\n", result.ChallengeID)
				fmt.Printf("   Goal: %s\n", result.GoalID)
				if result.Message != "" {
					fmt.Printf("   %s\n", result.Message)
				}
			}

			return nil
		},
	}

	// Add --active flag
	cmd.Flags().BoolVar(&isActive, "active", true, "Set goal active (true) or inactive (false)")

	return cmd
}

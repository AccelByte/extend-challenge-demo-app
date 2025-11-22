// Copyright (c) 2025 AccelByte Inc. All Rights Reserved.
// This is licensed software from AccelByte Inc, for limitations
// and restrictions contact your company contract manager.

package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/AccelByte/extend-challenge/extend-challenge-demo-app/internal/api"
	"github.com/AccelByte/extend-challenge/extend-challenge-demo-app/internal/cli"
	"github.com/spf13/cobra"
)

// NewBatchSelectCommand creates the batch-select command
func NewBatchSelectCommand() *cobra.Command {
	var (
		goalIDs         string
		replaceExisting bool
	)

	cmd := &cobra.Command{
		Use:   "batch-select <challenge-id>",
		Short: "Batch select multiple goals",
		Long: `Activate multiple goals at once (M4 feature).
Provide a comma-separated list of goal IDs to activate.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			challengeID := args[0]

			// Parse goal IDs
			goalIDList := strings.Split(goalIDs, ",")
			if len(goalIDList) == 0 {
				return fmt.Errorf("goal-ids cannot be empty")
			}

			// Trim whitespace from each goal ID
			for i := range goalIDList {
				goalIDList[i] = strings.TrimSpace(goalIDList[i])
			}

			// Get format flag
			format, _ := cmd.Flags().GetString("format")

			// Create container
			container := cli.GetContainerFromFlags(cmd)

			// Create request
			req := &api.BatchSelectRequest{
				GoalIDs:         goalIDList,
				ReplaceExisting: replaceExisting,
			}

			// Call API
			ctx := context.Background()
			result, err := container.APIClient.BatchSelectGoals(ctx, challengeID, req)
			if err != nil {
				return fmt.Errorf("failed to batch select goals: %w", err)
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
				fmt.Printf("Batch Goal Selection Completed\n")
				fmt.Println("─────────────────────────────────────────")
				fmt.Printf("Challenge ID:      %s\n", result.ChallengeID)
				fmt.Printf("Selected Goals:    %d\n", len(result.SelectedGoals))
				fmt.Printf("Total Active:      %d\n", result.TotalActiveGoals)
				fmt.Printf("Replaced Goals:    %d\n", len(result.ReplacedGoals))
				fmt.Println("─────────────────────────────────────────")
				fmt.Println("Selected Goals:")
				for _, goal := range result.SelectedGoals {
					fmt.Printf("  - %s (%s)\n", goal.Name, goal.ID)
				}

			default: // text
				fmt.Printf("✅ Successfully selected %d goals\n", len(result.SelectedGoals))
				fmt.Printf("   Challenge: %s\n", result.ChallengeID)
				fmt.Printf("   Total Active: %d\n", result.TotalActiveGoals)
				if len(result.ReplacedGoals) > 0 {
					fmt.Printf("   Replaced: %d goals\n", len(result.ReplacedGoals))
				}
			}

			return nil
		},
	}

	// Add flags
	cmd.Flags().StringVar(&goalIDs, "goal-ids", "", "Comma-separated goal IDs (required)")
	cmd.Flags().BoolVar(&replaceExisting, "replace-existing", false, "Deactivate existing goals first")
	_ = cmd.MarkFlagRequired("goal-ids")

	return cmd
}

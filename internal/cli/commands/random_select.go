// Copyright (c) 2025 AccelByte Inc. All Rights Reserved.
// This is licensed software from AccelByte Inc, for limitations
// and restrictions contact your company contract manager.

package commands

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/AccelByte/extend-challenge/extend-challenge-demo-app/internal/api"
	"github.com/AccelByte/extend-challenge/extend-challenge-demo-app/internal/cli"
	"github.com/spf13/cobra"
)

// NewRandomSelectCommand creates the random-select command
func NewRandomSelectCommand() *cobra.Command {
	var (
		count           int
		replaceExisting bool
		excludeActive   bool
	)

	cmd := &cobra.Command{
		Use:   "random-select <challenge-id>",
		Short: "Randomly select N goals",
		Long: `Randomly activate N goals from a challenge (M4 feature).
The system will automatically exclude completed/claimed goals and goals with unmet prerequisites.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			challengeID := args[0]

			// Validate count
			if count <= 0 {
				return fmt.Errorf("count must be greater than 0")
			}

			// Get format flag
			format, _ := cmd.Flags().GetString("format")

			// Create container
			container := cli.GetContainerFromFlags(cmd)

			// Create request
			req := &api.RandomSelectRequest{
				Count:           count,
				ReplaceExisting: replaceExisting,
				ExcludeActive:   excludeActive,
			}

			// Call API
			ctx := context.Background()
			result, err := container.APIClient.RandomSelectGoals(ctx, challengeID, req)
			if err != nil {
				return fmt.Errorf("failed to random select goals: %w", err)
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
				fmt.Printf("Random Goal Selection Completed\n")
				fmt.Println("─────────────────────────────────────────")
				fmt.Printf("Challenge ID:      %s\n", result.ChallengeID)
				fmt.Printf("Selected Goals:    %d\n", len(result.SelectedGoals))
				fmt.Printf("Total Active:      %d\n", result.TotalActiveGoals)
				fmt.Printf("Replaced Goals:    %d\n", len(result.ReplacedGoals))
				fmt.Println("─────────────────────────────────────────")
				fmt.Println("Randomly Selected Goals:")
				for _, goal := range result.SelectedGoals {
					fmt.Printf("  - %s (%s)\n", goal.Name, goal.ID)
				}

			default: // text
				fmt.Printf("✅ Successfully selected %d random goals\n", len(result.SelectedGoals))
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
	cmd.Flags().IntVar(&count, "count", 3, "Number of goals to select")
	cmd.Flags().BoolVar(&replaceExisting, "replace-existing", false, "Deactivate existing goals first")
	cmd.Flags().BoolVar(&excludeActive, "exclude-active", false, "Exclude already-active goals")

	return cmd
}

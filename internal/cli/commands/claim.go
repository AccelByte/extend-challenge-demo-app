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

// NewClaimCommand creates the claim-reward command
func NewClaimCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "claim-reward <challenge-id> <goal-id>",
		Short: "Claim reward for completed goal",
		Long:  "Claim the reward for a completed goal within a challenge.",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			challengeID := args[0]
			goalID := args[1]

			// Get format flag
			format, _ := cmd.Flags().GetString("format")

			// Create container
			container := cli.GetContainerFromFlags(cmd)

			// Call API
			ctx := context.Background()
			claimResult, err := container.APIClient.ClaimReward(ctx, challengeID, goalID)

			// Prepare output
			reward := &output.ClaimResult{
				ChallengeID: challengeID,
				GoalID:      goalID,
				Status:      "success",
				Timestamp:   time.Now(),
				Error:       err,
			}

			if err != nil {
				reward.Status = "error"
			} else if claimResult != nil {
				// Use reward from claim result
				reward.Reward = &claimResult.Reward
			}

			// Format output
			formatter := output.NewFormatter(format)
			result, formatErr := formatter.FormatClaimResult(reward)
			if formatErr != nil {
				return fmt.Errorf("failed to format output: %w", formatErr)
			}

			fmt.Print(result)

			if err != nil {
				return fmt.Errorf("claim failed: %w", err)
			}

			return nil
		},
	}

	return cmd
}

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

// NewGetRotationStatusCommand creates the get-rotation-status command
func NewGetRotationStatusCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get-rotation-status <challenge-id>",
		Short: "Get rotation status for a challenge",
		Long:  "Get rotation schedule and current period info for a challenge (M5).",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			challengeID := args[0]

			// Create container
			container := cli.GetContainerFromFlags(cmd)

			// Call API
			ctx := context.Background()
			result, err := container.APIClient.GetRotationStatus(ctx, challengeID)
			if err != nil {
				return fmt.Errorf("failed to get rotation status: %w", err)
			}

			// Always output JSON for this command (consistent with E2E test consumption)
			jsonBytes, err := json.MarshalIndent(result, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to format output: %w", err)
			}

			fmt.Println(string(jsonBytes))
			return nil
		},
	}

	return cmd
}

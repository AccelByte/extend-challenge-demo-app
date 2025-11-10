// Copyright (c) 2025 AccelByte Inc. All Rights Reserved.
// This is licensed software from AccelByte Inc, for limitations
// and restrictions contact your company contract manager.

package commands

import (
	"context"
	"fmt"

	"github.com/AccelByte/extend-challenge/extend-challenge-demo-app/internal/cli"
	"github.com/AccelByte/extend-challenge/extend-challenge-demo-app/internal/cli/output"
	"github.com/spf13/cobra"
)

// NewGetCommand creates the get-challenge command
func NewGetCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get-challenge <challenge-id>",
		Short: "Get specific challenge details",
		Long:  "Get details for a specific challenge including all goals.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			challengeID := args[0]

			// Get format flag
			format, _ := cmd.Flags().GetString("format")

			// Create container
			container := cli.GetContainerFromFlags(cmd)

			// Call API
			ctx := context.Background()
			challenge, err := container.APIClient.GetChallenge(ctx, challengeID)
			if err != nil {
				return fmt.Errorf("failed to get challenge: %w", err)
			}

			// Format output
			formatter := output.NewFormatter(format)
			result, err := formatter.FormatChallenge(challenge)
			if err != nil {
				return fmt.Errorf("failed to format output: %w", err)
			}

			fmt.Println(result)
			return nil
		},
	}

	return cmd
}

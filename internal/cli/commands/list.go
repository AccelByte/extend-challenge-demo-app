// Copyright (c) 2025 AccelByte Inc. All Rights Reserved.
// This is licensed software from AccelByte Inc, for limitations
// and restrictions contact your company contract manager.

package commands

import (
	"context"
	"fmt"

	"github.com/AccelByte/extend-challenge/extend-challenge-demo-app/internal/api"
	"github.com/AccelByte/extend-challenge/extend-challenge-demo-app/internal/cli"
	"github.com/AccelByte/extend-challenge/extend-challenge-demo-app/internal/cli/output"
	"github.com/spf13/cobra"
)

// NewListCommand creates the list-challenges command
func NewListCommand() *cobra.Command {
	var activeOnly bool

	cmd := &cobra.Command{
		Use:   "list-challenges",
		Short: "List all challenges with progress",
		Long:  "List all challenges available to the user with their current progress.",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get format flag
			format, _ := cmd.Flags().GetString("format")

			// Create container
			container := cli.GetContainerFromFlags(cmd)

			// Call API (M3: use filtered version if active_only is set)
			ctx := context.Background()
			var challenges []api.Challenge
			var err error

			if activeOnly {
				challenges, err = container.APIClient.ListChallengesWithFilter(ctx, true)
			} else {
				challenges, err = container.APIClient.ListChallenges(ctx)
			}

			if err != nil {
				return fmt.Errorf("failed to list challenges: %w", err)
			}

			// Format output
			formatter := output.NewFormatter(format)
			result, err := formatter.FormatChallenges(challenges)
			if err != nil {
				return fmt.Errorf("failed to format output: %w", err)
			}

			fmt.Println(result)
			return nil
		},
	}

	// M3: Add --active-only flag
	cmd.Flags().BoolVar(&activeOnly, "active-only", false, "Show only active goals (M3 feature)")

	return cmd
}

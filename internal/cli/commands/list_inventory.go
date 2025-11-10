// Copyright (c) 2025 AccelByte Inc. All Rights Reserved.
// This is licensed software from AccelByte Inc, for limitations
// and restrictions contact your company contract manager.

package commands

import (
	"fmt"

	"github.com/AccelByte/extend-challenge/extend-challenge-demo-app/internal/cli"
	"github.com/AccelByte/extend-challenge/extend-challenge-demo-app/internal/cli/output"
	"github.com/spf13/cobra"
)

// NewListInventoryCommand creates the list-inventory command
func NewListInventoryCommand() *cobra.Command {
	var status string

	cmd := &cobra.Command{
		Use:   "list-inventory",
		Short: "List all user entitlements",
		Long:  "List all item entitlements owned by the user from AGS Platform.",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get format flag
			format, _ := cmd.Flags().GetString("format")

			// Create container
			container := cli.GetContainerFromFlags(cmd)

			// Build filters
			filters := make(map[string]string)
			if status != "" {
				filters["status"] = status
			}

			// Query entitlements
			ents, err := container.RewardVerifier.QueryUserEntitlements(filters)
			if err != nil {
				return fmt.Errorf("failed to query entitlements: %w", err)
			}

			// Format output
			formatter := output.NewFormatter(format)
			result, err := formatter.FormatEntitlements(ents)
			if err != nil {
				return fmt.Errorf("failed to format output: %w", err)
			}

			fmt.Println(result)
			return nil
		},
	}

	cmd.Flags().StringVar(&status, "status", "", "Filter by status (ACTIVE, INACTIVE)")

	return cmd
}

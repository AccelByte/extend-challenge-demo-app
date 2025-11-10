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

// NewVerifyEntitlementCommand creates the verify-entitlement command
func NewVerifyEntitlementCommand() *cobra.Command {
	var itemID string

	cmd := &cobra.Command{
		Use:   "verify-entitlement",
		Short: "Verify item entitlement for user",
		Long:  "Check if a specific item entitlement exists for the user in AGS Platform.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if itemID == "" {
				return fmt.Errorf("--item-id is required")
			}

			// Get format flag
			format, _ := cmd.Flags().GetString("format")

			// Create container
			container := cli.GetContainerFromFlags(cmd)

			// Query entitlement
			ent, err := container.RewardVerifier.GetUserEntitlement(itemID)
			if err != nil {
				return fmt.Errorf("failed to get entitlement: %w", err)
			}

			// Format output
			formatter := output.NewFormatter(format)
			result, err := formatter.FormatEntitlement(ent)
			if err != nil {
				return fmt.Errorf("failed to format output: %w", err)
			}

			fmt.Println(result)
			return nil
		},
	}

	cmd.Flags().StringVar(&itemID, "item-id", "", "Item ID to query (required)")
	_ = cmd.MarkFlagRequired("item-id")

	return cmd
}

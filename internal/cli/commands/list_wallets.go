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

// NewListWalletsCommand creates the list-wallets command
func NewListWalletsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list-wallets",
		Short: "List all user wallets",
		Long:  "List all currency wallets and their balances for the user from AGS Platform.",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get format flag
			format, _ := cmd.Flags().GetString("format")

			// Create container
			container := cli.GetContainerFromFlags(cmd)

			// Query wallets
			wallets, err := container.RewardVerifier.QueryUserWallets()
			if err != nil {
				return fmt.Errorf("failed to query wallets: %w", err)
			}

			// Format output
			formatter := output.NewFormatter(format)
			result, err := formatter.FormatWallets(wallets)
			if err != nil {
				return fmt.Errorf("failed to format output: %w", err)
			}

			fmt.Println(result)
			return nil
		},
	}

	return cmd
}

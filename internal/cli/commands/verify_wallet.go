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

// NewVerifyWalletCommand creates the verify-wallet command
func NewVerifyWalletCommand() *cobra.Command {
	var currencyCode string

	cmd := &cobra.Command{
		Use:   "verify-wallet",
		Short: "Verify wallet balance for user",
		Long:  "Check wallet balance for a specific currency code in AGS Platform.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if currencyCode == "" {
				return fmt.Errorf("--currency is required")
			}

			// Get format flag
			format, _ := cmd.Flags().GetString("format")

			// Create container
			container := cli.GetContainerFromFlags(cmd)

			// Query wallet
			wallet, err := container.RewardVerifier.GetUserWallet(currencyCode)
			if err != nil {
				return fmt.Errorf("failed to get wallet: %w", err)
			}

			// Format output
			formatter := output.NewFormatter(format)
			result, err := formatter.FormatWallet(wallet)
			if err != nil {
				return fmt.Errorf("failed to format output: %w", err)
			}

			fmt.Println(result)
			return nil
		},
	}

	cmd.Flags().StringVar(&currencyCode, "currency", "", "Currency code to query (required)")
	_ = cmd.MarkFlagRequired("currency")

	return cmd
}

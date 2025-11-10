// Copyright (c) 2025 AccelByte Inc. All Rights Reserved.
// This is licensed software from AccelByte Inc, for limitations
// and restrictions contact your company contract manager.

package output

import (
	"fmt"
	"strings"

	"github.com/AccelByte/extend-challenge/extend-challenge-demo-app/internal/ags"
	"github.com/AccelByte/extend-challenge/extend-challenge-demo-app/internal/api"
)

// TableFormatter formats output as a table
type TableFormatter struct{}

// FormatChallenges formats challenges as a table
func (f *TableFormatter) FormatChallenges(challenges []api.Challenge) (string, error) {
	var b strings.Builder

	// Header
	b.WriteString(fmt.Sprintf("%-20s %-30s %-15s %-15s\n", "ID", "NAME", "PROGRESS", "STATUS"))
	b.WriteString(strings.Repeat("-", 80) + "\n")

	// Rows
	for _, c := range challenges {
		completed := 0
		for _, g := range c.Goals {
			if g.Status == "completed" || g.Status == "claimed" {
				completed++
			}
		}

		progress := fmt.Sprintf("%d/%d", completed, len(c.Goals))
		name := truncate(c.Name, 30)

		// Calculate status based on goals
		status := "not_started"
		if completed == len(c.Goals) {
			status = "completed"
		} else if completed > 0 {
			status = "in_progress"
		}

		b.WriteString(fmt.Sprintf("%-20s %-30s %-15s %-15s\n",
			c.ID, name, progress, status))
	}

	return b.String(), nil
}

// FormatChallenge formats a single challenge as a table
func (f *TableFormatter) FormatChallenge(challenge *api.Challenge) (string, error) {
	var b strings.Builder

	b.WriteString(fmt.Sprintf("Challenge: %s\n", challenge.Name))
	b.WriteString(fmt.Sprintf("ID: %s\n", challenge.ID))
	b.WriteString(fmt.Sprintf("Description: %s\n\n", challenge.Description))

	// Goals header
	b.WriteString(fmt.Sprintf("%-30s %-15s %-15s\n", "GOAL", "PROGRESS", "STATUS"))
	b.WriteString(strings.Repeat("-", 60) + "\n")

	// Goals
	for _, g := range challenge.Goals {
		progress := fmt.Sprintf("%d/%d", g.Progress, g.Requirement.TargetValue)
		name := truncate(g.Name, 30)
		b.WriteString(fmt.Sprintf("%-30s %-15s %-15s\n",
			name, progress, g.Status))
	}

	return b.String(), nil
}

// FormatEventResult formats an event result as a table
func (f *TableFormatter) FormatEventResult(result *EventResult) (string, error) {
	var b strings.Builder

	b.WriteString(fmt.Sprintf("Event:    %s\n", result.Event))
	b.WriteString(fmt.Sprintf("User ID:  %s\n", result.UserID))
	if result.StatCode != "" {
		b.WriteString(fmt.Sprintf("Stat:     %s = %d\n", result.StatCode, result.Value))
	}
	b.WriteString(fmt.Sprintf("Status:   %s\n", result.Status))
	b.WriteString(fmt.Sprintf("Duration: %dms\n", result.DurationMs))

	if result.Error != nil {
		b.WriteString(fmt.Sprintf("Error:    %v\n", result.Error))
	}

	return b.String(), nil
}

// FormatClaimResult formats a claim result as a table
func (f *TableFormatter) FormatClaimResult(result *ClaimResult) (string, error) {
	var b strings.Builder

	b.WriteString(fmt.Sprintf("Challenge: %s\n", result.ChallengeID))
	b.WriteString(fmt.Sprintf("Goal:      %s\n", result.GoalID))
	b.WriteString(fmt.Sprintf("Status:    %s\n", result.Status))

	if result.Reward != nil {
		b.WriteString(fmt.Sprintf("Reward:    %s %s", result.Reward.Type, result.Reward.RewardID))
		if result.Reward.Quantity > 1 {
			b.WriteString(fmt.Sprintf(" x%d", result.Reward.Quantity))
		}
		b.WriteString("\n")
	}

	if result.Error != nil {
		b.WriteString(fmt.Sprintf("Error:     %v\n", result.Error))
	}

	return b.String(), nil
}

// FormatEntitlement formats a single entitlement as a table
func (f *TableFormatter) FormatEntitlement(ent *ags.Entitlement) (string, error) {
	// Use JSON formatter for single items
	jsonFormatter := &JSONFormatter{}
	return jsonFormatter.FormatEntitlement(ent)
}

// FormatEntitlements formats entitlements as a table
func (f *TableFormatter) FormatEntitlements(ents []*ags.Entitlement) (string, error) {
	var b strings.Builder

	// Header
	b.WriteString(fmt.Sprintf("%-20s %-30s %-10s %-10s %-20s\n", "ENTITLEMENT_ID", "ITEM_ID", "STATUS", "QUANTITY", "GRANTED_AT"))
	b.WriteString(strings.Repeat("-", 90) + "\n")

	// Rows
	for _, ent := range ents {
		entID := truncate(ent.EntitlementID, 20)
		itemID := truncate(ent.ItemID, 30)
		grantedAt := ent.GrantedAt.Format("2006-01-02 15:04")

		b.WriteString(fmt.Sprintf("%-20s %-30s %-10s %-10d %-20s\n",
			entID, itemID, ent.Status, ent.Quantity, grantedAt))
	}

	b.WriteString(fmt.Sprintf("\nTotal: %d entitlements\n", len(ents)))

	return b.String(), nil
}

// FormatWallet formats a single wallet as a table
func (f *TableFormatter) FormatWallet(wallet *ags.Wallet) (string, error) {
	// Use JSON formatter for single items
	jsonFormatter := &JSONFormatter{}
	return jsonFormatter.FormatWallet(wallet)
}

// FormatWallets formats wallets as a table
func (f *TableFormatter) FormatWallets(wallets []*ags.Wallet) (string, error) {
	var b strings.Builder

	// Header
	b.WriteString(fmt.Sprintf("%-20s %-15s %-15s %-10s\n", "WALLET_ID", "CURRENCY", "BALANCE", "STATUS"))
	b.WriteString(strings.Repeat("-", 60) + "\n")

	// Rows
	for _, w := range wallets {
		walletID := truncate(w.WalletID, 20)

		b.WriteString(fmt.Sprintf("%-20s %-15s %-15d %-10s\n",
			walletID, w.CurrencyCode, w.Balance, w.Status))
	}

	b.WriteString(fmt.Sprintf("\nTotal: %d wallets\n", len(wallets)))

	return b.String(), nil
}

// truncate truncates a string to maxLen characters
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

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

// TextFormatter formats output as human-readable text
type TextFormatter struct{}

// FormatChallenges formats challenges as text
func (f *TextFormatter) FormatChallenges(challenges []api.Challenge) (string, error) {
	var b strings.Builder

	b.WriteString(fmt.Sprintf("Found %d challenge(s)\n\n", len(challenges)))

	for i, c := range challenges {
		completed := 0
		for _, g := range c.Goals {
			if g.Status == "completed" || g.Status == "claimed" {
				completed++
			}
		}

		// Calculate status based on goals
		status := "not_started"
		if completed == len(c.Goals) {
			status = "completed"
		} else if completed > 0 {
			status = "in_progress"
		}

		b.WriteString(fmt.Sprintf("%d. %s (%s)\n", i+1, c.Name, c.ID))
		b.WriteString(fmt.Sprintf("   %s\n", c.Description))
		b.WriteString(fmt.Sprintf("   Progress: %d/%d goals | Status: %s\n", completed, len(c.Goals), status))
		if i < len(challenges)-1 {
			b.WriteString("\n")
		}
	}

	return b.String(), nil
}

// FormatChallenge formats a single challenge as text
func (f *TextFormatter) FormatChallenge(challenge *api.Challenge) (string, error) {
	var b strings.Builder

	b.WriteString(fmt.Sprintf("Challenge: %s\n", challenge.Name))
	b.WriteString(fmt.Sprintf("ID: %s\n", challenge.ID))
	b.WriteString(fmt.Sprintf("Description: %s\n\n", challenge.Description))

	b.WriteString("Goals:\n")
	for _, g := range challenge.Goals {
		status := strings.ToUpper(g.Status)
		progress := fmt.Sprintf("(%d/%d)", g.Progress, g.Requirement.TargetValue)

		b.WriteString(fmt.Sprintf("  [%s] %s %s\n", status, g.Name, progress))

		if g.Description != "" {
			b.WriteString(fmt.Sprintf("    %s\n", g.Description))
		}

		// Reward is a struct, not a pointer
		b.WriteString(fmt.Sprintf("    Reward: %s %s", g.Reward.Type, g.Reward.RewardID))
		if g.Reward.Quantity > 1 {
			b.WriteString(fmt.Sprintf(" x%d", g.Reward.Quantity))
		}
		b.WriteString("\n")
		b.WriteString("\n")
	}

	return b.String(), nil
}

// FormatEventResult formats an event result as text
func (f *TextFormatter) FormatEventResult(result *EventResult) (string, error) {
	if result.Error != nil {
		return fmt.Sprintf("✗ Event failed: %v\n", result.Error), nil
	}

	msg := fmt.Sprintf("✓ Event triggered successfully (%dms)\n", result.DurationMs)
	msg += fmt.Sprintf("  Event: %s\n", result.Event)
	msg += fmt.Sprintf("  User: %s\n", result.UserID)

	if result.StatCode != "" {
		msg += fmt.Sprintf("  Stat: %s = %d\n", result.StatCode, result.Value)
	}

	return msg, nil
}

// FormatClaimResult formats a claim result as text
func (f *TextFormatter) FormatClaimResult(result *ClaimResult) (string, error) {
	if result.Error != nil {
		return fmt.Sprintf("✗ Claim failed: %v\n", result.Error), nil
	}

	msg := "✓ Reward claimed successfully\n"
	msg += fmt.Sprintf("  Challenge: %s\n", result.ChallengeID)
	msg += fmt.Sprintf("  Goal: %s\n", result.GoalID)

	if result.Reward != nil {
		msg += fmt.Sprintf("  Reward: %s %s", result.Reward.Type, result.Reward.RewardID)
		if result.Reward.Quantity > 1 {
			msg += fmt.Sprintf(" x%d", result.Reward.Quantity)
		}
		msg += "\n"
	}

	return msg, nil
}

// FormatEntitlement formats a single entitlement as text
func (f *TextFormatter) FormatEntitlement(ent *ags.Entitlement) (string, error) {
	msg := "✓ Entitlement found\n"
	msg += fmt.Sprintf("  Item ID: %s\n", ent.ItemID)
	msg += fmt.Sprintf("  Status: %s\n", ent.Status)
	msg += fmt.Sprintf("  Quantity: %d\n", ent.Quantity)
	msg += fmt.Sprintf("  Granted: %s\n", ent.GrantedAt.Format("2006-01-02 15:04"))
	return msg, nil
}

// FormatEntitlements formats entitlements as text
func (f *TextFormatter) FormatEntitlements(ents []*ags.Entitlement) (string, error) {
	if len(ents) == 0 {
		return "No entitlements found\n", nil
	}

	msg := fmt.Sprintf("Found %d entitlement(s):\n\n", len(ents))
	for i, ent := range ents {
		msg += fmt.Sprintf("%d. %s\n", i+1, ent.ItemID)
		msg += fmt.Sprintf("   Status: %s | Quantity: %d\n", ent.Status, ent.Quantity)
		msg += fmt.Sprintf("   Granted: %s\n", ent.GrantedAt.Format("2006-01-02 15:04"))
		if i < len(ents)-1 {
			msg += "\n"
		}
	}
	return msg, nil
}

// FormatWallet formats a single wallet as text
func (f *TextFormatter) FormatWallet(wallet *ags.Wallet) (string, error) {
	msg := "✓ Wallet found\n"
	msg += fmt.Sprintf("  Currency: %s\n", wallet.CurrencyCode)
	msg += fmt.Sprintf("  Balance: %d\n", wallet.Balance)
	msg += fmt.Sprintf("  Status: %s\n", wallet.Status)
	return msg, nil
}

// FormatWallets formats wallets as text
func (f *TextFormatter) FormatWallets(wallets []*ags.Wallet) (string, error) {
	if len(wallets) == 0 {
		return "No wallets found\n", nil
	}

	msg := fmt.Sprintf("Found %d wallet(s):\n\n", len(wallets))
	for i, w := range wallets {
		msg += fmt.Sprintf("%d. %s: %d (%s)\n", i+1, w.CurrencyCode, w.Balance, w.Status)
	}
	return msg, nil
}

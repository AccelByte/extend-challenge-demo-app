// Copyright (c) 2025 AccelByte Inc. All Rights Reserved.
// This is licensed software from AccelByte Inc, for limitations
// and restrictions contact your company contract manager.

package output

import (
	"encoding/json"

	"github.com/AccelByte/extend-challenge/extend-challenge-demo-app/internal/ags"
	"github.com/AccelByte/extend-challenge/extend-challenge-demo-app/internal/api"
)

// JSONFormatter formats output as JSON
type JSONFormatter struct{}

// FormatChallenges formats challenges as JSON
func (f *JSONFormatter) FormatChallenges(challenges []api.Challenge) (string, error) {
	output := map[string]interface{}{
		"challenges": challenges,
		"total":      len(challenges),
	}

	data, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return "", err
	}

	return string(data), nil
}

// FormatChallenge formats a single challenge as JSON
func (f *JSONFormatter) FormatChallenge(challenge *api.Challenge) (string, error) {
	data, err := json.MarshalIndent(challenge, "", "  ")
	if err != nil {
		return "", err
	}

	return string(data), nil
}

// FormatEventResult formats an event result as JSON
func (f *JSONFormatter) FormatEventResult(result *EventResult) (string, error) {
	// Convert error to string for JSON output
	output := map[string]interface{}{
		"event":       result.Event,
		"user_id":     result.UserID,
		"timestamp":   result.Timestamp,
		"status":      result.Status,
		"duration_ms": result.DurationMs,
	}

	if result.StatCode != "" {
		output["stat_code"] = result.StatCode
		output["value"] = result.Value
	}

	if result.Error != nil {
		output["error"] = result.Error.Error()
	}

	data, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return "", err
	}

	return string(data), nil
}

// FormatClaimResult formats a claim result as JSON
func (f *JSONFormatter) FormatClaimResult(result *ClaimResult) (string, error) {
	output := map[string]interface{}{
		"challenge_id": result.ChallengeID,
		"goal_id":      result.GoalID,
		"status":       result.Status,
		"timestamp":    result.Timestamp,
	}

	if result.Reward != nil {
		output["reward"] = result.Reward
	}

	if result.Error != nil {
		output["error"] = result.Error.Error()
	}

	data, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return "", err
	}

	return string(data), nil
}

// FormatEntitlement formats a single entitlement as JSON
func (f *JSONFormatter) FormatEntitlement(ent *ags.Entitlement) (string, error) {
	output := map[string]interface{}{
		"entitlement_id": ent.EntitlementID,
		"item_id":        ent.ItemID,
		"namespace":      ent.Namespace,
		"status":         ent.Status,
		"quantity":       ent.Quantity,
		"granted_at":     ent.GrantedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	data, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return "", err
	}

	return string(data), nil
}

// FormatEntitlements formats a list of entitlements as JSON
func (f *JSONFormatter) FormatEntitlements(ents []*ags.Entitlement) (string, error) {
	output := map[string]interface{}{
		"entitlements": ents,
		"total":        len(ents),
	}

	data, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return "", err
	}

	return string(data), nil
}

// FormatWallet formats a single wallet as JSON
func (f *JSONFormatter) FormatWallet(wallet *ags.Wallet) (string, error) {
	output := map[string]interface{}{
		"wallet_id":     wallet.WalletID,
		"currency_code": wallet.CurrencyCode,
		"namespace":     wallet.Namespace,
		"balance":       wallet.Balance,
		"status":        wallet.Status,
	}

	data, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return "", err
	}

	return string(data), nil
}

// FormatWallets formats a list of wallets as JSON
func (f *JSONFormatter) FormatWallets(wallets []*ags.Wallet) (string, error) {
	output := map[string]interface{}{
		"wallets": wallets,
		"total":   len(wallets),
	}

	data, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return "", err
	}

	return string(data), nil
}

// Copyright (c) 2025 AccelByte Inc. All Rights Reserved.
// This is licensed software from AccelByte Inc, for limitations
// and restrictions contact your company contract manager.

package output

import (
	"time"

	"github.com/AccelByte/extend-challenge/extend-challenge-demo-app/internal/ags"
	"github.com/AccelByte/extend-challenge/extend-challenge-demo-app/internal/api"
)

// Formatter formats API responses for CLI output
type Formatter interface {
	// FormatChallenges formats a list of challenges
	FormatChallenges(challenges []api.Challenge) (string, error)

	// FormatChallenge formats a single challenge
	FormatChallenge(challenge *api.Challenge) (string, error)

	// FormatEventResult formats an event trigger result
	FormatEventResult(result *EventResult) (string, error)

	// FormatClaimResult formats a claim reward result
	FormatClaimResult(result *ClaimResult) (string, error)

	// FormatEntitlement formats a single entitlement
	FormatEntitlement(ent *ags.Entitlement) (string, error)

	// FormatEntitlements formats a list of entitlements
	FormatEntitlements(ents []*ags.Entitlement) (string, error)

	// FormatWallet formats a single wallet
	FormatWallet(wallet *ags.Wallet) (string, error)

	// FormatWallets formats a list of wallets
	FormatWallets(wallets []*ags.Wallet) (string, error)
}

// EventResult represents the result of triggering an event
type EventResult struct {
	Event      string        `json:"event"`
	UserID     string        `json:"user_id"`
	StatCode   string        `json:"stat_code,omitempty"`
	Value      int           `json:"value,omitempty"`
	Timestamp  time.Time     `json:"timestamp"`
	Status     string        `json:"status"`
	DurationMs int64         `json:"duration_ms"`
	Error      error         `json:"error,omitempty"`
	ErrorMsg   string        `json:"error_msg,omitempty"`
}

// ClaimResult represents the result of claiming a reward
type ClaimResult struct {
	ChallengeID string     `json:"challenge_id"`
	GoalID      string     `json:"goal_id"`
	Status      string     `json:"status"`
	Reward      *api.Reward `json:"reward,omitempty"`
	Timestamp   time.Time  `json:"timestamp"`
	Error       error      `json:"error,omitempty"`
	ErrorMsg    string     `json:"error_msg,omitempty"`
}

// NewFormatter creates a formatter for the given format type
func NewFormatter(format string) Formatter {
	switch format {
	case "json":
		return &JSONFormatter{}
	case "table":
		return &TableFormatter{}
	case "text":
		return &TextFormatter{}
	default:
		return &JSONFormatter{}
	}
}

package ags

import (
	"time"
)

// Entitlement represents a user's item entitlement in AGS Platform
type Entitlement struct {
	EntitlementID string
	ItemID        string
	Namespace     string
	Status        string // ACTIVE, INACTIVE, etc.
	Quantity      int32
	GrantedAt     time.Time
}

// Wallet represents a user's currency wallet in AGS Platform
type Wallet struct {
	WalletID     string
	CurrencyCode string
	Namespace    string
	Balance      int64
	Status       string // ACTIVE, INACTIVE, etc.
}

// RewardVerifier queries user entitlements and wallets from AGS Platform
type RewardVerifier interface {
	// GetUserEntitlement retrieves a single entitlement by item ID
	GetUserEntitlement(itemID string) (*Entitlement, error)

	// QueryUserEntitlements retrieves all entitlements for the user
	// filters can include: status (ACTIVE/INACTIVE), entitlementClass (ENTITLEMENT/APP/CODE)
	QueryUserEntitlements(filters map[string]string) ([]*Entitlement, error)

	// GetUserWallet retrieves a single wallet by currency code
	GetUserWallet(currencyCode string) (*Wallet, error)

	// QueryUserWallets retrieves all wallets for the user
	QueryUserWallets() ([]*Wallet, error)
}

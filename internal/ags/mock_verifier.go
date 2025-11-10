// Copyright (c) 2025 AccelByte Inc. All Rights Reserved.
// This is licensed software from AccelByte Inc, for limitations
// and restrictions contact your company contract manager.

package ags

import (
	"fmt"
	"time"
)

// MockRewardVerifier is a mock implementation for testing
type MockRewardVerifier struct {
	Entitlements []*Entitlement
	Wallets      []*Wallet
	Error        error
}

// NewMockRewardVerifier creates a new mock verifier with sample data
func NewMockRewardVerifier() *MockRewardVerifier {
	return &MockRewardVerifier{
		Entitlements: []*Entitlement{
			{
				EntitlementID: "ent-mock-1",
				ItemID:        "winter_sword",
				Namespace:     "demo",
				Status:        "ACTIVE",
				Quantity:      1,
				GrantedAt:     time.Now().Add(-24 * time.Hour),
			},
			{
				EntitlementID: "ent-mock-2",
				ItemID:        "bronze_shield",
				Namespace:     "demo",
				Status:        "ACTIVE",
				Quantity:      2,
				GrantedAt:     time.Now().Add(-48 * time.Hour),
			},
		},
		Wallets: []*Wallet{
			{
				WalletID:     "wallet-mock-1",
				CurrencyCode: "GOLD",
				Namespace:    "demo",
				Balance:      150,
				Status:       "ACTIVE",
			},
			{
				WalletID:     "wallet-mock-2",
				CurrencyCode: "GEMS",
				Namespace:    "demo",
				Balance:      25,
				Status:       "ACTIVE",
			},
		},
	}
}

// GetUserEntitlement retrieves a single entitlement by item ID
func (m *MockRewardVerifier) GetUserEntitlement(itemID string) (*Entitlement, error) {
	if m.Error != nil {
		return nil, m.Error
	}

	for _, ent := range m.Entitlements {
		if ent.ItemID == itemID {
			return ent, nil
		}
	}

	return nil, fmt.Errorf("entitlement not found for item: %s", itemID)
}

// QueryUserEntitlements retrieves all entitlements for the user
func (m *MockRewardVerifier) QueryUserEntitlements(filters map[string]string) ([]*Entitlement, error) {
	if m.Error != nil {
		return nil, m.Error
	}

	// Apply filters if provided
	if status, ok := filters["status"]; ok {
		filtered := make([]*Entitlement, 0)
		for _, ent := range m.Entitlements {
			if ent.Status == status {
				filtered = append(filtered, ent)
			}
		}
		return filtered, nil
	}

	return m.Entitlements, nil
}

// GetUserWallet retrieves a single wallet by currency code
func (m *MockRewardVerifier) GetUserWallet(currencyCode string) (*Wallet, error) {
	if m.Error != nil {
		return nil, m.Error
	}

	for _, wallet := range m.Wallets {
		if wallet.CurrencyCode == currencyCode {
			return wallet, nil
		}
	}

	return nil, fmt.Errorf("wallet not found for currency: %s", currencyCode)
}

// QueryUserWallets retrieves all wallets for the user
func (m *MockRewardVerifier) QueryUserWallets() ([]*Wallet, error) {
	if m.Error != nil {
		return nil, m.Error
	}

	return m.Wallets, nil
}

// Copyright (c) 2025 AccelByte Inc. All Rights Reserved.
// This is licensed software from AccelByte Inc, for limitations
// and restrictions contact your company contract manager.

package ags

import (
	"context"
	"fmt"
	"time"

	"github.com/AccelByte/accelbyte-go-sdk/platform-sdk/pkg/platformclient/entitlement"
	"github.com/AccelByte/accelbyte-go-sdk/platform-sdk/pkg/platformclient/wallet"
	"github.com/AccelByte/accelbyte-go-sdk/services-api/pkg/service/platform"
)

// AGSRewardVerifier implements RewardVerifier using AccelByte Platform SDK
type AGSRewardVerifier struct {
	entitlementSvc    *platform.EntitlementService
	walletSvc         *platform.WalletService
	userID            string
	namespace         string
	maxRetries        int
	initialRetryDelay time.Duration
}

// NewAGSRewardVerifier creates a new AGS reward verifier
// Parameters:
//   - entitlementSvc: Platform SDK entitlement service (pre-configured with auth)
//   - walletSvc: Platform SDK wallet service (pre-configured with auth)
//   - userID: User ID to query rewards for
//   - namespace: AGS namespace
func NewAGSRewardVerifier(
	entitlementSvc *platform.EntitlementService,
	walletSvc *platform.WalletService,
	userID string,
	namespace string,
) *AGSRewardVerifier {
	return &AGSRewardVerifier{
		entitlementSvc:    entitlementSvc,
		walletSvc:         walletSvc,
		userID:            userID,
		namespace:         namespace,
		maxRetries:        3,
		initialRetryDelay: 500 * time.Millisecond,
	}
}

// GetUserEntitlement retrieves a single entitlement by item ID
func (v *AGSRewardVerifier) GetUserEntitlement(itemID string) (*Entitlement, error) {
	return v.getUserEntitlementWithRetry(itemID)
}

// QueryUserEntitlements retrieves all entitlements for the user
func (v *AGSRewardVerifier) QueryUserEntitlements(filters map[string]string) ([]*Entitlement, error) {
	return v.queryUserEntitlementsWithRetry(filters)
}

// GetUserWallet retrieves a single wallet by currency code
func (v *AGSRewardVerifier) GetUserWallet(currencyCode string) (*Wallet, error) {
	return v.getUserWalletWithRetry(currencyCode)
}

// QueryUserWallets retrieves all wallets for the user
func (v *AGSRewardVerifier) QueryUserWallets() ([]*Wallet, error) {
	return v.queryUserWalletsWithRetry()
}

// getUserEntitlementWithRetry implements retry logic for GetUserEntitlement
func (v *AGSRewardVerifier) getUserEntitlementWithRetry(itemID string) (*Entitlement, error) {
	var lastErr error
	retryDelay := v.initialRetryDelay

	for attempt := 0; attempt <= v.maxRetries; attempt++ {
		if attempt > 0 {
			time.Sleep(retryDelay)
			retryDelay *= 2 // Exponential backoff
		}

		ent, err := v.doGetUserEntitlement(itemID)
		if err == nil {
			return ent, nil
		}

		// Check if error is retryable
		if !isRetryable(err) {
			return nil, err
		}

		lastErr = err
	}

	return nil, fmt.Errorf("failed after %d retries: %w", v.maxRetries, lastErr)
}

// doGetUserEntitlement performs the actual API call
func (v *AGSRewardVerifier) doGetUserEntitlement(itemID string) (*Entitlement, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Create params
	params := &entitlement.GetUserEntitlementByItemIDParams{
		Namespace: v.namespace,
		UserID:    v.userID,
		ItemID:    itemID,
	}
	params.SetContext(ctx)

	// Call SDK (auth is handled by the service)
	resp, err := v.entitlementSvc.GetUserEntitlementByItemIDShort(params)
	if err != nil {
		return nil, fmt.Errorf("get entitlement failed: %w", err)
	}

	if resp == nil {
		return nil, fmt.Errorf("empty response")
	}

	// Convert to our domain model
	ent := &Entitlement{
		Namespace: v.namespace,
		ItemID:    itemID,
	}

	if resp.ID != nil {
		ent.EntitlementID = *resp.ID
	}
	if resp.Status != nil {
		ent.Status = string(*resp.Status)
	}
	if resp.UseCount != 0 {
		ent.Quantity = resp.UseCount
	}
	if resp.GrantedAt != nil {
		// Convert strfmt.DateTime to time.Time
		grantedTime, err := time.Parse(time.RFC3339, resp.GrantedAt.String())
		if err == nil {
			ent.GrantedAt = grantedTime
		}
	}

	return ent, nil
}

// queryUserEntitlementsWithRetry implements retry logic for QueryUserEntitlements
func (v *AGSRewardVerifier) queryUserEntitlementsWithRetry(filters map[string]string) ([]*Entitlement, error) {
	var lastErr error
	retryDelay := v.initialRetryDelay

	for attempt := 0; attempt <= v.maxRetries; attempt++ {
		if attempt > 0 {
			time.Sleep(retryDelay)
			retryDelay *= 2
		}

		ents, err := v.doQueryUserEntitlements(filters)
		if err == nil {
			return ents, nil
		}

		if !isRetryable(err) {
			return nil, err
		}

		lastErr = err
	}

	return nil, fmt.Errorf("failed after %d retries: %w", v.maxRetries, lastErr)
}

// doQueryUserEntitlements performs the actual API call
func (v *AGSRewardVerifier) doQueryUserEntitlements(filters map[string]string) ([]*Entitlement, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Prepare params
	params := &entitlement.QueryUserEntitlementsParams{
		Namespace: v.namespace,
		UserID:    v.userID,
	}
	params.SetContext(ctx)

	// Apply filters
	if entitlementClass, ok := filters["entitlementClass"]; ok {
		params.EntitlementClazz = &entitlementClass
	}
	if status, ok := filters["status"]; ok {
		params.EntitlementName = &status
	}

	// Call SDK
	resp, err := v.entitlementSvc.QueryUserEntitlementsShort(params)
	if err != nil {
		return nil, fmt.Errorf("query entitlements failed: %w", err)
	}

	if resp == nil || resp.Data == nil {
		// Empty list is valid
		return []*Entitlement{}, nil
	}

	// Convert to our domain models
	entitlements := make([]*Entitlement, 0, len(resp.Data))
	for _, e := range resp.Data {
		if e == nil {
			continue
		}

		ent := &Entitlement{
			Namespace: v.namespace,
		}

		if e.ID != nil {
			ent.EntitlementID = *e.ID
		}
		if e.ItemID != nil {
			ent.ItemID = *e.ItemID
		}
		if e.Status != nil {
			ent.Status = string(*e.Status)
		}
		if e.UseCount != 0 {
			ent.Quantity = e.UseCount
		}
		if e.GrantedAt != nil {
			// Convert strfmt.DateTime to time.Time
			grantedTime, err := time.Parse(time.RFC3339, e.GrantedAt.String())
			if err == nil {
				ent.GrantedAt = grantedTime
			}
		}

		entitlements = append(entitlements, ent)
	}

	return entitlements, nil
}

// getUserWalletWithRetry implements retry logic for GetUserWallet
func (v *AGSRewardVerifier) getUserWalletWithRetry(currencyCode string) (*Wallet, error) {
	var lastErr error
	retryDelay := v.initialRetryDelay

	for attempt := 0; attempt <= v.maxRetries; attempt++ {
		if attempt > 0 {
			time.Sleep(retryDelay)
			retryDelay *= 2
		}

		w, err := v.doGetUserWallet(currencyCode)
		if err == nil {
			return w, nil
		}

		if !isRetryable(err) {
			return nil, err
		}

		lastErr = err
	}

	return nil, fmt.Errorf("failed after %d retries: %w", v.maxRetries, lastErr)
}

// doGetUserWallet performs the actual API call
func (v *AGSRewardVerifier) doGetUserWallet(currencyCode string) (*Wallet, error) {
	// Note: The admin wallet endpoint requires wallet UUID, not currency code.
	// Instead, we query all wallets and filter by currency code.
	wallets, err := v.doQueryUserWallets()
	if err != nil {
		return nil, fmt.Errorf("query wallets failed: %w", err)
	}

	// Find wallet matching the currency code
	for _, w := range wallets {
		if w.CurrencyCode == currencyCode {
			return w, nil
		}
	}

	// Wallet not found for this currency
	return nil, fmt.Errorf("wallet with currency code %s not found", currencyCode)
}

// queryUserWalletsWithRetry implements retry logic for QueryUserWallets
func (v *AGSRewardVerifier) queryUserWalletsWithRetry() ([]*Wallet, error) {
	var lastErr error
	retryDelay := v.initialRetryDelay

	for attempt := 0; attempt <= v.maxRetries; attempt++ {
		if attempt > 0 {
			time.Sleep(retryDelay)
			retryDelay *= 2
		}

		wallets, err := v.doQueryUserWallets()
		if err == nil {
			return wallets, nil
		}

		if !isRetryable(err) {
			return nil, err
		}

		lastErr = err
	}

	return nil, fmt.Errorf("failed after %d retries: %w", v.maxRetries, lastErr)
}

// doQueryUserWallets performs the actual API call
func (v *AGSRewardVerifier) doQueryUserWallets() ([]*Wallet, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Call SDK
	params := &wallet.QueryUserCurrencyWalletsParams{
		Namespace: v.namespace,
		UserID:    v.userID,
	}
	params.SetContext(ctx)

	resp, err := v.walletSvc.QueryUserCurrencyWalletsShort(params)
	if err != nil {
		return nil, fmt.Errorf("query wallets failed: %w", err)
	}

	if resp == nil {
		// Empty list is valid
		return []*Wallet{}, nil
	}

	// Convert to our domain models
	wallets := make([]*Wallet, 0, len(resp))
	for _, w := range resp {
		if w == nil {
			continue
		}

		wallet := &Wallet{
			Namespace: v.namespace,
		}

		// Extract fields from CurrencyWallet (these are pointers in SDK)
		if w.CurrencyCode != nil {
			wallet.CurrencyCode = *w.CurrencyCode
		}
		if w.Balance != nil {
			wallet.Balance = *w.Balance
		}

		// Extract WalletID and Status from the first WalletInfo
		if len(w.WalletInfos) > 0 && w.WalletInfos[0] != nil {
			if w.WalletInfos[0].ID != nil {
				wallet.WalletID = *w.WalletInfos[0].ID
			}
			if w.WalletInfos[0].Status != nil {
				wallet.Status = *w.WalletInfos[0].Status
			}
		}

		wallets = append(wallets, wallet)
	}

	return wallets, nil
}

// isRetryable checks if an error should trigger a retry
func isRetryable(err error) bool {
	if err == nil {
		return false
	}

	// Check error message for retryable conditions
	errStr := err.Error()

	// Timeout errors are retryable
	if contains(errStr, "timeout") || contains(errStr, "deadline exceeded") {
		return true
	}

	// Connection errors are retryable
	if contains(errStr, "connection refused") || contains(errStr, "no such host") {
		return true
	}

	// 5xx errors are retryable
	if contains(errStr, "500") || contains(errStr, "502") || contains(errStr, "503") || contains(errStr, "504") {
		return true
	}

	// 429 rate limit is retryable
	if contains(errStr, "429") || contains(errStr, "rate limit") {
		return true
	}

	// Default: not retryable (4xx client errors, 404, etc.)
	return false
}

// contains checks if a string contains a substring (case-insensitive)
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) > 0 && containsImpl(s, substr))
}

func containsImpl(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

# Reward Verification Testing Guide

**Date:** 2025-10-22
**Status:** Ready for Real AGS Testing

## Overview

The reward verification feature is now **fully implemented with real AGS Platform SDK integration**. This guide shows you how to test with:
1. **Mock mode** (no credentials required) - for development
2. **Real AGS mode** (requires AGS credentials) - for production testing

---

## Quick Start: Mock Mode (No Setup Required)

Mock mode works immediately with no configuration:

```bash
# Verify entitlement (returns mock data)
./challenge-demo verify-entitlement --item-id=winter_sword --format=json

# Check wallet balance
./challenge-demo verify-wallet --currency=GOLD --format=table

# List all entitlements
./challenge-demo list-inventory --format=text

# List all wallets
./challenge-demo list-wallets --format=table
```

**Mock data includes:**
- 2 entitlements: `winter_sword` (qty: 1), `bronze_shield` (qty: 2)
- 2 wallets: `GOLD` (balance: 150), `GEMS` (balance: 25)

---

## Real AGS Testing

### Prerequisites

You need AGS credentials with permissions for:
- **IAM Service**: User authentication (OAuth2)
- **Platform Service**: Entitlements and Wallets (read access)

### Step 1: Set Environment Variables

The AccelByte SDK reads configuration from environment variables:

```bash
export AB_CLIENT_ID="your-oauth-client-id"
export AB_CLIENT_SECRET="your-oauth-client-secret"
export AB_BASE_URL="https://demo.accelbyte.io"  # Your AGS base URL
export AB_NAMESPACE="your-namespace"
```

**How to get these values:**
1. Log in to AccelByte Admin Portal
2. Go to **Settings → OAuth Clients**
3. Create or use existing client with permissions:
   - `NAMESPACE:{namespace}:ENTITLEMENT [READ]`
   - `NAMESPACE:{namespace}:WALLET [READ]`
4. Copy Client ID and Client Secret

### Step 2: Authenticate as User

You need to authenticate as a specific user to query their entitlements/wallets:

```bash
./challenge-demo verify-entitlement \
  --item-id=actual_item_id \
  --auth-mode=password \
  --email=user@example.com \
  --password=user_password \
  --namespace=your-namespace
```

**Auth Flow:**
1. Demo app uses `PasswordAuthProvider` to get user access token from IAM
2. Platform SDK uses this token to authenticate API calls
3. Returns real entitlements/wallets for that user

### Step 3: Test Commands

#### 1. Verify Single Entitlement

```bash
./challenge-demo verify-entitlement \
  --item-id=sword001 \
  --auth-mode=password \
  --email=testuser@example.com \
  --password=Test123! \
  --namespace=mygame \
  --format=json
```

**Expected Output (if user has entitlement):**
```json
{
  "entitlement_id": "ent-abc-123",
  "granted_at": "2025-10-20T15:30:00Z",
  "item_id": "sword001",
  "namespace": "mygame",
  "quantity": 1,
  "status": "ACTIVE"
}
```

**Expected Error (if not found):**
```
Error: failed to query entitlements: get entitlement failed: ...
```

#### 2. List All Entitlements

```bash
./challenge-demo list-inventory \
  --auth-mode=password \
  --email=testuser@example.com \
  --password=Test123! \
  --namespace=mygame \
  --format=table
```

**Expected Output:**
```
ENTITLEMENT_ID       ITEM_ID                        STATUS     QUANTITY   GRANTED_AT
-------------------------------------------------------------------------------------------
ent-abc-123          sword001                       ACTIVE     1          2025-10-20 15:30
ent-abc-456          shield002                      ACTIVE     2          2025-10-19 10:00

Total: 2 entitlements
```

#### 3. Verify Wallet Balance

```bash
./challenge-demo verify-wallet \
  --currency=GOLD \
  --auth-mode=password \
  --email=testuser@example.com \
  --password=Test123! \
  --namespace=mygame \
  --format=json
```

**Expected Output:**
```json
{
  "balance": 1500,
  "currency_code": "GOLD",
  "namespace": "mygame",
  "status": "ACTIVE",
  "wallet_id": "wallet-xyz-789"
}
```

#### 4. List All Wallets

```bash
./challenge-demo list-wallets \
  --auth-mode=password \
  --email=testuser@example.com \
  --password=Test123! \
  --namespace=mygame \
  --format=table
```

**Expected Output:**
```
WALLET_ID            CURRENCY        BALANCE         STATUS
------------------------------------------------------------
wallet-xyz-789       GOLD            1500            ACTIVE
wallet-xyz-790       GEMS            250             ACTIVE

Total: 2 wallets
```

---

## Troubleshooting

### Error: "401 Unauthorized"

**Problem:** SDK cannot authenticate with AGS

**Solutions:**
1. Check environment variables are set correctly:
   ```bash
   echo $AB_CLIENT_ID
   echo $AB_BASE_URL
   echo $AB_NAMESPACE
   ```

2. Verify OAuth client has correct permissions in Admin Portal

3. Check user credentials are correct (email/password)

4. Verify AB_BASE_URL format (should be `https://demo.accelbyte.io` without `/iam` suffix)

### Error: "connection refused" or "timeout"

**Problem:** Cannot reach AGS servers

**Solutions:**
1. Check network connectivity:
   ```bash
   curl https://demo.accelbyte.io/iam/healthz
   ```

2. Verify firewall/proxy settings allow HTTPS to AGS

3. Check AB_BASE_URL is correct

### Error: "failed to query entitlements" (200 OK but empty)

**Problem:** User has no entitlements/wallets

**This is normal!** User simply doesn't have any items/currency yet.

**To test:**
1. Grant entitlement via Admin Portal or Platform API
2. Credit wallet via Admin Portal
3. Re-run verification command

### Error: "item not found" for verify-entitlement

**Problem:** User doesn't have that specific item

**Solutions:**
1. Check item ID exists in your catalog (Admin Portal → Items)
2. Grant entitlement to user first
3. Use `list-inventory` to see what user actually has

---

## Implementation Details

### SDK Authentication Flow

```
1. Environment Variables (AB_CLIENT_ID, AB_CLIENT_SECRET, AB_BASE_URL, AB_NAMESPACE)
   ↓
2. ConfigRepository (sdkAuth.DefaultConfigRepositoryImpl)
   ↓
3. TokenRepository (sdkAuth.DefaultTokenRepositoryImpl)
   ↓
4. Platform SDK Client (factory.NewPlatformClient)
   ↓
5. Platform Services (EntitlementService, WalletService)
   ↓
6. AGSRewardVerifier (makes authenticated API calls)
```

### Key Files

- **`internal/app/container.go:104-124`** - SDK authentication setup
- **`internal/ags/ags_verifier.go`** - Real AGS implementation with retry logic
- **`internal/ags/mock_verifier.go`** - Mock implementation for testing

### Retry Logic

The AGS verifier implements **exponential backoff retry** for transient failures:
- **Max retries:** 3
- **Delays:** 500ms → 1s → 2s
- **Retryable errors:** 500, 502, 503, 504, timeouts, connection errors
- **Non-retryable:** 400, 401, 403, 404 (fail immediately)

---

## Next Steps

### For Development
- ✅ Use mock mode (no setup required)
- ✅ All commands work with sample data
- ✅ Perfect for local development

### For Production Testing
1. Set up AGS environment variables (see Step 1 above)
2. Create test user account in your namespace
3. Grant test entitlements/wallets via Admin Portal
4. Run commands with `--auth-mode=password`
5. Verify data matches Admin Portal

### For Automated Testing
- Unit tests for `MockRewardVerifier` (no external deps)
- Integration tests for `AGSRewardVerifier` (requires test AGS environment)
- E2E tests for CLI commands

---

## Summary

✅ **Implementation Status:** COMPLETE
✅ **Mock Mode:** Working perfectly
✅ **Real AGS SDK:** Ready for testing (authentication configured)
✅ **Build & Linter:** Zero issues

**What's needed to test with real AGS:**
1. Set 4 environment variables (`AB_CLIENT_ID`, `AB_CLIENT_SECRET`, `AB_BASE_URL`, `AB_NAMESPACE`)
2. Run commands with `--auth-mode=password --email=... --password=...`
3. Verify results match AGS Admin Portal

**No code changes needed** - just configuration!

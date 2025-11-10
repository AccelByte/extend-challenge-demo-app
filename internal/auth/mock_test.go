// Copyright (c) 2025 AccelByte Inc. All Rights Reserved.
// This is licensed software from AccelByte Inc, for limitations
// and restrictions contact your company contract manager.

package auth

import (
	"context"
	"testing"
	"time"
)

func TestNewMockAuthProvider(t *testing.T) {
	provider := NewMockAuthProvider("test-user-123", "demo")

	if provider == nil {
		t.Fatal("Expected non-nil provider")
	}

	if provider.token == nil {
		t.Fatal("Expected non-nil token")
	}

	if provider.token.TokenType != "Bearer" {
		t.Errorf("Expected token type 'Bearer', got '%s'", provider.token.TokenType)
	}

	if provider.token.AccessToken == "" {
		t.Error("Expected non-empty access token")
	}

	if provider.userID != "test-user-123" {
		t.Errorf("Expected userID 'test-user-123', got '%s'", provider.userID)
	}

	if provider.namespace != "demo" {
		t.Errorf("Expected namespace 'demo', got '%s'", provider.namespace)
	}
}

func TestMockAuthProvider_Authenticate(t *testing.T) {
	provider := NewMockAuthProvider("alice", "demo")
	ctx := context.Background()

	token, err := provider.Authenticate(ctx)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if token == nil {
		t.Fatal("Expected non-nil token")
	}

	if token.TokenType != "Bearer" {
		t.Errorf("Expected token type 'Bearer', got '%s'", token.TokenType)
	}
}

func TestMockAuthProvider_GetToken(t *testing.T) {
	provider := NewMockAuthProvider("bob", "test-ns")
	ctx := context.Background()

	token, err := provider.GetToken(ctx)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if token == nil {
		t.Fatal("Expected non-nil token")
	}

	if token.IsExpired() {
		t.Error("Token should not be expired")
	}
}

func TestMockAuthProvider_GetToken_Expired(t *testing.T) {
	provider := NewMockAuthProvider("charlie", "demo")
	// Force token to expire
	provider.token.ExpiresAt = time.Now().Add(-1 * time.Hour)

	ctx := context.Background()
	token, err := provider.GetToken(ctx)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if token.IsExpired() {
		t.Error("Token should have been refreshed and not be expired")
	}
}

func TestMockAuthProvider_RefreshToken(t *testing.T) {
	provider := NewMockAuthProvider("dave", "staging")
	oldToken := provider.token
	ctx := context.Background()

	newToken, err := provider.RefreshToken(ctx, oldToken)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if newToken == nil {
		t.Fatal("Expected non-nil token")
	}

	if newToken.AccessToken == "" {
		t.Error("Expected non-empty access token")
	}

	if newToken.TokenType != "Bearer" {
		t.Errorf("Expected token type 'Bearer', got '%s'", newToken.TokenType)
	}
}

func TestMockAuthProvider_IsTokenValid(t *testing.T) {
	provider := NewMockAuthProvider("eve", "prod")

	tests := []struct {
		name   string
		token  *Token
		expect bool
	}{
		{
			name:   "valid token",
			token:  provider.token,
			expect: true,
		},
		{
			name:   "nil token",
			token:  nil,
			expect: false,
		},
		{
			name: "expired token",
			token: &Token{
				AccessToken: "test",
				ExpiresAt:   time.Now().Add(-1 * time.Hour),
			},
			expect: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := provider.IsTokenValid(tt.token)
			if result != tt.expect {
				t.Errorf("Expected %v, got %v", tt.expect, result)
			}
		})
	}
}

// TestMockAuthProvider_DifferentUsers verifies different users get different tokens
func TestMockAuthProvider_DifferentUsers(t *testing.T) {
	providerAlice := NewMockAuthProvider("alice", "demo")
	providerBob := NewMockAuthProvider("bob", "demo")

	ctx := context.Background()
	tokenAlice, err := providerAlice.GetToken(ctx)
	if err != nil {
		t.Fatalf("Unexpected error for alice: %v", err)
	}

	tokenBob, err := providerBob.GetToken(ctx)
	if err != nil {
		t.Fatalf("Unexpected error for bob: %v", err)
	}

	// Tokens should be different (different user_id in payload)
	if tokenAlice.AccessToken == tokenBob.AccessToken {
		t.Error("Expected different tokens for different users")
	}
}

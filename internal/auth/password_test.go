// Copyright (c) 2025 AccelByte Inc. All Rights Reserved.
// This is licensed software from AccelByte Inc, for limitations
// and restrictions contact your company contract manager.

package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewPasswordAuthProvider(t *testing.T) {
	provider := NewPasswordAuthProvider(
		"https://demo.accelbyte.io/iam",
		"client-id",
		"client-secret",
		"demo",
		"alice@example.com",
		"password123",
	)

	if provider == nil {
		t.Fatal("Expected non-nil provider")
	}

	if provider.email != "alice@example.com" {
		t.Errorf("Expected email 'alice@example.com', got '%s'", provider.email)
	}

	if provider.namespace != "demo" {
		t.Errorf("Expected namespace 'demo', got '%s'", provider.namespace)
	}
}

func TestPasswordAuthProvider_Authenticate(t *testing.T) {
	// Create mock IAM server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request
		if r.Method != "POST" {
			t.Errorf("Expected POST, got %s", r.Method)
		}

		if r.URL.Path != "/iam/v3/oauth/token" {
			t.Errorf("Expected /iam/v3/oauth/token, got %s", r.URL.Path)
		}

		if r.Header.Get("Content-Type") != "application/x-www-form-urlencoded" {
			t.Errorf("Expected application/x-www-form-urlencoded, got %s", r.Header.Get("Content-Type"))
		}

		// Check Basic Auth
		username, password, ok := r.BasicAuth()
		if !ok {
			t.Error("Expected Basic Auth")
		}
		if username != "test-client" {
			t.Errorf("Expected username 'test-client', got '%s'", username)
		}
		if password != "test-secret" {
			t.Errorf("Expected password 'test-secret', got '%s'", password)
		}

		// Parse form
		if err := r.ParseForm(); err != nil {
			t.Errorf("Failed to parse form: %v", err)
		}

		// Verify grant_type
		if r.Form.Get("grant_type") != "password" {
			t.Errorf("Expected grant_type 'password', got '%s'", r.Form.Get("grant_type"))
		}

		// Verify username (email)
		if r.Form.Get("username") != "alice@example.com" {
			t.Errorf("Expected username 'alice@example.com', got '%s'", r.Form.Get("username"))
		}

		// Return token response
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"access_token":  "test-user-token",
			"token_type":    "Bearer",
			"expires_in":    3600,
			"refresh_token": "test-refresh-token",
			"user_id":       "user-123",
			"namespace":     "demo",
		})
	}))
	defer server.Close()

	// Create provider
	provider := NewPasswordAuthProvider(
		server.URL,
		"test-client",
		"test-secret",
		"demo",
		"alice@example.com",
		"password123",
	)

	// Test authenticate
	ctx := context.Background()
	token, err := provider.Authenticate(ctx)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if token == nil {
		t.Fatal("Expected non-nil token")
	}

	if token.AccessToken != "test-user-token" {
		t.Errorf("Expected access token 'test-user-token', got '%s'", token.AccessToken)
	}

	if token.TokenType != "Bearer" {
		t.Errorf("Expected token type 'Bearer', got '%s'", token.TokenType)
	}

	if token.RefreshToken != "test-refresh-token" {
		t.Errorf("Expected refresh token 'test-refresh-token', got '%s'", token.RefreshToken)
	}

	if token.IsExpired() {
		t.Error("Token should not be expired")
	}
}

func TestPasswordAuthProvider_Authenticate_InvalidCredentials(t *testing.T) {
	// Create mock IAM server that returns 401
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer server.Close()

	provider := NewPasswordAuthProvider(
		server.URL,
		"test-client",
		"test-secret",
		"demo",
		"alice@example.com",
		"wrong-password",
	)

	ctx := context.Background()
	token, err := provider.Authenticate(ctx)

	if err == nil {
		t.Fatal("Expected error for invalid credentials")
	}

	if token != nil {
		t.Error("Expected nil token for invalid credentials")
	}

	// SDK returns error with prefix "password grant failed:"
	// Just check that we got an error (SDK error format differs from manual HTTP)
	if err == nil {
		t.Error("Expected error for 401 response")
	}
}

func TestPasswordAuthProvider_RefreshToken(t *testing.T) {
	callCount := 0

	// Create mock IAM server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++

		// Parse form
		if err := r.ParseForm(); err != nil {
			t.Errorf("Failed to parse form: %v", err)
		}

		grantType := r.Form.Get("grant_type")

		switch grantType {
		case "password":
			// First call: authenticate
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"access_token":  "initial-token",
				"token_type":    "Bearer",
				"expires_in":    3600,
				"refresh_token": "refresh-token-1",
			})
		case "refresh_token":
			// Second call: refresh
			if r.Form.Get("refresh_token") != "refresh-token-1" {
				t.Errorf("Expected refresh_token 'refresh-token-1', got '%s'", r.Form.Get("refresh_token"))
			}

			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"access_token":  "refreshed-token",
				"token_type":    "Bearer",
				"expires_in":    3600,
				"refresh_token": "refresh-token-2",
			})
		}
	}))
	defer server.Close()

	provider := NewPasswordAuthProvider(
		server.URL,
		"test-client",
		"test-secret",
		"demo",
		"alice@example.com",
		"password123",
	)

	// First authenticate
	ctx := context.Background()
	initialToken, err := provider.Authenticate(ctx)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Then refresh
	refreshedToken, err := provider.RefreshToken(ctx, initialToken)
	if err != nil {
		t.Fatalf("Unexpected error during refresh: %v", err)
	}

	if refreshedToken.AccessToken != "refreshed-token" {
		t.Errorf("Expected refreshed token 'refreshed-token', got '%s'", refreshedToken.AccessToken)
	}

	if refreshedToken.RefreshToken != "refresh-token-2" {
		t.Errorf("Expected refresh token 'refresh-token-2', got '%s'", refreshedToken.RefreshToken)
	}

	if callCount != 2 {
		t.Errorf("Expected 2 calls to IAM server, got %d", callCount)
	}
}

func TestPasswordAuthProvider_GetToken(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"access_token":  "test-token",
			"token_type":    "Bearer",
			"expires_in":    3600,
			"refresh_token": "test-refresh",
		})
	}))
	defer server.Close()

	provider := NewPasswordAuthProvider(
		server.URL,
		"test-client",
		"test-secret",
		"demo",
		"alice@example.com",
		"password123",
	)

	ctx := context.Background()

	// First call should authenticate
	token1, err := provider.GetToken(ctx)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Second call should return cached token
	token2, err := provider.GetToken(ctx)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if token1.AccessToken != token2.AccessToken {
		t.Error("Expected same token on second call (cached)")
	}
}

func TestPasswordAuthProvider_GetToken_Expired(t *testing.T) {
	callCount := 0

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"access_token":  "new-token",
			"token_type":    "Bearer",
			"expires_in":    3600,
			"refresh_token": "new-refresh",
		})
	}))
	defer server.Close()

	provider := NewPasswordAuthProvider(
		server.URL,
		"test-client",
		"test-secret",
		"demo",
		"alice@example.com",
		"password123",
	)

	// Manually set expired token
	provider.currentToken = &Token{
		AccessToken:  "expired-token",
		TokenType:    "Bearer",
		ExpiresAt:    time.Now().Add(-1 * time.Hour),
		RefreshToken: "old-refresh",
	}

	ctx := context.Background()
	token, err := provider.GetToken(ctx)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if token.AccessToken == "expired-token" {
		t.Error("Expected new token, got expired token")
	}

	if token.AccessToken != "new-token" {
		t.Errorf("Expected 'new-token', got '%s'", token.AccessToken)
	}

	if callCount != 1 {
		t.Errorf("Expected 1 call to refresh, got %d", callCount)
	}
}

func TestPasswordAuthProvider_IsTokenValid(t *testing.T) {
	provider := NewPasswordAuthProvider(
		"https://demo.accelbyte.io/iam",
		"client-id",
		"client-secret",
		"demo",
		"alice@example.com",
		"password123",
	)

	tests := []struct {
		name   string
		token  *Token
		expect bool
	}{
		{
			name: "valid token",
			token: &Token{
				AccessToken: "test",
				ExpiresAt:   time.Now().Add(1 * time.Hour),
			},
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

// Copyright (c) 2025 AccelByte Inc. All Rights Reserved.
// This is licensed software from AccelByte Inc, for limitations
// and restrictions contact your company contract manager.

package auth

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"
)

// MockAuthProvider implements AuthProvider with a static token for local development
type MockAuthProvider struct {
	token     *Token
	userID    string // User ID to embed in JWT
	namespace string // Namespace to embed in JWT
}

// NewMockAuthProvider creates a new mock auth provider
// Parameters:
//   - userID: User ID to include in JWT "sub" claim (from --user-id CLI flag)
//   - namespace: Namespace to include in JWT "namespace" claim (from --namespace CLI flag)
func NewMockAuthProvider(userID, namespace string) *MockAuthProvider {
	// Create a static token that expires in 1 hour
	token := &Token{
		AccessToken:  generateMockJWT(userID, namespace),
		TokenType:    "Bearer",
		ExpiresAt:    time.Now().Add(1 * time.Hour),
		RefreshToken: "",
	}

	return &MockAuthProvider{
		token:     token,
		userID:    userID,
		namespace: namespace,
	}
}

// Authenticate returns the static token
func (p *MockAuthProvider) Authenticate(ctx context.Context) (*Token, error) {
	return p.token, nil
}

// RefreshToken returns a new static token
func (p *MockAuthProvider) RefreshToken(ctx context.Context, token *Token) (*Token, error) {
	// Generate new token with 1 hour expiry using stored userID and namespace
	newToken := &Token{
		AccessToken:  generateMockJWT(p.userID, p.namespace),
		TokenType:    "Bearer",
		ExpiresAt:    time.Now().Add(1 * time.Hour),
		RefreshToken: "",
	}

	p.token = newToken
	return newToken, nil
}

// GetToken returns the current static token
func (p *MockAuthProvider) GetToken(ctx context.Context) (*Token, error) {
	// If expired, refresh
	if p.token.IsExpired() {
		return p.RefreshToken(ctx, p.token)
	}
	return p.token, nil
}

// IsTokenValid checks if token is valid
func (p *MockAuthProvider) IsTokenValid(token *Token) bool {
	if token == nil {
		return false
	}
	return !token.IsExpired()
}

// generateMockJWT generates a mock JWT token
// Parameters:
//   - userID: User ID to embed in "sub" claim (allows testing different users)
//   - namespace: Namespace to embed in "namespace" claim
//
// Note: This is NOT cryptographically valid. For local development,
// run backend with PLUGIN_GRPC_SERVER_AUTH_ENABLED=false
func generateMockJWT(userID, namespace string) string {
	header := map[string]interface{}{
		"alg": "HS256",
		"typ": "JWT",
	}

	payload := map[string]interface{}{
		"sub":       userID,    // Use parameter, not hardcoded
		"namespace": namespace, // Use parameter, not hardcoded
		"iat":       time.Now().Unix(),
		"exp":       time.Now().Add(1 * time.Hour).Unix(),
	}

	// Encode header and payload (no signature for mock)
	headerJSON, _ := json.Marshal(header)
	payloadJSON, _ := json.Marshal(payload)

	headerB64 := base64.RawURLEncoding.EncodeToString(headerJSON)
	payloadB64 := base64.RawURLEncoding.EncodeToString(payloadJSON)

	// Mock JWT: header.payload.mock-signature
	return fmt.Sprintf("%s.%s.mock-signature", headerB64, payloadB64)
}

// Copyright (c) 2025 AccelByte Inc. All Rights Reserved.
// This is licensed software from AccelByte Inc, for limitations
// and restrictions contact your company contract manager.

package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

// ClientAuthProvider implements AuthProvider using AGS IAM OAuth2 Client Credentials
// This is used for SERVICE authentication (client_id + secret → service token)
// WARNING: This token does NOT have a user_id in the "sub" claim!
// Primarily used for admin operations that require service-level permissions.
type ClientAuthProvider struct {
	iamURL       string
	clientID     string
	clientSecret string
	namespace    string

	httpClient   *http.Client
	currentToken *Token
	mu           sync.RWMutex // Protects currentToken
}

// NewClientAuthProvider creates a new client auth provider
// Parameters:
//   - iamURL: AGS IAM base URL (e.g., "https://demo.accelbyte.io/iam")
//   - clientID: OAuth2 client ID for service account
//   - clientSecret: OAuth2 client secret for service account
//   - namespace: AGS namespace
func NewClientAuthProvider(iamURL, clientID, clientSecret, namespace string) *ClientAuthProvider {
	return &ClientAuthProvider{
		iamURL:       iamURL,
		clientID:     clientID,
		clientSecret: clientSecret,
		namespace:    namespace,
		httpClient:   &http.Client{Timeout: 10 * time.Second},
	}
}

// Authenticate performs OAuth2 Client Credentials flow
func (c *ClientAuthProvider) Authenticate(ctx context.Context) (*Token, error) {
	// Build request
	data := url.Values{}
	data.Set("grant_type", "client_credentials")

	tokenURL := fmt.Sprintf("%s/v3/oauth/token", c.iamURL)

	req, err := http.NewRequestWithContext(ctx, "POST", tokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(c.clientID, c.clientSecret)

	// Send request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	// Check status
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("authentication failed with status %d", resp.StatusCode)
	}

	// Parse response
	var tokenResp struct {
		AccessToken  string `json:"access_token"`
		TokenType    string `json:"token_type"`
		ExpiresIn    int    `json:"expires_in"`
		RefreshToken string `json:"refresh_token,omitempty"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	// Create token
	token := &Token{
		AccessToken:  tokenResp.AccessToken,
		TokenType:    tokenResp.TokenType,
		ExpiresAt:    time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second),
		RefreshToken: tokenResp.RefreshToken,
	}

	// Store current token
	c.mu.Lock()
	c.currentToken = token
	c.mu.Unlock()

	return token, nil
}

// RefreshToken refreshes an existing token
func (c *ClientAuthProvider) RefreshToken(ctx context.Context, token *Token) (*Token, error) {
	if token.RefreshToken == "" {
		// No refresh token, perform full authentication
		return c.Authenticate(ctx)
	}

	// Build refresh request
	data := url.Values{}
	data.Set("grant_type", "refresh_token")
	data.Set("refresh_token", token.RefreshToken)

	tokenURL := fmt.Sprintf("%s/v3/oauth/token", c.iamURL)

	req, err := http.NewRequestWithContext(ctx, "POST", tokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("create refresh request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(c.clientID, c.clientSecret)

	// Send request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("refresh request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	// Check status
	if resp.StatusCode != http.StatusOK {
		// Refresh failed, try full authentication
		return c.Authenticate(ctx)
	}

	// Parse response
	var tokenResp struct {
		AccessToken  string `json:"access_token"`
		TokenType    string `json:"token_type"`
		ExpiresIn    int    `json:"expires_in"`
		RefreshToken string `json:"refresh_token,omitempty"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, fmt.Errorf("decode refresh response: %w", err)
	}

	// Create new token
	newToken := &Token{
		AccessToken:  tokenResp.AccessToken,
		TokenType:    tokenResp.TokenType,
		ExpiresAt:    time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second),
		RefreshToken: tokenResp.RefreshToken,
	}

	// Store current token
	c.mu.Lock()
	c.currentToken = newToken
	c.mu.Unlock()

	return newToken, nil
}

// GetToken returns the current valid token, refreshing if necessary
func (c *ClientAuthProvider) GetToken(ctx context.Context) (*Token, error) {
	c.mu.RLock()
	token := c.currentToken
	c.mu.RUnlock()

	// No token yet
	if token == nil {
		return c.Authenticate(ctx)
	}

	// Token expired
	if token.IsExpired() {
		return c.RefreshToken(ctx, token)
	}

	// Token expiring soon (within 5 minutes)
	if token.ExpiresIn() < 5*time.Minute {
		// Try to refresh in background, but return current token
		go func() {
			refreshCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			_, _ = c.RefreshToken(refreshCtx, token)
		}()
	}

	return token, nil
}

// IsTokenValid checks if a token is still valid
func (c *ClientAuthProvider) IsTokenValid(token *Token) bool {
	if token == nil {
		return false
	}
	return !token.IsExpired()
}

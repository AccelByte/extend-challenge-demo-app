// Copyright (c) 2025 AccelByte Inc. All Rights Reserved.
// This is licensed software from AccelByte Inc, for limitations
// and restrictions contact your company contract manager.

package auth

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/AccelByte/accelbyte-go-sdk/iam-sdk/pkg/iamclient"
	"github.com/AccelByte/accelbyte-go-sdk/iam-sdk/pkg/iamclient/o_auth2_0"
	"github.com/go-openapi/runtime/client"
)

// PasswordAuthProvider implements AuthProvider using AGS IAM Password Grant
// This is used for USER authentication (email + password → user token)
type PasswordAuthProvider struct {
	iamURL       string
	clientID     string // Still required for Password Grant
	clientSecret string // Still required for Password Grant
	namespace    string
	email        string // User email
	password     string // User password

	currentToken *Token
	mu           sync.RWMutex // Protects currentToken
}

// NewPasswordAuthProvider creates a new password auth provider
// Parameters:
//   - iamURL: AGS IAM base URL (e.g., "https://demo.accelbyte.io/iam")
//   - clientID: OAuth2 client ID (required even for password grant)
//   - clientSecret: OAuth2 client secret (required even for password grant)
//   - namespace: AGS namespace
//   - email: User email for login
//   - password: User password for login
func NewPasswordAuthProvider(iamURL, clientID, clientSecret, namespace, email, password string) *PasswordAuthProvider {
	return &PasswordAuthProvider{
		iamURL:       iamURL,
		clientID:     clientID,
		clientSecret: clientSecret,
		namespace:    namespace,
		email:        email,
		password:     password,
	}
}

// Authenticate performs OAuth2 Password Grant flow using AccelByte Go SDK
func (p *PasswordAuthProvider) Authenticate(ctx context.Context) (*Token, error) {
	// Create IAM client from base URL
	iamClient := createIAMClient(p.iamURL)

	// Prepare token grant parameters for password grant
	params := &o_auth2_0.TokenGrantV3Params{
		GrantType: "password",
		Username:  &p.email,
		Password:  &p.password,
		Context:   ctx,
	}

	// Create Basic Auth with client credentials
	basicAuth := client.BasicAuth(p.clientID, p.clientSecret)

	// Call TokenGrantV3Short with Basic Auth
	ok, err := iamClient.OAuth20.TokenGrantV3Short(params, basicAuth)
	if err != nil {
		return nil, fmt.Errorf("password grant failed: %w", err)
	}

	// Extract token response
	if ok == nil {
		return nil, fmt.Errorf("empty SDK result")
	}

	tokenResp := ok.GetPayload()
	if tokenResp == nil {
		return nil, fmt.Errorf("empty token response")
	}

	// Validate required fields
	if tokenResp.AccessToken == nil || tokenResp.TokenType == nil || tokenResp.ExpiresIn == nil {
		return nil, fmt.Errorf("invalid token response: missing required fields")
	}

	// Create token from SDK response
	token := &Token{
		AccessToken:  *tokenResp.AccessToken,
		TokenType:    *tokenResp.TokenType,
		ExpiresAt:    time.Now().Add(time.Duration(*tokenResp.ExpiresIn) * time.Second),
		RefreshToken: "", // Initialize as empty
	}

	// Set refresh token if present
	if tokenResp.RefreshToken != "" {
		token.RefreshToken = tokenResp.RefreshToken
	}

	// Store current token
	p.mu.Lock()
	p.currentToken = token
	p.mu.Unlock()

	return token, nil
}

// RefreshToken refreshes an existing token using refresh_token grant with AccelByte Go SDK
func (p *PasswordAuthProvider) RefreshToken(ctx context.Context, token *Token) (*Token, error) {
	if token.RefreshToken == "" {
		// No refresh token, perform full authentication
		return p.Authenticate(ctx)
	}

	// Create IAM client from base URL
	iamClient := createIAMClient(p.iamURL)

	// Prepare token grant parameters for refresh token grant
	refreshToken := token.RefreshToken
	params := &o_auth2_0.TokenGrantV3Params{
		GrantType:    "refresh_token",
		RefreshToken: &refreshToken,
		Context:      ctx,
	}

	// Create Basic Auth with client credentials
	basicAuth := client.BasicAuth(p.clientID, p.clientSecret)

	// Call TokenGrantV3Short with Basic Auth
	ok, err := iamClient.OAuth20.TokenGrantV3Short(params, basicAuth)
	if err != nil {
		// Refresh failed, try full authentication
		return p.Authenticate(ctx)
	}

	// Extract token response
	if ok == nil {
		return nil, fmt.Errorf("empty SDK result")
	}

	tokenResp := ok.GetPayload()
	if tokenResp == nil {
		return nil, fmt.Errorf("empty token response")
	}

	// Validate required fields
	if tokenResp.AccessToken == nil || tokenResp.TokenType == nil || tokenResp.ExpiresIn == nil {
		return nil, fmt.Errorf("invalid token response: missing required fields")
	}

	// Create new token
	newToken := &Token{
		AccessToken:  *tokenResp.AccessToken,
		TokenType:    *tokenResp.TokenType,
		ExpiresAt:    time.Now().Add(time.Duration(*tokenResp.ExpiresIn) * time.Second),
		RefreshToken: "", // Initialize as empty
	}

	// Set refresh token if present
	if tokenResp.RefreshToken != "" {
		newToken.RefreshToken = tokenResp.RefreshToken
	}

	// Store current token
	p.mu.Lock()
	p.currentToken = newToken
	p.mu.Unlock()

	return newToken, nil
}

// GetToken returns the current valid token, refreshing if necessary
func (p *PasswordAuthProvider) GetToken(ctx context.Context) (*Token, error) {
	p.mu.RLock()
	token := p.currentToken
	p.mu.RUnlock()

	// No token yet
	if token == nil {
		return p.Authenticate(ctx)
	}

	// Token expired
	if token.IsExpired() {
		return p.RefreshToken(ctx, token)
	}

	// Token expiring soon (within 5 minutes)
	if token.ExpiresIn() < 5*time.Minute {
		// Try to refresh in background, but return current token
		go func() {
			refreshCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			_, _ = p.RefreshToken(refreshCtx, token)
		}()
	}

	return token, nil
}

// IsTokenValid checks if a token is still valid
func (p *PasswordAuthProvider) IsTokenValid(token *Token) bool {
	if token == nil {
		return false
	}
	return !token.IsExpired()
}

// createIAMClient creates an AccelByte IAM client from the IAM base URL
func createIAMClient(iamURL string) *iamclient.JusticeIamService {
	// Parse the IAM URL to extract scheme and host
	// Expected format: "https://demo.accelbyte.io/iam" or "http://localhost:8080/iam"
	scheme := "https"
	host := iamURL

	// Simple URL parsing
	if len(iamURL) > 8 && iamURL[:8] == "https://" {
		scheme = "https"
		host = iamURL[8:]
	} else if len(iamURL) > 7 && iamURL[:7] == "http://" {
		scheme = "http"
		host = iamURL[7:]
	}

	// Remove /iam suffix if present
	if len(host) > 4 && host[len(host)-4:] == "/iam" {
		host = host[:len(host)-4]
	}

	// Create IAM client configuration
	cfg := &iamclient.TransportConfig{
		Host:     host,
		BasePath: "",
		Schemes:  []string{scheme},
	}

	return iamclient.NewHTTPClientWithConfig(nil, cfg)
}

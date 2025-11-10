// Copyright (c) 2025 AccelByte Inc. All Rights Reserved.
// This is licensed software from AccelByte Inc, for limitations
// and restrictions contact your company contract manager.

package auth

import "context"

// AuthProvider handles authentication and token management
type AuthProvider interface {
	// Authenticate performs initial authentication and returns a token
	Authenticate(ctx context.Context) (*Token, error)

	// RefreshToken refreshes an existing token
	RefreshToken(ctx context.Context, token *Token) (*Token, error)

	// GetToken returns the current valid token (auto-refreshes if needed)
	GetToken(ctx context.Context) (*Token, error)

	// IsTokenValid checks if the token is still valid
	IsTokenValid(token *Token) bool
}

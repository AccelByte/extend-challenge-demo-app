// Copyright (c) 2025 AccelByte Inc. All Rights Reserved.
// This is licensed software from AccelByte Inc, for limitations
// and restrictions contact your company contract manager.

package auth

import "time"

// Token represents an authentication token
type Token struct {
	AccessToken  string
	TokenType    string // Usually "Bearer"
	ExpiresAt    time.Time
	RefreshToken string // Optional
}

// IsExpired checks if the token has expired
func (t *Token) IsExpired() bool {
	if t == nil {
		return true
	}
	return time.Now().After(t.ExpiresAt)
}

// ExpiresIn returns the duration until token expiration
func (t *Token) ExpiresIn() time.Duration {
	if t == nil {
		return 0
	}
	return time.Until(t.ExpiresAt)
}

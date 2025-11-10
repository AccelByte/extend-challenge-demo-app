// Copyright (c) 2025 AccelByte Inc. All Rights Reserved.
// This is licensed software from AccelByte Inc, for limitations
// and restrictions contact your company contract manager.

package auth

import (
	"testing"
	"time"
)

func TestToken_IsExpired(t *testing.T) {
	tests := []struct {
		name   string
		token  *Token
		expect bool
	}{
		{
			name:   "nil token",
			token:  nil,
			expect: true,
		},
		{
			name: "valid token",
			token: &Token{
				ExpiresAt: time.Now().Add(1 * time.Hour),
			},
			expect: false,
		},
		{
			name: "expired token",
			token: &Token{
				ExpiresAt: time.Now().Add(-1 * time.Hour),
			},
			expect: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.token.IsExpired()
			if result != tt.expect {
				t.Errorf("Expected %v, got %v", tt.expect, result)
			}
		})
	}
}

func TestToken_ExpiresIn(t *testing.T) {
	tests := []struct {
		name   string
		token  *Token
		expect time.Duration
	}{
		{
			name:   "nil token",
			token:  nil,
			expect: 0,
		},
		{
			name: "expires in 1 hour",
			token: &Token{
				ExpiresAt: time.Now().Add(1 * time.Hour),
			},
			expect: 1 * time.Hour,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.token.ExpiresIn()
			// Allow small time drift (1 second)
			diff := result - tt.expect
			if diff < 0 {
				diff = -diff
			}
			if diff > time.Second {
				t.Errorf("Expected ~%v, got %v (diff: %v)", tt.expect, result, diff)
			}
		})
	}
}

// Copyright (c) 2025 AccelByte Inc. All Rights Reserved.
// This is licensed software from AccelByte Inc, for limitations
// and restrictions contact your company contract manager.

package app

import "testing"

func TestNewContainer(t *testing.T) {
	container := NewContainer(
		"http://localhost:8080", // backendURL
		"mock",                  // authMode
		"",                      // eventHandlerURL
		"test-user",             // userID
		"demo",                  // namespace
		"",                      // email
		"",                      // password
		"",                      // clientID
		"",                      // clientSecret
		"",                      // iamURL
		"",                      // platformURL
		"",                      // adminClientID
		"",                      // adminClientSecret
	)

	if container == nil {
		t.Fatal("Expected non-nil container")
	}

	if container.AuthProvider == nil {
		t.Error("Expected non-nil AuthProvider")
	}

	if container.APIClient == nil {
		t.Error("Expected non-nil APIClient")
	}

	if container.UserID != "test-user" {
		t.Errorf("Expected UserID 'test-user', got '%s'", container.UserID)
	}

	if container.Namespace != "demo" {
		t.Errorf("Expected Namespace 'demo', got '%s'", container.Namespace)
	}
}

func TestNewContainer_DifferentAuthMode(t *testing.T) {
	// Test with different auth modes
	modes := []string{"mock", "password", "client", "invalid"}

	for _, mode := range modes {
		container := NewContainer(
			"http://localhost:8080", // backendURL
			mode,                    // authMode
			"",                      // eventHandlerURL
			"test-user",             // userID
			"demo",                  // namespace
			"alice@example.com",     // email (for password mode)
			"password123",           // password (for password mode)
			"client-id",             // clientID
			"client-secret",         // clientSecret
			"https://demo.accelbyte.io/iam", // iamURL
			"",                      // platformURL
			"",                      // adminClientID
			"",                      // adminClientSecret
		)

		if container == nil {
			t.Fatalf("Expected non-nil container for mode %s", mode)
		}

		if container.AuthProvider == nil {
			t.Errorf("Expected non-nil AuthProvider for mode %s", mode)
		}
	}
}

func TestNewContainer_WithEventHandler(t *testing.T) {
	// Note: This will fail to connect since there's no event handler running,
	// but should still create a container with nil EventTrigger
	container := NewContainer(
		"http://localhost:8080", // backendURL
		"mock",                  // authMode
		"localhost:9999",        // eventHandlerURL
		"test-user",             // userID
		"demo",                  // namespace
		"",                      // email
		"",                      // password
		"",                      // clientID
		"",                      // clientSecret
		"",                      // iamURL
		"",                      // platformURL
		"",                      // adminClientID
		"",                      // adminClientSecret
	)

	if container == nil {
		t.Fatal("Expected non-nil container")
	}

	// EventTrigger should be nil because connection fails (no event handler running)
	if container.EventTrigger != nil {
		t.Error("Expected nil EventTrigger when event handler is not running")
	}
}

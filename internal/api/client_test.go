// Copyright (c) 2025 AccelByte Inc. All Rights Reserved.
// This is licensed software from AccelByte Inc, for limitations
// and restrictions contact your company contract manager.

package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/AccelByte/extend-challenge/extend-challenge-demo-app/internal/auth"
)

func TestNewHTTPAPIClient(t *testing.T) {
	mockAuth := auth.NewMockAuthProvider("test-user", "demo")
	client := NewHTTPAPIClient("http://localhost:8080", mockAuth)

	if client == nil {
		t.Fatal("Expected non-nil client")
	}

	if client.baseURL != "http://localhost:8080" {
		t.Errorf("Expected baseURL 'http://localhost:8080', got '%s'", client.baseURL)
	}
}

func TestHTTPAPIClient_ListChallenges(t *testing.T) {
	mockAuth := auth.NewMockAuthProvider("test-user", "demo")

	tests := []struct {
		name           string
		serverResponse string
		statusCode     int
		expectError    bool
		expectCount    int
	}{
		{
			name:           "successful response",
			serverResponse: `{"challenges":[{"challengeId":"c1","name":"Challenge 1","description":"Test","goals":[]}]}`,
			statusCode:     http.StatusOK,
			expectError:    false,
			expectCount:    1,
		},
		{
			name:           "empty list",
			serverResponse: `{"challenges":[]}`,
			statusCode:     http.StatusOK,
			expectError:    false,
			expectCount:    0,
		},
		{
			name:           "server error",
			serverResponse: `Internal Server Error`,
			statusCode:     http.StatusInternalServerError,
			expectError:    true,
		},
		{
			name:           "not found",
			serverResponse: `Not Found`,
			statusCode:     http.StatusNotFound,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Verify authorization header
				auth := r.Header.Get("Authorization")
				if auth == "" {
					t.Error("Expected Authorization header")
				}

				w.WriteHeader(tt.statusCode)
				_, _ = w.Write([]byte(tt.serverResponse))
			}))
			defer server.Close()

			client := NewHTTPAPIClient(server.URL, mockAuth)
			challenges, err := client.ListChallenges(context.Background())

			if tt.expectError {
				if err == nil {
					t.Error("Expected error, got nil")
				}
			} else {
				if err != nil {
					t.Fatalf("Unexpected error: %v", err)
				}
				if len(challenges) != tt.expectCount {
					t.Errorf("Expected %d challenges, got %d", tt.expectCount, len(challenges))
				}
			}
		})
	}
}

func TestHTTPAPIClient_GetChallenge(t *testing.T) {
	mockAuth := auth.NewMockAuthProvider("test-user", "demo")

	challenge := Challenge{
		ID:          "c1",
		Name:        "Test Challenge",
		Description: "Test Description",
		Goals:       []Goal{},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/challenges/c1" {
			t.Errorf("Expected path '/v1/challenges/c1', got '%s'", r.URL.Path)
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(challenge)
	}))
	defer server.Close()

	client := NewHTTPAPIClient(server.URL, mockAuth)
	result, err := client.GetChallenge(context.Background(), "c1")

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("Expected non-nil challenge")
	}

	if result.ID != "c1" {
		t.Errorf("Expected ID 'c1', got '%s'", result.ID)
	}
}

func TestHTTPAPIClient_ClaimReward(t *testing.T) {
	mockAuth := auth.NewMockAuthProvider("test-user", "demo")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST request, got %s", r.Method)
		}

		if r.URL.Path != "/v1/challenges/c1/goals/g1/claim" {
			t.Errorf("Expected path '/v1/challenges/c1/goals/g1/claim', got '%s'", r.URL.Path)
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(ClaimResult{
			GoalID: "g1",
			Status: "claimed",
			Reward: Reward{
				Type:     "ITEM",
				RewardID: "item123",
				Quantity: 100,
			},
			ClaimedAt: "2025-01-01T00:00:00Z",
		})
	}))
	defer server.Close()

	client := NewHTTPAPIClient(server.URL, mockAuth)
	result, err := client.ClaimReward(context.Background(), "c1", "g1")

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("Expected non-nil result")
	}

	if result.GoalID != "g1" {
		t.Errorf("Expected goal_id 'g1', got '%s'", result.GoalID)
	}

	if result.Status != "claimed" {
		t.Errorf("Expected status 'claimed', got '%s'", result.Status)
	}
}

func TestHTTPAPIClient_GetLastRequest(t *testing.T) {
	mockAuth := auth.NewMockAuthProvider("test-user", "demo")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("[]"))
	}))
	defer server.Close()

	client := NewHTTPAPIClient(server.URL, mockAuth)

	// Make a request to populate lastRequest
	_, _ = client.ListChallenges(context.Background())

	lastRequest := client.GetLastRequest()
	if lastRequest == nil {
		t.Fatal("Expected non-nil lastRequest")
	}

	if lastRequest.Method != "GET" {
		t.Errorf("Expected method GET, got %s", lastRequest.Method)
	}
}

func TestHTTPAPIClient_GetLastResponse(t *testing.T) {
	mockAuth := auth.NewMockAuthProvider("test-user", "demo")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("[]"))
	}))
	defer server.Close()

	client := NewHTTPAPIClient(server.URL, mockAuth)

	// Make a request to populate lastResponse
	_, _ = client.ListChallenges(context.Background())

	lastResponse := client.GetLastResponse()
	if lastResponse == nil {
		t.Fatal("Expected non-nil lastResponse")
	}

	if lastResponse.StatusCode != http.StatusOK {
		t.Errorf("Expected status code 200, got %d", lastResponse.StatusCode)
	}
}

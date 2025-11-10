// Copyright (c) 2025 AccelByte Inc. All Rights Reserved.
// This is licensed software from AccelByte Inc, for limitations
// and restrictions contact your company contract manager.

package tui

import (
	"fmt"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/AccelByte/extend-challenge/extend-challenge-demo-app/internal/api"
	"github.com/AccelByte/extend-challenge/extend-challenge-demo-app/internal/auth"
)

func TestNewDashboardModel(t *testing.T) {
	mockAuth := auth.NewMockAuthProvider("test-user", "demo")
	apiClient := api.NewHTTPAPIClient("http://localhost:8080", mockAuth)

	model := NewDashboardModel(apiClient)

	if model == nil {
		t.Fatal("Expected non-nil model")
	}

	if model.challengeCursor != 0 {
		t.Errorf("Expected cursor 0, got %d", model.challengeCursor)
	}

	if model.loading {
		t.Error("Expected loading to be false")
	}
}

func TestDashboardModel_Update_KeyUp(t *testing.T) {
	mockAuth := auth.NewMockAuthProvider("test-user", "demo")
	apiClient := api.NewHTTPAPIClient("http://localhost:8080", mockAuth)
	model := NewDashboardModel(apiClient)

	// Set up some challenges
	model.challenges = []api.Challenge{
		{ID: "c1", Name: "Challenge 1"},
		{ID: "c2", Name: "Challenge 2"},
		{ID: "c3", Name: "Challenge 3"},
	}
	model.challengeCursor = 1

	// Send up key
	newModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyUp})
	updatedModel := newModel.(*DashboardModel)

	if updatedModel.challengeCursor != 0 {
		t.Errorf("Expected cursor 0, got %d", updatedModel.challengeCursor)
	}
}

func TestDashboardModel_Update_KeyDown(t *testing.T) {
	mockAuth := auth.NewMockAuthProvider("test-user", "demo")
	apiClient := api.NewHTTPAPIClient("http://localhost:8080", mockAuth)
	model := NewDashboardModel(apiClient)

	// Set up some challenges
	model.challenges = []api.Challenge{
		{ID: "c1", Name: "Challenge 1"},
		{ID: "c2", Name: "Challenge 2"},
		{ID: "c3", Name: "Challenge 3"},
	}
	model.challengeCursor = 0

	// Send down key
	newModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyDown})
	updatedModel := newModel.(*DashboardModel)

	if updatedModel.challengeCursor != 1 {
		t.Errorf("Expected cursor 1, got %d", updatedModel.challengeCursor)
	}
}

func TestDashboardModel_Update_ChallengesLoaded(t *testing.T) {
	mockAuth := auth.NewMockAuthProvider("test-user", "demo")
	apiClient := api.NewHTTPAPIClient("http://localhost:8080", mockAuth)
	model := NewDashboardModel(apiClient)
	model.loading = true

	challenges := []api.Challenge{
		{ID: "c1", Name: "Challenge 1"},
		{ID: "c2", Name: "Challenge 2"},
	}

	msg := ChallengesLoadedMsg{
		challenges: challenges,
		err:        nil,
	}

	newModel, _ := model.Update(msg)
	updatedModel := newModel.(*DashboardModel)

	if updatedModel.loading {
		t.Error("Expected loading to be false")
	}

	if len(updatedModel.challenges) != 2 {
		t.Errorf("Expected 2 challenges, got %d", len(updatedModel.challenges))
	}
}

func TestDashboardModel_View(t *testing.T) {
	mockAuth := auth.NewMockAuthProvider("test-user", "demo")
	apiClient := api.NewHTTPAPIClient("http://localhost:8080", mockAuth)
	model := NewDashboardModel(apiClient)

	// Test empty state
	view := model.View()
	if view == "" {
		t.Error("Expected non-empty view")
	}

	// Test loading state
	model.loading = true
	view = model.View()
	if view == "" {
		t.Error("Expected non-empty view for loading state")
	}

	// Test with challenges
	model.loading = false
	model.challenges = []api.Challenge{
		{ID: "c1", Name: "Challenge 1", Description: "Test challenge", Goals: []api.Goal{
			{ID: "g1", Name: "Goal 1", Status: "completed"},
			{ID: "g2", Name: "Goal 2", Status: "in_progress"},
			{ID: "g3", Name: "Goal 3", Status: "claimed"},
		}},
	}
	view = model.View()
	if view == "" {
		t.Error("Expected non-empty view with challenges")
	}

	// Test error state
	model.errorMsg = "Test error"
	view = model.View()
	if view == "" {
		t.Error("Expected non-empty view for error state")
	}
}

func TestDashboardModel_Update_KeyR(t *testing.T) {
	mockAuth := auth.NewMockAuthProvider("test-user", "demo")
	apiClient := api.NewHTTPAPIClient("http://localhost:8080", mockAuth)
	model := NewDashboardModel(apiClient)

	// Send 'r' key to refresh
	newModel, cmd := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
	updatedModel := newModel.(*DashboardModel)

	if !updatedModel.loading {
		t.Error("Expected loading to be true after refresh")
	}

	if cmd == nil {
		t.Error("Expected refresh command")
	}
}

func TestDashboardModel_Update_KeyUpAtTop(t *testing.T) {
	mockAuth := auth.NewMockAuthProvider("test-user", "demo")
	apiClient := api.NewHTTPAPIClient("http://localhost:8080", mockAuth)
	model := NewDashboardModel(apiClient)

	model.challenges = []api.Challenge{
		{ID: "c1", Name: "Challenge 1"},
		{ID: "c2", Name: "Challenge 2"},
	}
	model.challengeCursor = 0

	// Try to go up from top
	newModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyUp})
	updatedModel := newModel.(*DashboardModel)

	if updatedModel.challengeCursor != 0 {
		t.Errorf("Expected cursor to stay at 0, got %d", updatedModel.challengeCursor)
	}
}

func TestDashboardModel_Update_KeyDownAtBottom(t *testing.T) {
	mockAuth := auth.NewMockAuthProvider("test-user", "demo")
	apiClient := api.NewHTTPAPIClient("http://localhost:8080", mockAuth)
	model := NewDashboardModel(apiClient)

	model.challenges = []api.Challenge{
		{ID: "c1", Name: "Challenge 1"},
		{ID: "c2", Name: "Challenge 2"},
	}
	model.challengeCursor = 1

	// Try to go down from bottom
	newModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyDown})
	updatedModel := newModel.(*DashboardModel)

	if updatedModel.challengeCursor != 1 {
		t.Errorf("Expected cursor to stay at 1, got %d", updatedModel.challengeCursor)
	}
}

func TestDashboardModel_Update_ChallengesLoadedWithError(t *testing.T) {
	mockAuth := auth.NewMockAuthProvider("test-user", "demo")
	apiClient := api.NewHTTPAPIClient("http://localhost:8080", mockAuth)
	model := NewDashboardModel(apiClient)
	model.loading = true

	msg := ChallengesLoadedMsg{
		challenges: nil,
		err:        fmt.Errorf("test error"),
	}

	newModel, _ := model.Update(msg)
	updatedModel := newModel.(*DashboardModel)

	if updatedModel.loading {
		t.Error("Expected loading to be false")
	}

	if updatedModel.errorMsg == "" {
		t.Error("Expected error message to be set")
	}
}

func TestDashboardModel_Init(t *testing.T) {
	mockAuth := auth.NewMockAuthProvider("test-user", "demo")
	apiClient := api.NewHTTPAPIClient("http://localhost:8080", mockAuth)
	model := NewDashboardModel(apiClient)

	cmd := model.Init()
	if cmd == nil {
		t.Error("Expected init command")
	}
}

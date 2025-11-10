// Copyright (c) 2025 AccelByte Inc. All Rights Reserved.
// This is licensed software from AccelByte Inc, for limitations
// and restrictions contact your company contract manager.

package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/AccelByte/extend-challenge/extend-challenge-demo-app/internal/app"
)

func TestNewAppModel(t *testing.T) {
	container := app.NewContainer("http://localhost:8080", "mock", "", "test-user", "demo", "", "", "", "", "", "", "", "")
	model := NewAppModel(container)

	if model.container == nil {
		t.Fatal("Expected non-nil container")
	}

	if model.dashboard == nil {
		t.Fatal("Expected non-nil dashboard")
	}

	if model.width != 80 {
		t.Errorf("Expected width 80, got %d", model.width)
	}

	if model.height != 24 {
		t.Errorf("Expected height 24, got %d", model.height)
	}
}

func TestAppModel_Update_Quit(t *testing.T) {
	container := app.NewContainer("http://localhost:8080", "mock", "", "test-user", "demo", "", "", "", "", "", "", "", "")
	model := NewAppModel(container)

	// Send quit key
	newModel, cmd := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	updatedModel := newModel.(AppModel)

	if !updatedModel.quitting {
		t.Error("Expected quitting to be true")
	}

	if cmd == nil {
		t.Error("Expected quit command")
	}
}

func TestAppModel_Update_WindowSize(t *testing.T) {
	container := app.NewContainer("http://localhost:8080", "mock", "", "test-user", "demo", "", "", "", "", "", "", "", "")
	model := NewAppModel(container)

	// Send window size message
	newModel, _ := model.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	updatedModel := newModel.(AppModel)

	if updatedModel.width != 120 {
		t.Errorf("Expected width 120, got %d", updatedModel.width)
	}

	if updatedModel.height != 40 {
		t.Errorf("Expected height 40, got %d", updatedModel.height)
	}
}

func TestAppModel_View(t *testing.T) {
	container := app.NewContainer("http://localhost:8080", "mock", "", "test-user", "demo", "", "", "", "", "", "", "", "")
	model := NewAppModel(container)

	view := model.View()
	if view == "" {
		t.Error("Expected non-empty view")
	}
}

func TestAppModel_View_Quitting(t *testing.T) {
	container := app.NewContainer("http://localhost:8080", "mock", "", "test-user", "demo", "", "", "", "", "", "", "", "")
	model := NewAppModel(container)
	model.quitting = true

	view := model.View()
	if view == "" {
		t.Error("Expected non-empty view for quitting state")
	}
}

func TestNewApp(t *testing.T) {
	container := app.NewContainer("http://localhost:8080", "mock", "", "test-user", "demo", "", "", "", "", "", "", "", "")
	application := NewApp(container)

	if application == nil {
		t.Fatal("Expected non-nil app")
	}

	if application.container == nil {
		t.Fatal("Expected non-nil container")
	}
}

func TestAppModel_RenderHeader(t *testing.T) {
	container := app.NewContainer("http://localhost:8080", "mock", "", "test-user", "demo", "", "", "", "", "", "", "", "")
	model := NewAppModel(container)

	header := model.renderHeader()
	if header == "" {
		t.Error("Expected non-empty header")
	}
}

func TestAppModel_RenderFooter(t *testing.T) {
	container := app.NewContainer("http://localhost:8080", "mock", "", "test-user", "demo", "", "", "", "", "", "", "", "")
	model := NewAppModel(container)

	footer := model.renderFooter()
	if footer == "" {
		t.Error("Expected non-empty footer")
	}
}

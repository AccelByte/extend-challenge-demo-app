// Copyright (c) 2025 AccelByte Inc. All Rights Reserved.
// This is licensed software from AccelByte Inc, for limitations
// and restrictions contact your company contract manager.

package tui

import (
	"context"
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/AccelByte/extend-challenge/extend-challenge-demo-app/internal/app"
)

// TickMsg is sent periodically for token refresh checks
type TickMsg struct {
	time time.Time
}

// Screen represents the current active screen
type Screen int

const (
	ScreenDashboard Screen = iota
	ScreenEventSimulator
	ScreenInventory
)

// AppModel is the root model containing all screen models
type AppModel struct {
	container      *app.Container
	dashboard      *DashboardModel
	eventSimulator *EventSimulatorModel
	inventory      *InventoryModel
	currentScreen  Screen
	width          int
	height         int
	quitting       bool
}

// NewAppModel creates the initial app model
func NewAppModel(container *app.Container) AppModel {
	var eventSimulator *EventSimulatorModel
	if container.EventTrigger != nil {
		eventSimulator = NewEventSimulatorModel(container.EventTrigger, container.UserID, container.Namespace)
	}

	return AppModel{
		container:      container,
		dashboard:      NewDashboardModel(container.APIClient),
		eventSimulator: eventSimulator,
		inventory:      NewInventoryModel(container.RewardVerifier),
		currentScreen:  ScreenDashboard,
		width:          80,
		height:         24,
		quitting:       false,
	}
}

// Init initializes the model and returns initial commands
func (m AppModel) Init() tea.Cmd {
	return tea.Batch(
		m.dashboard.Init(),
		tokenRefreshTickCmd(), // Start token refresh ticker
	)
}

// Update handles messages and returns updated model
func (m AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	// Handle global messages first
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Skip global shortcuts if an input field is focused (to allow typing)
		skipGlobalShortcuts := false
		if m.currentScreen == ScreenEventSimulator && m.eventSimulator != nil {
			skipGlobalShortcuts = m.eventSimulator.IsInputFocused()
		}

		// Always allow Ctrl+C to quit (unconditional escape hatch)
		if msg.String() == "ctrl+c" {
			m.quitting = true
			return m, tea.Quit
		}

		// Skip navigation shortcuts (including 'q') if input is focused
		if !skipGlobalShortcuts {
			switch msg.String() {
			case "q":
				// Quit application
				m.quitting = true
				return m, tea.Quit

			case "1":
				// Switch to dashboard
				m.currentScreen = ScreenDashboard
				return m, nil

			case "2", "e":
				// Switch to event simulator (if available)
				if m.eventSimulator != nil {
					m.currentScreen = ScreenEventSimulator
					return m, nil
				}

			case "3", "i":
				// Switch to inventory screen
				m.currentScreen = ScreenInventory
				// Load inventory data when entering screen
				return m, func() tea.Msg { return LoadInventoryMsg{} }

			case "esc":
				// Return to dashboard (only from other screens, not from dashboard itself)
				if m.currentScreen != ScreenDashboard {
					m.currentScreen = ScreenDashboard
					return m, nil
				}
				// If already on dashboard, let the dashboard handle Esc (for detail view → list view)
			}
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case TickMsg:
		// Handle token refresh check (every 1 minute)
		return m, tokenRefreshTickCmd()
	}

	// Route message to current screen
	switch m.currentScreen {
	case ScreenDashboard:
		newDashboard, cmd := m.dashboard.Update(msg)
		m.dashboard = newDashboard.(*DashboardModel)
		return m, cmd

	case ScreenEventSimulator:
		if m.eventSimulator != nil {
			newSimulator, cmd := m.eventSimulator.Update(msg)
			m.eventSimulator = newSimulator.(*EventSimulatorModel)
			return m, cmd
		}

	case ScreenInventory:
		newInventory, cmd := m.inventory.Update(msg)
		m.inventory = newInventory.(*InventoryModel)
		return m, cmd
	}

	return m, cmd
}

// View renders the current screen
func (m AppModel) View() string {
	if m.quitting {
		return "Goodbye!\n"
	}

	// Render header
	header := m.renderHeader()

	// Render current screen content
	var content string
	switch m.currentScreen {
	case ScreenDashboard:
		content = m.dashboard.View()
	case ScreenEventSimulator:
		if m.eventSimulator != nil {
			content = m.eventSimulator.View()
		} else {
			content = "Event Simulator not available (event handler not connected)"
		}
	case ScreenInventory:
		content = m.inventory.View()
	}

	// Render footer
	footer := m.renderFooter()

	// Combine with spacing
	return lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		"\n",
		content,
		"\n",
		footer,
	)
}

// renderHeader renders the status bar
func (m AppModel) renderHeader() string {
	var screen string
	switch m.currentScreen {
	case ScreenDashboard:
		screen = "Dashboard"
	case ScreenEventSimulator:
		screen = "Event Simulator"
	case ScreenInventory:
		screen = "Inventory & Wallets"
	}

	// Get token status (user + optional admin)
	authStatus := "Auth: ✗ No token"
	ctx := context.Background()

	// User token status
	userTokenStatus := ""
	token, err := m.container.AuthProvider.GetToken(ctx)
	if err == nil && token != nil {
		if m.container.AuthProvider.IsTokenValid(token) {
			expiresIn := time.Until(token.ExpiresAt)
			if expiresIn > 0 {
				minutes := int(expiresIn.Minutes())
				if minutes > 60 {
					hours := minutes / 60
					userTokenStatus = fmt.Sprintf("User (%dh)", hours)
				} else {
					userTokenStatus = fmt.Sprintf("User (%dm)", minutes)
				}
			} else {
				userTokenStatus = "User (Expired)"
			}
		} else {
			userTokenStatus = "User (Invalid)"
		}
	}

	// Admin token status (if available)
	adminTokenStatus := ""
	if m.container.AdminAuthProvider != nil {
		adminToken, adminErr := m.container.AdminAuthProvider.GetToken(ctx)
		if adminErr == nil && adminToken != nil {
			if m.container.AdminAuthProvider.IsTokenValid(adminToken) {
				expiresIn := time.Until(adminToken.ExpiresAt)
				if expiresIn > 0 {
					minutes := int(expiresIn.Minutes())
					if minutes > 60 {
						hours := minutes / 60
						adminTokenStatus = fmt.Sprintf(" | Admin (%dh)", hours)
					} else {
						adminTokenStatus = fmt.Sprintf(" | Admin (%dm)", minutes)
					}
				} else {
					adminTokenStatus = " | Admin (Expired)"
				}
			} else {
				adminTokenStatus = " | Admin (Invalid)"
			}
		}
	}

	// Combine user and admin token status
	if userTokenStatus != "" {
		authStatus = "Auth: ✓ " + userTokenStatus + adminTokenStatus
	}

	// Check if input is focused (affects quit shortcut display)
	inputFocused := false
	if m.currentScreen == ScreenEventSimulator && m.eventSimulator != nil {
		inputFocused = m.eventSimulator.IsInputFocused()
	}

	quitHint := "[q] Quit"
	if inputFocused {
		quitHint = "[Ctrl+C] Quit"
	}

	return headerStyle.Render(fmt.Sprintf("Challenge Demo App - %s | %s | User: %s | %s", screen, authStatus, m.container.UserID, quitHint))
}

// renderFooter renders keyboard shortcuts (context-aware based on screen and focus state)
func (m AppModel) renderFooter() string {
	var shortcuts string

	// Check if input is focused (affects available shortcuts)
	inputFocused := false
	if m.currentScreen == ScreenEventSimulator && m.eventSimulator != nil {
		inputFocused = m.eventSimulator.IsInputFocused()
	}

	if inputFocused {
		// When input is focused, only Ctrl+C works for quit, other navigation disabled
		shortcuts = "⚠ Input Mode: Navigation disabled | [Esc] Unfocus | [Ctrl+C] Quit"
	} else {
		// Normal navigation mode - add screen-specific shortcuts
		baseShortcuts := "[1] Dashboard"
		if m.eventSimulator != nil {
			baseShortcuts += "  [2/e] Event Simulator"
		}
		baseShortcuts += "  [3/i] Inventory"

		// Add screen-specific shortcuts
		switch m.currentScreen {
		case ScreenInventory:
			shortcuts = baseShortcuts + "  [Tab] Switch Panel  [↑↓] Scroll  [r] Refresh  [Esc] Back  [q] Quit"
		default:
			shortcuts = baseShortcuts + "  [r] Refresh  [q] Quit"
		}
	}

	return footerStyle.Render(shortcuts)
}

// tokenRefreshTickCmd returns a command that ticks every minute for token checks
func tokenRefreshTickCmd() tea.Cmd {
	return tea.Tick(time.Minute, func(t time.Time) tea.Msg {
		return TickMsg{time: t}
	})
}

// App is the root Bubble Tea application
type App struct {
	container *app.Container
}

// NewApp creates a new TUI app
func NewApp(container *app.Container) *App {
	return &App{container: container}
}

// Run starts the TUI application
func (a *App) Run() error {
	// Create initial model
	model := NewAppModel(a.container)

	// Configure Bubble Tea program
	p := tea.NewProgram(
		model,
		tea.WithAltScreen(), // Use alternate screen buffer
	)

	// Start program
	finalModel, err := p.Run()
	if err != nil {
		return fmt.Errorf("error running TUI: %w", err)
	}

	// Type assert to access final state if needed
	_ = finalModel.(AppModel)

	return nil
}

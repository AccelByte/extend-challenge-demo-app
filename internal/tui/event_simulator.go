// Copyright (c) 2025 AccelByte Inc. All Rights Reserved.
// This is licensed software from AccelByte Inc, for limitations
// and restrictions contact your company contract manager.

package tui

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/AccelByte/extend-challenge/extend-challenge-demo-app/internal/events"
)

// EventType represents the type of event to trigger
type EventType int

const (
	EventTypeLogin EventType = iota
	EventTypeStatUpdate
)

// EventHistoryEntry represents a single event trigger in history
type EventHistoryEntry struct {
	EventType EventType
	StatCode  string
	Value     int
	Success   bool
	Duration  time.Duration
	Error     string
	Timestamp time.Time
}

// EventSimulatorModel manages the event simulator screen
type EventSimulatorModel struct {
	eventTrigger events.EventTrigger
	userID       string
	namespace    string

	// UI state
	selectedType EventType
	statCodeInput textinput.Model
	statValueInput textinput.Model
	focusedInput  int // 0 = event type, 1 = stat code, 2 = stat value

	// Event history (last 10 events)
	history []EventHistoryEntry

	// Status
	loading bool
	err     error
}

// NewEventSimulatorModel creates a new event simulator model
func NewEventSimulatorModel(eventTrigger events.EventTrigger, userID, namespace string) *EventSimulatorModel {
	statCodeInput := textinput.New()
	statCodeInput.Placeholder = "kills"
	statCodeInput.CharLimit = 50
	statCodeInput.Width = 30

	statValueInput := textinput.New()
	statValueInput.Placeholder = "10"
	statValueInput.CharLimit = 10
	statValueInput.Width = 30

	return &EventSimulatorModel{
		eventTrigger:   eventTrigger,
		userID:         userID,
		namespace:      namespace,
		selectedType:   EventTypeLogin,
		statCodeInput:  statCodeInput,
		statValueInput: statValueInput,
		focusedInput:   0,
		history:        make([]EventHistoryEntry, 0, 10),
	}
}

// Init initializes the model
func (m *EventSimulatorModel) Init() tea.Cmd {
	return nil
}

// Update handles messages and updates the model
func (m *EventSimulatorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Handle escape key to unfocus inputs
		if msg.String() == "esc" && m.IsInputFocused() {
			m.focusedInput = 0
			m.updateInputFocus()
			return m, nil
		}

		// Only handle navigation keys when NOT focused on text input
		// When focused, let text input handle all keys (including arrows for cursor movement)
		if !m.IsInputFocused() {
			switch msg.String() {
			case "tab":
				// Cycle through inputs
				m.focusedInput = (m.focusedInput + 1) % 3
				m.updateInputFocus()
				return m, nil

			case "up":
				// Toggle event type
				if m.selectedType == EventTypeStatUpdate {
					m.selectedType = EventTypeLogin
				}
				return m, nil

			case "down":
				// Toggle event type
				if m.selectedType == EventTypeLogin {
					m.selectedType = EventTypeStatUpdate
				}
				return m, nil

			case "enter":
				// Trigger event
				if m.eventTrigger == nil {
					m.err = fmt.Errorf("event trigger not available (event handler not connected)")
					return m, nil
				}

				m.loading = true
				m.err = nil
				return m, m.triggerEventCmd()
			}
		} else {
			// When input is focused, handle special keys
			switch msg.String() {
			case "tab":
				// Allow tab to cycle through inputs even when focused
				m.focusedInput = (m.focusedInput + 1) % 3
				m.updateInputFocus()
				return m, nil

			case "enter":
				// Allow enter to trigger event even when focused
				if m.eventTrigger == nil {
					m.err = fmt.Errorf("event trigger not available (event handler not connected)")
					return m, nil
				}

				m.loading = true
				m.err = nil
				return m, m.triggerEventCmd()
			}
		}

	case eventTriggeredMsg:
		// Event trigger completed
		m.loading = false

		// Add to history
		entry := EventHistoryEntry{
			EventType: msg.eventType,
			StatCode:  msg.statCode,
			Value:     msg.value,
			Success:   msg.err == nil,
			Duration:  msg.duration,
			Timestamp: time.Now(),
		}
		if msg.err != nil {
			entry.Error = msg.err.Error()
		}

		// Prepend to history (newest first)
		m.history = append([]EventHistoryEntry{entry}, m.history...)
		if len(m.history) > 10 {
			m.history = m.history[:10]
		}

		if msg.err != nil {
			m.err = msg.err
		}

		return m, nil
	}

	// Update text inputs
	switch m.focusedInput {
	case 1:
		m.statCodeInput, cmd = m.statCodeInput.Update(msg)
		return m, cmd
	case 2:
		m.statValueInput, cmd = m.statValueInput.Update(msg)
		return m, cmd
	}

	return m, nil
}

// View renders the event simulator screen
func (m *EventSimulatorModel) View() string {
	var s string

	// Title
	s += titleStyle.Render("Event Simulator") + "\n\n"

	// Event trigger availability check
	if m.eventTrigger == nil {
		s += errorStyle.Render("⚠ Event Handler Not Connected") + "\n"
		s += dimStyle.Render("Start the event handler service to enable event simulation.") + "\n\n"
		return s
	}

	// User context
	s += dimStyle.Render(fmt.Sprintf("User: %s | Namespace: %s", m.userID, m.namespace)) + "\n\n"

	// Event type selector
	s += boldStyle.Render("Event Type:") + "\n"
	if m.selectedType == EventTypeLogin {
		s += selectedStyle.Render("▶ Login Event") + "\n"
		s += "  Stat Update Event\n"
	} else {
		s += "  Login Event\n"
		s += selectedStyle.Render("▶ Stat Update Event") + "\n"
	}
	s += "\n"

	// Stat update inputs (only show for stat update events)
	if m.selectedType == EventTypeStatUpdate {
		s += boldStyle.Render("Stat Code:") + "\n"
		if m.focusedInput == 1 {
			s += focusedInputStyle.Render(m.statCodeInput.View()) + "\n\n"
		} else {
			s += m.statCodeInput.View() + "\n\n"
		}

		s += boldStyle.Render("Value:") + "\n"
		if m.focusedInput == 2 {
			s += focusedInputStyle.Render(m.statValueInput.View()) + "\n\n"
		} else {
			s += m.statValueInput.View() + "\n\n"
		}
	}

	// Trigger button
	if m.loading {
		s += loadingStyle.Render("⏳ Triggering event...") + "\n\n"
	} else {
		s += successStyle.Render("[Enter] Trigger Event") + "\n\n"
	}

	// Error message
	if m.err != nil {
		s += errorStyle.Render(fmt.Sprintf("Error: %v", m.err)) + "\n\n"
	}

	// Event history
	s += boldStyle.Render("Recent Events (Last 10):") + "\n"
	if len(m.history) == 0 {
		s += dimStyle.Render("No events triggered yet") + "\n"
	} else {
		for _, entry := range m.history {
			s += m.renderHistoryEntry(entry) + "\n"
		}
	}

	s += "\n"
	// Show context-aware shortcuts based on focus state
	if m.IsInputFocused() {
		s += dimStyle.Render("[←→] Move Cursor  [Tab] Next Field  [Enter] Trigger  [Esc] Unfocus  [Ctrl+C] Quit") + "\n"
	} else {
		s += dimStyle.Render("[↑↓] Select  [Tab] Next Field  [Enter] Trigger  [Esc] Back  [q] Quit") + "\n"
	}

	return s
}

// renderHistoryEntry renders a single history entry
func (m *EventSimulatorModel) renderHistoryEntry(entry EventHistoryEntry) string {
	var s string

	// Success/failure indicator
	if entry.Success {
		s += successStyle.Render("✓")
	} else {
		s += errorStyle.Render("✗")
	}

	// Event type and details
	if entry.EventType == EventTypeLogin {
		s += " Login Event"
	} else {
		s += fmt.Sprintf(" Stat Update: %s = %d", entry.StatCode, entry.Value)
	}

	// Duration
	s += dimStyle.Render(fmt.Sprintf(" (%dms)", entry.Duration.Milliseconds()))

	// Error (if any)
	if !entry.Success && entry.Error != "" {
		s += "\n  " + errorStyle.Render(entry.Error)
	}

	return s
}

// updateInputFocus updates which input is focused
func (m *EventSimulatorModel) updateInputFocus() {
	switch m.focusedInput {
	case 1:
		m.statCodeInput.Focus()
		m.statValueInput.Blur()
	case 2:
		m.statCodeInput.Blur()
		m.statValueInput.Focus()
	default:
		m.statCodeInput.Blur()
		m.statValueInput.Blur()
	}
}

// IsInputFocused returns true if any text input is currently focused
func (m *EventSimulatorModel) IsInputFocused() bool {
	return m.focusedInput == 1 || m.focusedInput == 2
}

// triggerEventCmd triggers an event and returns the result
func (m *EventSimulatorModel) triggerEventCmd() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		startTime := time.Now()
		var err error
		var eventType EventType
		var statCode string
		var value int

		switch m.selectedType {
		case EventTypeLogin:
			eventType = EventTypeLogin
			err = m.eventTrigger.TriggerLogin(ctx, m.userID, m.namespace)

		case EventTypeStatUpdate:
			eventType = EventTypeStatUpdate
			statCode = m.statCodeInput.Value()
			if statCode == "" {
				statCode = "kills" // Default
			}

			valueStr := m.statValueInput.Value()
			if valueStr == "" {
				value = 10 // Default
			} else {
				value, err = strconv.Atoi(valueStr)
				if err != nil {
					return eventTriggeredMsg{
						eventType: eventType,
						duration:  time.Since(startTime),
						err:       fmt.Errorf("invalid value: %w", err),
					}
				}
			}

			err = m.eventTrigger.TriggerStatUpdate(ctx, m.userID, m.namespace, statCode, value)
		}

		duration := time.Since(startTime)

		return eventTriggeredMsg{
			eventType: eventType,
			statCode:  statCode,
			value:     value,
			duration:  duration,
			err:       err,
		}
	}
}

// eventTriggeredMsg is sent when an event trigger completes
type eventTriggeredMsg struct {
	eventType EventType
	statCode  string
	value     int
	duration  time.Duration
	err       error
}

// Additional styles for event simulator
var (
	focusedInputStyle = lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")). // green
		Padding(0, 1)
)

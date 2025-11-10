// Copyright (c) 2025 AccelByte Inc. All Rights Reserved.
// This is licensed software from AccelByte Inc, for limitations
// and restrictions contact your company contract manager.

package tui

import (
	"context"
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/AccelByte/extend-challenge/extend-challenge-demo-app/internal/api"
)

// ViewMode represents the dashboard view mode
type ViewMode int

const (
	ViewModeList   ViewMode = iota // Challenge list view
	ViewModeDetail                  // Single challenge detail view
)

// ChallengesLoadedMsg is sent when challenges are loaded
type ChallengesLoadedMsg struct {
	challenges []api.Challenge
	err        error
}

// ClaimGoalMsg is sent when a goal claim is attempted
type ClaimGoalMsg struct {
	result *api.ClaimResult
	err    error
}

// DashboardModel represents the challenge dashboard screen
type DashboardModel struct {
	apiClient       api.APIClient
	challenges      []api.Challenge
	viewMode        ViewMode
	challengeCursor int
	goalCursor      int // Selected goal index in detail view
	loading         bool
	claiming        bool   // True when claiming a reward
	successMsg      string // Success message to display
	errorMsg        string
}

// NewDashboardModel creates a new dashboard model
func NewDashboardModel(apiClient api.APIClient) *DashboardModel {
	return &DashboardModel{
		apiClient:       apiClient,
		viewMode:        ViewModeList,
		challengeCursor: 0,
		goalCursor:      0,
		loading:         false,
	}
}

// Init loads challenges
func (m *DashboardModel) Init() tea.Cmd {
	m.loading = true
	return m.loadChallengesCmd()
}

// Update handles messages for the dashboard
func (m *DashboardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.viewMode == ViewModeList {
				// Navigate challenge list
				if m.challengeCursor > 0 {
					m.challengeCursor--
				}
			} else {
				// Navigate goal list in detail view
				if m.goalCursor > 0 {
					m.goalCursor--
				}
			}
			return m, nil

		case "down", "j":
			if m.viewMode == ViewModeList {
				// Navigate challenge list
				if m.challengeCursor < len(m.challenges)-1 {
					m.challengeCursor++
				}
			} else {
				// Navigate goal list in detail view
				if m.challengeCursor < len(m.challenges) {
					challenge := m.challenges[m.challengeCursor]
					if m.goalCursor < len(challenge.Goals)-1 {
						m.goalCursor++
					}
				}
			}
			return m, nil

		case "enter":
			// Drill down into selected challenge
			if m.viewMode == ViewModeList && len(m.challenges) > 0 {
				m.viewMode = ViewModeDetail
				m.goalCursor = 0 // Reset goal cursor
			}
			return m, nil

		case "esc":
			// Go back to challenge list
			if m.viewMode == ViewModeDetail {
				m.viewMode = ViewModeList
			}
			return m, nil

		case "r":
			// Refresh challenges
			m.loading = true
			m.successMsg = "" // Clear success message on refresh
			return m, m.loadChallengesCmd()

		case "c":
			// Claim reward for selected goal
			if m.viewMode == ViewModeDetail && m.challengeCursor < len(m.challenges) {
				challenge := m.challenges[m.challengeCursor]
				if m.goalCursor < len(challenge.Goals) {
					goal := challenge.Goals[m.goalCursor]
					if goal.Status == "completed" {
						m.claiming = true
						m.errorMsg = ""
						m.successMsg = ""
						return m, m.claimGoalCmd(challenge.ID, goal.ID)
					}
				}
			}
			return m, nil
		}

	case ChallengesLoadedMsg:
		m.loading = false
		if msg.err != nil {
			m.errorMsg = fmt.Sprintf("Failed to load challenges: %v", msg.err)
			return m, nil
		}

		m.challenges = msg.challenges
		m.errorMsg = ""
		// Reset cursor if out of bounds
		if m.challengeCursor >= len(m.challenges) {
			m.challengeCursor = 0
		}
		return m, nil

	case ClaimGoalMsg:
		m.claiming = false
		if msg.err != nil {
			m.errorMsg = fmt.Sprintf("Failed to claim reward: %v", msg.err)
			m.successMsg = ""
			return m, nil
		}

		// Show success message
		m.successMsg = "✓ Reward claimed successfully!"
		m.errorMsg = ""

		// Refresh challenges to show updated status
		m.loading = true
		return m, m.loadChallengesCmd()
	}

	return m, nil
}

// View renders the dashboard
func (m *DashboardModel) View() string {
	var b strings.Builder

	// Title
	b.WriteString(titleStyle.Render("Challenge Dashboard"))
	b.WriteString("\n\n")

	// Loading state
	if m.loading {
		b.WriteString(loadingStyle.Render("Loading challenges..."))
		return b.String()
	}

	// Claiming state
	if m.claiming {
		b.WriteString(loadingStyle.Render("Claiming reward..."))
		return b.String()
	}

	// Success message
	if m.successMsg != "" {
		b.WriteString(completedStyle.Render(m.successMsg))
		b.WriteString("\n\n")
	}

	// Error state
	if m.errorMsg != "" {
		b.WriteString(errorStyle.Render(m.errorMsg))
		b.WriteString("\n\n")
		b.WriteString(subtitleStyle.Render("Press 'r' to retry"))
		return b.String()
	}

	// Empty state
	if len(m.challenges) == 0 {
		b.WriteString(subtitleStyle.Render("No challenges available"))
		return b.String()
	}

	// Render based on view mode
	if m.viewMode == ViewModeList {
		return b.String() + m.renderChallengeList()
	}
	return b.String() + m.renderChallengeDetail()
}

// renderChallengeList renders the challenge list view
func (m *DashboardModel) renderChallengeList() string {
	var b strings.Builder

	// Challenge list
	for i, challenge := range m.challenges {
		cursor := " "
		style := itemStyle
		if i == m.challengeCursor {
			cursor = ">"
			style = selectedStyle
		}

		// Count completed goals
		completed := 0
		total := len(challenge.Goals)
		for _, goal := range challenge.Goals {
			if goal.Status == "completed" || goal.Status == "claimed" {
				completed++
			}
		}

		line := fmt.Sprintf("%s %s [%d/%d]", cursor, challenge.Name, completed, total)
		b.WriteString(style.Render(line))
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(subtitleStyle.Render("Use ↑↓ to navigate, Enter to view details, 'r' to refresh, 'q' to quit"))

	return b.String()
}

// renderChallengeDetail renders the detail view for selected challenge
func (m *DashboardModel) renderChallengeDetail() string {
	if m.challengeCursor >= len(m.challenges) {
		return ""
	}

	challenge := m.challenges[m.challengeCursor]

	var b strings.Builder
	b.WriteString(titleStyle.Render(challenge.Name))
	b.WriteString("\n")
	b.WriteString(subtitleStyle.Render(challenge.Description))
	b.WriteString("\n\n")

	b.WriteString(subtitleStyle.Render("Goals:"))
	b.WriteString("\n\n")

	for i, goal := range challenge.Goals {
		b.WriteString(m.renderGoalDetailed(goal, i == m.goalCursor))
	}

	b.WriteString("\n")
	b.WriteString(subtitleStyle.Render("Use ↑↓ to navigate goals, Esc to go back, 'r' to refresh"))

	return b.String()
}

// renderGoalDetailed renders a single goal with full details
func (m *DashboardModel) renderGoalDetailed(goal api.Goal, selected bool) string {
	var b strings.Builder

	// Status icon and styling
	var icon string
	var statusStyle = itemStyle
	switch goal.Status {
	case "not_started":
		icon = "○"
		statusStyle = subtitleStyle
	case "in_progress":
		icon = "●"
		statusStyle = progressStyle
	case "completed":
		icon = "✓"
		statusStyle = completedStyle
	case "claimed":
		icon = "⚡"
		statusStyle = claimedStyle
	}

	// Cursor indicator
	cursor := " "
	if selected {
		cursor = "►"
	}

	// Progress bar (20 characters for detail view)
	progressBar := m.renderProgressBar(int(goal.Progress), int(goal.Requirement.TargetValue), 20)

	// Claim button hint
	claimHint := ""
	if goal.Status == "completed" && selected {
		claimHint = " " + highlightStyle.Render("[c] Claim")
	}

	// Build output
	nameStyle := statusStyle
	if selected {
		nameStyle = selectedStyle
	}

	b.WriteString(fmt.Sprintf("%s %s %s\n", cursor, icon, nameStyle.Render(goal.Name)))
	b.WriteString(fmt.Sprintf("  %s\n", subtitleStyle.Render(goal.Description)))

	// Show requirement details (stat code and operator)
	if goal.Requirement.StatCode != "" {
		operatorSymbol := goal.Requirement.Operator
		switch goal.Requirement.Operator {
		case "gte":
			operatorSymbol = ">="
		case "lte":
			operatorSymbol = "<="
		case "eq":
			operatorSymbol = "=="
		}
		requirementInfo := fmt.Sprintf("Requirement: %s %s %d",
			goal.Requirement.StatCode, operatorSymbol, goal.Requirement.TargetValue)
		b.WriteString(fmt.Sprintf("  %s\n", dimStyle.Render(requirementInfo)))
	}

	b.WriteString(fmt.Sprintf("  %s %d/%d%s\n", progressBar, goal.Progress, goal.Requirement.TargetValue, claimHint))

	// Show reward info
	if goal.Reward.Type != "" {
		rewardInfo := fmt.Sprintf("Reward: %s %s", goal.Reward.Type, goal.Reward.RewardID)
		if goal.Reward.Quantity > 0 {
			rewardInfo = fmt.Sprintf("%s x%d", rewardInfo, goal.Reward.Quantity)
		}
		b.WriteString(fmt.Sprintf("  %s\n", subtitleStyle.Render(rewardInfo)))
	}
	b.WriteString("\n")

	return b.String()
}

// renderProgressBar renders a progress bar using block characters
func (m *DashboardModel) renderProgressBar(current, target, width int) string {
	if target == 0 {
		return "[" + strings.Repeat("░", width) + "]"
	}

	filled := (current * width) / target
	if filled > width {
		filled = width
	}

	return fmt.Sprintf("[%s%s]",
		strings.Repeat("█", filled),
		strings.Repeat("░", width-filled))
}

// loadChallengesCmd returns a command to fetch challenges
func (m *DashboardModel) loadChallengesCmd() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		challenges, err := m.apiClient.ListChallenges(ctx)
		return ChallengesLoadedMsg{challenges: challenges, err: err}
	}
}

// claimGoalCmd returns a command to claim a goal reward
func (m *DashboardModel) claimGoalCmd(challengeID, goalID string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		result, err := m.apiClient.ClaimReward(ctx, challengeID, goalID)
		return ClaimGoalMsg{result: result, err: err}
	}
}

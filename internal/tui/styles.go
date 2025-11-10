// Copyright (c) 2025 AccelByte Inc. All Rights Reserved.
// This is licensed software from AccelByte Inc, for limitations
// and restrictions contact your company contract manager.

package tui

import "github.com/charmbracelet/lipgloss"

var (
	// Colors
	primaryColor   = lipgloss.Color("99")  // Purple
	secondaryColor = lipgloss.Color("86")  // Cyan
	warningColor   = lipgloss.Color("220") // Yellow
	errorColor     = lipgloss.Color("196") // Red
	mutedColor     = lipgloss.Color("245") // Gray

	// Header style
	headerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("230")).
			Background(primaryColor).
			Padding(0, 1).
			Bold(true)

	// Footer style
	footerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("255")). // White for better contrast
			Background(lipgloss.Color("31")).  // Darker cyan for better contrast
			Padding(0, 1)

	// Title style
	titleStyle = lipgloss.NewStyle().
			Foreground(primaryColor).
			Bold(true).
			Padding(1, 0)

	// Subtitle style
	subtitleStyle = lipgloss.NewStyle().
			Foreground(mutedColor).
			Italic(true)

	// Selected item style
	selectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("230")).
			Background(primaryColor).
			Bold(true).
			Padding(0, 1)

	// Normal item style
	itemStyle = lipgloss.NewStyle().
			Padding(0, 1)

	// Error message style
	errorStyle = lipgloss.NewStyle().
			Foreground(errorColor).
			Bold(true)

	// Loading style
	loadingStyle = lipgloss.NewStyle().
			Foreground(warningColor).
			Italic(true)

	// Success style
	successStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("2")). // green
			Bold(true)

	// Bold style
	boldStyle = lipgloss.NewStyle().
			Bold(true)

	// Dim style
	dimStyle = lipgloss.NewStyle().
			Foreground(mutedColor)

	// Progress (in progress) style
	progressStyle = lipgloss.NewStyle().
			Foreground(secondaryColor)

	// Completed style
	completedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("2")) // Green

	// Claimed style
	claimedStyle = lipgloss.NewStyle().
			Foreground(warningColor) // Yellow/Gold

	// Highlight style
	highlightStyle = lipgloss.NewStyle().
			Foreground(warningColor).
			Bold(true)
)

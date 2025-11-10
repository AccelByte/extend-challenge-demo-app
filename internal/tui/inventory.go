// Copyright (c) 2025 AccelByte Inc. All Rights Reserved.
// This is licensed software from AccelByte Inc, for limitations
// and restrictions contact your company contract manager.

package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/AccelByte/extend-challenge/extend-challenge-demo-app/internal/ags"
)

// LoadInventoryMsg triggers data loading
type LoadInventoryMsg struct{}

// InventoryLoadedMsg contains loaded data
type InventoryLoadedMsg struct {
	Entitlements []*ags.Entitlement
	Wallets      []*ags.Wallet
}

// InventoryErrorMsg contains load error
type InventoryErrorMsg struct {
	Err error
}

// InventoryModel shows entitlements and wallets
type InventoryModel struct {
	verifier     ags.RewardVerifier
	entitlements []*ags.Entitlement
	wallets      []*ags.Wallet
	loading      bool
	err          error

	// UI state
	scrollOffset int
	focusedPanel string // "entitlements" or "wallets"
}

// NewInventoryModel creates a new inventory model
func NewInventoryModel(verifier ags.RewardVerifier) *InventoryModel {
	return &InventoryModel{
		verifier:     verifier,
		focusedPanel: "entitlements",
		scrollOffset: 0,
	}
}

// Init initializes the inventory model and loads data
func (m *InventoryModel) Init() tea.Cmd {
	return m.loadInventoryCmd()
}

// Update handles messages for the inventory screen
func (m *InventoryModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "r":
			// Refresh data
			m.loading = true
			m.err = nil
			return m, m.loadInventoryCmd()

		case "tab":
			// Switch between panels
			if m.focusedPanel == "entitlements" {
				m.focusedPanel = "wallets"
			} else {
				m.focusedPanel = "entitlements"
			}
			m.scrollOffset = 0
			return m, nil

		case "up", "k":
			// Scroll up
			if m.scrollOffset > 0 {
				m.scrollOffset--
			}
			return m, nil

		case "down", "j":
			// Scroll down
			maxItems := len(m.entitlements)
			if m.focusedPanel == "wallets" {
				maxItems = len(m.wallets)
			}
			if m.scrollOffset < maxItems-1 && maxItems > 0 {
				m.scrollOffset++
			}
			return m, nil
		}

	case LoadInventoryMsg:
		m.loading = true
		m.err = nil
		return m, m.loadInventoryCmd()

	case InventoryLoadedMsg:
		m.loading = false
		m.entitlements = msg.Entitlements
		m.wallets = msg.Wallets
		m.err = nil
		return m, nil

	case InventoryErrorMsg:
		m.loading = false
		m.err = msg.Err
		return m, nil
	}

	return m, nil
}

// View renders the inventory screen
func (m *InventoryModel) View() string {
	if m.loading {
		return m.renderLoading()
	}

	if m.err != nil {
		return m.renderError()
	}

	return m.renderInventory()
}

// renderLoading renders the loading state
func (m *InventoryModel) renderLoading() string {
	return lipgloss.NewStyle().
		Padding(2).
		Render("Loading inventory data...")
}

// renderError renders the error state
func (m *InventoryModel) renderError() string {
	errorStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("9")).
		Padding(1)

	return errorStyle.Render(fmt.Sprintf("Error loading inventory: %v\n\nPress 'r' to retry", m.err))
}

// renderInventory renders the two-panel layout
func (m *InventoryModel) renderInventory() string {
	// Render entitlements panel
	entitlementsPanel := m.renderEntitlementsPanel()

	// Render wallets panel
	walletsPanel := m.renderWalletsPanel()

	// Join panels side by side
	panels := lipgloss.JoinHorizontal(
		lipgloss.Top,
		entitlementsPanel,
		"  ", // Spacing between panels
		walletsPanel,
	)

	// Summary
	summary := fmt.Sprintf("\nShowing %d entitlement(s), %d wallet(s)",
		len(m.entitlements), len(m.wallets))

	return panels + summary
}

// renderEntitlementsPanel renders the entitlements list
func (m *InventoryModel) renderEntitlementsPanel() string {
	focused := m.focusedPanel == "entitlements"

	// Panel style
	panelStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		Width(35).
		Height(15).
		Padding(1)

	if focused {
		panelStyle = panelStyle.BorderForeground(lipgloss.Color("12"))
	} else {
		panelStyle = panelStyle.BorderForeground(lipgloss.Color("8"))
	}

	// Header
	header := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("12")).
		Render("Item Entitlements")

	// Content
	var content strings.Builder

	if len(m.entitlements) == 0 {
		content.WriteString("\n(No entitlements)")
	} else {
		for i, ent := range m.entitlements {
			// Skip items before scroll offset
			if i < m.scrollOffset {
				continue
			}

			// Stop if we've rendered enough items
			if content.Len() > 300 {
				content.WriteString("\n...")
				break
			}

			// Status badge
			statusColor := "10" // Green for ACTIVE
			if ent.Status != "ACTIVE" {
				statusColor = "8" // Gray for INACTIVE
			}

			statusBadge := lipgloss.NewStyle().
				Foreground(lipgloss.Color(statusColor)).
				Render(fmt.Sprintf("[%s]", ent.Status))

			content.WriteString(fmt.Sprintf("\n%s %s\n", statusBadge, ent.ItemID))
			content.WriteString(fmt.Sprintf("  Quantity: %d\n", ent.Quantity))
			content.WriteString(fmt.Sprintf("  Granted: %s\n", ent.GrantedAt.Format("2006-01-02 15:04")))
		}
	}

	return panelStyle.Render(header + "\n" + content.String())
}

// renderWalletsPanel renders the wallets list
func (m *InventoryModel) renderWalletsPanel() string {
	focused := m.focusedPanel == "wallets"

	// Panel style
	panelStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		Width(30).
		Height(15).
		Padding(1)

	if focused {
		panelStyle = panelStyle.BorderForeground(lipgloss.Color("12"))
	} else {
		panelStyle = panelStyle.BorderForeground(lipgloss.Color("8"))
	}

	// Header
	header := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("12")).
		Render("Wallet Balances")

	// Content
	var content strings.Builder

	if len(m.wallets) == 0 {
		content.WriteString("\n(No wallets)")
	} else {
		for i, wallet := range m.wallets {
			// Skip items before scroll offset
			if i < m.scrollOffset && focused {
				continue
			}

			// Stop if we've rendered enough items
			if content.Len() > 300 {
				content.WriteString("\n...")
				break
			}

			// Status indicator
			statusIndicator := "✓"
			if wallet.Status != "ACTIVE" {
				statusIndicator = "✗"
			}

			content.WriteString(fmt.Sprintf("\n%s: %d %s\n", wallet.CurrencyCode, wallet.Balance, statusIndicator))
			content.WriteString(fmt.Sprintf("  Status: %s\n", wallet.Status))
		}
	}

	return panelStyle.Render(header + "\n" + content.String())
}

// loadInventoryCmd loads entitlements and wallets
func (m *InventoryModel) loadInventoryCmd() tea.Cmd {
	return func() tea.Msg {
		// Query entitlements
		entitlements, err := m.verifier.QueryUserEntitlements(nil)
		if err != nil {
			return InventoryErrorMsg{Err: fmt.Errorf("failed to load entitlements: %w", err)}
		}

		// Query wallets
		wallets, err := m.verifier.QueryUserWallets()
		if err != nil {
			return InventoryErrorMsg{Err: fmt.Errorf("failed to load wallets: %w", err)}
		}

		return InventoryLoadedMsg{
			Entitlements: entitlements,
			Wallets:      wallets,
		}
	}
}

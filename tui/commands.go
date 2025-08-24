package tui

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/kkrll/the-terminal-budget/data"

	tea "github.com/charmbracelet/bubbletea"
)

func (m *model) HandleCommand(cmd string) string {
	parts := strings.Fields(cmd)
	if len(parts) == 0 {
		// Reload wallets to refresh the display
		m.wallets, m.err = m.loadWallets()
		return "Display refreshed"
	}

	switch parts[0] {
	case "help":
		return "Available commands:\nadjust 0 +100 | delete 1 | hide 0,2\nnew | filter owner alice | currency USD"

	case "filter":
		if len(parts) < 2 {
			return "Usage: filter owner <name> | filter type <type> | filter currency <code> | filter reset"
		}
		if parts[1] == "reset" {
			m.filterOwner = ""
			m.filterType = ""
			m.filterCurrency = ""
			m.hiddenIndexes = make(map[int]bool)
			return "Filters cleared"
		}
		if len(parts) < 3 {
			return "Usage: filter owner <name> | filter type <type> | filter currency <code>"
		}

		filterType := parts[1]
		filterValue := strings.Join(parts[2:], " ") // Support multi-word values

		switch filterType {
		case "owner":
			m.filterOwner = filterValue
			return fmt.Sprintf("Filtering by owner: %s", filterValue)
		case "type":
			m.filterType = filterValue
			return fmt.Sprintf("Filtering by type: %s", filterValue)
		case "currency":
			m.filterCurrency = strings.ToUpper(filterValue)
			return fmt.Sprintf("Filtering by currency: %s", strings.ToUpper(filterValue))
		default:
			return "Usage: filter owner <name> | filter type <type> | filter currency <code>"
		}

	case "hide":
		if len(parts) < 2 {
			return "Usage: hide 0,2,3 (comma-separated indexes)"
		}
		return m.handleHideCommand(parts[1])

	case "currency":
		if len(parts) < 2 {
			return "Usage: currency <CURRENCY_CODE>"
		}
		return m.handleCurrencyCommand(parts[1])

	case "new":
		return m.handleNewWalletCommand()

	case "adjust":
		if len(parts) < 3 {
			return "Usage: adjust <index> <amount> (e.g., adjust 0 +100, adjust 1 -50, adjust 2 500)"
		}
		return m.handleAdjustCommand(parts[1], parts[2])

	case "delete":
		if len(parts) < 2 {
			return "Usage: delete <index>"
		}
		return m.handleDeleteCommand(parts[1])

	default:
		return fmt.Sprintf("Unknown command: %s. Type 'help' for available commands.", parts[0])
	}
}

func (m *model) handleHideCommand(indexStr string) string {
	indexes := strings.Split(indexStr, ",")
	var hiddenCount int

	for _, idxStr := range indexes {
		idxStr = strings.TrimSpace(idxStr)
		idx, err := strconv.Atoi(idxStr)
		if err != nil {
			return fmt.Sprintf("Invalid index: %s", idxStr)
		}
		if idx < 0 || idx >= len(m.wallets) {
			return fmt.Sprintf("Index %d is out of range (0-%d)", idx, len(m.wallets)-1)
		}
		m.hiddenIndexes[idx] = true
		hiddenCount++
	}

	return fmt.Sprintf("Hidden %d wallet(s)", hiddenCount)
}

func (m *model) handleCurrencyCommand(currency string) string {
	currency = strings.ToUpper(currency)
	m.displayCurrency = currency
	return fmt.Sprintf("Display currency changed to %s", currency)
}

func (m *model) handleNewWalletCommand() string {
	m.creationStep = 0
	m.creationData = Wallet{}
	m.creationInput = ""
	m.creationCursorPos = 0
	m.creationPrefilled = false

	m.creationOptions = []string{}
	m.selectedOption = 0
	m.isCustomInput = false

	m.currentScreen = walletCreationScreen

	return ""
}

func (m *model) handleAdjustCommand(indexStr, amountStr string) string {
	idx, err := strconv.Atoi(indexStr)
	if err != nil {
		return fmt.Sprintf("Invalid index: %s", indexStr)
	}

	if idx < 0 || idx >= len(m.wallets) {
		return fmt.Sprintf("Index %d is out of range (0-%d)", idx, len(m.wallets)-1)
	}

	var amount float64
	var isSet bool

	if strings.HasPrefix(amountStr, "+") || strings.HasPrefix(amountStr, "-") {
		amount, err = strconv.ParseFloat(amountStr, 64)
		isSet = false
	} else {
		amount, err = strconv.ParseFloat(amountStr, 64)
		isSet = true
	}

	if err != nil {
		return fmt.Sprintf("Invalid amount: %s", amountStr)
	}

	var dataErr error
	if isSet {
		dataErr = data.SetWalletBalanceByIndex(m.currentPath, idx, amount)
	} else {
		dataErr = data.AdjustWalletByIndex(m.currentPath, idx, amount)
	}

	if dataErr != nil {
		return fmt.Sprintf("Failed to adjust wallet: %v", dataErr)
	}

	m.wallets, m.err = m.loadWallets()
	if m.err != nil {
		return fmt.Sprintf("Wallet adjusted, but failed to reload: %v", m.err)
	}

	walletName := m.wallets[idx].Name
	if isSet {
		return fmt.Sprintf("Set %s balance to %.2f", walletName, amount)
	} else {
		return fmt.Sprintf("Adjusted %s by %.2f", walletName, amount)
	}
}

func (m *model) handleDeleteCommand(indexStr string) string {
	idx, err := strconv.Atoi(indexStr)
	if err != nil {
		return fmt.Sprintf("Invalid index: %s", indexStr)
	}

	if idx < 0 || idx >= len(m.wallets) {
		return fmt.Sprintf("Index %d is out of range (0-%d)", idx, len(m.wallets)-1)
	}

	walletName := m.wallets[idx].Name
	walletOwner := m.wallets[idx].Owner

	// Set up confirmation dialog instead of immediate deletion
	m.confirmationMessage = fmt.Sprintf("Are you sure you want to delete wallet '%s' (owned by %s)?", walletName, walletOwner)
	m.originScreen = walletScreen
	m.confirmationAction = func() error {
		return data.DeleteWalletByIndex(m.currentPath, idx)
	}
	m.onConfirm = func(m *model) (tea.Model, tea.Cmd) {
		// After successful deletion, reload wallets and clear hidden indexes
		m.wallets, m.err = m.loadWallets()
		if m.err != nil {
			m.err = fmt.Errorf("wallet deleted, but failed to reload: %v", m.err)
		}

		// Clear hidden indexes since wallet positions have changed
		m.hiddenIndexes = make(map[int]bool)

		return m, nil
	}
	m.currentScreen = confirmationScreen

	return "" // Return empty string since we're switching to confirmation screen
}

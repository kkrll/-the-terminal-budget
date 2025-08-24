package tui

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/kkrll/the-terminal-budget/data"

	tea "github.com/charmbracelet/bubbletea"
)

func (m model) loadWallets() ([]Wallet, error) {
	if m.currentPath == "" {
		return []Wallet{}, nil
	}
	budgetFile, err := data.LoadBudgetFile(m.currentPath)
	if err != nil {
		return nil, err
	}
	return budgetFile.Wallets, nil
}

func (m *model) handleCreationStep() (tea.Model, tea.Cmd) {
	input := strings.TrimSpace(m.creationInput)

	switch m.creationStep {
	case 0: // Name
		if input == "" {
			return m, nil // Name is required
		}
		m.creationData.Name = input

	case 1: // Type
		if m.isCustomInput {
			if input == "" {
				return m, nil // Custom type required
			}
			m.creationData.Type = input
		} else if m.selectedOption < len(m.creationOptions)-1 {
			m.creationData.Type = m.creationOptions[m.selectedOption]
		} else {
			// User selected "custom" option - switch to input mode
			m.isCustomInput = true
			return m, nil
		}

	case 2: // Currency
		if m.isCustomInput {
			if input == "" {
				return m, nil // Custom currency required
			}
			m.creationData.Currency = strings.ToUpper(input)
		} else if m.selectedOption < len(m.creationOptions)-1 {
			m.creationData.Currency = m.creationOptions[m.selectedOption]
		} else {
			m.isCustomInput = true
			return m, nil // Switch to custom input mode
		}

	case 3: // Owner
		if m.isCustomInput {
			if input == "" {
				return m, nil // Custom owner required
			}
			m.creationData.Owner = input
		} else if m.selectedOption < len(m.creationOptions)-1 {
			m.creationData.Owner = m.creationOptions[m.selectedOption]
		} else {
			m.isCustomInput = true
			return m, nil // Switch to custom input mode
		}

	case 4: // Balance
		var balance float64
		if input == "" {
			balance = 0.0
		} else {
			var err error
			balance, err = strconv.ParseFloat(input, 64)
			if err != nil {
				return m, nil // Invalid balance, stay on this step
			}
		}
		m.creationData.Balance = balance

		// Final step - create the wallet
		err := data.CreateWallet(
			m.currentPath,
			m.creationData.Name,
			m.creationData.Owner,
			m.creationData.Type,
			m.creationData.Currency,
			m.creationData.Balance,
		)

		if err != nil {
			// Could show error, but for now just stay on creation screen
			return m, nil
		}

		// Success - return to main screen
		m.currentScreen = walletScreen
		m.wallets, m.err = m.loadWallets()
		return m, nil
	}

	// Move to next step
	m.creationStep++
	m.creationInput = ""
	m.creationCursorPos = 0
	m.selectedOption = 0
	m.isCustomInput = false

	// Populate options for the new step
	switch m.creationStep {
	case 1:
		m.populateTypeOptions()
	case 2:
		m.populateCurrencyOptions()
	case 3:
		m.populateOwnerOptions()
	default:
		m.creationOptions = []string{}
	}

	return m, nil
}

func (m *model) populateTypeOptions() {
	budgetFile, err := data.LoadBudgetFile(m.currentPath)
	if err != nil {
		m.creationOptions = []string{"cash", "bank", "invest", "custom: enter new type..."}
		return
	}

	// Get unique types from existing wallets
	typeSet := make(map[string]bool)
	for _, wallet := range budgetFile.Wallets {
		typeSet[wallet.Type] = true
	}

	var options []string
	if len(budgetFile.Wallets) == 0 {
		options = []string{"bank", "cash", "invest"}
	} else {
		for walletType := range typeSet {
			options = append(options, walletType)
		}
	}

	// Add custom option
	options = append(options, "custom: enter new type...")
	m.creationOptions = options

	if m.selectedOption >= len(m.creationOptions) {
		m.selectedOption = 0
	}
}

func (m *model) populateCurrencyOptions() {
	budgetFile, err := data.LoadBudgetFile(m.currentPath)
	if err != nil {
		m.creationOptions = []string{"USD", "EUR", "GBP", "custom: enter currency code..."}
		return
	}

	// Get unique currencies from existing wallets
	currencySet := make(map[string]bool)
	for _, wallet := range budgetFile.Wallets {
		currencySet[wallet.Currency] = true
	}

	var options []string
	if len(budgetFile.Wallets) == 0 {
		options = []string{"USD", "EUR", "GBP"}
	} else {
		for walletType := range currencySet {
			options = append(options, walletType)
		}
	}

	// Add custom option
	options = append(options, "custom: enter currency code...")
	m.creationOptions = options

	if m.selectedOption >= len(m.creationOptions) {
		m.selectedOption = 0
	}
}

func (m *model) populateOwnerOptions() {
	budgetFile, err := data.LoadBudgetFile(m.currentPath)
	if err != nil {
		m.creationOptions = []string{"User", "custom: enter owner name..."}
		return
	}

	// Get unique owners from existing wallets
	ownerSet := make(map[string]bool)
	for _, wallet := range budgetFile.Wallets {
		ownerSet[wallet.Owner] = true
	}

	var options []string
	if len(budgetFile.Wallets) == 0 {
		options = []string{"me"}
	} else {
		for walletType := range ownerSet {
			options = append(options, walletType)
		}
	}

	// Add custom option
	options = append(options, "custom: enter owner name...")
	m.creationOptions = options

	if m.selectedOption >= len(m.creationOptions) {
		m.selectedOption = 0
	}
}

func (m *model) handleNewFileCreation() (tea.Model, tea.Cmd) {
	m.currentScreen = budgetCreationScreen
	m.creationCursorPos = 0
	m.creationInput = ""
	m.creationStep = 0
	return m, nil
}

// Screen-specific input handlers

// Greeting screen input handling
func (m *model) handleGreetingInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		if m.isNewFile {
			return m.handleNewFileCreation()
		} else {
			// Load selected budget file
			selectedFile := m.availableFiles[m.selectedFileIndex]
			m.currentPath = selectedFile.Name
			m.wallets = selectedFile.Wallets
			m.currentScreen = walletScreen
		}
		return m, nil
	case "esc":
		return m, tea.Quit
	case "d", "delete":
		if len(m.availableFiles) > 0 && !m.isNewFile {
			// Delete selected budget file
			selectedFile := m.availableFiles[m.selectedFileIndex]
			m.confirmationMessage = fmt.Sprintf("Are you sure you want to delete the budget file '%s'?", selectedFile.Name)
			m.originScreen = greetingScreen
			m.confirmationAction = func() error {
				return data.DeleteBudgetFile(selectedFile.Name)
			}
			m.onConfirm = func(m *model) (tea.Model, tea.Cmd) {
				// After deletion, refresh the file list and return to greeting screen
				m.availableFiles, m.err = data.ListBudgetFiles()
				if m.err != nil {
					m.err = fmt.Errorf("failed to list budget files: %v", m.err)
				}
				m.selectedFileIndex = 0
				m.isNewFile = false
				return m, nil
			}
			m.currentScreen = confirmationScreen
			return m, nil
		}
	case "up", "down":
		totalOptions := len(m.availableFiles) + 1 // +1 for "Create new budget..."
		if msg.String() == "up" && m.selectedFileIndex > 0 {
			m.selectedFileIndex--
		} else if msg.String() == "down" && m.selectedFileIndex < totalOptions-1 {
			m.selectedFileIndex++
		}

		// Check if "Create new budget..." is selected
		m.isNewFile = (m.selectedFileIndex == len(m.availableFiles))
		if !m.isNewFile {
			m.creationInput = ""
			m.creationCursorPos = 0
		}
		return m, nil
	}
	return m, nil
}

// Budget creation screen input handling
func (m *model) handleBudgetCreationInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		return m.handleBudgetFileCreation()
	case "esc":
		m.currentScreen = greetingScreen
		m.availableFiles, m.err = data.ListBudgetFiles()
		if m.err != nil {
			m.err = fmt.Errorf("failed to list budget files: %v", m.err)
		}
		CleanSlates(m)
		return m, nil
	case "backspace":
		if len(m.creationInput) > 0 && m.creationCursorPos > 0 {
			m.creationInput = m.creationInput[:m.creationCursorPos-1] + m.creationInput[m.creationCursorPos:]
			m.creationCursorPos--
		}
		return m, nil
	case "left":
		if m.creationCursorPos > 0 {
			m.creationCursorPos--
		}
		return m, nil
	case "right":
		if m.creationCursorPos < len(m.creationInput) {
			m.creationCursorPos++
		}
		return m, nil
	default:
		// Handle text input
		m.creationInput = m.creationInput[:m.creationCursorPos] + msg.String() + m.creationInput[m.creationCursorPos:]
		m.creationCursorPos++
		return m, nil
	}
}

// Confirmation screen input handling
func (m *model) handleConfirmationInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y", "Y":
		if m.confirmationAction != nil {
			err := m.confirmationAction()
			if err != nil {
				m.err = fmt.Errorf("action failed: %v", err)
				originScreen := m.originScreen // Save origin screen before cleanup
				CleanSlates(m)
				m.currentScreen = originScreen
				return m, nil
			} else if m.onConfirm != nil {
				onConfirmFunc := m.onConfirm
				originScreen := m.originScreen // Save origin screen before cleanup
				CleanSlates(m)
				m.currentScreen = originScreen
				return onConfirmFunc(m)
			}
		}
		originScreen := m.originScreen // Save origin screen before cleanup
		CleanSlates(m)
		m.currentScreen = originScreen
		return m, nil
	case "n", "N", "esc":
		originScreen := m.originScreen // Save origin screen before cleanup
		CleanSlates(m)
		m.currentScreen = originScreen
		return m, nil
	}
	return m, nil
}

// Wallet screen input handling
func (m *model) handleWalletInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		m.commandResult = m.HandleCommand(m.commandInput)
		m.commandInput = ""
		m.cursorPos = 0
		return m, nil
	case "esc":
		m.currentScreen = greetingScreen
		m.availableFiles, m.err = data.ListBudgetFiles()
		if m.err != nil {
			m.err = fmt.Errorf("failed to list budget files: %v", m.err)
		}
		CleanSlates(m)
		return m, nil
	case "backspace":
		if len(m.commandInput) > 0 && m.cursorPos > 0 {
			m.commandInput = m.commandInput[:m.cursorPos-1] + m.commandInput[m.cursorPos:]
			m.cursorPos--
		}
		return m, nil
	case "left":
		if m.cursorPos > 0 {
			m.cursorPos--
		}
		return m, nil
	case "right":
		if m.cursorPos < len(m.commandInput) {
			m.cursorPos++
		}
		return m, nil
	default:
		// Handle text input
		m.commandInput = m.commandInput[:m.cursorPos] + msg.String() + m.commandInput[m.cursorPos:]
		m.cursorPos++
		return m, nil
	}
}

// Wallet creation screen input handling
func (m *model) handleWalletCreationInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		return m.handleCreationStep()
	case "esc":
		m.currentScreen = walletScreen
		m.wallets, m.err = m.loadWallets()
		CleanSlates(m)
		return m, nil
	case "backspace":
		if len(m.creationInput) > 0 && m.creationCursorPos > 0 && (m.creationStep == 0 || m.creationStep == 4 || m.isCustomInput) {
			m.creationInput = m.creationInput[:m.creationCursorPos-1] + m.creationInput[m.creationCursorPos:]
			m.creationCursorPos--
		}
		return m, nil
	case "left":
		if m.creationCursorPos > 0 && (m.creationStep == 0 || m.creationStep == 4 || m.isCustomInput) {
			m.creationCursorPos--
		}
		return m, nil
	case "right":
		if m.creationCursorPos < len(m.creationInput) && (m.creationStep == 0 || m.creationStep == 4 || m.isCustomInput) {
			m.creationCursorPos++
		}
		return m, nil
	case "up", "down":
		if len(m.creationOptions) > 0 {
			if msg.String() == "up" && m.selectedOption > 0 {
				m.selectedOption--
			} else if msg.String() == "down" && m.selectedOption < len(m.creationOptions)-1 {
				m.selectedOption++
			}

			// Check if "custom input" option is selected
			m.isCustomInput = (m.selectedOption == len(m.creationOptions)-1)
			if !m.isCustomInput {
				m.creationInput = ""
				m.creationCursorPos = 0
			}
		}
		return m, nil
	default:
		// Handle text input only for appropriate steps
		if m.creationStep == 0 || m.creationStep == 4 || m.isCustomInput {
			m.creationInput = m.creationInput[:m.creationCursorPos] + msg.String() + m.creationInput[m.creationCursorPos:]
			m.creationCursorPos++
		}
		return m, nil
	}
}

func (m model) handleBudgetFileCreation() (tea.Model, tea.Cmd) {
	filename := m.creationInput

	if filename == "" {
		return m, nil
	}

	err := data.CreateBudgetFile(filename)
	if err != nil {
		return m, nil
	}

	m.currentPath = filename
	m.wallets = []Wallet{}
	m.currentScreen = walletScreen

	m.creationCursorPos = 0
	m.creationInput = ""
	m.isNewFile = false

	return m, nil
}

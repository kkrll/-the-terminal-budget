package tui

import (
	"github.com/kkrll/the-terminal-budget/data"

	tea "github.com/charmbracelet/bubbletea"
)

type screen int

const (
	greetingScreen screen = iota
	budgetCreationScreen
	walletScreen
	walletCreationScreen
	confirmationScreen
)

// Use the shared Wallet type from data package
type Wallet = data.Wallet

type model struct {
	currentScreen   screen
	greeting        string
	currentPath     string
	wallets         []Wallet
	err             error
	width           int
	height          int
	commandInput    string
	commandResult   string
	cursorPos       int
	hiddenIndexes   map[int]bool
	filterOwner     string
	filterType      string
	filterCurrency  string
	displayCurrency string

	// File selection state
	availableFiles    []data.BudgetFile
	selectedFileIndex int
	isNewFile         bool

	// Wallet creation state
	creationStep      int
	creationData      Wallet
	creationInput     string
	creationCursorPos int
	creationPrefilled bool

	// Selection state for option-based steps
	creationOptions []string
	selectedOption  int
	isCustomInput   bool

	//Comfirmation dialog state
	confirmationMessage string
	confirmationAction  func() error
	onConfirm           func(*model) (tea.Model, tea.Cmd)
	originScreen        screen
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	case tea.KeyMsg:
		// Handle global quit first
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}

		// Then handle screen-specific input
		switch m.currentScreen {
		case greetingScreen:
			return (&m).handleGreetingInput(msg)
		case budgetCreationScreen:
			return (&m).handleBudgetCreationInput(msg)
		case confirmationScreen:
			return (&m).handleConfirmationInput(msg)
		case walletScreen:
			return (&m).handleWalletInput(msg)
		case walletCreationScreen:
			return (&m).handleWalletCreationInput(msg)
		}
	}
	return m, nil
}

func (m model) View() string {
	switch m.currentScreen {
	case greetingScreen:
		return m.GreetingView()
	case confirmationScreen:
		return m.ConfirmationView()
	case budgetCreationScreen:
		return m.BudgetCreationView()
	case walletScreen:
		return m.walletsView()
	case walletCreationScreen:
		return m.walletCreationView()
	}
	return ""
}

func RunTUI() error {
	// Load available budget files
	availableFiles, err := data.ListBudgetFiles()
	if err != nil {
		availableFiles = []data.BudgetFile{} // Start with empty list if error
	}

	// Auto-create example budget if no files exist (first-time user experience)
	if len(availableFiles) == 0 {
		err := data.CreateExampleBudget()
		if err != nil {
			// If example creation fails, continue anyway - user can create manually
			// This ensures the app doesn't crash on first run
		} else {
			// Reload files to include the new example budget
			availableFiles, _ = data.ListBudgetFiles()
		}
	}

	initialModel := model{
		currentScreen:     greetingScreen,
		greeting:          getRandomGreeting(),
		width:             80,
		height:            24,
		hiddenIndexes:     make(map[int]bool),
		availableFiles:    availableFiles,
		selectedFileIndex: 0,
		isNewFile:         false,
	}

	p := tea.NewProgram(initialModel, tea.WithAltScreen())
	_, err = p.Run()
	return err
}

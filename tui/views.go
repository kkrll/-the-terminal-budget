package tui

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/kkrll/the-terminal-budget/data"

	"github.com/charmbracelet/lipgloss"
)

var greetings = []string{
	"hey, how are you?",
	"hi there!",
	"ah, that's you",
	"long time no see, human",
	"here you are",
	"welcome back, friend",
	"good to see you, friend",
	"greetings, traveller",
	"oh, it's you again",
	"look who showed up",
	"hey stranger",
	"nice to have you here",
	"salutations",
	"ahoy!",
	"welcome, human",
	"system online: user detected",
	"hey, commander",
	"ready for action?",
	"glad you made it",
	"hi, friend",
	"hi, wanderer",
	"welcome, adventurer",
	"hail, wayfarer",
	"well met, explorer",
	"ah, a seeker arrives",
	"welcome back, wanderer",
	"hello, kindred spirit",
	"salutations, voyager",
	"the path brings you here again",
	"ah, a fellow traveler of the terminal",
	"welcome, lost soul",
	"good to see you, pilgrim",
	"the journey continues, friend",
	"back from your quest?",
	"hello, drifter",
	"well met, companion",
	"the road greets you once more",
}

func (m model) GreetingView() string {
	greeting := lipgloss.NewStyle().
		Bold(true).
		Align(lipgloss.Center).
		Render(m.greeting)

	question := lipgloss.NewStyle().
		Align(lipgloss.Center).
		Render("What budget are we dealing with today?")

	fileList := m.createFileSelectionList()

	separator := lipgloss.NewStyle().
		Align(lipgloss.Center).
		Foreground(lipgloss.Color("#626262")).
		Render("──────────────────────────────────────────────")

	instructionLine1 := lipgloss.NewStyle().
		Align(lipgloss.Center).
		Foreground(lipgloss.Color("#626262")).
		Render("Use ↑↓ to select, ⏎ to open  |  [d] or Delete to delete")

	instructionLine2 := lipgloss.NewStyle().
		Align(lipgloss.Center).
		Foreground(lipgloss.Color("#626262")).
		Render("[ctrl+c] to quit the app")

	content := lipgloss.JoinVertical(
		lipgloss.Center,
		greeting,
		"",
		question,
		"",
		fileList,
		"",
		separator,
		"",
		instructionLine1,
		instructionLine2,
	)

	return lipgloss.Place(
		m.width, m.height,
		lipgloss.Center,
		lipgloss.Center,
		content,
	)
}

func getRandomGreeting() string {
	rand.Seed(time.Now().UnixNano())
	return greetings[rand.Intn(len(greetings))]
}

func (m model) createFileSelectionList() string {
	var items []string

	// Add existing files
	for i, file := range m.availableFiles {
		var prefix string
		if i == m.selectedFileIndex {
			prefix = "[x] "
		} else {
			prefix = "[ ] "
		}

		displayName := truncate(file.Name, 20)
		// Format the time nicely
		timeAgo := formatTimeAgo(file.UpdatedAt)
		item := fmt.Sprintf("%s%-20s Updated %s", prefix, displayName, timeAgo)
		items = append(items, item)
	}

	// Add "Create new budget..." option
	var prefix string
	if m.selectedFileIndex == len(m.availableFiles) {
		prefix = "[x] "
	} else {
		prefix = "[ ] "
	}
	items = append(items, prefix+"Create new budget...")

	return lipgloss.NewStyle().
		Align(lipgloss.Left).
		Render(strings.Join(items, "\n"))
}

func (m model) ConfirmationView() string {
	title := lipgloss.NewStyle().
		Bold(true).
		Render(m.confirmationMessage)

	options := "[Y]es | [N]o / Esc"

	instructions := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#626262")).
		Render("Press Y to confirm, N or Esc to cancel")

	content := lipgloss.JoinVertical(
		lipgloss.Center,
		title,
		"",
		options,
		"",
		instructions,
	)

	return lipgloss.Place(
		m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		content,
	)
}

func (m model) BudgetCreationView() string {
	title := lipgloss.NewStyle().
		Bold(true).
		Render("CREATE NEW BUDGET")

	content := lipgloss.JoinVertical(
		lipgloss.Center,
		title,
		"",
		"Enter budget name:",
		"",
		m.createTextInput(),
		"",
		m.createCreationInstructions())

	return lipgloss.Place(
		m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		content,
	)
}

func (m model) walletsView() string {
	if m.err != nil {
		errorMsg := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF0000")).
			Bold(true).
			Align(lipgloss.Center).
			Render(fmt.Sprintf("Error loading wallets: %v", m.err))
		return lipgloss.Place(
			m.width, m.height,
			lipgloss.Center, lipgloss.Center,
			errorMsg,
		)
	}

	table := m.createWalletTable()

	inputBox := m.createInputBox()

	commandHints := m.createCommandHints()

	if len(m.wallets) == 0 {
		emptyMsg := lipgloss.NewStyle().
			Render("No wallets found. Would you like to create one?")

		content := lipgloss.JoinVertical(
			lipgloss.Center,
			emptyMsg,
			"",
			"",
			"",
			inputBox,
			commandHints,
		)

		return lipgloss.Place(
			m.width, m.height,
			lipgloss.Center, lipgloss.Center,
			content,
		)
	}

	content := lipgloss.JoinVertical(
		lipgloss.Center,
		table,
		"",
		"",
		"",
		inputBox,
		commandHints,
	)

	return lipgloss.Place(
		m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		content,
	)
}

func (m model) createWalletTable() string {
	title := lipgloss.NewStyle().
		Bold(true).
		Render("YOUR BUDGET")

	headers := []string{"Name", "Owner", "Type", "Balance", "Currency"}

	headerRow := fmt.Sprintf("    %-15s %-12s %-10s %10s  %-8s",
		headers[0], headers[1], headers[2], headers[3], headers[4])

	separator := strings.Repeat("-", len(headerRow))

	var rows []string
	rows = append(rows, title)
	rows = append(rows, "")
	rows = append(rows, "")
	rows = append(rows, headerRow)
	rows = append(rows, separator)

	for i, wallet := range m.wallets {
		row := fmt.Sprintf("%2d. %-15s %-12s %-10s %10.2f  %-8s",
			i,
			truncate(wallet.Name, 15),
			truncate(wallet.Owner, 12),
			truncate(wallet.Type, 10),
			wallet.Balance,
			wallet.Currency,
		)

		var isExscluded bool = false
		if m.hiddenIndexes[i] {
			isExscluded = true
		}
		if m.filterOwner != "" && wallet.Owner != m.filterOwner {
			isExscluded = true
		}
		if m.filterType != "" && wallet.Type != m.filterType {
			isExscluded = true
		}
		if m.filterCurrency != "" && wallet.Currency != m.filterCurrency {
			isExscluded = true
		}

		displayRow := row
		if isExscluded {
			displayRow = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#626262")).
				Strikethrough(true).
				Italic(true).
				Render(row)
		}

		rows = append(rows, displayRow)
	}

	total := m.calculateTotal()
	rows = append(rows, separator)
	rows = append(rows, total)

	return strings.Join(rows, "\n")
}

func (m model) calculateTotal() string {
	if len(m.wallets) == 0 {
		return "0 wallets                                     0.00"
	}

	budgetData, err := data.LoadBudgetFile(m.currentPath)
	if err != nil {
		return fmt.Sprintf("Error loading budget data: %v", err)
	}

	defaultCurrency, err := data.GetDefaultCurrency(budgetData)
	if err != nil {
		return fmt.Sprintf("Error getting default currency: %v", err)
	}

	// Use display currency if set, otherwise use default currency
	targetCurrency := defaultCurrency
	if m.displayCurrency != "" {
		targetCurrency = m.displayCurrency
	}

	total := 0.0
	visibleCount := 0

	for i, wallet := range m.wallets {
		// Skip hidden or filtered wallets
		if m.hiddenIndexes[i] {
			continue
		}
		if m.filterOwner != "" && wallet.Owner != m.filterOwner {
			continue
		}
		if m.filterType != "" && wallet.Type != m.filterType {
			continue
		}
		if m.filterCurrency != "" && wallet.Currency != m.filterCurrency {
			continue
		}

		visibleCount++

		if wallet.Currency == targetCurrency {
			total += wallet.Balance
		} else {
			converted, err := data.ConvertCurrency(wallet.Balance, wallet.Currency, targetCurrency, defaultCurrency)
			if err != nil {
				total += wallet.Balance
			} else {
				total += converted
			}
		}
	}

	walletCount := fmt.Sprintf("%d wallet", visibleCount)
	if visibleCount != 1 {
		walletCount += "s"
	}
	totalWidth := 64
	leftSide := walletCount
	rightSide := fmt.Sprintf("%.2f %s", total, targetCurrency)
	spacing := totalWidth - len(leftSide) - len(rightSide)
	if spacing < 1 {
		spacing = 1
	}

	return fmt.Sprintf("%s%s%s", leftSide, strings.Repeat(" ", spacing), rightSide)
}

func (m model) createInputBox() string {
	var inputContent string
	if m.commandInput == "" {
		inputContent = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262")).
			Render("█Enter command...")
	} else {
		displayText := m.commandInput
		if m.cursorPos < len(displayText) {
			displayText = displayText[:m.cursorPos] + "█" + displayText[m.cursorPos:]
		} else {
			displayText = displayText + "█"
		}
		inputContent = displayText
	}

	inputBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		Padding(1, 2).
		Width(64).
		Render(inputContent)

	return inputBox
}

func (m model) createCommandHints() string {

	var line1, line2, line3 string
	line3 = ""
	currentCommand := firstN(m.commandInput, 2)

	switch currentCommand {
	case "fi":
		line1 = "Filter calculated wallets by owner, type, or currency:"
		line2 = "'filter owner <name>' | 'filter type <type>' | 'filter currency <code>' | 'filter reset'"
		line3 = "'filter reset' clears all the filters applied."
	case "hi":
		line1 = "Exclude wallets from calculations by index:"
		line2 = "'hide 0,2,3' (comma-separated indexes)"
	case "cu":
		line1 = "Set display currency for total calculation:"
		line2 = "'currency <CURRENCY_CODE>' (e.g., USD, EUR, GBP)"
	case "ne":
		line1 = "Create new wallet:"
		line2 = "command 'new' launches wallet creation wizard"
	case "ad":
		line1 = "Adjust wallet balance by index:"
		line2 = "'adjust <index> <amount>'"
		line3 = "(e.g., 'adjust 0 +100', 'adjust 1 -50', 'adjust 2 500')"
	case "de":
		line1 = "Delete wallet by index:"
		line2 = "'delete <index>'"
	default:
		line1 = "Available commands:"
		line2 = "new |  adjust  |  hide  |  filter  |  currency  |  delete"
	}

	hints := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#626262")).
		Width(64).
		Render(lipgloss.JoinVertical(lipgloss.Left, line1, line2, line3))

	return hints
}

func (m model) walletCreationView() string {
	title := lipgloss.NewStyle().
		Bold(true).
		Render("CREATE NEW WALLET")

	var content []string
	content = append(content, title)
	content = append(content, "")

	switch m.creationStep {
	case 0: // Name input
		content = append(content, "Step 1 of 5. Name your new wallet")
		content = append(content, "")
		content = append(content, m.createTextInput())

	case 1, 2, 3: // Type, Currency, Owner selection
		var prompt string
		switch m.creationStep {
		case 1:
			prompt = "Step 2 of 5. What's your wallet type?"
		case 2:
			prompt = "Step 3 of 5. What currency?"
		case 3:
			prompt = "Step 4 of 5. Who owns this wallet?"
		}

		content = append(content, prompt)
		content = append(content, "")
		content = append(content, m.createSelectionList())
		content = append(content, "")
		content = append(content, m.createTextInput()) // Always show input

	case 4: // Balance input
		content = append(content, "Step 5 of 5. Initial balance (or press Enter for 0.00)")
		content = append(content, "")
		content = append(content, m.createTextInput())
	}

	// Current values display
	if m.creationStep > 0 {
		content = append(content, "")
		content = append(content, m.createCurrentValues())
	}

	content = append(content, "")
	content = append(content, m.createCreationInstructions())

	finalContent := lipgloss.JoinVertical(lipgloss.Center, content...)

	return lipgloss.Place(
		m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		finalContent,
	)
}

func (m model) createSelectionList() string {
	var items []string

	for i, option := range m.creationOptions {
		var prefix string
		if i == m.selectedOption {
			prefix = "[x] "
		} else {
			prefix = "[ ] "
		}

		items = append(items, prefix+option)
	}

	return lipgloss.NewStyle().
		Align(lipgloss.Left).
		Render(strings.Join(items, "\n"))
}

func (m model) createTextInput() string {
	var inputContent string
	if m.creationInput == "" {
		inputContent = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262")).
			Render("█Type here...")
	} else {
		displayText := m.creationInput
		if m.creationCursorPos < len(displayText) {
			displayText = displayText[:m.creationCursorPos] + "█" + displayText[m.creationCursorPos:]
		} else {
			displayText = displayText + "█"
		}
		inputContent = displayText
	}

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		Padding(0, 1).
		Width(40).
		Render(inputContent)
}

func (m model) createCurrentValues() string {
	var values []string

	if m.creationData.Name != "" {
		values = append(values, fmt.Sprintf("Name: %s", m.creationData.Name))
	}
	if m.creationData.Type != "" {
		values = append(values, fmt.Sprintf("Type: %s", m.creationData.Type))
	}
	if m.creationData.Currency != "" {
		values = append(values, fmt.Sprintf("Currency: %s", m.creationData.Currency))
	}
	if m.creationData.Owner != "" {
		values = append(values, fmt.Sprintf("Owner: %s", m.creationData.Owner))
	}

	if len(values) == 0 {
		return ""
	}

	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("#888888")).
		Render("Current: " + strings.Join(values, " | "))
}

func (m model) createCreationInstructions() string {
	var instructions []string

	if m.creationStep == 0 || (len(m.creationOptions) > 0 && m.isCustomInput) {
		instructions = append(instructions, "Type and press Enter")
	} else if len(m.creationOptions) > 0 {
		instructions = append(instructions, "Use ↑↓ to select, Enter to continue")
	} else {
		instructions = append(instructions, "Press Enter to continue")
	}

	instructions = append(instructions, "Esc to cancel")

	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("#626262")).
		Render(strings.Join(instructions, "  |  "))
}

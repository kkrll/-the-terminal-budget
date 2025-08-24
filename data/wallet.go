package data

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type Wallet struct {
	Name     string  `json:"name"`
	Owner    string  `json:"owner"`
	Type     string  `json:"type"`
	Currency string  `json:"currency"`
	Balance  float64 `json:"balance"`
}

type BudgetData struct {
	Wallets         []Wallet `json:"wallets"`
	DefaultCurrency string   `json:"default_currency"`
}

type BudgetFile struct {
	Name            string    `json:"name"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
	Wallets         []Wallet  `json:"wallets"`
	DefaultCurrency string    `json:"default_currency"`
}

func GetFilesDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".budget")
}

func ListBudgetFiles() ([]BudgetFile, error) {
	budgetsDir := filepath.Join(GetFilesDir(), "files")

	// Create directory if it doesn't exist
	if err := os.MkdirAll(budgetsDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create budgets directory: %v", err)
	}

	// Read directory contents
	files, err := os.ReadDir(budgetsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read budgets directory: %v", err)
	}

	var budgetFiles []BudgetFile
	for _, file := range files {
		if !file.IsDir() && filepath.Ext(file.Name()) == ".json" {
			filename := strings.TrimSuffix(file.Name(), ".json")
			budgetFile, err := LoadBudgetFile(filename)
			if err != nil {
				// Skip files that can't be loaded
				continue
			}
			budgetFiles = append(budgetFiles, *budgetFile)
		}
	}

	// Sort by UpdatedAt (most recent first)
	sort.Slice(budgetFiles, func(i, j int) bool {
		return budgetFiles[i].UpdatedAt.After(budgetFiles[j].UpdatedAt)
	})

	return budgetFiles, nil
}

func LoadBudgetFile(filename string) (*BudgetFile, error) {
	budgetsDir := filepath.Join(GetFilesDir(), "files")
	filePath := filepath.Join(budgetsDir, filename+".json")

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("budget file '%s' does not exist", filename)
	}

	file, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read budget file '%s': %v", filename, err)
	}

	// Try to parse as new BudgetFile format first
	var budgetFile BudgetFile
	if err := json.Unmarshal(file, &budgetFile); err == nil && budgetFile.Name != "" {
		return &budgetFile, nil
	}

	// Fall back to old BudgetData format for backward compatibility
	var oldData BudgetData
	if err := json.Unmarshal(file, &oldData); err != nil {
		return nil, fmt.Errorf("failed to parse budget file '%s' in any known format: %v", filename, err)
	}

	// Convert old format to new format
	now := time.Now()
	newBudgetFile := &BudgetFile{
		Name:            filename,
		CreatedAt:       now,
		UpdatedAt:       now,
		Wallets:         oldData.Wallets,
		DefaultCurrency: oldData.DefaultCurrency,
	}

	// If no default currency, try to infer from first wallet
	if newBudgetFile.DefaultCurrency == "" && len(newBudgetFile.Wallets) > 0 {
		newBudgetFile.DefaultCurrency = newBudgetFile.Wallets[0].Currency
	}

	// Save in new format for future use
	SaveBudgetFile(newBudgetFile)

	return newBudgetFile, nil
}

func CreateBudgetFile(filename string) error {
	budgetsDir := filepath.Join(GetFilesDir(), "files")

	// Create directory if it doesn't exist
	if err := os.MkdirAll(budgetsDir, 0755); err != nil {
		return fmt.Errorf("failed to create budgets directory: %v", err)
	}

	filePath := filepath.Join(budgetsDir, filename+".json")

	// Check if file already exists
	if _, err := os.Stat(filePath); err == nil {
		return fmt.Errorf("budget file '%s' already exists", filename)
	}

	now := time.Now()
	budgetFile := BudgetFile{
		Name:            filename,
		CreatedAt:       now,
		UpdatedAt:       now,
		Wallets:         []Wallet{},
		DefaultCurrency: "USD", // Default fallback
	}

	jsonData, err := json.MarshalIndent(budgetFile, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal budget file: %v", err)
	}

	if err := os.WriteFile(filePath, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to write budget file '%s': %v", filename, err)
	}

	return nil
}

func SaveBudgetFile(budgetFile *BudgetFile) error {
	budgetsDir := filepath.Join(GetFilesDir(), "files")

	// Create directory if it doesn't exist
	if err := os.MkdirAll(budgetsDir, 0755); err != nil {
		return fmt.Errorf("failed to create budgets directory: %v", err)
	}

	filePath := filepath.Join(budgetsDir, budgetFile.Name+".json")

	// Update the timestamp
	budgetFile.UpdatedAt = time.Now()

	jsonData, err := json.MarshalIndent(budgetFile, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal budget file: %v", err)
	}

	if err := os.WriteFile(filePath, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to write budget file '%s': %v", budgetFile.Name, err)
	}

	return nil
}

func GetExistingCurrencies(data *BudgetFile) []string {
	currencies := make(map[string]bool)
	for _, wallet := range data.Wallets {
		currencies[wallet.Currency] = true
	}

	var currenciesList []string
	for currency := range currencies {
		currenciesList = append(currenciesList, currency)
	}

	return currenciesList
}

func GetDefaultCurrency(data *BudgetFile) (string, error) {
	if data.DefaultCurrency != "" {
		return data.DefaultCurrency, nil
	}
	if len(data.Wallets) > 0 {
		return data.Wallets[0].Currency, nil
	}

	return "", fmt.Errorf("no default currency set and no wallets exist")
}

func SetDefaultCurrency(currency string, data *BudgetFile) error {
	data.DefaultCurrency = currency
	return SaveBudgetFile(data)
}

func CreateWallet(filename, name, owner, walletType, currency string, balance float64) error {
	data, err := LoadBudgetFile(filename)
	if err != nil {
		return err
	}

	for _, wallet := range data.Wallets {
		if wallet.Name == name {
			return fmt.Errorf("wallet with name '%s' already exists", name)
		}
	}

	currency, validationErr := normalizeCurrency(currency)
	if validationErr != nil {
		return fmt.Errorf("invalid currency: %v", validationErr)
	}

	newWallet := Wallet{
		Name:     name,
		Owner:    owner,
		Type:     walletType,
		Currency: currency,
		Balance:  balance,
	}

	data.Wallets = append(data.Wallets, newWallet)
	return SaveBudgetFile(data)
}

func AdjustWalletByIndex(filename string, index int, amount float64) error {
	data, err := LoadBudgetFile(filename)
	if err != nil {
		return err
	}

	if index < 0 || index >= len(data.Wallets) {
		return fmt.Errorf("wallet index %d is out of range (0-%d)", index, len(data.Wallets)-1)
	}

	data.Wallets[index].Balance += amount
	return SaveBudgetFile(data)
}

func SetWalletBalanceByIndex(filename string, index int, balance float64) error {
	data, err := LoadBudgetFile(filename)
	if err != nil {
		return err
	}

	if index < 0 || index >= len(data.Wallets) {
		return fmt.Errorf("wallet index %d is out of range (0-%d)", index, len(data.Wallets)-1)
	}

	data.Wallets[index].Balance = balance
	return SaveBudgetFile(data)
}

func DeleteWalletByIndex(filename string, index int) error {
	data, err := LoadBudgetFile(filename)
	if err != nil {
		return err
	}

	if index < 0 || index >= len(data.Wallets) {
		return fmt.Errorf("wallet index %d is out of range (0-%d)", index, len(data.Wallets)-1)
	}

	data.Wallets = append(data.Wallets[:index], data.Wallets[index+1:]...)
	return SaveBudgetFile(data)
}

func DeleteBudgetFile(filename string) error {
	budgetsDir := filepath.Join(GetFilesDir(), "files")
	filePath := filepath.Join(budgetsDir, filename+".json")

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("budget file '%s' does not exist", filename)
	}

	if err := os.Remove(filePath); err != nil {
		return fmt.Errorf("failed to delete budget file '%s': %v", filename, err)
	}

	return nil
}

// CreateExampleBudget creates a sample budget with example wallets for new users
func CreateExampleBudget() error {
	// Create the example budget file
	err := CreateBudgetFile("example")
	if err != nil {
		return fmt.Errorf("failed to create example budget file: %v", err)
	}

	// Add sample wallets
	sampleWallets := []struct {
		name, owner, walletType, currency string
		balance                           float64
	}{
		{"Cash Wallet", "User", "cash", "USD", 250.75},
		{"Bank Account", "User", "bank", "USD", 1500.00},
		{"Savings Fund", "User", "bank", "EUR", 800.50},
		{"Investment Portfolio", "User", "invest", "USD", 5000.00},
		{"Emergency Fund", "Family", "bank", "USD", 2000.00},
	}

	for _, wallet := range sampleWallets {
		err := CreateWallet("example", wallet.name, wallet.owner, wallet.walletType, wallet.currency, wallet.balance)
		if err != nil {
			return fmt.Errorf("failed to create sample wallet '%s': %v", wallet.name, err)
		}
	}

	return nil
}

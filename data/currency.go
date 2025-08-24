package data

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type FrankfurterResponse struct {
	Amount float64            `json:"amount"`
	Base   string             `json:"base"`
	Date   string             `json:"date"`
	Rates  map[string]float64 `json:"rates"`
}

type OpenERResponse struct {
	Disclaimer string             `json:"disclaimer"`
	License    string             `json:"license"`
	Timestamp  int64              `json:"timestamp"`
	Base       string             `json:"base"`
	Rates      map[string]float64 `json:"rates"`
}

type ExchangeRateCache struct {
	Rates     map[string]float64 `json:"rates"`
	Base      string             `json:"base"`
	Timestamp int64              `json:"timestamp"`
	TTL       int64              `json:"ttl"`
}

type CurrencyConfig struct {
	PrimaryAPI  string `json:"primary_api"`
	BackupAPI   string `json:"backup_api"`
	CacheTTL    int64  `json:"cache_ttl"`
	DefaultBase string `json:"default_base"`
}

func fetchFromFrankfurter(baseCurrency string) (*FrankfurterResponse, error) {
	url := fmt.Sprintf("https://api.frankfurter.dev/v1/latest?base=%s", baseCurrency)

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch from Frankfurter: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Frankfurter API returned status %d", resp.StatusCode)
	}

	var frankResp FrankfurterResponse
	err = json.NewDecoder(resp.Body).Decode(&frankResp)
	if err != nil {
		return nil, fmt.Errorf("failed to decode Frankfurter response: %v", err)
	}

	return &frankResp, nil
}

func fetchFromOpenER(baseCurrency string) (*OpenERResponse, error) {
	url := fmt.Sprintf("https://open.er-api.com/v6/latest/%s", baseCurrency)

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch from Open ER: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Open ER API returned status %d", resp.StatusCode)
	}

	var openResp OpenERResponse
	err = json.NewDecoder(resp.Body).Decode(&openResp)
	if err != nil {
		return nil, fmt.Errorf("failed to decode Open ER response: %v", err)
	}

	return &openResp, nil
}

func fetchExchangeRates(baseCurrency string) (map[string]float64, error) {
	baseCurrency, err := normalizeCurrency(baseCurrency)
	if err != nil {
		return nil, err // Add this check
	}

	frankResp, err := fetchFromFrankfurter(baseCurrency)
	if err == nil {
		return frankResp.Rates, nil
	}

	openResp, err := fetchFromOpenER(baseCurrency)
	if err == nil {
		return openResp.Rates, nil
	}

	return nil, fmt.Errorf("both APIs failed: %v", err)
}

func getCachePath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".budget", "exchange_cache.json")
}

func loadCache() (*ExchangeRateCache, error) {
	cachePath := getCachePath()

	if _, err := os.Stat(cachePath); os.IsNotExist(err) {
		return nil, nil
	}

	file, err := os.ReadFile(cachePath)
	if err != nil {
		return nil, err
	}

	var cache ExchangeRateCache
	err = json.Unmarshal(file, &cache)
	if err != nil {
		return nil, err
	}

	return &cache, nil
}

func saveCache(cache *ExchangeRateCache) error {
	cachePath := getCachePath()
	dir := filepath.Dir(cachePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	jsonData, err := json.MarshalIndent(cache, "", " ")
	if err != nil {
		return err
	}

	return os.WriteFile(cachePath, jsonData, 0644)
}

func isCacheValid(cache *ExchangeRateCache) bool {
	if cache == nil {
		return false
	}

	now := time.Now().Unix()
	return now-cache.Timestamp < cache.TTL
}

func getExchangeRates(baseCurrency string) (map[string]float64, error) {
	cache, err := loadCache()
	if err == nil && cache != nil && cache.Base == baseCurrency && isCacheValid(cache) {
		return cache.Rates, nil
	}

	rates, err := fetchExchangeRates(baseCurrency)
	if err != nil {
		return nil, err
	}

	newCache := &ExchangeRateCache{
		Rates:     rates,
		Base:      baseCurrency,
		Timestamp: time.Now().Unix(),
		TTL:       3600,
	}

	saveCache(newCache)

	return rates, nil
}

func normalizeCurrency(currency string) (string, error) {
	normalized := strings.ToUpper(strings.TrimSpace(currency))

	if len(normalized) != 3 {
		return "", fmt.Errorf("currency code must be exactly 3 characters, got '%s'", currency)
	}

	for _, char := range normalized {
		if char < 'A' || char > 'Z' {
			return "", fmt.Errorf("currency code must contain only letters, got '%s'", currency)
		}
	}

	return normalized, nil
}

func ConvertCurrency(amount float64, fromCurrency, toCurrency, baseCurrency string) (float64, error) {
	fromCurrency, err := normalizeCurrency(fromCurrency)
	if err != nil {
		return 0, err
	}
	toCurrency, err = normalizeCurrency(toCurrency)
	if err != nil {
		return 0, err
	}
	baseCurrency, err = normalizeCurrency(baseCurrency)
	if err != nil {
		return 0, err
	}

	if fromCurrency == toCurrency {
		return amount, nil
	}

	rates, err := getExchangeRates(baseCurrency)
	if err != nil {
		return 0, err
	}

	var baseAmount float64
	if fromCurrency == baseCurrency {
		baseAmount = amount
	} else {
		fromRate, exists := rates[fromCurrency]
		if !exists {
			return 0, fmt.Errorf("no exchange rate found for %s", fromCurrency)
		}
		baseAmount = amount / fromRate
	}

	if toCurrency == baseCurrency {
		return baseAmount, nil
	}

	toRate, exists := rates[toCurrency]
	if !exists {
		return 0, fmt.Errorf("no exchange rate found for %s", toCurrency)
	}

	return baseAmount * toRate, nil

}

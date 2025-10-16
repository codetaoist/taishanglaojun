// AI
// 
package currency

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"strings"
	"sync"
	"time"
)

// CurrencyManager 
// 
type CurrencyManager struct {
	baseCurrency    string
	exchangeRates   map[string]float64
	lastUpdate      time.Time
	mutex           sync.RWMutex
	rateProvider    ExchangeRateProvider
}

// ExchangeRateProvider 
// 
type ExchangeRateProvider interface {
	GetExchangeRates(ctx context.Context, baseCurrency string) (map[string]float64, error)
}

// CurrencyInfo 
type CurrencyInfo struct {
	Code         string `json:"code"`         //  (USD, EUR, CNY, JPY, GBP, KRW, SGD, HKD, AUD, CAD)
	Name         string `json:"name"`         // 
	Symbol       string `json:"symbol"`       //  ($, , , , , S$, HK$, A$, CAD$)
	DecimalPlaces int   `json:"decimal_places"` // 
	Country      string `json:"country"`      // /
	Region       string `json:"region"`       // 
}

// SupportedCurrencies 
// 
var SupportedCurrencies = map[string]CurrencyInfo{
	// 
	"USD": {
		Code:         "USD",
		Name:         "US Dollar",
		Symbol:       "$",
		DecimalPlaces: 2,
		Country:      "US",
		Region:       "North America",
	},
	"EUR": {
		Code:         "EUR",
		Name:         "Euro",
		Symbol:       "",
		DecimalPlaces: 2,
		Country:      "EU",
		Region:       "Europe",
	},
	"CNY": {
		Code:         "CNY",
		Name:         "Chinese Yuan",
		Symbol:       "",
		DecimalPlaces: 2,
		Country:      "CN",
		Region:       "Asia",
	},
	"JPY": {
		Code:         "JPY",
		Name:         "Japanese Yen",
		Symbol:       "",
		DecimalPlaces: 0,
		Country:      "JP",
		Region:       "Asia",
	},
	"GBP": {
		Code:         "GBP",
		Name:         "British Pound",
		Symbol:       "",
		DecimalPlaces: 2,
		Country:      "GB",
		Region:       "Europe",
	},
	"KRW": {
		Code:         "KRW",
		Name:         "South Korean Won",
		Symbol:       "",
		DecimalPlaces: 0,
		Country:      "KR",
		Region:       "Asia",
	},
	"SGD": {
		Code:         "SGD",
		Name:         "Singapore Dollar",
		Symbol:       "S$",
		DecimalPlaces: 2,
		Country:      "SG",
		Region:       "Asia",
	},
	"HKD": {
		Code:         "HKD",
		Name:         "Hong Kong Dollar",
		Symbol:       "HK$",
		DecimalPlaces: 2,
		Country:      "HK",
		Region:       "Asia",
	},
	"AUD": {
		Code:         "AUD",
		Name:         "Australian Dollar",
		Symbol:       "A$",
		DecimalPlaces: 2,
		Country:      "AU",
		Region:       "Oceania",
	},
	"CAD": {
		Code:         "CAD",
		Name:         "Canadian Dollar",
		Symbol:       "C$",
		DecimalPlaces: 2,
		Country:      "CA",
		Region:       "North America",
	},
	"CHF": {
		Code:         "CHF",
		Name:         "Swiss Franc",
		Symbol:       "CHF",
		DecimalPlaces: 2,
		Country:      "CH",
		Region:       "Europe",
	},
	"SEK": {
		Code:         "SEK",
		Name:         "Swedish Krona",
		Symbol:       "kr",
		DecimalPlaces: 2,
		Country:      "SE",
		Region:       "Europe",
	},
	"NOK": {
		Code:         "NOK",
		Name:         "Norwegian Krone",
		Symbol:       "kr",
		DecimalPlaces: 2,
		Country:      "NO",
		Region:       "Europe",
	},
	"DKK": {
		Code:         "DKK",
		Name:         "Danish Krone",
		Symbol:       "kr",
		DecimalPlaces: 2,
		Country:      "DK",
		Region:       "Europe",
	},
	"PLN": {
		Code:         "PLN",
		Name:         "Polish Zloty",
		Symbol:       "z",
		DecimalPlaces: 2,
		Country:      "PL",
		Region:       "Europe",
	},
	"CZK": {
		Code:         "CZK",
		Name:         "Czech Koruna",
		Symbol:       "K",
		DecimalPlaces: 2,
		Country:      "CZ",
		Region:       "Europe",
	},
	"HUF": {
		Code:         "HUF",
		Name:         "Hungarian Forint",
		Symbol:       "Ft",
		DecimalPlaces: 0,
		Country:      "HU",
		Region:       "Europe",
	},
	"RUB": {
		Code:         "RUB",
		Name:         "Russian Ruble",
		Symbol:       "",
		DecimalPlaces: 2,
		Country:      "RU",
		Region:       "Europe",
	},
	"INR": {
		Code:         "INR",
		Name:         "Indian Rupee",
		Symbol:       "",
		DecimalPlaces: 2,
		Country:      "IN",
		Region:       "Asia",
	},
	"THB": {
		Code:         "THB",
		Name:         "Thai Baht",
		Symbol:       "",
		DecimalPlaces: 2,
		Country:      "TH",
		Region:       "Asia",
	},
	"MYR": {
		Code:         "MYR",
		Name:         "Malaysian Ringgit",
		Symbol:       "RM",
		DecimalPlaces: 2,
		Country:      "MY",
		Region:       "Asia",
	},
	"IDR": {
		Code:         "IDR",
		Name:         "Indonesian Rupiah",
		Symbol:       "Rp",
		DecimalPlaces: 0,
		Country:      "ID",
		Region:       "Asia",
	},
	"PHP": {
		Code:         "PHP",
		Name:         "Philippine Peso",
		Symbol:       "",
		DecimalPlaces: 2,
		Country:      "PH",
		Region:       "Asia",
	},
	"VND": {
		Code:         "VND",
		Name:         "Vietnamese Dong",
		Symbol:       "",
		DecimalPlaces: 0,
		Country:      "VN",
		Region:       "Asia",
	},
	"TWD": {
		Code:         "TWD",
		Name:         "Taiwan Dollar",
		Symbol:       "NT$",
		DecimalPlaces: 0,
		Country:      "TW",
		Region:       "Asia",
	},
	"NZD": {
		Code:         "NZD",
		Name:         "New Zealand Dollar",
		Symbol:       "NZ$",
		DecimalPlaces: 2,
		Country:      "NZ",
		Region:       "Oceania",
	},
	"BRL": {
		Code:         "BRL",
		Name:         "Brazilian Real",
		Symbol:       "R$",
		DecimalPlaces: 2,
		Country:      "BR",
		Region:       "South America",
	},
	"MXN": {
		Code:         "MXN",
		Name:         "Mexican Peso",
		Symbol:       "$",
		DecimalPlaces: 2,
		Country:      "MX",
		Region:       "North America",
	},
	"ARS": {
		Code:         "ARS",
		Name:         "Argentine Peso",
		Symbol:       "$",
		DecimalPlaces: 2,
		Country:      "AR",
		Region:       "South America",
	},
	"CLP": {
		Code:         "CLP",
		Name:         "Chilean Peso",
		Symbol:       "$",
		DecimalPlaces: 0,
		Country:      "CL",
		Region:       "South America",
	},
	"ZAR": {
		Code:         "ZAR",
		Name:         "South African Rand",
		Symbol:       "R",
		DecimalPlaces: 2,
		Country:      "ZA",
		Region:       "Africa",
	},
	"EGP": {
		Code:         "EGP",
		Name:         "Egyptian Pound",
		Symbol:       "",
		DecimalPlaces: 2,
		Country:      "EG",
		Region:       "Africa",
	},
	"AED": {
		Code:         "AED",
		Name:         "UAE Dirham",
		Symbol:       ".",
		DecimalPlaces: 2,
		Country:      "AE",
		Region:       "Middle East",
	},
	"SAR": {
		Code:         "SAR",
		Name:         "Saudi Riyal",
		Symbol:       ".",
		DecimalPlaces: 2,
		Country:      "SA",
		Region:       "Middle East",
	},
	"ILS": {
		Code:         "ILS",
		Name:         "Israeli Shekel",
		Symbol:       "",
		DecimalPlaces: 2,
		Country:      "IL",
		Region:       "Middle East",
	},
	"TRY": {
		Code:         "TRY",
		Name:         "Turkish Lira",
		Symbol:       "",
		DecimalPlaces: 2,
		Country:      "TR",
		Region:       "Europe",
	},
}

// NewCurrencyManager 
func NewCurrencyManager(baseCurrency string, rateProvider ExchangeRateProvider) *CurrencyManager {
	return &CurrencyManager{
		baseCurrency:  baseCurrency,
		exchangeRates: make(map[string]float64),
		rateProvider:  rateProvider,
	}
}

// UpdateExchangeRates 
func (cm *CurrencyManager) UpdateExchangeRates(ctx context.Context) error {
	if cm.rateProvider == nil {
		return fmt.Errorf("exchange rate provider not configured")
	}

	rates, err := cm.rateProvider.GetExchangeRates(ctx, cm.baseCurrency)
	if err != nil {
		return fmt.Errorf("failed to get exchange rates: %w", err)
	}

	cm.mutex.Lock()
	cm.exchangeRates = rates
	cm.lastUpdate = time.Now()
	cm.mutex.Unlock()

	return nil
}

// ConvertCurrency 
func (cm *CurrencyManager) ConvertCurrency(amount float64, fromCurrency, toCurrency string) (float64, error) {
	if fromCurrency == toCurrency {
		return amount, nil
	}

	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	// 1
	if time.Since(cm.lastUpdate) > time.Hour {
		return 0, fmt.Errorf("exchange rates are outdated")
	}

	var fromRate, toRate float64 = 1.0, 1.0

	// 
	if fromCurrency != cm.baseCurrency {
		rate, exists := cm.exchangeRates[fromCurrency]
		if !exists {
			return 0, fmt.Errorf("exchange rate not found for currency: %s", fromCurrency)
		}
		fromRate = rate
	}

	// 
	if toCurrency != cm.baseCurrency {
		rate, exists := cm.exchangeRates[toCurrency]
		if !exists {
			return 0, fmt.Errorf("exchange rate not found for currency: %s", toCurrency)
		}
		toRate = rate
	}

	// amount * (1/fromRate) * toRate
	convertedAmount := amount * (toRate / fromRate)
	return convertedAmount, nil
}

// FormatCurrency 
func (cm *CurrencyManager) FormatCurrency(amount float64, currency, locale string) (string, error) {
	currencyInfo, exists := SupportedCurrencies[currency]
	if !exists {
		return "", fmt.Errorf("unsupported currency: %s", currency)
	}

	// 
	roundedAmount := math.Round(amount*math.Pow(10, float64(currencyInfo.DecimalPlaces))) / math.Pow(10, float64(currencyInfo.DecimalPlaces))

	switch locale {
	case "zh-CN":
		return cm.formatCurrencyZhCN(roundedAmount, currencyInfo), nil
	case "en-US":
		return cm.formatCurrencyEnUS(roundedAmount, currencyInfo), nil
	case "ja-JP":
		return cm.formatCurrencyJaJP(roundedAmount, currencyInfo), nil
	case "ko-KR":
		return cm.formatCurrencyKoKR(roundedAmount, currencyInfo), nil
	case "fr-FR":
		return cm.formatCurrencyFrFR(roundedAmount, currencyInfo), nil
	case "de-DE":
		return cm.formatCurrencyDeDE(roundedAmount, currencyInfo), nil
	case "es-ES":
		return cm.formatCurrencyEsES(roundedAmount, currencyInfo), nil
	case "it-IT":
		return cm.formatCurrencyItIT(roundedAmount, currencyInfo), nil
	case "pt-BR":
		return cm.formatCurrencyPtBR(roundedAmount, currencyInfo), nil
	case "ru-RU":
		return cm.formatCurrencyRuRU(roundedAmount, currencyInfo), nil
	case "ar-SA":
		return cm.formatCurrencyArSA(roundedAmount, currencyInfo), nil
	case "th-TH":
		return cm.formatCurrencyThTH(roundedAmount, currencyInfo), nil
	default:
		return cm.formatCurrencyDefault(roundedAmount, currencyInfo), nil
	}
}

// formatCurrencyZhCN 
func (cm *CurrencyManager) formatCurrencyZhCN(amount float64, info CurrencyInfo) string {
	amountStr := cm.formatNumber(amount, info.DecimalPlaces, ",", ".")
	
	switch info.Code {
	case "CNY":
		return fmt.Sprintf("%s", amountStr)
	case "USD":
		return fmt.Sprintf("$%s", amountStr)
	case "EUR":
		return fmt.Sprintf("%s", amountStr)
	case "JPY":
		return fmt.Sprintf("%s", amountStr)
	case "GBP":
		return fmt.Sprintf("%s", amountStr)
	case "HKD":
		return fmt.Sprintf("HK$%s", amountStr)
	default:
		return fmt.Sprintf("%s %s", info.Symbol, amountStr)
	}
}

// formatCurrencyEnUS 
func (cm *CurrencyManager) formatCurrencyEnUS(amount float64, info CurrencyInfo) string {
	amountStr := cm.formatNumber(amount, info.DecimalPlaces, ",", ".")
	
	switch info.Code {
	case "USD":
		return fmt.Sprintf("$%s", amountStr)
	case "EUR":
		return fmt.Sprintf("%s", amountStr)
	case "GBP":
		return fmt.Sprintf("%s", amountStr)
	case "JPY":
		return fmt.Sprintf("%s", amountStr)
	case "CNY":
		return fmt.Sprintf("%s", amountStr)
	default:
		return fmt.Sprintf("%s%s", info.Symbol, amountStr)
	}
}

// formatCurrencyJaJP 
func (cm *CurrencyManager) formatCurrencyJaJP(amount float64, info CurrencyInfo) string {
	amountStr := cm.formatNumber(amount, info.DecimalPlaces, ",", ".")
	
	switch info.Code {
	case "JPY":
		return fmt.Sprintf("%s", amountStr)
	case "USD":
		return fmt.Sprintf("$%s", amountStr)
	case "EUR":
		return fmt.Sprintf("%s", amountStr)
	case "CNY":
		return fmt.Sprintf("%s", amountStr)
	default:
		return fmt.Sprintf("%s%s", info.Symbol, amountStr)
	}
}

// formatCurrencyKoKR 
func (cm *CurrencyManager) formatCurrencyKoKR(amount float64, info CurrencyInfo) string {
	amountStr := cm.formatNumber(amount, info.DecimalPlaces, ",", ".")
	
	switch info.Code {
	case "KRW":
		return fmt.Sprintf("%s", amountStr)
	case "USD":
		return fmt.Sprintf("$%s", amountStr)
	case "EUR":
		return fmt.Sprintf("%s", amountStr)
	case "JPY":
		return fmt.Sprintf("%s", amountStr)
	case "CNY":
		return fmt.Sprintf("%s", amountStr)
	default:
		return fmt.Sprintf("%s%s", info.Symbol, amountStr)
	}
}

// formatCurrencyFrFR 
func (cm *CurrencyManager) formatCurrencyFrFR(amount float64, info CurrencyInfo) string {
	amountStr := cm.formatNumber(amount, info.DecimalPlaces, " ", ",")
	
	switch info.Code {
	case "EUR":
		return fmt.Sprintf("%s ", amountStr)
	case "USD":
		return fmt.Sprintf("%s $", amountStr)
	case "GBP":
		return fmt.Sprintf("%s ", amountStr)
	default:
		return fmt.Sprintf("%s %s", amountStr, info.Symbol)
	}
}

// formatCurrencyDeDE 
func (cm *CurrencyManager) formatCurrencyDeDE(amount float64, info CurrencyInfo) string {
	amountStr := cm.formatNumber(amount, info.DecimalPlaces, ".", ",")
	
	switch info.Code {
	case "EUR":
		return fmt.Sprintf("%s ", amountStr)
	case "USD":
		return fmt.Sprintf("%s $", amountStr)
	case "CHF":
		return fmt.Sprintf("CHF %s", amountStr)
	default:
		return fmt.Sprintf("%s %s", amountStr, info.Symbol)
	}
}

// formatCurrencyEsES 
func (cm *CurrencyManager) formatCurrencyEsES(amount float64, info CurrencyInfo) string {
	amountStr := cm.formatNumber(amount, info.DecimalPlaces, ".", ",")
	
	switch info.Code {
	case "EUR":
		return fmt.Sprintf("%s ", amountStr)
	case "USD":
		return fmt.Sprintf("%s $", amountStr)
	default:
		return fmt.Sprintf("%s %s", amountStr, info.Symbol)
	}
}

// formatCurrencyItIT 
func (cm *CurrencyManager) formatCurrencyItIT(amount float64, info CurrencyInfo) string {
	amountStr := cm.formatNumber(amount, info.DecimalPlaces, ".", ",")
	
	switch info.Code {
	case "EUR":
		return fmt.Sprintf("%s", amountStr)
	case "USD":
		return fmt.Sprintf("$ %s", amountStr)
	default:
		return fmt.Sprintf("%s %s", info.Symbol, amountStr)
	}
}

// formatCurrencyPtBR İ
func (cm *CurrencyManager) formatCurrencyPtBR(amount float64, info CurrencyInfo) string {
	amountStr := cm.formatNumber(amount, info.DecimalPlaces, ".", ",")
	
	switch info.Code {
	case "BRL":
		return fmt.Sprintf("R$ %s", amountStr)
	case "USD":
		return fmt.Sprintf("US$ %s", amountStr)
	case "EUR":
		return fmt.Sprintf("%s", amountStr)
	default:
		return fmt.Sprintf("%s %s", info.Symbol, amountStr)
	}
}

// formatCurrencyRuRU 
func (cm *CurrencyManager) formatCurrencyRuRU(amount float64, info CurrencyInfo) string {
	amountStr := cm.formatNumber(amount, info.DecimalPlaces, " ", ",")
	
	switch info.Code {
	case "RUB":
		return fmt.Sprintf("%s ", amountStr)
	case "USD":
		return fmt.Sprintf("%s $", amountStr)
	case "EUR":
		return fmt.Sprintf("%s ", amountStr)
	default:
		return fmt.Sprintf("%s %s", amountStr, info.Symbol)
	}
}

// formatCurrencyArSA 
func (cm *CurrencyManager) formatCurrencyArSA(amount float64, info CurrencyInfo) string {
	amountStr := cm.formatNumber(amount, info.DecimalPlaces, ",", ".")
	
	switch info.Code {
	case "SAR":
		return fmt.Sprintf("%s .", amountStr)
	case "AED":
		return fmt.Sprintf("%s .", amountStr)
	case "USD":
		return fmt.Sprintf("%s $", amountStr)
	case "EUR":
		return fmt.Sprintf("%s ", amountStr)
	default:
		return fmt.Sprintf("%s %s", amountStr, info.Symbol)
	}
}

// formatCurrencyThTH 
func (cm *CurrencyManager) formatCurrencyThTH(amount float64, info CurrencyInfo) string {
	amountStr := cm.formatNumber(amount, info.DecimalPlaces, ",", ".")
	
	switch info.Code {
	case "THB":
		return fmt.Sprintf("%s", amountStr)
	case "USD":
		return fmt.Sprintf("$%s", amountStr)
	case "EUR":
		return fmt.Sprintf("%s", amountStr)
	default:
		return fmt.Sprintf("%s%s", info.Symbol, amountStr)
	}
}

// formatCurrencyDefault 
func (cm *CurrencyManager) formatCurrencyDefault(amount float64, info CurrencyInfo) string {
	amountStr := cm.formatNumber(amount, info.DecimalPlaces, ",", ".")
	return fmt.Sprintf("%s%s", info.Symbol, amountStr)
}

// formatNumber 
func (cm *CurrencyManager) formatNumber(amount float64, decimalPlaces int, thousandSep, decimalSep string) string {
	// 
	format := fmt.Sprintf("%%.%df", decimalPlaces)
	amountStr := fmt.Sprintf(format, amount)
	
	// 
	parts := strings.Split(amountStr, ".")
	integerPart := parts[0]
	decimalPart := ""
	if len(parts) > 1 && decimalPlaces > 0 {
		decimalPart = parts[1]
	}
	
	// 
	if len(integerPart) > 3 {
		var result []string
		for i, digit := range reverse(integerPart) {
			if i > 0 && i%3 == 0 {
				result = append(result, thousandSep)
			}
			result = append(result, string(digit))
		}
		integerPart = reverse(strings.Join(result, ""))
	}
	
	// 
	if decimalPart != "" {
		return fmt.Sprintf("%s%s%s", integerPart, decimalSep, decimalPart)
	}
	return integerPart
}

// reverse 
func reverse(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

// GetCurrencyInfo 
func (cm *CurrencyManager) GetCurrencyInfo(currency string) (CurrencyInfo, error) {
	info, exists := SupportedCurrencies[currency]
	if !exists {
		return CurrencyInfo{}, fmt.Errorf("unsupported currency: %s", currency)
	}
	return info, nil
}

// GetCurrenciesByRegion 
func (cm *CurrencyManager) GetCurrenciesByRegion(region string) []CurrencyInfo {
	var currencies []CurrencyInfo
	for _, info := range SupportedCurrencies {
		if strings.EqualFold(info.Region, region) {
			currencies = append(currencies, info)
		}
	}
	return currencies
}

// GetCurrenciesByCountry 
func (cm *CurrencyManager) GetCurrenciesByCountry(country string) []CurrencyInfo {
	var currencies []CurrencyInfo
	for _, info := range SupportedCurrencies {
		if strings.EqualFold(info.Country, country) {
			currencies = append(currencies, info)
		}
	}
	return currencies
}

// GetAllCurrencies 
func (cm *CurrencyManager) GetAllCurrencies() []CurrencyInfo {
	var currencies []CurrencyInfo
	for _, info := range SupportedCurrencies {
		currencies = append(currencies, info)
	}
	return currencies
}

// IsCurrencySupported 
func (cm *CurrencyManager) IsCurrencySupported(currency string) bool {
	_, exists := SupportedCurrencies[currency]
	return exists
}

// GetExchangeRate 
func (cm *CurrencyManager) GetExchangeRate(fromCurrency, toCurrency string) (float64, error) {
	if fromCurrency == toCurrency {
		return 1.0, nil
	}

	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	// 
	if time.Since(cm.lastUpdate) > time.Hour {
		return 0, fmt.Errorf("exchange rates are outdated")
	}

	var fromRate, toRate float64 = 1.0, 1.0

	if fromCurrency != cm.baseCurrency {
		rate, exists := cm.exchangeRates[fromCurrency]
		if !exists {
			return 0, fmt.Errorf("exchange rate not found for currency: %s", fromCurrency)
		}
		fromRate = rate
	}

	if toCurrency != cm.baseCurrency {
		rate, exists := cm.exchangeRates[toCurrency]
		if !exists {
			return 0, fmt.Errorf("exchange rate not found for currency: %s", toCurrency)
		}
		toRate = rate
	}

	return toRate / fromRate, nil
}

// GetLastUpdateTime 
func (cm *CurrencyManager) GetLastUpdateTime() time.Time {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()
	return cm.lastUpdate
}

// ParseCurrency 
func (cm *CurrencyManager) ParseCurrency(currencyStr, locale string) (float64, string, error) {
	// 
	currencyStr = strings.TrimSpace(currencyStr)
	
	// 
	for code, info := range SupportedCurrencies {
		if strings.HasPrefix(currencyStr, info.Symbol) {
			amountStr := strings.TrimSpace(strings.TrimPrefix(currencyStr, info.Symbol))
			amount, err := cm.parseAmount(amountStr, locale)
			if err != nil {
				return 0, "", err
			}
			return amount, code, nil
		}
		if strings.HasSuffix(currencyStr, info.Symbol) {
			amountStr := strings.TrimSpace(strings.TrimSuffix(currencyStr, info.Symbol))
			amount, err := cm.parseAmount(amountStr, locale)
			if err != nil {
				return 0, "", err
			}
			return amount, code, nil
		}
	}
	
	return 0, "", fmt.Errorf("unable to parse currency: %s", currencyStr)
}

// parseAmount 
func (cm *CurrencyManager) parseAmount(amountStr, locale string) (float64, error) {
	// locale
	switch locale {
	case "fr-FR", "de-DE", "es-ES", "it-IT":
		// 
		amountStr = strings.ReplaceAll(amountStr, " ", "")
		if strings.Contains(amountStr, ",") && strings.Contains(amountStr, ".") {
			// 
			lastComma := strings.LastIndex(amountStr, ",")
			lastDot := strings.LastIndex(amountStr, ".")
			if lastComma > lastDot {
				// 
				amountStr = strings.ReplaceAll(amountStr[:lastComma], ".", "")
				amountStr = strings.Replace(amountStr, ",", ".", 1)
			} else {
				// 
				amountStr = strings.ReplaceAll(amountStr[:lastDot], ",", "")
			}
		} else if strings.Contains(amountStr, ",") {
			// 
			commaCount := strings.Count(amountStr, ",")
			if commaCount == 1 {
				parts := strings.Split(amountStr, ",")
				if len(parts[1]) <= 3 {
					// 
					amountStr = strings.Replace(amountStr, ",", ".", 1)
				} else {
					// 
					amountStr = strings.ReplaceAll(amountStr, ",", "")
				}
			} else {
				// 
				amountStr = strings.ReplaceAll(amountStr, ",", "")
			}
		}
	default:
		// 
		amountStr = strings.ReplaceAll(amountStr, ",", "")
	}
	
	return strconv.ParseFloat(amountStr, 64)
}

// MockExchangeRateProvider 
type MockExchangeRateProvider struct{}

// GetExchangeRates 
func (m *MockExchangeRateProvider) GetExchangeRates(ctx context.Context, baseCurrency string) (map[string]float64, error) {
	// USD
	rates := map[string]float64{
		"EUR": 0.85,
		"CNY": 7.20,
		"JPY": 110.0,
		"GBP": 0.73,
		"KRW": 1180.0,
		"SGD": 1.35,
		"HKD": 7.80,
		"AUD": 1.40,
		"CAD": 1.25,
		"CHF": 0.92,
		"SEK": 8.50,
		"NOK": 8.80,
		"DKK": 6.35,
		"PLN": 3.90,
		"CZK": 21.5,
		"HUF": 295.0,
		"RUB": 75.0,
		"INR": 74.0,
		"THB": 31.0,
		"MYR": 4.15,
		"IDR": 14250.0,
		"PHP": 50.0,
		"VND": 23000.0,
		"TWD": 28.0,
		"NZD": 1.45,
		"BRL": 5.20,
		"MXN": 20.0,
		"ARS": 98.0,
		"CLP": 800.0,
		"ZAR": 14.5,
		"EGP": 15.7,
		"AED": 3.67,
		"SAR": 3.75,
		"ILS": 3.25,
		"TRY": 8.50,
	}
	
	if baseCurrency != "USD" {
		// USD
	baseRate, exists := rates[baseCurrency]
		if !exists {
			return nil, fmt.Errorf("unsupported base currency: %s", baseCurrency)
		}
		
		convertedRates := make(map[string]float64)
		convertedRates["USD"] = 1.0 / baseRate
		
		for currency, rate := range rates {
			if currency != baseCurrency {
				convertedRates[currency] = rate / baseRate
			}
		}
		
		return convertedRates, nil
	}
	
	return rates, nil
}


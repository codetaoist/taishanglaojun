// 本地化服务
package localization

import (
	"context"
	"fmt"
	"time"

	"github.com/codetaoist/taishanglaojun/core-services/localization/culture"
	"github.com/codetaoist/taishanglaojun/core-services/localization/currency"
	"github.com/codetaoist/taishanglaojun/core-services/localization/i18n"
	"github.com/codetaoist/taishanglaojun/core-services/localization/timezone"
)

// LocalizationService 本地化服务
type LocalizationService struct {
	i18nManager      *i18n.I18nManager
	timezoneManager  *timezone.TimezoneManager
	currencyManager  *currency.CurrencyManager
	cultureManager   *culture.CultureManager
	defaultLocale    string
	defaultTimezone  string
	defaultCurrency  string
	defaultCulture   string
}

// LocalizationConfig 本地化配置
type LocalizationConfig struct {
	DefaultLocale    string `json:"default_locale"`
	DefaultTimezone  string `json:"default_timezone"`
	DefaultCurrency  string `json:"default_currency"`
	DefaultCulture   string `json:"default_culture"`
	TranslationsPath string `json:"translations_path"`
	ExchangeRateAPI  string `json:"exchange_rate_api"`
	CacheEnabled     bool   `json:"cache_enabled"`
	CacheTTL         int    `json:"cache_ttl"`
}

// UserLocalizationContext 用户本地化上下文
type UserLocalizationContext struct {
	UserID       string    `json:"user_id"`
	Locale       string    `json:"locale"`
	Timezone     string    `json:"timezone"`
	Currency     string    `json:"currency"`
	Culture      string    `json:"culture"`
	Region       string    `json:"region"`
	Country      string    `json:"country"`
	Language     string    `json:"language"`
	DateFormat   string    `json:"date_format"`
	TimeFormat   string    `json:"time_format"`
	NumberFormat string    `json:"number_format"`
	RTL          bool      `json:"rtl"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// LocalizedContent 本地化内容
type LocalizedContent struct {
	Text         string                 `json:"text"`
	HTML         string                 `json:"html"`
	Metadata     map[string]interface{} `json:"metadata"`
	Locale       string                 `json:"locale"`
	Culture      string                 `json:"culture"`
	Direction    string                 `json:"direction"` // ltr, rtl
	GeneratedAt  time.Time              `json:"generated_at"`
}

// LocalizedDateTime 本地化日期时间
type LocalizedDateTime struct {
	Original     time.Time `json:"original"`
	Localized    time.Time `json:"localized"`
	Formatted    string    `json:"formatted"`
	Timezone     string    `json:"timezone"`
	Culture      string    `json:"culture"`
	Format       string    `json:"format"`
}

// LocalizedNumber 本地化数字
type LocalizedNumber struct {
	Original  float64 `json:"original"`
	Formatted string  `json:"formatted"`
	Culture   string  `json:"culture"`
	Type      string  `json:"type"` // number, currency, percentage
}

// LocalizedAddress 
type LocalizedAddress struct {
	Original  map[string]string `json:"original"`
	Formatted string            `json:"formatted"`
	Culture   string            `json:"culture"`
	Country   string            `json:"country"`
}

// NewLocalizationService 创建本地化服务
func NewLocalizationService(config LocalizationConfig) (*LocalizationService, error) {
	// 
	i18nMgr, err := i18n.NewI18nManager(config.TranslationsPath, config.DefaultLocale)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize i18n manager: %w", err)
	}

	timezoneMgr := timezone.NewTimezoneManager(config.DefaultTimezone)
	currencyMgr := currency.NewCurrencyManager(config.DefaultCurrency)
	cultureMgr := culture.NewCultureManager(config.DefaultCulture)

	// 
	if config.ExchangeRateAPI != "" {
		if err := currencyMgr.UpdateExchangeRates(); err != nil {
			// 浫
			fmt.Printf("Warning: failed to update exchange rates: %v\n", err)
		}
	}

	return &LocalizationService{
		i18nManager:     i18nMgr,
		timezoneManager: timezoneMgr,
		currencyManager: currencyMgr,
		cultureManager:  cultureMgr,
		defaultLocale:   config.DefaultLocale,
		defaultTimezone: config.DefaultTimezone,
		defaultCurrency: config.DefaultCurrency,
		defaultCulture:  config.DefaultCulture,
	}, nil
}

// GetUserContext 
func (ls *LocalizationService) GetUserContext(ctx context.Context, userID string) (*UserLocalizationContext, error) {
	// 
	// 
	return &UserLocalizationContext{
		UserID:       userID,
		Locale:       ls.defaultLocale,
		Timezone:     ls.defaultTimezone,
		Currency:     ls.defaultCurrency,
		Culture:      ls.defaultCulture,
		Region:       "Asia",
		Country:      "CN",
		Language:     "zh",
		DateFormat:   "2006-01-02",
		TimeFormat:   "15:04",
		NumberFormat: "decimal",
		RTL:          false,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}, nil
}

// UpdateUserContext 
func (ls *LocalizationService) UpdateUserContext(ctx context.Context, userContext *UserLocalizationContext) error {
	// 
	if !ls.i18nManager.IsLocaleSupported(userContext.Locale) {
		return fmt.Errorf("unsupported locale: %s", userContext.Locale)
	}

	if !ls.timezoneManager.IsTimezoneSupported(userContext.Timezone) {
		return fmt.Errorf("unsupported timezone: %s", userContext.Timezone)
	}

	if !ls.currencyManager.IsCurrencySupported(userContext.Currency) {
		return fmt.Errorf("unsupported currency: %s", userContext.Currency)
	}

	if !ls.cultureManager.IsCultureSupported(userContext.Culture) {
		return fmt.Errorf("unsupported culture: %s", userContext.Culture)
	}

	// ?	userContext.UpdatedAt = time.Now()

	// 浽
	// ?	return nil
}

// LocalizeText 本地化文本
func (ls *LocalizationService) LocalizeText(ctx context.Context, key string, params map[string]interface{}, userContext *UserLocalizationContext) (*LocalizedContent, error) {
	// 
	text, err := ls.i18nManager.Translate(key, userContext.Locale, params)
	if err != nil {
		return nil, fmt.Errorf("failed to translate text: %w", err)
	}

	// 
	cultureInfo, err := ls.cultureManager.GetCultureInfo(userContext.Culture)
	if err != nil {
		return nil, fmt.Errorf("failed to get culture info: %w", err)
	}

	direction := "ltr"
	if cultureInfo.RTL {
		direction = "rtl"
	}

	return &LocalizedContent{
		Text:        text,
		HTML:        fmt.Sprintf(`<span dir="%s" lang="%s">%s</span>`, direction, userContext.Language, text),
		Metadata:    map[string]interface{}{"key": key, "params": params},
		Locale:      userContext.Locale,
		Culture:     userContext.Culture,
		Direction:   direction,
		GeneratedAt: time.Now(),
	}, nil
}

// LocalizeDateTime 本地化日期时间
func (ls *LocalizationService) LocalizeDateTime(ctx context.Context, dt time.Time, format string, userContext *UserLocalizationContext) (*LocalizedDateTime, error) {
	// 
	localTime, err := ls.timezoneManager.ConvertTimezone(dt, "UTC", userContext.Timezone)
	if err != nil {
		return nil, fmt.Errorf("failed to convert timezone: %w", err)
	}

	// ?	formatted, err := ls.timezoneManager.FormatTime(localTime, format, userContext.Locale)
	if err != nil {
		return nil, fmt.Errorf("failed to format time: %w", err)
	}

	return &LocalizedDateTime{
		Original:  dt,
		Localized: localTime,
		Formatted: formatted,
		Timezone:  userContext.Timezone,
		Culture:   userContext.Culture,
		Format:    format,
	}, nil
}

// LocalizeNumber 本地化数字
func (ls *LocalizationService) LocalizeNumber(ctx context.Context, number float64, numberType string, userContext *UserLocalizationContext) (*LocalizedNumber, error) {
	var formatted string
	var err error

	switch numberType {
	case "currency":
		formatted, err = ls.currencyManager.FormatCurrency(number, userContext.Currency, userContext.Locale)
	case "percentage":
		formatted = fmt.Sprintf("%.2f%%", number*100)
	default:
		// 
		cultureInfo, err := ls.cultureManager.GetCultureInfo(userContext.Culture)
		if err != nil {
			return nil, fmt.Errorf("failed to get culture info: %w", err)
		}
		
		// ?		if cultureInfo.NumberFormat.ThousandSeparator == "," {
			formatted = fmt.Sprintf("%.2f", number)
			// 
		} else {
			formatted = fmt.Sprintf("%.2f", number)
		}
	}

	if err != nil {
		return nil, fmt.Errorf("failed to format number: %w", err)
	}

	return &LocalizedNumber{
		Original:  number,
		Formatted: formatted,
		Culture:   userContext.Culture,
		Type:      numberType,
	}, nil
}

// LocalizeCurrency 本地化货币
func (ls *LocalizationService) LocalizeCurrency(ctx context.Context, amount float64, fromCurrency, toCurrency string, userContext *UserLocalizationContext) (*LocalizedNumber, error) {
	// 
	convertedAmount, err := ls.currencyManager.ConvertCurrency(amount, fromCurrency, toCurrency)
	if err != nil {
		return nil, fmt.Errorf("failed to convert currency: %w", err)
	}

	// ?	formatted, err := ls.currencyManager.FormatCurrency(convertedAmount, toCurrency, userContext.Locale)
	if err != nil {
		return nil, fmt.Errorf("failed to format currency: %w", err)
	}

	return &LocalizedNumber{
		Original:  convertedAmount,
		Formatted: formatted,
		Culture:   userContext.Culture,
		Type:      "currency",
	}, nil
}

// LocalizeAddress 
func (ls *LocalizationService) LocalizeAddress(ctx context.Context, addressData map[string]string, userContext *UserLocalizationContext) (*LocalizedAddress, error) {
	// 
	formatted, err := ls.cultureManager.FormatAddress(addressData, userContext.Culture)
	if err != nil {
		return nil, fmt.Errorf("failed to format address: %w", err)
	}

	return &LocalizedAddress{
		Original:  addressData,
		Formatted: formatted,
		Culture:   userContext.Culture,
		Country:   userContext.Country,
	}, nil
}

// LocalizeName 本地化姓名
func (ls *LocalizationService) LocalizeName(ctx context.Context, firstName, lastName, middleName, honorific string, userContext *UserLocalizationContext) (string, error) {
	// 
	name, err := ls.cultureManager.FormatName(firstName, lastName, middleName, honorific, userContext.Culture)
	if err != nil {
		return "", fmt.Errorf("failed to format name: %w", err)
	}

	return name, nil
}

// LocalizePhoneNumber 本地化电话号码
func (ls *LocalizationService) LocalizePhoneNumber(ctx context.Context, phoneNumber string, userContext *UserLocalizationContext) (string, error) {
	return ls.cultureManager.FormatPhoneNumber(phoneNumber, userContext.Culture)
}

// GetBusinessHours 本地化营业时间
func (ls *LocalizationService) GetBusinessHours(ctx context.Context, userContext *UserLocalizationContext) (string, error) {
	return ls.cultureManager.GetBusinessHours(userContext.Culture)
}

// IsWorkingDay 本地化是否为工作日
func (ls *LocalizationService) IsWorkingDay(ctx context.Context, date time.Time, userContext *UserLocalizationContext) (bool, error) {
	// 
	isWorkingDay, err := ls.cultureManager.IsWorkingDay(date, userContext.Culture)
	if err != nil {
		return false, fmt.Errorf("failed to check working day: %w", err)
	}

	return isWorkingDay, nil
}

// GetHolidays 本地化节假日
func (ls *LocalizationService) GetHolidays(ctx context.Context, userContext *UserLocalizationContext) ([]culture.Holiday, error) {
	return ls.cultureManager.GetHolidays(userContext.Culture)
}

// IsHoliday 本地化是否为节假日
func (ls *LocalizationService) IsHoliday(ctx context.Context, date time.Time, userContext *UserLocalizationContext) (bool, culture.Holiday, error) {
	isHoliday, holiday, err := ls.cultureManager.IsHoliday(date, userContext.Culture)
	if err != nil {
		return false, culture.Holiday{}, fmt.Errorf("failed to check holiday: %w", err)
	}

	return isHoliday, holiday, nil
}

// GetColorMeaning 本地化颜色含义
func (ls *LocalizationService) GetColorMeaning(ctx context.Context, color string, userContext *UserLocalizationContext) (string, error) {
	return ls.cultureManager.GetColorMeaning(color, userContext.Culture)
}

// IsTabooTopic 本地化是否为禁忌话题
func (ls *LocalizationService) IsTabooTopic(ctx context.Context, topic string, userContext *UserLocalizationContext) (bool, error) {
	return ls.cultureManager.IsTabooTopic(topic, userContext.Culture)
}

// GetFoodRestrictions 本地化食物限制
func (ls *LocalizationService) GetFoodRestrictions(ctx context.Context, userContext *UserLocalizationContext) ([]string, error) {
	return ls.cultureManager.GetFoodRestrictions(userContext.Culture)
}

// DetectUserLocalization 检测用户本地化设置	
func (ls *LocalizationService) DetectUserLocalization(ctx context.Context, acceptLanguage, userAgent, ipAddress string) (*UserLocalizationContext, error) {
	// 
	locale := ls.i18nManager.DetectLanguage(acceptLanguage)
	if locale == "" {
		locale = ls.defaultLocale
	}

	// IP
	timezone := ls.defaultTimezone

	// 
	currency := ls.defaultCurrency

	// ?	culture := ls.defaultCulture

	return &UserLocalizationContext{
		Locale:       locale,
		Timezone:     timezone,
		Currency:     currency,
		Culture:      culture,
		Region:       "Asia",
		Country:      "CN",
		Language:     locale[:2],
		DateFormat:   "2006-01-02",
		TimeFormat:   "15:04",
		NumberFormat: "decimal",
		RTL:          false,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}, nil
}

// GetSupportedLocales 
func (ls *LocalizationService) GetSupportedLocales() []string {
	return ls.i18nManager.GetSupportedLocales()
}

// GetSupportedTimezones ?func (ls *LocalizationService) GetSupportedTimezones() []string {
	return ls.timezoneManager.GetSupportedTimezones()
}

// GetSupportedCurrencies ?func (ls *LocalizationService) GetSupportedCurrencies() []currency.CurrencyInfo {
	return ls.currencyManager.GetSupportedCurrencies()
}

// GetSupportedCultures ?func (ls *LocalizationService) GetSupportedCultures() []culture.CultureInfo {
	return ls.cultureManager.GetAllCultures()
}

// ValidateLocalizationSettings ?func (ls *LocalizationService) ValidateLocalizationSettings(locale, timezone, currency, culture string) error {
	if !ls.i18nManager.IsLocaleSupported(locale) {
		return fmt.Errorf("unsupported locale: %s", locale)
	}

	if !ls.timezoneManager.IsTimezoneSupported(timezone) {
		return fmt.Errorf("unsupported timezone: %s", timezone)
	}

	if !ls.currencyManager.IsCurrencySupported(currency) {
		return fmt.Errorf("unsupported currency: %s", currency)
	}

	if !ls.cultureManager.IsCultureSupported(culture) {
		return fmt.Errorf("unsupported culture: %s", culture)
	}

	return nil
}

// RefreshExchangeRates 
func (ls *LocalizationService) RefreshExchangeRates(ctx context.Context) error {
	return ls.currencyManager.UpdateExchangeRates()
}

// ReloadTranslations 
func (ls *LocalizationService) ReloadTranslations(ctx context.Context) error {
	return ls.i18nManager.ReloadTranslations()
}

// GetLocalizationStats ?func (ls *LocalizationService) GetLocalizationStats(ctx context.Context) map[string]interface{} {
	return map[string]interface{}{
		"supported_locales":    len(ls.GetSupportedLocales()),
		"supported_timezones":  len(ls.GetSupportedTimezones()),
		"supported_currencies": len(ls.GetSupportedCurrencies()),
		"supported_cultures":   len(ls.GetSupportedCultures()),
		"default_locale":       ls.defaultLocale,
		"default_timezone":     ls.defaultTimezone,
		"default_currency":     ls.defaultCurrency,
		"default_culture":      ls.defaultCulture,
	}
}


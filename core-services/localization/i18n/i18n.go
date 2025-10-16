// 国际化包
package i18n

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"golang.org/x/text/message/catalog"
)

// SupportedLanguages 
var SupportedLanguages = []string{
	"zh-CN", // ?	"zh-TW", // 
	"en-US", // 
	"en-GB", // 
	"ja-JP", // 
	"ko-KR", // 
	"fr-FR", // 
	"de-DE", // 
	"es-ES", // 
	"it-IT", // 
	"pt-BR", // 
	"ru-RU", // 
	"ar-SA", // 
	"hi-IN", // ?	"th-TH", // 
	"vi-VN", // ?	"id-ID", // ?	"ms-MY", // ?}

// I18nManager 
type I18nManager struct {
	bundle      *i18n.Bundle
	localizers  map[string]*i18n.Localizer
	catalog     catalog.Builder
	printers    map[string]*message.Printer
	mutex       sync.RWMutex
	fallbackLang string
	loadPath    string
}

// TranslationMessage 
type TranslationMessage struct {
	ID          string                 `json:"id"`
	Description string                 `json:"description,omitempty"`
	Hash        string                 `json:"hash,omitempty"`
	LeftDelim   string                 `json:"leftDelim,omitempty"`
	RightDelim  string                 `json:"rightDelim,omitempty"`
	Zero        string                 `json:"zero,omitempty"`
	One         string                 `json:"one,omitempty"`
	Two         string                 `json:"two,omitempty"`
	Few         string                 `json:"few,omitempty"`
	Many        string                 `json:"many,omitempty"`
	Other       string                 `json:"other,omitempty"`
	Translation string                 `json:"translation,omitempty"`
	Vars        map[string]interface{} `json:"vars,omitempty"`
}

// LocalizationConfig 本地化配置
type LocalizationConfig struct {
	DefaultLanguage string   `json:"default_language"`
	FallbackLanguage string  `json:"fallback_language"`
	SupportedLanguages []string `json:"supported_languages"`
	LoadPath        string   `json:"load_path"`
	AutoReload      bool     `json:"auto_reload"`
	CacheEnabled    bool     `json:"cache_enabled"`
	CacheTTL        time.Duration `json:"cache_ttl"`
}

// NewI18nManager 
func NewI18nManager(config LocalizationConfig) (*I18nManager, error) {
	bundle := i18n.NewBundle(language.English)
	bundle.RegisterUnmarshalFunc("json", json.Unmarshal)

	manager := &I18nManager{
		bundle:      bundle,
		localizers:  make(map[string]*i18n.Localizer),
		printers:    make(map[string]*message.Printer),
		fallbackLang: config.FallbackLanguage,
		loadPath:    config.LoadPath,
	}

	// catalog builder
	manager.catalog = catalog.NewBuilder()

	// 
	if err := manager.loadTranslations(); err != nil {
		return nil, fmt.Errorf("failed to load translations: %w", err)
	}

	// localizersprinters
	manager.initializeLocalizers()

	return manager, nil
}

// loadTranslations 
func (m *I18nManager) loadTranslations() error {
	for _, lang := range SupportedLanguages {
		filePath := filepath.Join(m.loadPath, fmt.Sprintf("%s.json", lang))
		
		// ?		if _, err := ioutil.ReadFile(filePath); err != nil {
			// 
			if err := m.createDefaultTranslationFile(lang, filePath); err != nil {
				return fmt.Errorf("failed to create default translation file for %s: %w", lang, err)
			}
		}

		// 
		if _, err := m.bundle.LoadMessageFile(filePath); err != nil {
			return fmt.Errorf("failed to load translation file %s: %w", filePath, err)
		}
	}

	return nil
}

// createDefaultTranslationFile 
func (m *I18nManager) createDefaultTranslationFile(lang, filePath string) error {
	defaultMessages := m.getDefaultMessages(lang)
	
	data, err := json.MarshalIndent(defaultMessages, "", "  ")
	if err != nil {
		return err
	}

	return ioutil.WriteFile(filePath, data, 0644)
}

// getDefaultMessages 
func (m *I18nManager) getDefaultMessages(lang string) map[string]TranslationMessage {
	messages := make(map[string]TranslationMessage)

	// 
	switch lang {
	case "zh-CN":
		messages = map[string]TranslationMessage{
			"welcome": {
				ID:          "welcome",
				Description: "",
				Translation: "AI",
			},
			"login": {
				ID:          "login",
				Description: "",
				Translation: "",
			},
			"logout": {
				ID:          "logout",
				Description: "?,
				Translation: "?,
			},
			"error.not_found": {
				ID:          "error.not_found",
				Description: "?,
				Translation: "",
			},
			"error.internal_server": {
				ID:          "error.internal_server",
				Description: "?,
				Translation: "?,
			},
		}
	case "en-US":
		messages = map[string]TranslationMessage{
			"welcome": {
				ID:          "welcome",
				Description: "Welcome message",
				Translation: "Welcome to Taishang Laojun AI Platform",
			},
			"login": {
				ID:          "login",
				Description: "Login button",
				Translation: "Login",
			},
			"logout": {
				ID:          "logout",
				Description: "Logout button",
				Translation: "Logout",
			},
			"error.not_found": {
				ID:          "error.not_found",
				Description: "Not found error",
				Translation: "The requested resource was not found",
			},
			"error.internal_server": {
				ID:          "error.internal_server",
				Description: "Internal server error",
				Translation: "Internal server error, please try again later",
			},
		}
	case "ja-JP":
		messages = map[string]TranslationMessage{
			"welcome": {
				ID:          "welcome",
				Description: "?,
				Translation: "AI?,
			},
			"login": {
				ID:          "login",
				Description: "?,
				Translation: "",
			},
			"logout": {
				ID:          "logout",
				Description: "",
				Translation: "?,
			},
			"error.not_found": {
				ID:          "error.not_found",
				Description: "?,
				Translation: "?,
			},
			"error.internal_server": {
				ID:          "error.internal_server",
				Description: "?,
				Translation: "",
			},
		}
	default:
		// 
		return m.getDefaultMessages("en-US")
	}

	return messages
}

// initializeLocalizers localizers
func (m *I18nManager) initializeLocalizers() {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	for _, lang := range SupportedLanguages {
		// localizer
		localizer := i18n.NewLocalizer(m.bundle, lang, m.fallbackLang)
		m.localizers[lang] = localizer

		// message printer
		tag, err := language.Parse(lang)
		if err != nil {
			tag = language.English
		}
		printer := message.NewPrinter(tag)
		m.printers[lang] = printer
	}
}

// Translate 
func (m *I18nManager) Translate(ctx context.Context, lang, messageID string, templateData map[string]interface{}) (string, error) {
	m.mutex.RLock()
	localizer, exists := m.localizers[lang]
	m.mutex.RUnlock()

	if !exists {
		// fallback
		m.mutex.RLock()
		localizer = m.localizers[m.fallbackLang]
		m.mutex.RUnlock()
	}

	if localizer == nil {
		return messageID, fmt.Errorf("no localizer available for language %s", lang)
	}

	// ?	localizeConfig := &i18n.LocalizeConfig{
		MessageID:    messageID,
		TemplateData: templateData,
	}

	// 
	translation, err := localizer.Localize(localizeConfig)
	if err != nil {
		// messageIDfallback
		return messageID, err
	}

	return translation, nil
}

// TranslateWithCount 
func (m *I18nManager) TranslateWithCount(ctx context.Context, lang, messageID string, count int, templateData map[string]interface{}) (string, error) {
	m.mutex.RLock()
	localizer, exists := m.localizers[lang]
	m.mutex.RUnlock()

	if !exists {
		m.mutex.RLock()
		localizer = m.localizers[m.fallbackLang]
		m.mutex.RUnlock()
	}

	if localizer == nil {
		return messageID, fmt.Errorf("no localizer available for language %s", lang)
	}

	if templateData == nil {
		templateData = make(map[string]interface{})
	}
	templateData["Count"] = count

	localizeConfig := &i18n.LocalizeConfig{
		MessageID:    messageID,
		PluralCount:  count,
		TemplateData: templateData,
	}

	translation, err := localizer.Localize(localizeConfig)
	if err != nil {
		return messageID, err
	}

	return translation, nil
}

// FormatNumber 格式化数字
func (m *I18nManager) FormatNumber(lang string, number interface{}) string {
	m.mutex.RLock()
	printer, exists := m.printers[lang]
	m.mutex.RUnlock()

	if !exists {
		m.mutex.RLock()
		printer = m.printers[m.fallbackLang]
		m.mutex.RUnlock()
	}

	if printer == nil {
		return fmt.Sprintf("%v", number)
	}

	switch v := number.(type) {
	case int:
		return printer.Sprintf("%d", v)
	case int64:
		return printer.Sprintf("%d", v)
	case float64:
		return printer.Sprintf("%.2f", v)
	case float32:
		return printer.Sprintf("%.2f", v)
	default:
		return fmt.Sprintf("%v", v)
	}
}

// FormatCurrency 格式化货币
func (m *I18nManager) FormatCurrency(lang string, amount float64, currency string) string {
	m.mutex.RLock()
	printer, exists := m.printers[lang]
	m.mutex.RUnlock()

	if !exists {
		m.mutex.RLock()
		printer = m.printers[m.fallbackLang]
		m.mutex.RUnlock()
	}

	if printer == nil {
		return fmt.Sprintf("%.2f %s", amount, currency)
	}

	// 
	switch lang {
	case "zh-CN":
		switch currency {
		case "CNY", "RMB":
			return printer.Sprintf("%.2f", amount)
		case "USD":
			return printer.Sprintf("$%.2f", amount)
		case "EUR":
			return printer.Sprintf("?.2f", amount)
		default:
			return printer.Sprintf("%.2f %s", amount, currency)
		}
	case "en-US":
		switch currency {
		case "USD":
			return printer.Sprintf("$%.2f", amount)
		case "EUR":
			return printer.Sprintf("?.2f", amount)
		case "GBP":
			return printer.Sprintf("%.2f", amount)
		case "JPY":
			return printer.Sprintf("%.0f", amount)
		default:
			return printer.Sprintf("%.2f %s", amount, currency)
		}
	case "ja-JP":
		switch currency {
		case "JPY":
			return printer.Sprintf("%.0f", amount)
		case "USD":
			return printer.Sprintf("$%.2f", amount)
		default:
			return printer.Sprintf("%.2f %s", amount, currency)
		}
	default:
		return printer.Sprintf("%.2f %s", amount, currency)
	}
}

// GetSupportedLanguages 
func (m *I18nManager) GetSupportedLanguages() []string {
	return SupportedLanguages
}

// IsLanguageSupported 
func (m *I18nManager) IsLanguageSupported(lang string) bool {
	for _, supported := range SupportedLanguages {
		if supported == lang {
			return true
		}
	}
	return false
}

// DetectLanguageFromAcceptLanguage Accept-Language
func (m *I18nManager) DetectLanguageFromAcceptLanguage(acceptLanguage string) string {
	if acceptLanguage == "" {
		return m.fallbackLang
	}

	// Accept-Language?	languages := strings.Split(acceptLanguage, ",")
	for _, lang := range languages {
		// 
		lang = strings.TrimSpace(strings.Split(lang, ";")[0])
		
		// ?		if m.IsLanguageSupported(lang) {
			return lang
		}
		
		// 
		if len(lang) >= 2 {
			langCode := lang[:2]
			for _, supported := range SupportedLanguages {
				if strings.HasPrefix(supported, langCode) {
					return supported
				}
			}
		}
	}

	return m.fallbackLang
}

// ReloadTranslations 
func (m *I18nManager) ReloadTranslations() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// bundle
	m.bundle = i18n.NewBundle(language.English)
	m.bundle.RegisterUnmarshalFunc("json", json.Unmarshal)

	// 
	if err := m.loadTranslations(); err != nil {
		return err
	}

	// localizers
	m.initializeLocalizers()

	return nil
}

// AddTranslation 
func (m *I18nManager) AddTranslation(lang, messageID, translation string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// 
	message := &i18n.Message{
		ID:    messageID,
		Other: translation,
	}

	// bundle
	return m.bundle.AddMessages(language.MustParse(lang), message)
}

// GetLanguageInfo 
func (m *I18nManager) GetLanguageInfo(lang string) LanguageInfo {
	languageInfoMap := map[string]LanguageInfo{
		"zh-CN": {Code: "zh-CN", Name: "?, NativeName: "?, Direction: "ltr"},
		"zh-TW": {Code: "zh-TW", Name: "Traditional Chinese", NativeName: "w", Direction: "ltr"},
		"en-US": {Code: "en-US", Name: "English (US)", NativeName: "English (US)", Direction: "ltr"},
		"en-GB": {Code: "en-GB", Name: "English (UK)", NativeName: "English (UK)", Direction: "ltr"},
		"ja-JP": {Code: "ja-JP", Name: "Japanese", NativeName: "?, Direction: "ltr"},
		"ko-KR": {Code: "ko-KR", Name: "Korean", NativeName: "?, Direction: "ltr"},
		"fr-FR": {Code: "fr-FR", Name: "French", NativeName: "Franais", Direction: "ltr"},
		"de-DE": {Code: "de-DE", Name: "German", NativeName: "Deutsch", Direction: "ltr"},
		"es-ES": {Code: "es-ES", Name: "Spanish", NativeName: "Espaol", Direction: "ltr"},
		"it-IT": {Code: "it-IT", Name: "Italian", NativeName: "Italiano", Direction: "ltr"},
		"pt-BR": {Code: "pt-BR", Name: "Portuguese (Brazil)", NativeName: "Portugus (Brasil)", Direction: "ltr"},
		"ru-RU": {Code: "ru-RU", Name: "Russian", NativeName: "", Direction: "ltr"},
		"ar-SA": {Code: "ar-SA", Name: "Arabic", NativeName: "", Direction: "rtl"},
		"hi-IN": {Code: "hi-IN", Name: "Hindi", NativeName: "", Direction: "ltr"},
		"th-TH": {Code: "th-TH", Name: "Thai", NativeName: "?, Direction: "ltr"},
		"vi-VN": {Code: "vi-VN", Name: "Vietnamese", NativeName: "Ting Vit", Direction: "ltr"},
		"id-ID": {Code: "id-ID", Name: "Indonesian", NativeName: "Bahasa Indonesia", Direction: "ltr"},
		"ms-MY": {Code: "ms-MY", Name: "Malay", NativeName: "Bahasa Melayu", Direction: "ltr"},
	}

	if info, exists := languageInfoMap[lang]; exists {
		return info
	}

	// 
	return LanguageInfo{
		Code:       lang,
		Name:       lang,
		NativeName: lang,
		Direction:  "ltr",
	}
}

// LanguageInfo 语言信息
type LanguageInfo struct {
	Code       string `json:"code"`
	Name       string `json:"name"`
	NativeName string `json:"native_name"`
	Direction  string `json:"direction"` // ltr ?rtl
}

// GetAllLanguageInfo 获取所有语言信息
func (m *I18nManager) GetAllLanguageInfo() []LanguageInfo {
	var languages []LanguageInfo
	for _, lang := range SupportedLanguages {
		languages = append(languages, m.GetLanguageInfo(lang))
	}
	return languages
}


// AI
package culture

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// CultureManager 
type CultureManager struct {
	defaultCulture string
	cultures       map[string]CultureInfo
}

// CultureInfo 
type CultureInfo struct {
	Code            string            `json:"code"`             //  (zh-CN, en-US, etc.)
	Name            string            `json:"name"`             // 
	Language        string            `json:"language"`         // 
	Country         string            `json:"country"`          // 
	Region          string            `json:"region"`           // 
	DateFormat      DateFormats       `json:"date_format"`      // 
	NumberFormat    NumberFormats     `json:"number_format"`    // 
	AddressFormat   AddressFormat     `json:"address_format"`   // 
	NameFormat      NameFormat        `json:"name_format"`      // 
	PhoneFormat     PhoneFormat       `json:"phone_format"`     // 绰
	BusinessCulture BusinessCulture   `json:"business_culture"` // 
	SocialNorms     SocialNorms       `json:"social_norms"`     // 淶
	Holidays        []Holiday         `json:"holidays"`         // 
	WorkingDays     []time.Weekday    `json:"working_days"`     // 
	WeekendDays     []time.Weekday    `json:"weekend_days"`     // 
	FirstDayOfWeek  time.Weekday      `json:"first_day_of_week"` // 
	RTL             bool              `json:"rtl"`              // 
	ColorMeanings   map[string]string `json:"color_meanings"`   // 
	TabooTopics     []string          `json:"taboo_topics"`     // 
	Preferences     CulturePrefs      `json:"preferences"`      // 
}

// DateFormats 
type DateFormats struct {
	Short    string `json:"short"`    // 
	Medium   string `json:"medium"`   // 
	Long     string `json:"long"`     // 
	Full     string `json:"full"`     // 
	Time12   string `json:"time_12"`  // 12
	Time24   string `json:"time_24"`  // 24
	DateTime string `json:"datetime"` // 
}

// NumberFormats 
type NumberFormats struct {
	DecimalSeparator  string `json:"decimal_separator"`  // 
	ThousandSeparator string `json:"thousand_separator"` // 
	PercentFormat     string `json:"percent_format"`     // 
	NegativeFormat    string `json:"negative_format"`    // 
}

// AddressFormat 
type AddressFormat struct {
	Format      string   `json:"format"`       // 
	Fields      []string `json:"fields"`       // 
	Required    []string `json:"required"`     // 
	PostalCode  string   `json:"postal_code"`  // 
	PhonePrefix string   `json:"phone_prefix"` // 绰
}

// NameFormat 
type NameFormat struct {
	Order       string `json:"order"`        //  (first_last, last_first)
	Honorifics  []string `json:"honorifics"` // /
	Separators  string `json:"separators"`   // 
	FirstName   string `json:"first_name"`   // 
	LastName    string `json:"last_name"`    // 
	MiddleName  bool   `json:"middle_name"`  // 
	FamilyFirst bool   `json:"family_first"` // 
}

// PhoneFormat 绰
type PhoneFormat struct {
	CountryCode   string `json:"country_code"`   // 
	Format        string `json:"format"`         // 绰
	MobilePrefix  []string `json:"mobile_prefix"` // 
	LandlineFormat string `json:"landline_format"` // 
}

// BusinessCulture 
type BusinessCulture struct {
	MeetingStyle    string   `json:"meeting_style"`    // 
	DecisionMaking  string   `json:"decision_making"`  // 
	Hierarchy       string   `json:"hierarchy"`        // 
	Communication   string   `json:"communication"`    // 
	Punctuality     string   `json:"punctuality"`      // 
	BusinessHours   string   `json:"business_hours"`   // 
	Greetings       []string `json:"greetings"`        // 
	GiftGiving      string   `json:"gift_giving"`      // 
	DressCode       string   `json:"dress_code"`       // 
	BusinessCards   string   `json:"business_cards"`   // 
}

// SocialNorms 淶
type SocialNorms struct {
	PersonalSpace   string   `json:"personal_space"`   // 
	EyeContact      string   `json:"eye_contact"`      // 
	Gestures        []string `json:"gestures"`         // 
	TouchingRules   string   `json:"touching_rules"`   // 
	ConversationStyle string `json:"conversation_style"` // 
	Silence         string   `json:"silence"`          // 
	Humor           string   `json:"humor"`            // 
	Compliments     string   `json:"compliments"`      // 
	Criticism       string   `json:"criticism"`        // 
	Privacy         string   `json:"privacy"`          // 
}

// Holiday 
type Holiday struct {
	Name        string    `json:"name"`         // 
	Date        string    `json:"date"`         // 
	Type        string    `json:"type"`         // national, religious, cultural
	Description string    `json:"description"`  // 
	Traditions  []string  `json:"traditions"`   // 
	IsPublic    bool      `json:"is_public"`    // 
	Duration    int       `json:"duration"`     // 
}

// CulturePrefs 
type CulturePrefs struct {
	PreferredColors    []string          `json:"preferred_colors"`     // 
	AvoidColors        []string          `json:"avoid_colors"`         // 
	LuckyNumbers       []int             `json:"lucky_numbers"`        // 
	UnluckyNumbers     []int             `json:"unlucky_numbers"`      // 
	ImagePreferences   map[string]string `json:"image_preferences"`    // 
	SymbolMeanings     map[string]string `json:"symbol_meanings"`      // 
	FoodRestrictions   []string          `json:"food_restrictions"`    // 
	ReligiousConsiderations []string     `json:"religious_considerations"` // 
}

// SupportedCultures 
var SupportedCultures = map[string]CultureInfo{
	"zh-CN": {
		Code:     "zh-CN",
		Name:     "",
		Language: "zh",
		Country:  "CN",
		Region:   "Asia",
		DateFormat: DateFormats{
			Short:    "2006/01/02",
			Medium:   "2006-01-02",
			Long:     "2006-01-02 ",
			Full:     "2006-01-02 ",
			Time12:   "3:04",
			Time24:   "15:04",
			DateTime: "2006-01-02 15:04",
		},
		NumberFormat: NumberFormats{
			DecimalSeparator:  ".",
			ThousandSeparator: ",",
			PercentFormat:     "%.2f%%",
			NegativeFormat:    "-%.2f",
		},
		AddressFormat: AddressFormat{
			Format:      "{country} {province} {city} {district} {street} {building}",
			Fields:      []string{"country", "province", "city", "district", "street", "building", "postal_code"},
			Required:    []string{"province", "city", "street"},
			PostalCode:  "\\d{6}",
			PhonePrefix: "+86",
		},
		NameFormat: NameFormat{
			Order:       "last_first",
			Honorifics:  []string{"", "", "", "", "", "", "", "", ""},
			Separators:  "",
			MiddleName:  false,
			FamilyFirst: true,
		},
		PhoneFormat: PhoneFormat{
			CountryCode:    "+86",
			Format:         "1XX-XXXX-XXXX",
			MobilePrefix:   []string{"13", "14", "15", "16", "17", "18", "19"},
			LandlineFormat: "0XX-XXXXXXXX",
		},
		BusinessCulture: BusinessCulture{
			MeetingStyle:   "formal",
			DecisionMaking: "hierarchical",
			Hierarchy:      "high",
			Communication:  "indirect",
			Punctuality:    "important",
			BusinessHours:  "09:00-18:00",
			Greetings:      []string{"", "", "", ""},
			GiftGiving:     "important",
			DressCode:      "formal",
			BusinessCards:  "two_hands",
		},
		SocialNorms: SocialNorms{
			PersonalSpace:     "close",
			EyeContact:        "respectful",
			Gestures:          []string{"bow", "handshake"},
			TouchingRules:     "minimal",
			ConversationStyle: "polite",
			Silence:           "comfortable",
			Humor:             "subtle",
			Compliments:       "modest",
			Criticism:         "indirect",
			Privacy:           "important",
		},
		Holidays: []Holiday{
			{Name: "", Date: "lunar_new_year", Type: "national", IsPublic: true, Duration: 7},
			{Name: "", Date: "qingming", Type: "traditional", IsPublic: true, Duration: 1},
			{Name: "", Date: "05-01", Type: "national", IsPublic: true, Duration: 3},
			{Name: "", Date: "dragon_boat", Type: "traditional", IsPublic: true, Duration: 1},
			{Name: "", Date: "mid_autumn", Type: "traditional", IsPublic: true, Duration: 1},
			{Name: "", Date: "10-01", Type: "national", IsPublic: true, Duration: 7},
		},
		WorkingDays:    []time.Weekday{time.Monday, time.Tuesday, time.Wednesday, time.Thursday, time.Friday},
		WeekendDays:    []time.Weekday{time.Saturday, time.Sunday},
		FirstDayOfWeek: time.Monday,
		RTL:            false,
		ColorMeanings: map[string]string{
			"red":    "luck,prosperity,joy",
			"gold":   "wealth,prosperity",
			"white":  "purity,mourning",
			"black":  "formality,mourning",
			"green":  "growth,harmony",
			"blue":   "trust,stability",
		},
		TabooTopics: []string{"politics", "personal_income", "age", "weight"},
		Preferences: CulturePrefs{
			PreferredColors:    []string{"red", "gold", "green"},
			AvoidColors:        []string{"white_for_celebration"},
			LuckyNumbers:       []int{6, 8, 9},
			UnluckyNumbers:     []int{4},
			ImagePreferences:   map[string]string{"dragon": "power", "phoenix": "beauty", "lotus": "purity"},
			SymbolMeanings:     map[string]string{"*": "power", "**": "beauty", "***": "purity"},
			FoodRestrictions:   []string{},
			ReligiousConsiderations: []string{"buddhist_vegetarian", "halal_options"},
		},
	},
	"en-US": {
		Code:     "en-US",
		Name:     "English (United States)",
		Language: "en",
		Country:  "US",
		Region:   "North America",
		DateFormat: DateFormats{
			Short:    "01/02/2006",
			Medium:   "Jan 2, 2006",
			Long:     "January 2, 2006",
			Full:     "Monday, January 2, 2006",
			Time12:   "3:04 PM",
			Time24:   "15:04",
			DateTime: "01/02/2006 3:04 PM",
		},
		NumberFormat: NumberFormats{
			DecimalSeparator:  ".",
			ThousandSeparator: ",",
			PercentFormat:     "%.2f%%",
			NegativeFormat:    "-%.2f",
		},
		AddressFormat: AddressFormat{
			Format:      "{street} {city}, {state} {postal_code} {country}",
			Fields:      []string{"street", "city", "state", "postal_code", "country"},
			Required:    []string{"street", "city", "state", "postal_code"},
			PostalCode:  "\\d{5}(-\\d{4})?",
			PhonePrefix: "+1",
		},
		NameFormat: NameFormat{
			Order:       "first_last",
			Honorifics:  []string{"Mr.", "Mrs.", "Ms.", "Dr.", "Prof.", "Rev."},
			Separators:  " ",
			MiddleName:  true,
			FamilyFirst: false,
		},
		PhoneFormat: PhoneFormat{
			CountryCode:    "+1",
			Format:         "(XXX) XXX-XXXX",
			MobilePrefix:   []string{},
			LandlineFormat: "(XXX) XXX-XXXX",
		},
		BusinessCulture: BusinessCulture{
			MeetingStyle:   "direct",
			DecisionMaking: "collaborative",
			Hierarchy:      "low",
			Communication:  "direct",
			Punctuality:    "very_important",
			BusinessHours:  "09:00-17:00",
			Greetings:      []string{"Hello", "Good morning", "Good afternoon"},
			GiftGiving:     "minimal",
			DressCode:      "business_casual",
			BusinessCards:  "casual",
		},
		SocialNorms: SocialNorms{
			PersonalSpace:     "wide",
			EyeContact:        "direct",
			Gestures:          []string{"handshake", "wave"},
			TouchingRules:     "minimal",
			ConversationStyle: "friendly",
			Silence:           "uncomfortable",
			Humor:             "common",
			Compliments:       "direct",
			Criticism:         "constructive",
			Privacy:           "very_important",
		},
		Holidays: []Holiday{
			{Name: "New Year's Day", Date: "01-01", Type: "national", IsPublic: true, Duration: 1},
			{Name: "Independence Day", Date: "07-04", Type: "national", IsPublic: true, Duration: 1},
			{Name: "Thanksgiving", Date: "thanksgiving", Type: "national", IsPublic: true, Duration: 1},
			{Name: "Christmas", Date: "12-25", Type: "religious", IsPublic: true, Duration: 1},
		},
		WorkingDays:    []time.Weekday{time.Monday, time.Tuesday, time.Wednesday, time.Thursday, time.Friday},
		WeekendDays:    []time.Weekday{time.Saturday, time.Sunday},
		FirstDayOfWeek: time.Sunday,
		RTL:            false,
		ColorMeanings: map[string]string{
			"red":   "passion,danger",
			"blue":  "trust,calm",
			"green": "nature,money",
			"white": "purity,peace",
			"black": "elegance,formality",
		},
		TabooTopics: []string{"personal_income", "weight", "age", "religion", "politics"},
		Preferences: CulturePrefs{
			PreferredColors:    []string{"blue", "red", "white"},
			AvoidColors:        []string{},
			LuckyNumbers:       []int{7},
			UnluckyNumbers:     []int{13},
			ImagePreferences:   map[string]string{"eagle": "freedom", "flag": "patriotism"},
			SymbolMeanings:     map[string]string{"eagle": "freedom", "star": "excellence"},
			FoodRestrictions:   []string{"vegetarian", "vegan", "gluten_free"},
			ReligiousConsiderations: []string{"christian", "jewish", "muslim", "hindu"},
		},
	},
	"ja-JP": {
		Code:     "ja-JP",
		Name:     "Z",
		Language: "ja",
		Country:  "JP",
		Region:   "Asia",
		DateFormat: DateFormats{
			Short:    "2006/01/02",
			Medium:   "2006-01-02",
			Long:     "2006-01-02",
			Full:     "2006-01-02 ",
			Time12:   "3:04",
			Time24:   "15:04",
			DateTime: "2006-01-02 15:04",
		},
		NumberFormat: NumberFormats{
			DecimalSeparator:  ".",
			ThousandSeparator: ",",
			PercentFormat:     "%.2f%%",
			NegativeFormat:    "-%.2f",
		},
		AddressFormat: AddressFormat{
			Format:      "{postal_code} {prefecture} {city} {district} {street} {building}",
			Fields:      []string{"postal_code", "prefecture", "city", "district", "street", "building"},
			Required:    []string{"prefecture", "city", "street"},
			PostalCode:  "\\d{3}-\\d{4}",
			PhonePrefix: "+81",
		},
		NameFormat: NameFormat{
			Order:       "last_first",
			Honorifics:  []string{"", "", "", "", "", ""},
			Separators:  "",
			MiddleName:  false,
			FamilyFirst: true,
		},
		PhoneFormat: PhoneFormat{
			CountryCode:    "+81",
			Format:         "0XX-XXXX-XXXX",
			MobilePrefix:   []string{"070", "080", "090"},
			LandlineFormat: "0X-XXXX-XXXX",
		},
		BusinessCulture: BusinessCulture{
			MeetingStyle:   "formal",
			DecisionMaking: "consensus",
			Hierarchy:      "high",
			Communication:  "indirect",
			Punctuality:    "extremely_important",
			BusinessHours:  "09:00-18:00",
			Greetings:      []string{"", "", ""},
			GiftGiving:     "very_important",
			DressCode:      "formal",
			BusinessCards:  "meishi_ritual",
		},
		SocialNorms: SocialNorms{
			PersonalSpace:     "respectful",
			EyeContact:        "minimal",
			Gestures:          []string{"bow", "minimal_handshake"},
			TouchingRules:     "no_touching",
			ConversationStyle: "polite",
			Silence:           "comfortable",
			Humor:             "subtle",
			Compliments:       "humble",
			Criticism:         "very_indirect",
			Privacy:           "extremely_important",
		},
		Holidays: []Holiday{
			{Name: "", Date: "01-01", Type: "national", IsPublic: true, Duration: 1},
			{Name: "", Date: "coming_of_age", Type: "national", IsPublic: true, Duration: 1},
			{Name: "", Date: "02-11", Type: "national", IsPublic: true, Duration: 1},
			{Name: "", Date: "04-29", Type: "national", IsPublic: true, Duration: 1},
			{Name: "", Date: "05-03", Type: "national", IsPublic: true, Duration: 1},
			{Name: "", Date: "05-05", Type: "national", IsPublic: true, Duration: 1},
		},
		WorkingDays:    []time.Weekday{time.Monday, time.Tuesday, time.Wednesday, time.Thursday, time.Friday},
		WeekendDays:    []time.Weekday{time.Saturday, time.Sunday},
		FirstDayOfWeek: time.Sunday,
		RTL:            false,
		ColorMeanings: map[string]string{
			"red":    "life,vitality",
			"white":  "purity,death",
			"black":  "formality",
			"blue":   "calm,stability",
			"green":  "nature,youth",
		},
		TabooTopics: []string{"personal_income", "age", "weight", "world_war_ii", "politics"},
		Preferences: CulturePrefs{
			PreferredColors:    []string{"red", "white", "blue"},
			AvoidColors:        []string{"white_for_celebration"},
			LuckyNumbers:       []int{7, 8},
			UnluckyNumbers:     []int{4, 9},
			ImagePreferences:   map[string]string{"cherry_blossom": "beauty", "crane": "longevity", "turtle": "longevity"},
			SymbolMeanings:     map[string]string{"": "beauty", "w": "longevity"},
			FoodRestrictions:   []string{"vegetarian", "halal"},
			ReligiousConsiderations: []string{"buddhist", "shinto"},
		},
	},
	"ko-KR": {
		Code:     "ko-KR",
		Name:     "",
		Language: "ko",
		Country:  "KR",
		Region:   "Asia",
		DateFormat: DateFormats{
			Short:    "2006. 01. 02.",
			Medium:   "2006-01-02",
			Long:     "2006-01-02 ",
			Full:     "2006-01-02 ",
			Time12:   " 3:04",
			Time24:   "15:04",
			DateTime: "2006-01-02 15:04",
		},
		NumberFormat: NumberFormats{
			DecimalSeparator:  ".",
			ThousandSeparator: ",",
			PercentFormat:     "%.2f%%",
			NegativeFormat:    "-%.2f",
		},
		AddressFormat: AddressFormat{
			Format:      "{postal_code} {province} {city} {district} {street} {building}",
			Fields:      []string{"postal_code", "province", "city", "district", "street", "building"},
			Required:    []string{"province", "city", "street"},
			PostalCode:  "\\d{5}",
			PhonePrefix: "+82",
		},
		NameFormat: NameFormat{
			Order:       "last_first",
			Honorifics:  []string{"", "", "", "", "", ""},
			Separators:  "",
			MiddleName:  false,
			FamilyFirst: true,
		},
		PhoneFormat: PhoneFormat{
			CountryCode:    "+82",
			Format:         "010-XXXX-XXXX",
			MobilePrefix:   []string{"010", "011", "016", "017", "018", "019"},
			LandlineFormat: "0X-XXX-XXXX",
		},
		BusinessCulture: BusinessCulture{
			MeetingStyle:   "formal",
			DecisionMaking: "hierarchical",
			Hierarchy:      "very_high",
			Communication:  "indirect",
			Punctuality:    "important",
			BusinessHours:  "09:00-18:00",
			Greetings:      []string{"", " ", ""},
			GiftGiving:     "important",
			DressCode:      "formal",
			BusinessCards:  "two_hands",
		},
		SocialNorms: SocialNorms{
			PersonalSpace:     "close",
			EyeContact:        "respectful",
			Gestures:          []string{"bow", "handshake"},
			TouchingRules:     "minimal",
			ConversationStyle: "respectful",
			Silence:           "comfortable",
			Humor:             "careful",
			Compliments:       "humble",
			Criticism:         "indirect",
			Privacy:           "important",
		},
		Holidays: []Holiday{
			{Name: "", Date: "01-01", Type: "national", IsPublic: true, Duration: 1},
			{Name: "", Date: "lunar_new_year", Type: "traditional", IsPublic: true, Duration: 3},
			{Name: "", Date: "03-01", Type: "national", IsPublic: true, Duration: 1},
			{Name: "", Date: "05-05", Type: "national", IsPublic: true, Duration: 1},
			{Name: "", Date: "buddha_birthday", Type: "religious", IsPublic: true, Duration: 1},
			{Name: "", Date: "mid_autumn", Type: "traditional", IsPublic: true, Duration: 3},
		},
		WorkingDays:    []time.Weekday{time.Monday, time.Tuesday, time.Wednesday, time.Thursday, time.Friday},
		WeekendDays:    []time.Weekday{time.Saturday, time.Sunday},
		FirstDayOfWeek: time.Sunday,
		RTL:            false,
		ColorMeanings: map[string]string{
			"red":    "passion,good_fortune",
			"blue":   "trust,stability",
			"white":  "purity,death",
			"black":  "formality",
			"yellow": "prosperity",
		},
		TabooTopics: []string{"personal_income", "age", "north_korea", "politics"},
		Preferences: CulturePrefs{
			PreferredColors:    []string{"red", "blue", "white"},
			AvoidColors:        []string{"white_for_celebration"},
			LuckyNumbers:       []int{7, 8, 9},
			UnluckyNumbers:     []int{4},
			ImagePreferences:   map[string]string{"tiger": "strength", "crane": "longevity", "pine": "endurance"},
			SymbolMeanings:     map[string]string{"": "strength", "w": "longevity", "": "endurance"},
			FoodRestrictions:   []string{"vegetarian", "halal"},
			ReligiousConsiderations: []string{"buddhist", "christian", "confucian"},
		},
	},
	"ar-SA": {
		Code:     "ar-SA",
		Name:     " (  )",
		Language: "ar",
		Country:  "SA",
		Region:   "Middle East",
		DateFormat: DateFormats{
			Short:    "02/01/2006",
			Medium:   "02  2006",
			Long:     "02  2006",
			Full:     " 02  2006",
			Time12:   "3:04 ",
			Time24:   "15:04",
			DateTime: "02/01/2006 3:04 ",
		},
		NumberFormat: NumberFormats{
			DecimalSeparator:  ".",
			ThousandSeparator: ",",
			PercentFormat:     "%.2f%%",
			NegativeFormat:    "-%.2f",
		},
		AddressFormat: AddressFormat{
			Format:      "{building} {street} {district} {city} {postal_code} {country}",
			Fields:      []string{"building", "street", "district", "city", "postal_code", "country"},
			Required:    []string{"city", "street"},
			PostalCode:  "\\d{5}",
			PhonePrefix: "+966",
		},
		NameFormat: NameFormat{
			Order:       "first_last",
			Honorifics:  []string{"", "", "", "", ""},
			Separators:  " ",
			MiddleName:  true,
			FamilyFirst: false,
		},
		PhoneFormat: PhoneFormat{
			CountryCode:    "+966",
			Format:         "05X XXX XXXX",
			MobilePrefix:   []string{"050", "051", "052", "053", "054", "055", "056", "057", "058", "059"},
			LandlineFormat: "01X XXX XXXX",
		},
		BusinessCulture: BusinessCulture{
			MeetingStyle:   "formal",
			DecisionMaking: "hierarchical",
			Hierarchy:      "very_high",
			Communication:  "indirect",
			Punctuality:    "flexible",
			BusinessHours:  "08:00-17:00",
			Greetings:      []string{" ", " ", ""},
			GiftGiving:     "important",
			DressCode:      "conservative",
			BusinessCards:  "right_hand",
		},
		SocialNorms: SocialNorms{
			PersonalSpace:     "close_same_gender",
			EyeContact:        "same_gender",
			Gestures:          []string{"handshake_same_gender"},
			TouchingRules:     "same_gender_only",
			ConversationStyle: "respectful",
			Silence:           "comfortable",
			Humor:             "careful",
			Compliments:       "appreciated",
			Criticism:        "very_indirect",
			Privacy:          "very_important",
		},
		Holidays: []Holiday{
			{Name: "  ", Date: "hijri_new_year", Type: "religious", IsPublic: true, Duration: 1},
			{Name: " ", Date: "eid_al_fitr", Type: "religious", IsPublic: true, Duration: 3},
			{Name: " ", Date: "eid_al_adha", Type: "religious", IsPublic: true, Duration: 4},
			{Name: " ", Date: "09-23", Type: "national", IsPublic: true, Duration: 1},
		},
		WorkingDays:    []time.Weekday{time.Sunday, time.Monday, time.Tuesday, time.Wednesday, time.Thursday},
		WeekendDays:    []time.Weekday{time.Friday, time.Saturday},
		FirstDayOfWeek: time.Saturday,
		RTL:            true,
		ColorMeanings: map[string]string{
			"green":  "islam,paradise",
			"white":  "purity,peace",
			"black":  "dignity",
			"gold":   "wealth",
			"blue":   "trust",
		},
		TabooTopics: []string{"alcohol", "pork", "politics", "israel", "personal_relationships"},
		Preferences: CulturePrefs{
			PreferredColors:    []string{"green", "white", "gold"},
			AvoidColors:        []string{"pink", "purple"},
			LuckyNumbers:       []int{7},
			UnluckyNumbers:     []int{13},
			ImagePreferences:   map[string]string{"geometric": "islamic_art", "calligraphy": "beauty"},
			SymbolMeanings:     map[string]string{"": "islam", "": "guidance"},
			FoodRestrictions:   []string{"halal_only", "no_alcohol", "no_pork"},
			ReligiousConsiderations: []string{"islamic", "prayer_times", "ramadan", "halal"},
		},
	},
}

// NewCultureManager 
func NewCultureManager(defaultCulture string) *CultureManager {
	return &CultureManager{
		defaultCulture: defaultCulture,
		cultures:       SupportedCultures,
	}
}

// GetCultureInfo 
func (cm *CultureManager) GetCultureInfo(cultureCode string) (CultureInfo, error) {
	culture, exists := cm.cultures[cultureCode]
	if !exists {
		return CultureInfo{}, fmt.Errorf("unsupported culture: %s", cultureCode)
	}
	return culture, nil
}

// FormatName 
func (cm *CultureManager) FormatName(firstName, lastName, middleName, honorific, cultureCode string) (string, error) {
	culture, err := cm.GetCultureInfo(cultureCode)
	if err != nil {
		return "", err
	}

	nameFormat := culture.NameFormat
	var parts []string

	if honorific != "" {
		parts = append(parts, honorific)
	}

	if nameFormat.FamilyFirst {
		// 		if lastName != "" {
			parts = append(parts, lastName)
		}
		if middleName != "" && nameFormat.MiddleName {
			parts = append(parts, middleName)
		}
		if firstName != "" {
			parts = append(parts, firstName)
		}
	} else {
		// 		if firstName != "" {
			parts = append(parts, firstName)
		}
		if middleName != "" && nameFormat.MiddleName {
			parts = append(parts, middleName)
		}
		if lastName != "" {
			parts = append(parts, lastName)
		}
	}

	separator := nameFormat.Separators
	if separator == "" {
		separator = " "
	}

	return strings.Join(parts, separator), nil
}

// FormatAddress 
func (cm *CultureManager) FormatAddress(addressData map[string]string, cultureCode string) (string, error) {
	culture, err := cm.GetCultureInfo(cultureCode)
	if err != nil {
		return "", err
	}

	addressFormat := culture.AddressFormat
	result := addressFormat.Format

	// 滻
	for field, value := range addressData {
		placeholder := fmt.Sprintf("{%s}", field)
		result = strings.ReplaceAll(result, placeholder, value)
	}

	// 
	for _, field := range addressFormat.Fields {
		placeholder := fmt.Sprintf("{%s}", field)
		result = strings.ReplaceAll(result, placeholder, "")
	}

	// 
	result = strings.TrimSpace(result)
	result = strings.ReplaceAll(result, "  ", " ")

	return result, nil
}

// FormatPhoneNumber 绰
func (cm *CultureManager) FormatPhoneNumber(phoneNumber, cultureCode string) (string, error) {
	culture, err := cm.GetCultureInfo(cultureCode)
	if err != nil {
		return "", err
	}

	phoneFormat := culture.PhoneFormat
	
	// 
	cleanNumber := strings.ReplaceAll(phoneNumber, " ", "")
	cleanNumber = strings.ReplaceAll(cleanNumber, "-", "")
	cleanNumber = strings.ReplaceAll(cleanNumber, "(", "")
	cleanNumber = strings.ReplaceAll(cleanNumber, ")", "")
	cleanNumber = strings.ReplaceAll(cleanNumber, "+", "")

	// 
	format := phoneFormat.Format
	result := ""
	numberIndex := 0

	for _, char := range format {
		if char == 'X' && numberIndex < len(cleanNumber) {
			result += string(cleanNumber[numberIndex])
			numberIndex++
		} else if char != 'X' {
			result += string(char)
		}
	}

	return result, nil
}

// GetBusinessHours 
func (cm *CultureManager) GetBusinessHours(cultureCode string) (string, error) {
	culture, err := cm.GetCultureInfo(cultureCode)
	if err != nil {
		return "", err
	}
	return culture.BusinessCulture.BusinessHours, nil
}

// GetWorkingDays 
// 
func (cm *CultureManager) GetWorkingDays(cultureCode string) ([]time.Weekday, error) {
	culture, err := cm.GetCultureInfo(cultureCode)
	if err != nil {
		return nil, err
	}
	return culture.WorkingDays, nil
}

// GetWeekendDays 
func (cm *CultureManager) GetWeekendDays(cultureCode string) ([]time.Weekday, error) {
	culture, err := cm.GetCultureInfo(cultureCode)
	if err != nil {
		return nil, err
	}
	return culture.WeekendDays, nil
}

// IsWorkingDay 
// 
func (cm *CultureManager) IsWorkingDay(date time.Time, cultureCode string) (bool, error) {
	workingDays, err := cm.GetWorkingDays(cultureCode)
	if err != nil {
		return false, err
	}

	for _, day := range workingDays {
		if date.Weekday() == day {
			return true, nil
		}
	}
	return false, nil
}

// GetHolidays 
func (cm *CultureManager) GetHolidays(cultureCode string) ([]Holiday, error) {
	culture, err := cm.GetCultureInfo(cultureCode)
	if err != nil {
		return nil, err
	}
	return culture.Holidays, nil
}

// IsHoliday 
// 
func (cm *CultureManager) IsHoliday(date time.Time, cultureCode string) (bool, Holiday, error) {
	holidays, err := cm.GetHolidays(cultureCode)
	if err != nil {
		return false, Holiday{}, err
	}

	for _, holiday := range holidays {
		// 
		// 
		if cm.matchHolidayDate(date, holiday) {
			return true, holiday, nil
		}
	}

	return false, Holiday{}, nil
}

// matchHolidayDate 
//  MM-DD
func (cm *CultureManager) matchHolidayDate(date time.Time, holiday Holiday) bool {
	//  MM-DD
	if strings.Contains(holiday.Date, "-") && len(holiday.Date) == 5 {
		holidayDate := fmt.Sprintf("%02d-%02d", date.Month(), date.Day())
		return holidayDate == holiday.Date
	}
	
	// 
	return false
}

// GetColorMeaning 
func (cm *CultureManager) GetColorMeaning(color, cultureCode string) (string, error) {
	culture, err := cm.GetCultureInfo(cultureCode)
	if err != nil {
		return "", err
	}

	meaning, exists := culture.ColorMeanings[color]
	if !exists {
		return "", fmt.Errorf("color meaning not found for %s in culture %s", color, cultureCode)
	}

	return meaning, nil
}

// IsTabooTopic 
func (cm *CultureManager) IsTabooTopic(topic, cultureCode string) (bool, error) {
	culture, err := cm.GetCultureInfo(cultureCode)
	if err != nil {
		return false, err
	}

	for _, taboo := range culture.TabooTopics {
		if strings.EqualFold(topic, taboo) {
			return true, nil
		}
	}

	return false, nil
}

// GetPreferredColors 
func (cm *CultureManager) GetPreferredColors(cultureCode string) ([]string, error) {
	culture, err := cm.GetCultureInfo(cultureCode)
	if err != nil {
		return nil, err
	}
	return culture.Preferences.PreferredColors, nil
}

// GetAvoidColors 
func (cm *CultureManager) GetAvoidColors(cultureCode string) ([]string, error) {
	culture, err := cm.GetCultureInfo(cultureCode)
	if err != nil {
		return nil, err
	}
	return culture.Preferences.AvoidColors, nil
}

// IsLuckyNumber 
func (cm *CultureManager) IsLuckyNumber(number int, cultureCode string) (bool, error) {
	culture, err := cm.GetCultureInfo(cultureCode)
	if err != nil {
		return false, err
	}

	for _, lucky := range culture.Preferences.LuckyNumbers {
		if number == lucky {
			return true, nil
		}
	}

	return false, nil
}

// IsUnluckyNumber 
// 
func (cm *CultureManager) IsUnluckyNumber(number int, cultureCode string) (bool, error) {
	culture, err := cm.GetCultureInfo(cultureCode)
	if err != nil {
		return false, err
	}

	for _, unlucky := range culture.Preferences.UnluckyNumbers {
		if number == unlucky {
			return true, nil
		}
	}

	return false, nil
}

// GetFoodRestrictions 
func (cm *CultureManager) GetFoodRestrictions(cultureCode string) ([]string, error) {
	culture, err := cm.GetCultureInfo(cultureCode)
	if err != nil {
		return nil, err
	}
	return culture.Preferences.FoodRestrictions, nil
}

// GetReligiousConsiderations 
func (cm *CultureManager) GetReligiousConsiderations(cultureCode string) ([]string, error) {
	culture, err := cm.GetCultureInfo(cultureCode)
	if err != nil {
		return nil, err
	}
	return culture.Preferences.ReligiousConsiderations, nil
}

// IsRTL 
func (cm *CultureManager) IsRTL(cultureCode string) (bool, error) {
	culture, err := cm.GetCultureInfo(cultureCode)
	if err != nil {
		return false, err
	}
	return culture.RTL, nil
}

// GetFirstDayOfWeek 
// 
func (cm *CultureManager) GetFirstDayOfWeek(cultureCode string) (time.Weekday, error) {
	culture, err := cm.GetCultureInfo(cultureCode)
	if err != nil {
		return time.Sunday, err
	}
	return culture.FirstDayOfWeek, nil
}

// GetAllCultures 
func (cm *CultureManager) GetAllCultures() []CultureInfo {
	var cultures []CultureInfo
	for _, culture := range cm.cultures {
		cultures = append(cultures, culture)
	}
	return cultures
}

// GetCulturesByRegion 
// 
func (cm *CultureManager) GetCulturesByRegion(region string) []CultureInfo {
	var cultures []CultureInfo
	for _, culture := range cm.cultures {
		if strings.EqualFold(culture.Region, region) {
			cultures = append(cultures, culture)
		}
	}
	return cultures
}

// GetCulturesByLanguage 
func (cm *CultureManager) GetCulturesByLanguage(language string) []CultureInfo {
	var cultures []CultureInfo
	for _, culture := range cm.cultures {
		if strings.EqualFold(culture.Language, language) {
			cultures = append(cultures, culture)
		}
	}
	return cultures
}

// IsCultureSupported 
// 
func (cm *CultureManager) IsCultureSupported(cultureCode string) bool {
	_, exists := cm.cultures[cultureCode]
	return exists
}

// DetectCultureFromContext 
// HTTP
func (cm *CultureManager) DetectCultureFromContext(ctx context.Context) (string, error) {
	// HTTP
	// 
	return cm.defaultCulture, nil
}


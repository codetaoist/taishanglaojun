// AI
package timezone

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"
)

// TimezoneManager ?type TimezoneManager struct {
	defaultTimezone string
	timezones       map[string]*time.Location
	mutex           sync.RWMutex
}

// TimezoneInfo 
type TimezoneInfo struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Abbreviation string `json:"abbreviation"`
	Offset       int    `json:"offset"`       // UTC
	DSTOffset    int    `json:"dst_offset"`   // ?	Country      string `json:"country"`
	Region       string `json:"region"`
}

// SupportedTimezones ?var SupportedTimezones = map[string]TimezoneInfo{
	// 
	"Asia/Shanghai": {
		ID:           "Asia/Shanghai",
		Name:         "China Standard Time",
		Abbreviation: "CST",
		Offset:       28800, // UTC+8
		DSTOffset:    0,
		Country:      "CN",
		Region:       "Asia",
	},
	"Asia/Hong_Kong": {
		ID:           "Asia/Hong_Kong",
		Name:         "Hong Kong Time",
		Abbreviation: "HKT",
		Offset:       28800, // UTC+8
		DSTOffset:    0,
		Country:      "HK",
		Region:       "Asia",
	},
	"Asia/Tokyo": {
		ID:           "Asia/Tokyo",
		Name:         "Japan Standard Time",
		Abbreviation: "JST",
		Offset:       32400, // UTC+9
		DSTOffset:    0,
		Country:      "JP",
		Region:       "Asia",
	},
	"Asia/Seoul": {
		ID:           "Asia/Seoul",
		Name:         "Korea Standard Time",
		Abbreviation: "KST",
		Offset:       32400, // UTC+9
		DSTOffset:    0,
		Country:      "KR",
		Region:       "Asia",
	},
	"Asia/Singapore": {
		ID:           "Asia/Singapore",
		Name:         "Singapore Standard Time",
		Abbreviation: "SGT",
		Offset:       28800, // UTC+8
		DSTOffset:    0,
		Country:      "SG",
		Region:       "Asia",
	},
	"Asia/Bangkok": {
		ID:           "Asia/Bangkok",
		Name:         "Indochina Time",
		Abbreviation: "ICT",
		Offset:       25200, // UTC+7
		DSTOffset:    0,
		Country:      "TH",
		Region:       "Asia",
	},
	"Asia/Jakarta": {
		ID:           "Asia/Jakarta",
		Name:         "Western Indonesia Time",
		Abbreviation: "WIB",
		Offset:       25200, // UTC+7
		DSTOffset:    0,
		Country:      "ID",
		Region:       "Asia",
	},
	"Asia/Manila": {
		ID:           "Asia/Manila",
		Name:         "Philippine Standard Time",
		Abbreviation: "PST",
		Offset:       28800, // UTC+8
		DSTOffset:    0,
		Country:      "PH",
		Region:       "Asia",
	},
	"Asia/Kolkata": {
		ID:           "Asia/Kolkata",
		Name:         "India Standard Time",
		Abbreviation: "IST",
		Offset:       19800, // UTC+5:30
		DSTOffset:    0,
		Country:      "IN",
		Region:       "Asia",
	},

	// 
	"Europe/London": {
		ID:           "Europe/London",
		Name:         "Greenwich Mean Time",
		Abbreviation: "GMT",
		Offset:       0, // UTC+0
		DSTOffset:    3600, // UTC+1 (BST)
		Country:      "GB",
		Region:       "Europe",
	},
	"Europe/Paris": {
		ID:           "Europe/Paris",
		Name:         "Central European Time",
		Abbreviation: "CET",
		Offset:       3600, // UTC+1
		DSTOffset:    7200, // UTC+2 (CEST)
		Country:      "FR",
		Region:       "Europe",
	},
	"Europe/Berlin": {
		ID:           "Europe/Berlin",
		Name:         "Central European Time",
		Abbreviation: "CET",
		Offset:       3600, // UTC+1
		DSTOffset:    7200, // UTC+2 (CEST)
		Country:      "DE",
		Region:       "Europe",
	},
	"Europe/Rome": {
		ID:           "Europe/Rome",
		Name:         "Central European Time",
		Abbreviation: "CET",
		Offset:       3600, // UTC+1
		DSTOffset:    7200, // UTC+2 (CEST)
		Country:      "IT",
		Region:       "Europe",
	},
	"Europe/Madrid": {
		ID:           "Europe/Madrid",
		Name:         "Central European Time",
		Abbreviation: "CET",
		Offset:       3600, // UTC+1
		DSTOffset:    7200, // UTC+2 (CEST)
		Country:      "ES",
		Region:       "Europe",
	},
	"Europe/Dublin": {
		ID:           "Europe/Dublin",
		Name:         "Greenwich Mean Time",
		Abbreviation: "GMT",
		Offset:       0, // UTC+0
		DSTOffset:    3600, // UTC+1 (IST)
		Country:      "IE",
		Region:       "Europe",
	},
	"Europe/Moscow": {
		ID:           "Europe/Moscow",
		Name:         "Moscow Standard Time",
		Abbreviation: "MSK",
		Offset:       10800, // UTC+3
		DSTOffset:    0,
		Country:      "RU",
		Region:       "Europe",
	},

	// 
	"America/New_York": {
		ID:           "America/New_York",
		Name:         "Eastern Standard Time",
		Abbreviation: "EST",
		Offset:       -18000, // UTC-5
		DSTOffset:    -14400, // UTC-4 (EDT)
		Country:      "US",
		Region:       "America",
	},
	"America/Chicago": {
		ID:           "America/Chicago",
		Name:         "Central Standard Time",
		Abbreviation: "CST",
		Offset:       -21600, // UTC-6
		DSTOffset:    -18000, // UTC-5 (CDT)
		Country:      "US",
		Region:       "America",
	},
	"America/Denver": {
		ID:           "America/Denver",
		Name:         "Mountain Standard Time",
		Abbreviation: "MST",
		Offset:       -25200, // UTC-7
		DSTOffset:    -21600, // UTC-6 (MDT)
		Country:      "US",
		Region:       "America",
	},
	"America/Los_Angeles": {
		ID:           "America/Los_Angeles",
		Name:         "Pacific Standard Time",
		Abbreviation: "PST",
		Offset:       -28800, // UTC-8
		DSTOffset:    -25200, // UTC-7 (PDT)
		Country:      "US",
		Region:       "America",
	},
	"America/Toronto": {
		ID:           "America/Toronto",
		Name:         "Eastern Standard Time",
		Abbreviation: "EST",
		Offset:       -18000, // UTC-5
		DSTOffset:    -14400, // UTC-4 (EDT)
		Country:      "CA",
		Region:       "America",
	},
	"America/Vancouver": {
		ID:           "America/Vancouver",
		Name:         "Pacific Standard Time",
		Abbreviation: "PST",
		Offset:       -28800, // UTC-8
		DSTOffset:    -25200, // UTC-7 (PDT)
		Country:      "CA",
		Region:       "America",
	},

	// ?	"Australia/Sydney": {
		ID:           "Australia/Sydney",
		Name:         "Australian Eastern Standard Time",
		Abbreviation: "AEST",
		Offset:       36000, // UTC+10
		DSTOffset:    39600, // UTC+11 (AEDT)
		Country:      "AU",
		Region:       "Australia",
	},
	"Australia/Melbourne": {
		ID:           "Australia/Melbourne",
		Name:         "Australian Eastern Standard Time",
		Abbreviation: "AEST",
		Offset:       36000, // UTC+10
		DSTOffset:    39600, // UTC+11 (AEDT)
		Country:      "AU",
		Region:       "Australia",
	},
	"Pacific/Auckland": {
		ID:           "Pacific/Auckland",
		Name:         "New Zealand Standard Time",
		Abbreviation: "NZST",
		Offset:       43200, // UTC+12
		DSTOffset:    46800, // UTC+13 (NZDT)
		Country:      "NZ",
		Region:       "Pacific",
	},
}

// NewTimezoneManager ?func NewTimezoneManager(defaultTimezone string) (*TimezoneManager, error) {
	manager := &TimezoneManager{
		defaultTimezone: defaultTimezone,
		timezones:       make(map[string]*time.Location),
	}

	// 
	for tzID := range SupportedTimezones {
		loc, err := time.LoadLocation(tzID)
		if err != nil {
			return nil, fmt.Errorf("failed to load timezone %s: %w", tzID, err)
		}
		manager.timezones[tzID] = loc
	}

	return manager, nil
}

// ConvertTime 䵽?func (tm *TimezoneManager) ConvertTime(t time.Time, targetTimezone string) (time.Time, error) {
	tm.mutex.RLock()
	loc, exists := tm.timezones[targetTimezone]
	tm.mutex.RUnlock()

	if !exists {
		return t, fmt.Errorf("unsupported timezone: %s", targetTimezone)
	}

	return t.In(loc), nil
}

// ConvertTimeFromUTC UTC?func (tm *TimezoneManager) ConvertTimeFromUTC(utcTime time.Time, targetTimezone string) (time.Time, error) {
	return tm.ConvertTime(utcTime.UTC(), targetTimezone)
}

// ConvertTimeToUTC UTC
func (tm *TimezoneManager) ConvertTimeToUTC(localTime time.Time, sourceTimezone string) (time.Time, error) {
	tm.mutex.RLock()
	loc, exists := tm.timezones[sourceTimezone]
	tm.mutex.RUnlock()

	if !exists {
		return localTime, fmt.Errorf("unsupported timezone: %s", sourceTimezone)
	}

	// 
	if localTime.Location() == time.UTC {
		localTime = time.Date(
			localTime.Year(), localTime.Month(), localTime.Day(),
			localTime.Hour(), localTime.Minute(), localTime.Second(),
			localTime.Nanosecond(), loc,
		)
	}

	return localTime.UTC(), nil
}

// FormatTime ?func (tm *TimezoneManager) FormatTime(t time.Time, timezone, format, locale string) (string, error) {
	// ?	localTime, err := tm.ConvertTime(t, timezone)
	if err != nil {
		return "", err
	}

	// locale
	switch locale {
	case "zh-CN":
		return tm.formatTimeZhCN(localTime, format), nil
	case "en-US":
		return tm.formatTimeEnUS(localTime, format), nil
	case "ja-JP":
		return tm.formatTimeJaJP(localTime, format), nil
	case "ko-KR":
		return tm.formatTimeKoKR(localTime, format), nil
	case "fr-FR":
		return tm.formatTimeFrFR(localTime, format), nil
	case "de-DE":
		return tm.formatTimeDeDE(localTime, format), nil
	default:
		return localTime.Format(format), nil
	}
}

// formatTimeZhCN ?func (tm *TimezoneManager) formatTimeZhCN(t time.Time, format string) string {
	switch format {
	case "date":
		return t.Format("2006?1?2?)
	case "time":
		return t.Format("15:04:05")
	case "datetime":
		return t.Format("2006?1?2?15:04:05")
	case "short":
		return t.Format("01-02 15:04")
	case "long":
		weekdays := []string{"", "", "?, "?, "?, "?, "?, "?}
		weekday := weekdays[t.Weekday()]
		return fmt.Sprintf("%s %s", t.Format("2006?1?2?), weekday)
	default:
		return t.Format(format)
	}
}

// formatTimeEnUS ?func (tm *TimezoneManager) formatTimeEnUS(t time.Time, format string) string {
	switch format {
	case "date":
		return t.Format("January 2, 2006")
	case "time":
		return t.Format("3:04:05 PM")
	case "datetime":
		return t.Format("January 2, 2006 3:04:05 PM")
	case "short":
		return t.Format("01/02 3:04 PM")
	case "long":
		return t.Format("Monday, January 2, 2006")
	default:
		return t.Format(format)
	}
}

// formatTimeJaJP ?func (tm *TimezoneManager) formatTimeJaJP(t time.Time, format string) string {
	switch format {
	case "date":
		return t.Format("2006?1?2?)
	case "time":
		return t.Format("15?4?5?)
	case "datetime":
		return t.Format("2006?1?2?15?4?5?)
	case "short":
		return t.Format("01/02 15:04")
	case "long":
		weekdays := []string{"", "?, "?, "?, "?, "?, "?, "?}
		weekday := weekdays[t.Weekday()]
		return fmt.Sprintf("%s %s", t.Format("2006?1?2?), weekday)
	default:
		return t.Format(format)
	}
}

// formatTimeKoKR ?func (tm *TimezoneManager) formatTimeKoKR(t time.Time, format string) string {
	switch format {
	case "date":
		return t.Format("2006?01?02?)
	case "time":
		return t.Format("15?04?05?)
	case "datetime":
		return t.Format("2006?01?02?15?04?05?)
	case "short":
		return t.Format("01/02 15:04")
	case "long":
		weekdays := []string{"", "?, "?, "?, "?, "?, "?, "?}
		weekday := weekdays[t.Weekday()]
		return fmt.Sprintf("%s %s", t.Format("2006?01?02?), weekday)
	default:
		return t.Format(format)
	}
}

// formatTimeFrFR ?func (tm *TimezoneManager) formatTimeFrFR(t time.Time, format string) string {
	switch format {
	case "date":
		months := []string{"", "janvier", "fvrier", "mars", "avril", "mai", "juin",
			"juillet", "aot", "septembre", "octobre", "novembre", "dcembre"}
		return fmt.Sprintf("%d %s %d", t.Day(), months[t.Month()], t.Year())
	case "time":
		return t.Format("15:04:05")
	case "datetime":
		months := []string{"", "janvier", "fvrier", "mars", "avril", "mai", "juin",
			"juillet", "aot", "septembre", "octobre", "novembre", "dcembre"}
		return fmt.Sprintf("%d %s %d %s", t.Day(), months[t.Month()], t.Year(), t.Format("15:04:05"))
	case "short":
		return t.Format("02/01 15:04")
	case "long":
		weekdays := []string{"", "lundi", "mardi", "mercredi", "jeudi", "vendredi", "samedi", "dimanche"}
		months := []string{"", "janvier", "fvrier", "mars", "avril", "mai", "juin",
			"juillet", "aot", "septembre", "octobre", "novembre", "dcembre"}
		return fmt.Sprintf("%s %d %s %d", weekdays[t.Weekday()], t.Day(), months[t.Month()], t.Year())
	default:
		return t.Format(format)
	}
}

// formatTimeDeDE ?func (tm *TimezoneManager) formatTimeDeDE(t time.Time, format string) string {
	switch format {
	case "date":
		return t.Format("02.01.2006")
	case "time":
		return t.Format("15:04:05")
	case "datetime":
		return t.Format("02.01.2006 15:04:05")
	case "short":
		return t.Format("02.01 15:04")
	case "long":
		weekdays := []string{"", "Montag", "Dienstag", "Mittwoch", "Donnerstag", "Freitag", "Samstag", "Sonntag"}
		return fmt.Sprintf("%s, %s", weekdays[t.Weekday()], t.Format("02.01.2006"))
	default:
		return t.Format(format)
	}
}

// GetTimezoneInfo 
func (tm *TimezoneManager) GetTimezoneInfo(timezone string) (TimezoneInfo, error) {
	info, exists := SupportedTimezones[timezone]
	if !exists {
		return TimezoneInfo{}, fmt.Errorf("unsupported timezone: %s", timezone)
	}
	return info, nil
}

// GetTimezonesByRegion ?func (tm *TimezoneManager) GetTimezonesByRegion(region string) []TimezoneInfo {
	var timezones []TimezoneInfo
	for _, info := range SupportedTimezones {
		if strings.EqualFold(info.Region, region) {
			timezones = append(timezones, info)
		}
	}
	return timezones
}

// GetTimezonesByCountry ?func (tm *TimezoneManager) GetTimezonesByCountry(country string) []TimezoneInfo {
	var timezones []TimezoneInfo
	for _, info := range SupportedTimezones {
		if strings.EqualFold(info.Country, country) {
			timezones = append(timezones, info)
		}
	}
	return timezones
}

// GetAllTimezones 
func (tm *TimezoneManager) GetAllTimezones() []TimezoneInfo {
	var timezones []TimezoneInfo
	for _, info := range SupportedTimezones {
		timezones = append(timezones, info)
	}
	return timezones
}

// DetectTimezoneFromIP IPGeoIP?func (tm *TimezoneManager) DetectTimezoneFromIP(ctx context.Context, ip string) (string, error) {
	// GeoIP?	// 
	return tm.defaultTimezone, nil
}

// GetCurrentTime ?func (tm *TimezoneManager) GetCurrentTime(timezone string) (time.Time, error) {
	return tm.ConvertTime(time.Now(), timezone)
}

// GetCurrentTimeUTC UTC
func (tm *TimezoneManager) GetCurrentTimeUTC() time.Time {
	return time.Now().UTC()
}

// CalculateTimeDifference ?func (tm *TimezoneManager) CalculateTimeDifference(timezone1, timezone2 string) (time.Duration, error) {
	now := time.Now()
	
	time1, err := tm.ConvertTime(now, timezone1)
	if err != nil {
		return 0, err
	}
	
	time2, err := tm.ConvertTime(now, timezone2)
	if err != nil {
		return 0, err
	}
	
	return time1.Sub(time2), nil
}

// IsTimezoneSupported ?func (tm *TimezoneManager) IsTimezoneSupported(timezone string) bool {
	_, exists := SupportedTimezones[timezone]
	return exists
}

// GetBusinessHours ?func (tm *TimezoneManager) GetBusinessHours(timezone string) (BusinessHours, error) {
	// 
	businessHours := BusinessHours{
		Timezone: timezone,
		Monday:    DayHours{Start: "09:00", End: "18:00", Enabled: true},
		Tuesday:   DayHours{Start: "09:00", End: "18:00", Enabled: true},
		Wednesday: DayHours{Start: "09:00", End: "18:00", Enabled: true},
		Thursday:  DayHours{Start: "09:00", End: "18:00", Enabled: true},
		Friday:    DayHours{Start: "09:00", End: "18:00", Enabled: true},
		Saturday:  DayHours{Start: "10:00", End: "16:00", Enabled: false},
		Sunday:    DayHours{Start: "10:00", End: "16:00", Enabled: false},
	}
	
	return businessHours, nil
}

// BusinessHours 
type BusinessHours struct {
	Timezone  string   `json:"timezone"`
	Monday    DayHours `json:"monday"`
	Tuesday   DayHours `json:"tuesday"`
	Wednesday DayHours `json:"wednesday"`
	Thursday  DayHours `json:"thursday"`
	Friday    DayHours `json:"friday"`
	Saturday  DayHours `json:"saturday"`
	Sunday    DayHours `json:"sunday"`
}

// DayHours 
type DayHours struct {
	Start   string `json:"start"`   // : "HH:MM"
	End     string `json:"end"`     // : "HH:MM"
	Enabled bool   `json:"enabled"` // 
}

// IsBusinessTime ?func (tm *TimezoneManager) IsBusinessTime(t time.Time, timezone string) (bool, error) {
	businessHours, err := tm.GetBusinessHours(timezone)
	if err != nil {
		return false, err
	}
	
	localTime, err := tm.ConvertTime(t, timezone)
	if err != nil {
		return false, err
	}
	
	var dayHours DayHours
	switch localTime.Weekday() {
	case time.Monday:
		dayHours = businessHours.Monday
	case time.Tuesday:
		dayHours = businessHours.Tuesday
	case time.Wednesday:
		dayHours = businessHours.Wednesday
	case time.Thursday:
		dayHours = businessHours.Thursday
	case time.Friday:
		dayHours = businessHours.Friday
	case time.Saturday:
		dayHours = businessHours.Saturday
	case time.Sunday:
		dayHours = businessHours.Sunday
	}
	
	if !dayHours.Enabled {
		return false, nil
	}
	
	currentTime := localTime.Format("15:04")
	return currentTime >= dayHours.Start && currentTime <= dayHours.End, nil
}


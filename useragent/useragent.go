// Package useragent provides utilities for parsing and analyzing HTTP User-Agent strings.
package useragent

import (
	"fmt"
	"regexp"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// UserAgent contains the parsed information from a user agent string
type UserAgent struct {
	userAgent  string
	deviceType string
	deviceModel string
	os         string
	browserName string
	browserVer  string
	isBot      bool
	isMobile   bool
	isDesktop  bool
	isTablet   bool
	isTV       bool
	isConsole  bool
	isUnknown  bool
}

// String returns the user agent as a string
func (ua UserAgent) String() string { return ua.userAgent }

// UserAgent returns the full user agent string
func (ua UserAgent) UserAgent() string { return ua.userAgent }

// DeviceType returns the device type (mobile, desktop, tablet, bot, unknown)
func (ua UserAgent) DeviceType() string { return ua.deviceType }

// DeviceModel returns the specific device model if available
func (ua UserAgent) DeviceModel() string { return ua.deviceModel }

// OS returns the operating system name
func (ua UserAgent) OS() string { return ua.os }

// BrowserName returns the browser name
func (ua UserAgent) BrowserName() string { return ua.browserName }

// BrowserVer returns the browser version
func (ua UserAgent) BrowserVer() string { return ua.browserVer }

// BrowserInfo returns the browser name and version
func (ua UserAgent) BrowserInfo() Browser { return Browser{Name: ua.browserName, Version: ua.browserVer} }

// IsBot returns true if the user agent is a bot
func (ua UserAgent) IsBot() bool { return ua.isBot }

// IsMobile returns true if the user agent is a mobile device
func (ua UserAgent) IsMobile() bool { return ua.isMobile }

// IsDesktop returns true if the user agent is a desktop device
func (ua UserAgent) IsDesktop() bool { return ua.isDesktop }

// IsTablet returns true if the user agent is a tablet device
func (ua UserAgent) IsTablet() bool { return ua.isTablet }

// IsTV returns true if the user agent is a TV device
func (ua UserAgent) IsTV() bool { return ua.isTV }

// IsConsole returns true if the user agent is a gaming console
func (ua UserAgent) IsConsole() bool { return ua.isConsole }

// IsUnknown returns true if the user agent is unknown
func (ua UserAgent) IsUnknown() bool { return ua.isUnknown }

// Bot name extraction keywords - direct mapping for common bots
var botNameMap = map[string]string{
	"googlebot":             "Googlebot",
	"bingbot":               "Bingbot",
	"yandexbot":             "Yandexbot",
	"baidubot":              "Baidubot",
	"twitterbot":            "Twitterbot",
	"facebookbot":           "Facebookbot",
	"facebookexternalhit":   "Facebook",
	"linkedinbot":           "Linkedinbot",
	"slackbot":              "Slackbot",
	"telegrambot":           "Telegrambot",
	"adsbot":                "AdsBot",
}

// Common bot name patterns compiled only once for efficiency
var botNamePatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)([a-z0-9\-_]+bot)`),
	regexp.MustCompile(`(?i)(google-structured-data)`),
	regexp.MustCompile(`(?i)([a-z0-9\-_]+spider)`),
	regexp.MustCompile(`(?i)([a-z0-9\-_]+crawler)`),
}

// extractBotName extracts the bot name from a user agent string
// Optimized version with fast-path checks for common bots
func extractBotName(userAgent string) string {
	defaultName := "Unknown Bot"
	lowerUA := strings.ToLower(userAgent)
	
	// Fast path: direct checks for most common bots
	if strings.Contains(lowerUA, "googlebot") {
		return "Googlebot"
	}
	
	// Check for other common bots directly
	for keyword, name := range botNameMap {
		if strings.Contains(lowerUA, keyword) {
			return name
		}
	}
	
	// Slower path: regex matching for dynamic extraction
	for _, pattern := range botNamePatterns {
		matches := pattern.FindStringSubmatch(userAgent)
		if len(matches) > 1 {
			// Use the first captured group as the bot name
			title := cases.Title(language.English)
			return title.String(strings.ToLower(matches[1]))
		} else if len(matches) == 1 {
			// Use the whole match if no capture group
			title := cases.Title(language.English)
			return title.String(strings.ToLower(matches[0]))
		}
	}
	
	return defaultName
}
	
// GetShortIdentifier returns a short human-readable identifier for the session
// Format: Browser/Version (OS, DeviceType) or Bot: BotName for bots
func (ua UserAgent) GetShortIdentifier() string {
	// Special case for bots
	if ua.IsBot() {
		botName := extractBotName(ua.userAgent)
		return fmt.Sprintf("Bot: %s", botName)
	}
	
	// For normal browsers
	browserName := ua.BrowserName()
	if browserName == "" || browserName == BrowserUnknown {
		browserName = "Unknown"
	}
	
	browserVersion := ua.BrowserVer()
	if browserVersion == "" {
		browserVersion = "?"
	} else if strings.Contains(browserVersion, ".") {
		// Get only the first 10 characters but make sure we don't end with a dot
		if len(browserVersion) > 10 {
			browserVersion = browserVersion[:10]
			// Make sure we don't end with a dot
			if browserVersion[len(browserVersion)-1] == '.' {
				browserVersion = browserVersion[:len(browserVersion)-1] + "1"
			}
		}
	}
	
	osName := ua.OS()
	if osName == "" || osName == OSUnknown {
		osName = "Unknown OS"
	}
	
	deviceType := ua.DeviceType()
	if deviceType == "" {
		deviceType = "unknown"
	}
	
	return fmt.Sprintf("%s/%s (%s, %s)", browserName, browserVersion, osName, deviceType)
}

// Parse parses a user agent string and returns a UserAgent struct
func Parse(ua string) (UserAgent, error) {
	if ua == "" {
		return New("", DeviceTypeUnknown, "", OSUnknown, BrowserUnknown, ""), ErrEmptyUserAgent
	}

	// Convert to lowercase for consistency in string matching
	lowerUA := strings.ToLower(ua)

	// Parse device type
	deviceType := ParseDeviceType(lowerUA)

	// Get device model for mobile and tablet devices
	deviceModel := GetDeviceModel(lowerUA, deviceType)

	// Parse OS
	os := ParseOS(lowerUA)

	// Parse browser
	browser := ParseBrowser(lowerUA)

	return New(ua, deviceType, deviceModel, os, browser.Name, browser.Version), nil
}

// New creates a new UserAgent with the provided parameters
func New(ua, deviceType, deviceModel, os, browserName, browserVer string) UserAgent {
	result := UserAgent{
		userAgent:  ua,
		deviceType: deviceType,
		deviceModel: deviceModel,
		os:         os,
		browserName: browserName,
		browserVer:  browserVer,
	}

	// Set boolean flags
	setDeviceFlags(&result)

	return result
}

// setDeviceFlags sets the boolean flags based on device type
func setDeviceFlags(ua *UserAgent) {
	switch ua.deviceType {
	case DeviceTypeBot:
		ua.isBot = true
	case DeviceTypeMobile:
		ua.isMobile = true
	case DeviceTypeTablet:
		ua.isTablet = true
	case DeviceTypeDesktop:
		ua.isDesktop = true
	case DeviceTypeTV:
		ua.isTV = true
	case DeviceTypeConsole:
		ua.isConsole = true
	case DeviceTypeUnknown:
		ua.isUnknown = true
	}
}
package useragent

import (
	"fmt"
	"strings"
)

type UserAgent struct {
	userAgent   string
	deviceType  string
	os          string
	browserName string
	browserVer  string
	isBot       bool
	isMobile    bool
	isDesktop   bool
	isTablet    bool
	isUnknown   bool
}

func (ua UserAgent) String() string      { return ua.userAgent }
func (ua UserAgent) UserAgent() string   { return ua.userAgent }
func (ua UserAgent) DeviceType() string  { return ua.deviceType }
func (ua UserAgent) OS() string          { return ua.os }
func (ua UserAgent) BrowserName() string { return ua.browserName }
func (ua UserAgent) BrowserVer() string  { return ua.browserVer }
func (ua UserAgent) BrowserInfo() string { return fmt.Sprintf("%s %s", ua.browserName, ua.browserVer) }
func (ua UserAgent) IsBot() bool         { return ua.isBot }
func (ua UserAgent) IsMobile() bool      { return ua.isMobile }
func (ua UserAgent) IsDesktop() bool     { return ua.isDesktop }
func (ua UserAgent) IsTablet() bool      { return ua.isTablet }
func (ua UserAgent) IsUnknown() bool     { return ua.isUnknown }

// Parse analyzes a User-Agent string and returns a UserAgent struct
// containing all parsed information about the user agent, including device type,
// operating system, browser details, and various boolean flags.
func Parse(userAgentString string) UserAgent {
	if userAgentString == "" {
		return UserAgent{
			userAgent:   "",
			deviceType:  "unknown",
			os:          "unknown",
			browserName: "unknown",
			browserVer:  "",
			isUnknown:   true,
		}
	}

	// Initialize UserAgent struct
	ua := UserAgent{
		userAgent: userAgentString,
	}

	// Parse device type
	ua.deviceType = ParseDeviceType(userAgentString)

	// Set boolean flags based on device type
	switch ua.deviceType {
	case "bot":
		ua.isBot = true
	case "mobile":
		ua.isMobile = true
	case "tablet":
		ua.isTablet = true
	case "desktop":
		ua.isDesktop = true
	case "unknown":
		ua.isUnknown = true
	}

	// Parse operating system
	ua.os = ParseOS(userAgentString)

	// Parse browser information
	browser := ParseBrowser(userAgentString)
	ua.browserName = browser.Name
	ua.browserVer = browser.Version

	return ua
}

// New creates a new UserAgent instance with the given parameters
func New(userAgent string, deviceType string, os string, browserName string, browserVer string) UserAgent {
	ua := UserAgent{
		userAgent:   userAgent,
		deviceType:  deviceType,
		os:          os,
		browserName: browserName,
		browserVer:  browserVer,
	}

	// Set boolean flags based on device type
	switch ua.deviceType {
	case "bot":
		ua.isBot = true
	case "mobile":
		ua.isMobile = true
	case "tablet":
		ua.isTablet = true
	case "desktop":
		ua.isDesktop = true
	case "unknown":
		ua.isUnknown = true
	}

	return ua
}

// ParseDeviceType analyzes a User-Agent string and identifies the device type.
// It returns one of the following values:
// - "bot": Search bot, crawler, or spider (e.g. Googlebot, Bingbot)
// - "tablet": Tablet device (e.g. iPad, Android tablet)
// - "mobile": Mobile phone (e.g. iPhone, Android phone)
// - "desktop": Desktop/laptop computer (e.g. Windows, Mac)
// - "unknown": Unable to determine device type
func ParseDeviceType(ua string) string {
	ua = strings.ToLower(ua)

	// Check for bots first
	botKeywords := []string{
		"bot", "crawler", "spider", "googlebot", "bingbot", "baiduspider",
		"yandexbot", "slurp", "duckduckbot", "facebookexternalhit", "applebot",
		"twitterbot",
	}
	for _, keyword := range botKeywords {
		if strings.Contains(ua, keyword) {
			return "bot"
		}
	}

	// Check for tablets
	tabletKeywords := []string{
		"tablet", "ipad", "kindle fire", "nexus 7", "nexus 9", "nexus 10",
		"galaxy tab", "surface",
	}
	for _, keyword := range tabletKeywords {
		if strings.Contains(ua, keyword) {
			return "tablet"
		}
	}

	// Check for mobile devices
	mobileKeywords := []string{
		"mobile", "android", "iphone", "ipod", "blackberry", "windows phone",
		"opera mini", "opera mobi", "webos",
	}
	for _, keyword := range mobileKeywords {
		if strings.Contains(ua, keyword) {
			return "mobile"
		}
	}

	// Check for desktop
	desktopKeywords := []string{
		"windows", "macintosh", "linux", "ubuntu", "debian", "fedora", "redhat",
	}
	for _, keyword := range desktopKeywords {
		if strings.Contains(ua, keyword) {
			return "desktop"
		}
	}

	return "unknown"
}

// ParseOS analyzes a User-Agent string and identifies the operating system.
// It returns one of the following values:
// - "windows": Microsoft Windows
// - "macos": Apple macOS
// - "ios": Apple iOS
// - "android": Google Android
// - "linux": Linux distributions
// - "chrome_os": Google Chrome OS
// - "unknown": Unable to determine operating system
func ParseOS(ua string) string {
	ua = strings.ToLower(ua)

	switch {
	case strings.Contains(ua, "windows phone"):
		return "windows_phone"
	case strings.Contains(ua, "windows") || strings.Contains(ua, "win64") || strings.Contains(ua, "win32"):
		return "windows"
	case strings.Contains(ua, "android"):
		return "android"
	case strings.Contains(ua, "ipad") || strings.Contains(ua, "ipod") || strings.Contains(ua, "iphone"):
		return "ios"
	case strings.Contains(ua, "macintosh") || strings.Contains(ua, "mac os x"):
		return "macos"
	case strings.Contains(ua, "chrome os"):
		return "chrome_os"
	case strings.Contains(ua, "linux") ||
		strings.Contains(ua, "ubuntu") ||
		strings.Contains(ua, "debian") ||
		strings.Contains(ua, "fedora") ||
		strings.Contains(ua, "red hat") ||
		strings.Contains(ua, "suse"):
		return "linux"
	default:
		return "unknown"
	}
}

// Browser represents a web browser with its name and version
type Browser struct {
	Name    string // Browser name (e.g., "chrome", "firefox", "safari")
	Version string // Browser version (e.g., "58.0.3029.110")
}

// ParseBrowser analyzes a User-Agent string and identifies the web browser and version.
// It returns a Browser struct containing the browser name and version.
// Possible browser names include:
// - "chrome": Google Chrome
// - "firefox": Mozilla Firefox
// - "safari": Apple Safari
// - "edge": Microsoft Edge
// - "opera": Opera
// - "ie": Internet Explorer
// - "unknown": Unable to determine browser
func ParseBrowser(ua string) Browser {
	ua = strings.ToLower(ua)

	// Helper function to extract version from UA string
	extractVersion := func(ua, marker string) string {
		if idx := strings.Index(ua, marker); idx != -1 {
			versionStart := idx + len(marker)
			versionEnd := versionStart
			for i := versionStart; i < len(ua); i++ {
				if ua[i] == ' ' || ua[i] == ';' || ua[i] == ')' {
					break
				}
				versionEnd++
			}
			if versionEnd > versionStart {
				return ua[versionStart:versionEnd]
			}
		}
		return ""
	}

	switch {
	case strings.Contains(ua, "edge/") || strings.Contains(ua, "edg/"):
		return Browser{
			Name:    "edge",
			Version: extractVersion(ua, "edg/"),
		}

	case strings.Contains(ua, "chrome/"):
		return Browser{
			Name:    "chrome",
			Version: extractVersion(ua, "chrome/"),
		}

	case strings.Contains(ua, "firefox/"):
		return Browser{
			Name:    "firefox",
			Version: extractVersion(ua, "firefox/"),
		}

	case strings.Contains(ua, "safari/"):
		// Make sure it's not Chrome or Firefox masquerading as Safari
		if !strings.Contains(ua, "chrome/") && !strings.Contains(ua, "firefox/") {
			// Safari version is usually specified in "Version/" rather than "Safari/"
			version := extractVersion(ua, "version/")
			if version == "" {
				version = extractVersion(ua, "safari/")
			}
			return Browser{
				Name:    "safari",
				Version: version,
			}
		}

	case strings.Contains(ua, "opr/") || strings.Contains(ua, "opera/"):
		version := extractVersion(ua, "opr/")
		if version == "" {
			version = extractVersion(ua, "opera/")
		}
		return Browser{
			Name:    "opera",
			Version: version,
		}

	case strings.Contains(ua, "msie "):
		return Browser{
			Name:    "ie",
			Version: extractVersion(ua, "msie "),
		}

	case strings.Contains(ua, "trident/"):
		// IE 11 doesn't use "MSIE" anymore
		return Browser{
			Name:    "ie",
			Version: "11.0",
		}
	}

	return Browser{
		Name:    "unknown",
		Version: "",
	}
}

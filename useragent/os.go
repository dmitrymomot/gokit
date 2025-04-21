package useragent

import (
	"strings"
)

// OS detection keyword maps for faster lookups
var (
	windowsPhoneKeywords = makeKeywordMap([]string{"windows phone"})
	windowsKeywords     = makeKeywordMap([]string{"windows"})
	iOSKeywords         = makeKeywordMap([]string{"iphone", "ipad", "ipod"})
	macOSKeywords       = makeKeywordMap([]string{"macintosh", "mac os x"})
	harmonyOSKeywords   = makeKeywordMap([]string{"harmonyos"})
	androidKeywords     = makeKeywordMap([]string{"android"})
	fireOSKeywords      = makeKeywordMap([]string{"kindle", "silk"})
	chromeOSKeywords    = makeKeywordMap([]string{"cros", "chromeos", "chrome os"})
	linuxKeywords       = makeKeywordMap([]string{"linux", "ubuntu", "debian", "fedora", "mint", "x11"})
)

// ParseOS determines the operating system from a user agent string
// Optimized version using map-based lookups for faster performance
func ParseOS(lowerUA string) string {
	if lowerUA == "" {
		return OSUnknown
	}

	// Order checks by frequency for typical traffic patterns
	// Windows is most common in desktop traffic
	if strings.Contains(lowerUA, "windows") {
		if strings.Contains(lowerUA, "windows phone") {
			return OSWindowsPhone
		}
		return OSWindows
	}

	// iOS and macOS checks
	if strings.Contains(lowerUA, "iphone") || strings.Contains(lowerUA, "ipad") || strings.Contains(lowerUA, "ipod") {
		return OSiOS
	}

	if strings.Contains(lowerUA, "macintosh") || strings.Contains(lowerUA, "mac os x") {
		return OSMacOS
	}

	// Android is very common in mobile
	if strings.Contains(lowerUA, "android") {
		return OSAndroid
	}

	// Less common OS checks
	// Use direct string checks for frequently occurring patterns
	if strings.Contains(lowerUA, "linux") || strings.Contains(lowerUA, "ubuntu") {
		return OSLinux
	}

	// Use map lookups for less common patterns to reduce code size
	if hasAnyKeyword(lowerUA, harmonyOSKeywords) {
		return OSHarmonyOS
	}

	if hasAnyKeyword(lowerUA, fireOSKeywords) {
		return OSFireOS
	}

	if hasAnyKeyword(lowerUA, chromeOSKeywords) {
		return OSChromeOS
	}

	if hasAnyKeyword(lowerUA, linuxKeywords) {
		return OSLinux
	}

	return OSUnknown
}
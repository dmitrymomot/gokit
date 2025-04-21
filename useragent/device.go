package useragent

import (
	"strings"
)

// Device type keyword maps for faster lookups
var (
	// Maps for device type detection
	botKeywords     = makeKeywordMap([]string{"bot", "spider", "crawler", "archiver", "ping", "lighthouse", "slurp", "daum", "sogou", "yeti", "facebook", "twitter", "slack", "linkedin", "whatsapp", "telegram", "discord", "camo asset", "generator", "monitor", "analyzer", "validator", "fetcher", "scraper", "check"})
	tvKeywords      = makeKeywordMap([]string{"tv", "appletv", "smarttv", "googletv", "android tv", "webos", "tizen"})
	consoleKeywords = makeKeywordMap([]string{"playstation", "xbox", "nintendo", "wiiu", "switch"})
	tabletKeywords  = makeKeywordMap([]string{"tablet", "kindle", "silk"})
	mobileKeywords  = makeKeywordMap([]string{"mobile", "iphone", "android", "windows phone", "iemobile", "blackberry", "nokia"})
	desktopKeywords = makeKeywordMap([]string{"windows", "macintosh", "mac os x", "linux", "x11", "ubuntu", "fedora", "debian", "chromeos", "cros"})

	// Maps for mobile device models
	iPhoneKeywords     = makeKeywordMap([]string{"iphone"})
	samsungMobileWords = makeKeywordMap([]string{"samsung", "sm-g", "sm-a", "sm-n", "samsungbrowser"})
	huaweiMobileWords  = makeKeywordMap([]string{"huawei", "hwa-", "honor", "h60-", "h30-"})
	xiaomiMobileWords  = makeKeywordMap([]string{"xiaomi", "mi ", "redmi", "miui"})
	oppoMobileWords    = makeKeywordMap([]string{"oppo", "cph1", "cph2", "f1f"})
	vivoMobileWords    = makeKeywordMap([]string{"vivo", "viv-", "v1730", "v1731"})

	// Maps for tablet device models
	iPadKeywords      = makeKeywordMap([]string{"ipad"})
	surfaceWords      = makeKeywordMap([]string{"windows touch", "windows tablet"})
	samsungTabletWords = makeKeywordMap([]string{"sm-t", "gt-p", "sm-p"})
	huaweiTabletWords  = makeKeywordMap([]string{"mediapad", "agassi"})
	kindleWords        = makeKeywordMap([]string{"kindle", "silk", "kftt", "kfjwi"})
)

// makeKeywordMap creates a map from a slice of keywords for fast lookups
func makeKeywordMap(keywords []string) map[string]struct{} {
	result := make(map[string]struct{}, len(keywords))
	for _, word := range keywords {
		result[word] = struct{}{}
	}
	return result
}

// hasAnyKeyword checks if the user agent contains any of the keywords in the map
func hasAnyKeyword(ua string, keywordMap map[string]struct{}) bool {
	for keyword := range keywordMap {
		if strings.Contains(ua, keyword) {
			return true
		}
	}
	return false
}

// ParseDeviceType determines the device type from a user agent string
// Optimized version using hash map lookups instead of regex
func ParseDeviceType(lowerUA string) string {
	if lowerUA == "" {
		return DeviceTypeUnknown
	}

	// First check for iOS devices which are very common
	// iPad is always a tablet
	if strings.Contains(lowerUA, "ipad") {
		return DeviceTypeTablet
	}

	// iPhone is always mobile
	if strings.Contains(lowerUA, "iphone") {
		return DeviceTypeMobile
	}

	// Check for bots
	if hasAnyKeyword(lowerUA, botKeywords) {
		return DeviceTypeBot
	}

	// Android tablets don't have "Mobile" in their user agent string
	if strings.Contains(lowerUA, "android") {
		if !strings.Contains(lowerUA, "mobile") {
			return DeviceTypeTablet
		} else {
			return DeviceTypeMobile
		}
	}

	// Check for tablets
	if hasAnyKeyword(lowerUA, tabletKeywords) {
		return DeviceTypeTablet
	}

	// Check for mobile devices
	if hasAnyKeyword(lowerUA, mobileKeywords) {
		return DeviceTypeMobile
	}

	// Check for TV
	if hasAnyKeyword(lowerUA, tvKeywords) {
		return DeviceTypeTV
	}

	// Check for gaming consoles
	if hasAnyKeyword(lowerUA, consoleKeywords) {
		return DeviceTypeConsole
	}

	// Windows tablets check - must come before desktop check
	if strings.Contains(lowerUA, "windows") && 
		(strings.Contains(lowerUA, "touch") || strings.Contains(lowerUA, "tablet")) {
		return DeviceTypeTablet
	}
	
	// Check for desktop (most common)
	if hasAnyKeyword(lowerUA, desktopKeywords) {
		return DeviceTypeDesktop
	}

	return DeviceTypeUnknown
}

// GetDeviceModel determines the specific device model from a user agent string
// Optimized version using hash map lookups instead of regex
func GetDeviceModel(lowerUA, deviceType string) string {
	// If not a mobile or tablet device, return empty string
	if deviceType != DeviceTypeMobile && deviceType != DeviceTypeTablet {
		return ""
	}

	// Check for mobile device models
	if deviceType == DeviceTypeMobile {
		// Check in order of popularity for early returns
		if strings.Contains(lowerUA, "iphone") {
			return MobileDeviceIPhone
		}

		if hasAnyKeyword(lowerUA, samsungMobileWords) {
			return MobileDeviceSamsung
		}

		if hasAnyKeyword(lowerUA, huaweiMobileWords) {
			return MobileDeviceHuawei
		}

		if hasAnyKeyword(lowerUA, xiaomiMobileWords) {
			return MobileDeviceXiaomi
		}

		if hasAnyKeyword(lowerUA, oppoMobileWords) {
			return MobileDeviceOppo
		}

		if hasAnyKeyword(lowerUA, vivoMobileWords) {
			return MobileDeviceVivo
		}

		// Generic Android
		if strings.Contains(lowerUA, "android") {
			return MobileDeviceAndroid
		}

		return MobileDeviceUnknown
	}

	// Check for tablet device models
	if deviceType == DeviceTypeTablet {
		// Check in order of popularity
		if strings.Contains(lowerUA, "ipad") {
			return TabletDeviceIPad
		}

		// Surface tablets
		if strings.Contains(lowerUA, "windows") && 
			(strings.Contains(lowerUA, "touch") || strings.Contains(lowerUA, "tablet")) {
			return TabletDeviceSurface
		}

		if strings.Contains(lowerUA, "samsung") || hasAnyKeyword(lowerUA, samsungTabletWords) {
			return TabletDeviceSamsung
		}

		if strings.Contains(lowerUA, "huawei") || hasAnyKeyword(lowerUA, huaweiTabletWords) {
			return TabletDeviceHuawei
		}

		// Kindle Fire
		if hasAnyKeyword(lowerUA, kindleWords) {
			return TabletDeviceKindleFire
		}

		// Generic Android tablet
		if strings.Contains(lowerUA, "android") {
			return TabletDeviceAndroid
		}

		return TabletDeviceUnknown
	}

	return ""
}
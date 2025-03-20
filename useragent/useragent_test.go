package useragent_test

import (
	"testing"

	"github.com/dmitrymomot/gokit/useragent"
	"github.com/stretchr/testify/assert"
)

func TestParseDeviceType(t *testing.T) {
	tests := []struct {
		name     string
		ua       string
		expected string
	}{
		{
			name:     "Googlebot",
			ua:       "Mozilla/5.0 (compatible; Googlebot/2.1; +http://www.google.com/bot.html)",
			expected: "bot",
		},
		{
			name:     "iPad",
			ua:       "Mozilla/5.0 (iPad; CPU OS 14_4 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/14.0 Mobile/15E148 Safari/604.1",
			expected: "tablet",
		},
		{
			name:     "iPhone",
			ua:       "Mozilla/5.0 (iPhone; CPU iPhone OS 14_4 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/14.0 Mobile/15E148 Safari/604.1",
			expected: "mobile",
		},
		{
			name:     "Windows Desktop",
			ua:       "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36",
			expected: "desktop",
		},
		{
			name:     "Empty UA",
			ua:       "",
			expected: "unknown",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := useragent.ParseDeviceType(tc.ua)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestParseOS(t *testing.T) {
	tests := []struct {
		name     string
		ua       string
		expected string
	}{
		{
			name:     "Windows 10",
			ua:       "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36",
			expected: "windows",
		},
		{
			name:     "macOS",
			ua:       "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36",
			expected: "macos",
		},
		{
			name:     "iOS",
			ua:       "Mozilla/5.0 (iPhone; CPU iPhone OS 14_4 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/14.0 Mobile/15E148 Safari/604.1",
			expected: "ios",
		},
		{
			name:     "Android",
			ua:       "Mozilla/5.0 (Linux; Android 11; Pixel 5) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Mobile Safari/537.36",
			expected: "android",
		},
		{
			name:     "Linux",
			ua:       "Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:89.0) Gecko/20100101 Firefox/89.0",
			expected: "linux",
		},
		{
			name:     "Empty UA",
			ua:       "",
			expected: "unknown",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := useragent.ParseOS(tc.ua)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestParseBrowser(t *testing.T) {
	tests := []struct {
		name     string
		ua       string
		expected useragent.Browser
	}{
		{
			name: "Chrome",
			ua:   "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36",
			expected: useragent.Browser{
				Name:    "chrome",
				Version: "91.0.4472.124",
			},
		},
		{
			name: "Firefox",
			ua:   "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:89.0) Gecko/20100101 Firefox/89.0",
			expected: useragent.Browser{
				Name:    "firefox",
				Version: "89.0",
			},
		},
		{
			name: "Safari",
			ua:   "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/14.0.3 Safari/605.1.15",
			expected: useragent.Browser{
				Name:    "safari",
				Version: "14.0.3",
			},
		},
		{
			name: "Edge",
			ua:   "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36 Edg/91.0.864.59",
			expected: useragent.Browser{
				Name:    "edge",
				Version: "91.0.864.59",
			},
		},
		{
			name: "Empty UA",
			ua:   "",
			expected: useragent.Browser{
				Name:    "unknown",
				Version: "",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := useragent.ParseBrowser(tc.ua)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestParseUserAgent(t *testing.T) {
	tests := []struct {
		name     string
		ua       string
		expected useragent.UserAgent
	}{
		{
			name: "Desktop Chrome on Windows",
			ua:   "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36",
			expected: useragent.New(
				"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36",
				"desktop",
				"windows",
				"chrome",
				"91.0.4472.124",
			),
		},
		{
			name: "Mobile Safari on iPhone",
			ua:   "Mozilla/5.0 (iPhone; CPU iPhone OS 14_4 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/14.0 Mobile/15E148 Safari/604.1",
			expected: useragent.New(
				"Mozilla/5.0 (iPhone; CPU iPhone OS 14_4 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/14.0 Mobile/15E148 Safari/604.1",
				"mobile",
				"ios",
				"safari",
				"14.0",
			),
		},
		{
			name: "Googlebot",
			ua:   "Mozilla/5.0 (compatible; Googlebot/2.1; +http://www.google.com/bot.html)",
			expected: useragent.New(
				"Mozilla/5.0 (compatible; Googlebot/2.1; +http://www.google.com/bot.html)",
				"bot",
				"unknown",
				"unknown",
				"",
			),
		},
		{
			name: "Empty UA",
			ua:   "",
			expected: useragent.New(
				"",
				"unknown",
				"unknown",
				"unknown",
				"",
			),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := useragent.Parse(tc.ua)

			// Use getter methods to compare values
			assert.Equal(t, tc.expected.UserAgent(), result.UserAgent())
			assert.Equal(t, tc.expected.DeviceType(), result.DeviceType())
			assert.Equal(t, tc.expected.OS(), result.OS())
			assert.Equal(t, tc.expected.BrowserName(), result.BrowserName())
			assert.Equal(t, tc.expected.BrowserVer(), result.BrowserVer())
			assert.Equal(t, tc.expected.IsBot(), result.IsBot())
			assert.Equal(t, tc.expected.IsMobile(), result.IsMobile())
			assert.Equal(t, tc.expected.IsDesktop(), result.IsDesktop())
			assert.Equal(t, tc.expected.IsTablet(), result.IsTablet())
			assert.Equal(t, tc.expected.IsUnknown(), result.IsUnknown())
		})
	}
}

// TestNewUserAgent tests the NewUserAgent constructor
func TestNewUserAgent(t *testing.T) {
	ua := useragent.New(
		"test-ua",
		"mobile",
		"ios",
		"safari",
		"15.0",
	)

	assert.Equal(t, "test-ua", ua.UserAgent())
	assert.Equal(t, "mobile", ua.DeviceType())
	assert.Equal(t, "ios", ua.OS())
	assert.Equal(t, "safari", ua.BrowserName())
	assert.Equal(t, "15.0", ua.BrowserVer())
	assert.True(t, ua.IsMobile())
	assert.False(t, ua.IsDesktop())
	assert.False(t, ua.IsTablet())
	assert.False(t, ua.IsBot())
	assert.False(t, ua.IsUnknown())
}

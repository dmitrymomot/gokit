# User Agent Parser

A lightweight, efficient Go package for parsing and analyzing User-Agent strings from HTTP requests.

## Overview

The `useragent` package provides tools to parse and extract information from User-Agent strings, including:

- Device type detection (mobile, desktop, tablet, bot)
- Operating system identification
- Browser name and version extraction
- Convenient accessor methods for all parsed information

## Usage

### Basic Usage

```go
import "github.com/yourusername/gokit/useragent"

func handleRequest(w http.ResponseWriter, r *http.Request) {
    // Parse the User-Agent string
    ua, err := useragent.Parse(r.Header.Get("User-Agent"))
    if err != nil {
        // Handle error (empty user agent string)
        return
    }
    
    // Access information about the user agent
    fmt.Println("Device Type:", ua.DeviceType())
    fmt.Println("Operating System:", ua.OS())
    fmt.Println("Browser:", ua.BrowserInfo())
    
    // Get a human-readable identifier for logging or session display
    fmt.Println("Session:", ua.GetShortIdentifier()) // e.g. "Chrome/91.0 (Windows, desktop)"
    
    // Use boolean flags
    if ua.IsMobile() {
        // Handle mobile-specific logic
    } else if ua.IsDesktop() {
        // Handle desktop-specific logic
    } else if ua.IsBot() {
        // Handle bot-specific logic
    }
}
```

### Creating Custom User Agents

```go
// Create a custom user agent
ua := useragent.New(
    "Custom User Agent String",
    useragent.DeviceTypeMobile,
    useragent.OSAndroid,
    useragent.BrowserChrome,
    "88.0.4324.181"
)

// Use the custom user agent
fmt.Println(ua.String())
fmt.Println(ua.BrowserInfo())
```

### Direct Device Type Detection

```go
uaString := "Mozilla/5.0 (iPhone; CPU iPhone OS 14_4 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/14.0.3 Mobile/15E148 Safari/604.1"

// Just get the device type
deviceType := useragent.ParseDeviceType(strings.ToLower(uaString))
fmt.Println("Device Type:", deviceType) // Output: mobile

// Just get the operating system
os := useragent.ParseOS(strings.ToLower(uaString))
fmt.Println("OS:", os) // Output: ios

// Just get the browser info
browser := useragent.ParseBrowser(strings.ToLower(uaString))
fmt.Println("Browser:", browser.Name, browser.Version) // Output: safari 14.0.3
```

## Constants

The package provides various constants for device types, browser names, and operating systems:

```go
// Device types
const (
    DeviceTypeBot     = "bot"
    DeviceTypeMobile  = "mobile"
    DeviceTypeTablet  = "tablet"
    DeviceTypeDesktop = "desktop"
    DeviceTypeUnknown = "unknown"
)

// Browser names
const (
    BrowserChrome  = "chrome"
    BrowserFirefox = "firefox"
    BrowserSafari  = "safari"
    BrowserEdge    = "edge"
    BrowserOpera   = "opera"
    BrowserIE      = "ie"
    BrowserUnknown = "unknown"
)

// OS names
const (
    OSWindows      = "windows"
    OSWindowsPhone = "windows_phone"
    OSMacOS        = "macos"
    OSiOS          = "ios"
    OSAndroid      = "android"
    OSLinux        = "linux"
    OSChromeOS     = "chrome_os"
    OSUnknown      = "unknown"
)
```
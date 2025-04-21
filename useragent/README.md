# User Agent Parser

A high-performance, memory-efficient Go package for parsing and analyzing User-Agent strings from HTTP requests.

## Overview

The `useragent` package provides tools to parse and extract information from User-Agent strings, including:

- Device type detection (mobile, desktop, tablet, bot, TV, console)
- Device model identification (iPhone, Samsung, Huawei, etc.)
- Operating system identification (Windows, macOS, iOS, Android, HarmonyOS, FireOS, etc.)
- Browser detection with extensive support for multiple browsers
- Human-readable session identifiers for logging
- Optimized performance with minimal allocations

## Package Organization

The package has been organized into logical components for better maintainability:

- `constants.go` - All constant definitions
- `errors.go` - Error definitions
- `useragent.go` - Core UserAgent struct and methods
- `device.go` - Device type and model detection
- `os.go` - Operating system detection
- `browser.go` - Browser detection

## Performance

This package is highly optimized for performance. Benchmarks show impressive parsing speeds:

```
BenchmarkParse_ChromeDesktop     	  237394	      4923 ns/op	     161 B/op	       2 allocs/op
BenchmarkParse_Bot               	  643960	      1857 ns/op	      80 B/op	       1 allocs/op
BenchmarkParseDeviceType         	  810921	      1484 ns/op	     124 B/op	       1 allocs/op
BenchmarkGetDeviceModel          	 2893405	       413.9 ns/op	     136 B/op	       1 allocs/op
BenchmarkParseOS                 	 1998298	       601.0 ns/op	     124 B/op	       1 allocs/op
BenchmarkParseBrowser            	  437187	      2768 ns/op	     177 B/op	       2 allocs/op
BenchmarkGetShortIdentifier      	 5514026	       218.1 ns/op	     112 B/op	       4 allocs/op
```

## Usage

### Basic Usage

```go
import "github.com/dmitrymomot/gokit/useragent"

func handleRequest(w http.ResponseWriter, r *http.Request) {
    // Parse the User-Agent string
    ua, err := useragent.Parse(r.Header.Get("User-Agent"))
    if err != nil {
        // Handle error (empty user agent string)
        return
    }
    
    // Access information about the user agent
    fmt.Println("Device Type:", ua.DeviceType())
    fmt.Println("Device Model:", ua.DeviceModel()) // For mobile/tablet devices
    fmt.Println("Operating System:", ua.OS())
    fmt.Println("Browser:", ua.BrowserName(), ua.BrowserVer())
    
    // Get a human-readable identifier for logging or session display
    fmt.Println("Session:", ua.GetShortIdentifier()) // e.g. "Chrome/91.0 (Windows, desktop)"
    
    // Use boolean flags
    if ua.IsMobile() {
        // Handle mobile-specific logic
    } else if ua.IsDesktop() {
        // Handle desktop-specific logic
    } else if ua.IsTablet() {
        // Handle tablet-specific logic  
    } else if ua.IsBot() {
        // Handle bot-specific logic
    } else if ua.IsTV() {
        // Handle TV-specific logic
    } else if ua.IsConsole() {
        // Handle gaming console-specific logic
    }
}
```

### Creating Custom User Agents

```go
// Create a custom user agent
ua := useragent.New(
    "Custom User Agent String",
    useragent.DeviceTypeMobile,
    useragent.MobileDeviceIPhone, // Device model
    useragent.OSiOS,
    useragent.BrowserSafari,
    "15.0"
)

// Use the custom user agent
fmt.Println(ua.String())
fmt.Println(ua.DeviceModel())
fmt.Println(ua.BrowserName(), ua.BrowserVer())
```

### Direct Component Detection

```go
uaString := "Mozilla/5.0 (iPhone; CPU iPhone OS 14_4 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/14.0.3 Mobile/15E148 Safari/604.1"
lowerUA := strings.ToLower(uaString) // Convert once for efficiency

// Just get the device type
deviceType := useragent.ParseDeviceType(lowerUA)
fmt.Println("Device Type:", deviceType) // Output: mobile

// Get device model (when applicable)
deviceModel := useragent.GetDeviceModel(lowerUA, deviceType)
fmt.Println("Device Model:", deviceModel) // Output: iphone

// Just get the operating system
os := useragent.ParseOS(lowerUA)
fmt.Println("OS:", os) // Output: ios

// Just get the browser info
browser := useragent.ParseBrowser(lowerUA)
fmt.Println("Browser:", browser.Name, browser.Version) // Output: safari 14.0.3
```

## Constants

The package provides an extensive set of constants for device types, device models, browser names, and operating systems:

### Device Types

```go
const (
    DeviceTypeBot     = "bot"
    DeviceTypeMobile  = "mobile"
    DeviceTypeTablet  = "tablet"
    DeviceTypeDesktop = "desktop"
    DeviceTypeTV      = "tv"
    DeviceTypeConsole = "console"
    DeviceTypeUnknown = "unknown"
)
```

### Device Models (Mobile)

```go
const (
    MobileDeviceIPhone   = "iphone"
    MobileDeviceAndroid  = "android"
    MobileDeviceSamsung  = "samsung"
    MobileDeviceHuawei   = "huawei"
    MobileDeviceXiaomi   = "xiaomi"
    MobileDeviceOppo     = "oppo"
    MobileDeviceVivo     = "vivo"
    MobileDeviceUnknown  = "unknown"
)
```

### Device Models (Tablet)

```go
const (
    TabletDeviceIPad       = "ipad"
    TabletDeviceAndroid    = "android"
    TabletDeviceSamsung    = "samsung"
    TabletDeviceHuawei     = "huawei"
    TabletDeviceKindleFire = "kindle"
    TabletDeviceSurface    = "surface"
    TabletDeviceUnknown    = "unknown"
)
```

### Browser Names

```go
const (
    BrowserChrome   = "chrome"
    BrowserFirefox  = "firefox"
    BrowserSafari   = "safari"
    BrowserEdge     = "edge"
    BrowserOpera    = "opera"
    BrowserIE       = "ie"
    BrowserSamsung  = "samsung"
    BrowserUC       = "uc"
    BrowserQQ       = "qq"
    BrowserHuawei   = "huawei"
    BrowserVivo     = "vivo"
    BrowserMIUI     = "miui"
    BrowserBrave    = "brave"
    BrowserVivaldi  = "vivaldi"
    BrowserYandex   = "yandex"
    BrowserUnknown  = "unknown"
)
```

### OS Names

```go
const (
    OSWindows      = "windows"
    OSWindowsPhone = "windows phone"
    OSMacOS        = "macos"
    OSiOS          = "ios"
    OSAndroid      = "android"
    OSLinux        = "linux"
    OSChromeOS     = "chromeos"
    OSHarmonyOS    = "harmonyos"
    OSFireOS       = "fireos"
    OSUnknown      = "unknown"
)
```
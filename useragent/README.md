# User Agent Parser

A high-performance, memory-efficient package for parsing and analyzing HTTP User-Agent strings.

## Installation

```bash
go get github.com/dmitrymomot/gokit/useragent
```

## Overview

The `useragent` package provides a fast, memory-efficient way to parse User-Agent strings from HTTP requests, extracting detailed information about browsers, operating systems, and devices with minimal allocations.

## Features

- Optimized for high performance and low memory usage
- Comprehensive device type detection (mobile/tablet/desktop/TV/console/bot)
- Accurate device model identification (iPhone, Samsung, Huawei, etc.)
- Operating system detection with version extraction
- Browser identification with version parsing
- Human-readable session identifiers for logging
- Minimal memory allocations and zero external dependencies

## Usage

### Basic Parsing

```go
import "github.com/dmitrymomot/gokit/useragent"

// Parse the User-Agent string from an HTTP request
ua, err := useragent.Parse(r.Header.Get("User-Agent"))
if err != nil {
    // Handle invalid or empty user agent
    return
}

// Access parsed information
fmt.Println("Device Type:", ua.DeviceType())        // "mobile", "desktop", "tablet", etc.
fmt.Println("Device Model:", ua.DeviceModel())      // "iphone", "samsung", "huawei", etc.
fmt.Println("OS:", ua.OS())                         // "ios", "android", "windows", etc.
fmt.Println("Browser:", ua.BrowserName())           // "chrome", "safari", "firefox", etc.
fmt.Println("Browser Version:", ua.BrowserVer())    // "91.0.4472.124", "15.0", etc.

// Get a concise identifier for logging
fmt.Println("Session:", ua.GetShortIdentifier())    // "Chrome/91.0 (Windows, desktop)"
```

### Device Type Detection

```go
// Check device type with boolean helpers
if ua.IsMobile() {
    // Handle mobile device logic
} else if ua.IsTablet() {
    // Handle tablet device logic
} else if ua.IsDesktop() {
    // Handle desktop logic
} else if ua.IsBot() {
    // Handle bot/crawler logic
} else if ua.IsTV() {
    // Handle smart TV logic
} else if ua.IsConsole() {
    // Handle gaming console logic
}

// Use the device type string
switch ua.DeviceType() {
case useragent.DeviceTypeMobile:
    // Mobile device
case useragent.DeviceTypeTablet:
    // Tablet device
case useragent.DeviceTypeDesktop:
    // Desktop computer
case useragent.DeviceTypeBot:
    // Bot/crawler
case useragent.DeviceTypeTV:
    // Smart TV
case useragent.DeviceTypeConsole:
    // Gaming console
}
```

### Individual Component Parsing

```go
// Parse only the components you need (more efficient)
uaString := r.Header.Get("User-Agent")
lowerUA := strings.ToLower(uaString) // Convert once for efficiency

// Get the device type
deviceType := useragent.ParseDeviceType(lowerUA)

// Get the device model when applicable
if deviceType == useragent.DeviceTypeMobile || deviceType == useragent.DeviceTypeTablet {
    deviceModel := useragent.GetDeviceModel(lowerUA, deviceType)
    fmt.Println("Model:", deviceModel) // "iphone", "samsung", etc.
}

// Get just the operating system
os := useragent.ParseOS(lowerUA)
fmt.Println("OS:", os) // "windows", "ios", "android", etc.

// Get just the browser information
browser := useragent.ParseBrowser(lowerUA)
fmt.Println("Browser:", browser.Name, browser.Version)
```

### Custom User Agents

```go
// Create a custom user agent for testing or modeling
customUA := useragent.New(
    "Custom User Agent String",
    useragent.DeviceTypeMobile,
    useragent.MobileDeviceIPhone,
    useragent.OSiOS,
    useragent.BrowserSafari,
    "15.4"
)

// Use the custom user agent
fmt.Println(customUA.String())           // Original string
fmt.Println(customUA.GetShortIdentifier()) // Short identifier
```

## API Reference

### Core Types

```go
// UserAgent represents a parsed user agent string
type UserAgent struct {
    // String representation of the user agent
    UserAgentString string
    
    // Device information
    deviceType  string
    deviceModel string
    
    // OS information
    os string
    
    // Browser information
    browser struct {
        Name    string
        Version string
    }
}

// Browser represents browser information
type Browser struct {
    Name    string
    Version string
}
```

### Main Functions

```go
// Parse a user agent string into a UserAgent struct
func Parse(userAgent string) (*UserAgent, error)

// Create a new UserAgent with specified attributes
func New(userAgent, deviceType, deviceModel, os, browserName, browserVer string) *UserAgent

// Parse only the device type from a lowercase user agent string
func ParseDeviceType(lowerUA string) string

// Determine the device model from a lowercase user agent and device type
func GetDeviceModel(lowerUA, deviceType string) string

// Parse only the operating system from a lowercase user agent string
func ParseOS(lowerUA string) string

// Parse only the browser information from a lowercase user agent string
func ParseBrowser(lowerUA string) Browser

// Get a short, human-readable identifier for the user agent
func (ua *UserAgent) GetShortIdentifier() string
```

### User Agent Methods

```go
// Get the device type (mobile, desktop, tablet, bot, tv, console)
func (ua *UserAgent) DeviceType() string

// Get the device model (iphone, samsung, etc.)
func (ua *UserAgent) DeviceModel() string

// Get the operating system name
func (ua *UserAgent) OS() string

// Get the browser name
func (ua *UserAgent) BrowserName() string

// Get the browser version
func (ua *UserAgent) BrowserVer() string

// Check if the device is mobile
func (ua *UserAgent) IsMobile() bool

// Check if the device is a tablet
func (ua *UserAgent) IsTablet() bool

// Check if the device is a desktop computer
func (ua *UserAgent) IsDesktop() bool

// Check if the device is a bot/crawler
func (ua *UserAgent) IsBot() bool

// Check if the device is a smart TV
func (ua *UserAgent) IsTV() bool

// Check if the device is a gaming console
func (ua *UserAgent) IsConsole() bool
```

### Constants

The package defines constants for device types, device models, browsers, and operating systems:

```go
// Device types
const (
    DeviceTypeBot     = "bot"
    DeviceTypeMobile  = "mobile"
    DeviceTypeTablet  = "tablet"
    DeviceTypeDesktop = "desktop"
    DeviceTypeTV      = "tv"
    DeviceTypeConsole = "console"
    DeviceTypeUnknown = "unknown"
)

// Mobile device models
const (
    MobileDeviceIPhone   = "iphone"
    MobileDeviceAndroid  = "android"
    MobileDeviceSamsung  = "samsung"
    MobileDeviceHuawei   = "huawei"
    MobileDeviceXiaomi   = "xiaomi"
    // ...and many more
)

// Browser names
const (
    BrowserChrome  = "chrome"
    BrowserFirefox = "firefox"
    BrowserSafari  = "safari"
    BrowserEdge    = "edge"
    // ...and many more
)

// Operating systems
const (
    OSWindows    = "windows"
    OSMacOS      = "macos"
    OSiOS        = "ios"
    OSAndroid    = "android"
    OSLinux      = "linux"
    OSHarmonyOS  = "harmonyos"
    // ...and many more
)
```

## Best Practices

1. **Performance Optimization**:
   - Convert the user agent string to lowercase only once
   - Use individual component parsers if you only need specific information
   - Reuse the UserAgent object when making multiple checks

2. **Context Usage**:
   - Store the UserAgent in the request context for reuse across handlers
   - Include the short identifier in logs for easy session tracking

3. **Error Handling**:
   - Always check for errors when parsing user agent strings
   - Handle the case of empty or invalid user agent strings

4. **Device Detection**:
   - Combine device type with OS information for the most accurate device detection
   - For responsive design, make decisions based on device type rather than specific models
# Semantic Versioning Package

A robust, SemVer 2.0.0 compliant version parser, validator, and comparator for Go.

## Installation

```bash
go get github.com/dmitrymomot/gokit/semver
```

## Overview

The `semver` package provides a clean, immutable implementation of the [Semantic Versioning 2.0.0 specification](https://semver.org/). It offers reliable tools for parsing, validating, comparing, and manipulating version strings in Go applications. This package is thread-safe due to its immutable design and suitable for concurrent use in production systems.

## Features

- Full SemVer 2.0.0 specification compliance
- Immutable `Version` type for thread safety
- Parse version strings with or without leading 'v'
- Compare versions for precedence with proper prerelease handling
- Increment major, minor, and patch components
- Manipulate version elements independently
- Check if a version is within a specific range
- Clear error types for validation failures

## Usage

### Parsing Versions

```go
import (
    "fmt"
    "github.com/dmitrymomot/gokit/semver"
)

// Standard parsing with error checking
version, err := semver.Parse("1.2.3")
if err != nil {
    // Handle error
    return fmt.Errorf("failed to parse version: %w", err)
}
// version represents "1.2.3"

// Also supports leading 'v'
version, err = semver.Parse("v2.1.0")
// Returns version representing "2.1.0"

// Parse with prerelease and build metadata
version, err = semver.Parse("1.2.3-beta.1+20130313144700")
// Returns version representing "1.2.3-beta.1+20130313144700"

// Parse without error checking (panics on invalid versions)
version = semver.MustParse("1.2.3")
// Returns version representing "1.2.3"
```

### Validating Versions

```go
// Check if a string is a valid SemVer
isValid := semver.Validate("1.2.3")
// Returns: true

isValid = semver.Validate("v1.2.3-beta+exp.sha.5114f85")
// Returns: true

// Validate a Version struct
version := semver.Version{Major: 1, Minor: 2, Patch: 3}
isValid = version.IsValid()
// Returns: true
```

### Comparing Versions

```go
v1 := semver.MustParse("1.2.3")
v2 := semver.MustParse("2.0.0")

// Core comparison methods
if v1.LessThan(v2) {
    // This code will execute since 1.2.3 < 2.0.0
}

if v1.GreaterThan(v2) {
    // This code will not execute
}

if v1.Equal(v2) {
    // This code will not execute
}

if v1.LessThanOrEqual(v2) {
    // This code will execute
}

if v1.GreaterThanOrEqual(v2) {
    // This code will not execute
}

// Check if a version is within a range (inclusive)
v3 := semver.MustParse("1.5.0")
if v3.InRange(v1, v2) {
    // This code will execute since 1.2.3 <= 1.5.0 <= 2.0.0
}

// Manual comparison
result := v1.Compare(v2)
// Returns: -1 because v1 < v2
```

### Manipulating Versions

```go
version := semver.MustParse("1.2.3-alpha+001")

// Create new versions with modified components
v2 := version.WithMajor(2)        // Returns: 2.2.3-alpha+001
v2 = version.WithMinor(5)         // Returns: 1.5.3-alpha+001
v2 = version.WithPatch(7)         // Returns: 1.2.7-alpha+001
v2 = version.WithPrerelease("beta") // Returns: 1.2.3-beta+001
v2 = version.WithBuild("002")     // Returns: 1.2.3-alpha+002

// Increment components (resets prerelease and build metadata)
v2, err := version.Increment("major") // Returns: 2.0.0
if err != nil {
    // Handle error
}

v2, err = version.Increment("minor")  // Returns: 1.3.0
if err != nil {
    // Handle error
}

v2, err = version.Increment("patch")  // Returns: 1.2.4
if err != nil {
    // Handle error
}

// Get string representations
str := version.String()         // Returns: "1.2.3-alpha+001"
str = version.MajorMinorPatch() // Returns: "1.2.3"
```

### Error Handling

```go
import (
    "errors"
    "fmt"
    "github.com/dmitrymomot/gokit/semver"
)

// Handling parsing errors
_, err := semver.Parse("invalid.version")
if err != nil {
    switch {
    case errors.Is(err, semver.ErrInvalidVersion):
        // Handle invalid version format
        fmt.Println("The version string format is invalid")
    case errors.Is(err, semver.ErrEmptyVersion):
        // Handle empty version string
        fmt.Println("The version string is empty")
    case errors.Is(err, semver.ErrInvalidPrerelease):
        // Handle invalid prerelease identifier
        fmt.Println("The prerelease identifier is invalid")
    case errors.Is(err, semver.ErrInvalidBuild):
        // Handle invalid build metadata
        fmt.Println("The build metadata is invalid")
    default:
        // Handle other errors
        fmt.Println("An unexpected error occurred:", err)
    }
}

// Checking version requirements
clientVersion := semver.MustParse("2.1.3")
minRequired := semver.MustParse("2.0.0")

if clientVersion.LessThan(minRequired) {
    return fmt.Errorf("client version %s is less than minimum required %s", 
        clientVersion, minRequired)
}
// Client version 2.1.3 is not less than minimum required 2.0.0, so this error won't be returned

// Check compatibility with a version range
minSupported := semver.MustParse("1.5.0")
maxSupported := semver.MustParse("3.0.0")

if !clientVersion.InRange(minSupported, maxSupported) {
    return errors.New("unsupported client version")
}
// Client version 2.1.3 is in range 1.5.0 to 3.0.0, so this error won't be returned
```

## Best Practices

1. **Version Semantics**:
   - Use major version zero (0.y.z) for initial development only
   - Increment major version (x.0.0) when making incompatible API changes
   - Increment minor version (x.y.0) when adding functionality in a backward compatible manner
   - Increment patch version (x.y.z) when making backward compatible bug fixes

2. **Prerelease and Build Identifiers**:
   - Use prerelease versions (e.g., `1.0.0-alpha.1`) for pre-release testing
   - Use build metadata (e.g., `1.0.0+20130313144700`) for build identification
   - Remember that build metadata does not affect version precedence

3. **Error Handling**:
   - Always check for errors when parsing user-supplied version strings
   - Use MustParse only for hard-coded versions or when panics are acceptable
   - Check specific error types for better error messages

4. **Comparisons**:
   - Use the proper comparison methods for semantic version comparison
   - Avoid string-based version comparisons which can lead to incorrect results
   - For version ranges, always use InRange rather than separate comparisons

## API Reference

### Types

```go
type Version struct {
    Major      uint64 // Major version component
    Minor      uint64 // Minor version component
    Patch      uint64 // Patch version component
    Prerelease string // Prerelease version component (e.g., "alpha.1")
    Build      string // Build metadata (e.g., "20130313144700")
}
```

### Functions

```go
func Parse(version string) (Version, error)
```
Parses a version string into a Version struct.

```go
func MustParse(version string) Version
```
Parse without error checking (panics on invalid input).

```go
func Validate(version string) bool
```
Validates a version string.

### Methods

```go
func (v Version) IsValid() bool
```
Checks if a Version struct is valid.

```go
func (v Version) Compare(other Version) int
```
Compares versions and returns -1, 0, or 1.

```go
func (v Version) LessThan(other Version) bool
```
Checks if version is less than other.

```go
func (v Version) GreaterThan(other Version) bool
```
Checks if version is greater than other.

```go
func (v Version) Equal(other Version) bool
```
Checks if versions are equal.

```go
func (v Version) LessThanOrEqual(other Version) bool
```
Checks if version is less than or equal to other.

```go
func (v Version) GreaterThanOrEqual(other Version) bool
```
Checks if version is greater than or equal to other.

```go
func (v Version) InRange(lower, upper Version) bool
```
Checks if version is within a range (inclusive).

```go
func (v Version) WithMajor(major uint64) Version
```
Creates a new version with a different major component.

```go
func (v Version) WithMinor(minor uint64) Version
```
Creates a new version with a different minor component.

```go
func (v Version) WithPatch(patch uint64) Version
```
Creates a new version with a different patch component.

```go
func (v Version) WithPrerelease(prerelease string) Version
```
Creates a new version with a different prerelease component.

```go
func (v Version) WithBuild(build string) Version
```
Creates a new version with a different build metadata.

```go
func (v Version) Increment(component string) (Version, error)
```
Increments a version component (major, minor, or patch).

```go
func (v Version) String() string
```
Returns the full version as a string.

```go
func (v Version) MajorMinorPatch() string
```
Returns the major.minor.patch portion as a string.

### Error Types

```go
var ErrInvalidVersion = errors.New("invalid semantic version")
var ErrInvalidPrerelease = errors.New("invalid prerelease identifier")
var ErrInvalidBuild = errors.New("invalid build metadata")
var ErrEmptyVersion = errors.New("empty version string")
var ErrNegativeValue = errors.New("negative value in version")
```

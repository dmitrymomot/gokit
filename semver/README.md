# Semantic Versioning Package

A robust, SemVer 2.0.0 compliant version parser, validator, and comparator for Go.

## Installation

```bash
go get github.com/dmitrymomot/gokit/semver
```

## Overview

The `semver` package provides a clean, immutable implementation of the [Semantic Versioning 2.0.0 specification](https://semver.org/). It offers reliable tools for parsing, validating, comparing, and manipulating version strings in Go applications.

## Features

- Full SemVer 2.0.0 specification compliance
- Immutable `Version` type for thread safety
- Parse version strings with or without leading 'v'
- Compare versions for precedence
- Increment major, minor, and patch components
- Manipulate version elements independently
- Check if a version is within a specific range
- Clear error types for validation failures

## Usage

### Parsing Versions

```go
import "github.com/dmitrymomot/gokit/semver"

// Standard parsing with error checking
version, err := semver.Parse("1.2.3")
if err != nil {
    // Handle error
}

// Also supports leading 'v'
version, err = semver.Parse("v2.1.0")

// Parse with prerelease and build metadata
version, err = semver.Parse("1.2.3-beta.1+20130313144700")

// Parse without error checking (panics on invalid versions)
version = semver.MustParse("1.2.3")
```

### Validating Versions

```go
// Check if a string is a valid SemVer
isValid := semver.Validate("1.2.3")
isValid = semver.Validate("v1.2.3-beta+exp.sha.5114f85")

// Validate a Version struct
version := semver.Version{Major: 1, Minor: 2, Patch: 3}
isValid = version.IsValid()
```

### Comparing Versions

```go
v1 := semver.MustParse("1.2.3")
v2 := semver.MustParse("2.0.0")

// Core comparison methods
if v1.LessThan(v2) {
    // v1 < v2
}

if v1.GreaterThan(v2) {
    // v1 > v2
}

if v1.Equal(v2) {
    // v1 == v2
}

if v1.LessThanOrEqual(v2) {
    // v1 <= v2
}

if v1.GreaterThanOrEqual(v2) {
    // v1 >= v2
}

// Check if a version is within a range (inclusive)
v3 := semver.MustParse("1.5.0")
if v3.InRange(v1, v2) {
    // v1 <= v3 <= v2
}

// Manual comparison
result := v1.Compare(v2)
// result is -1 if v1 < v2
// result is  0 if v1 == v2
// result is  1 if v1 > v2
```

### Manipulating Versions

```go
version := semver.MustParse("1.2.3-alpha+001")

// Create new versions with modified components
v2 := version.WithMajor(2)        // 2.2.3-alpha+001
v2 = version.WithMinor(5)         // 1.5.3-alpha+001
v2 = version.WithPatch(7)         // 1.2.7-alpha+001
v2 = version.WithPrerelease("beta") // 1.2.3-beta+001
v2 = version.WithBuild("002")     // 1.2.3-alpha+002

// Increment components (resets prerelease and build metadata)
v2, err := version.Increment("major") // 2.0.0
v2, err = version.Increment("minor")  // 1.3.0
v2, err = version.Increment("patch")  // 1.2.4

// Get string representations
str := version.String()         // "1.2.3-alpha+001"
str = version.MajorMinorPatch() // "1.2.3"
```

### Working with APIs

```go
// API versioning with constraints
apiMinVersion := semver.MustParse("2.0.0")
clientVersion := semver.MustParse("2.1.3")

if clientVersion.LessThan(apiMinVersion) {
    return errors.New("client version too old, please upgrade")
}

// Check compatibility with a version range
minSupported := semver.MustParse("1.5.0")
maxSupported := semver.MustParse("3.0.0")

if !clientVersion.InRange(minSupported, maxSupported) {
    return errors.New("unsupported client version")
}
```

## API Reference

### Core Types

```go
// Main version type
type Version struct {
    Major      uint64 // Major version component
    Minor      uint64 // Minor version component
    Patch      uint64 // Patch version component
    Prerelease string // Prerelease version component (e.g., "alpha.1")
    Build      string // Build metadata (e.g., "20130313144700")
}
```

### Parsing and Validation

```go
// Parse a version string into a Version struct
func Parse(version string) (Version, error)

// Parse without error checking (panics on invalid input)
func MustParse(version string) Version

// Validate a version string
func Validate(version string) bool

// Check if a Version struct is valid
func (v Version) IsValid() bool
```

### Comparison Methods

```go
// Compare versions and return -1, 0, or 1
func (v Version) Compare(other Version) int

// Convenience comparison methods
func (v Version) LessThan(other Version) bool
func (v Version) GreaterThan(other Version) bool
func (v Version) Equal(other Version) bool
func (v Version) LessThanOrEqual(other Version) bool
func (v Version) GreaterThanOrEqual(other Version) bool
func (v Version) InRange(lower, upper Version) bool
```

### Manipulation Methods

```go
// Create new versions with modified components
func (v Version) WithMajor(major uint64) Version
func (v Version) WithMinor(minor uint64) Version
func (v Version) WithPatch(patch uint64) Version
func (v Version) WithPrerelease(prerelease string) Version
func (v Version) WithBuild(build string) Version

// Increment a version component
func (v Version) Increment(component string) (Version, error)

// String representations
func (v Version) String() string
func (v Version) MajorMinorPatch() string
```

### Error Types

```go
var (
    ErrInvalidVersion    = errors.New("invalid semantic version")
    ErrInvalidPrerelease = errors.New("invalid prerelease identifier")
    ErrInvalidBuild      = errors.New("invalid build metadata")
    ErrEmptyVersion      = errors.New("empty version string")
    ErrNegativeValue     = errors.New("negative value in version")
)
```

## SemVer Best Practices

1. **Major version zero (0.y.z)** is for initial development and may have breaking changes at any time
2. **Increment major version (x.0.0)** when making incompatible API changes
3. **Increment minor version (x.y.0)** when adding functionality in a backward compatible manner
4. **Increment patch version (x.y.z)** when making backward compatible bug fixes
5. **Use prerelease versions** (e.g., `1.0.0-alpha.1`) for pre-release testing
6. **Use build metadata** (e.g., `1.0.0+20130313144700`) for build identification

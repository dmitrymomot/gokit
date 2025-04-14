# Semantic Versioning (SemVer) Package

This package provides a robust implementation of the [Semantic Versioning 2.0.0 specification](https://semver.org/spec/v2.0.0.html) for Go applications.

## Overview

The semver package allows you to parse, validate, compare, and manipulate semantic versions. It provides a clean API for working with versions that adhere to the SemVer spec, with support for:

- Parsing version strings with or without leading 'v'
- Validating version strings against the SemVer spec
- Comparing versions for precedence
- Incrementing major, minor, and patch components
- Modifying version components
- Checking if a version is within a specified range

## Installation

```go
import "github.com/dmitrymomot/gokit/semver"
```

## Usage

### Parsing Versions

```go
// Parse a version string
version, err := semver.Parse("1.2.3")
if err != nil {
    // Handle error
}

// Parse with leading 'v'
version, err = semver.Parse("v1.2.3")

// Parse with prerelease and build metadata
version, err = semver.Parse("1.2.3-alpha.1+20130313144700")

// Parse without error checking (will panic on invalid versions)
version = semver.MustParse("1.2.3")
```

### Validating Versions

```go
// Check if a string is a valid semantic version
isValid := semver.Validate("1.2.3")
isValid = semver.Validate("v1.2.3-beta+exp.sha.5114f85")

// Check if a Version struct represents a valid version
version := semver.Version{Major: 1, Minor: 2, Patch: 3}
isValid = version.IsValid()
```

### Comparing Versions

```go
v1 := semver.MustParse("1.2.3")
v2 := semver.MustParse("2.0.0")

// Comparison methods
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

// Modify components
v2 := version.WithMajor(2)        // 2.2.3-alpha+001
v2 = version.WithMinor(5)         // 1.5.3-alpha+001
v2 = version.WithPatch(7)         // 1.2.7-alpha+001
v2 = version.WithPrerelease("beta") // 1.2.3-beta+001
v2 = version.WithBuild("002")     // 1.2.3-alpha+002

// Increment versions (resets prerelease and build metadata)
v2, err := version.Increment("major") // 2.0.0
v2, err = version.Increment("minor")  // 1.3.0
v2, err = version.Increment("patch")  // 1.2.4

// Get string representation
str := version.String()                // "1.2.3-alpha+001"
str = version.MajorMinorPatch()        // "1.2.3"
```

## Error Handling

The package defines several error types to help diagnose issues:

```go
var (
    ErrInvalidVersion    = errors.New("invalid semantic version")
    ErrInvalidPrerelease = errors.New("invalid prerelease identifier")
    ErrInvalidBuild      = errors.New("invalid build metadata")
    ErrEmptyVersion      = errors.New("empty version string")
    ErrNegativeValue     = errors.New("negative value in version")
)
```

## Thread Safety

The `Version` type is immutable by design. All methods that modify versions return a new instance, making the package safe for concurrent use.

## Best Practices

1. For API versioning, use major version zero (0.y.z) for initial development and expect frequent breaking changes.
2. Increment the major version when making incompatible API changes.
3. Increment the minor version when adding functionality in a backwards compatible manner.
4. Increment the patch version when making backwards compatible bug fixes.

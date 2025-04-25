package semver

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// semverRegex is a regular expression for validating semantic version strings
var semverRegex = regexp.MustCompile(`^v?(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)(?:-((?:0|[1-9]\d*|[a-zA-Z-][0-9a-zA-Z-]*)(?:\.(?:0|[1-9]\d*|[a-zA-Z-][0-9a-zA-Z-]*))*))?(?:\+([0-9a-zA-Z-]+(?:\.[0-9a-zA-Z-]+)*))?$`)

// Version represents a semantic version as defined by the SemVer 2.0.0 specification.
type Version struct {
	Major      uint64
	Minor      uint64
	Patch      uint64
	Prerelease string
	Build      string
}

// Parse parses a version string into a Version struct.
// It accepts versions with or without the leading 'v' character.
func Parse(version string) (Version, error) {
	if version == "" {
		return Version{}, ErrEmptyVersion
	}

	// Trim leading 'v' if present
	if version[0] == 'v' {
		version = version[1:]
	}

	matches := semverRegex.FindStringSubmatch(version)
	if matches == nil {
		return Version{}, fmt.Errorf("%w: %s", ErrInvalidVersion, version)
	}

	major, err := strconv.ParseUint(matches[1], 10, 64)
	if err != nil {
		return Version{}, fmt.Errorf("%w: invalid major version: %s", ErrInvalidVersion, matches[1])
	}

	minor, err := strconv.ParseUint(matches[2], 10, 64)
	if err != nil {
		return Version{}, fmt.Errorf("%w: invalid minor version: %s", ErrInvalidVersion, matches[2])
	}

	patch, err := strconv.ParseUint(matches[3], 10, 64)
	if err != nil {
		return Version{}, fmt.Errorf("%w: invalid patch version: %s", ErrInvalidVersion, matches[3])
	}

	v := Version{
		Major:      major,
		Minor:      minor,
		Patch:      patch,
		Prerelease: matches[4],
		Build:      matches[5],
	}

	return v, nil
}

// MustParse parses a version string into a Version struct.
// It panics if the version is invalid.
func MustParse(version string) Version {
	v, err := Parse(version)
	if err != nil {
		panic(err)
	}
	return v
}

// Validate checks if a version string is a valid semantic version.
func Validate(version string) bool {
	_, err := Parse(version)
	return err == nil
}

// String returns the string representation of the version.
func (v Version) String() string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "%d.%d.%d", v.Major, v.Minor, v.Patch)

	if v.Prerelease != "" {
		fmt.Fprintf(&sb, "-%s", v.Prerelease)
	}

	if v.Build != "" {
		fmt.Fprintf(&sb, "+%s", v.Build)
	}

	return sb.String()
}

// Compare compares two versions and returns:
//
//	-1 if v < other
//	 0 if v == other
//	 1 if v > other
//
// Prerelease versions are considered less than the corresponding release version.
// Build metadata is ignored when comparing versions for precedence.
func (v Version) Compare(other Version) int {
	// Compare major version
	if v.Major != other.Major {
		if v.Major < other.Major {
			return -1
		}
		return 1
	}

	// Compare minor version
	if v.Minor != other.Minor {
		if v.Minor < other.Minor {
			return -1
		}
		return 1
	}

	// Compare patch version
	if v.Patch != other.Patch {
		if v.Patch < other.Patch {
			return -1
		}
		return 1
	}

	// Compare prerelease - a version with a prerelease has lower precedence
	if v.Prerelease != other.Prerelease {
		// If one has a prerelease and the other doesn't
		if v.Prerelease == "" {
			return 1
		}
		if other.Prerelease == "" {
			return -1
		}

		// Compare prerelease identifiers
		return comparePrerelease(v.Prerelease, other.Prerelease)
	}

	// Versions are equal (build metadata doesn't affect precedence)
	return 0
}

// comparePrerelease compares prerelease strings according to SemVer rules
func comparePrerelease(a, b string) int {
	aParts := strings.Split(a, ".")
	bParts := strings.Split(b, ".")

	// Compare each identifier until a difference is found
	for i := 0; i < len(aParts) && i < len(bParts); i++ {
		if comp := compareIdentifier(aParts[i], bParts[i]); comp != 0 {
			return comp
		}
	}

	// If all identifiers are equal but one has more, the shorter one has precedence
	if len(aParts) < len(bParts) {
		return -1
	}
	if len(aParts) > len(bParts) {
		return 1
	}

	return 0
}

// compareIdentifier compares individual prerelease identifiers
func compareIdentifier(a, b string) int {
	// Check if both are numeric
	aNum, aErr := strconv.ParseUint(a, 10, 64)
	bNum, bErr := strconv.ParseUint(b, 10, 64)

	// If both are numeric, compare numerically
	if aErr == nil && bErr == nil {
		if aNum < bNum {
			return -1
		}
		if aNum > bNum {
			return 1
		}
		return 0
	}

	// If only a is numeric, it has lower precedence
	if aErr == nil {
		return -1
	}

	// If only b is numeric, it has lower precedence
	if bErr == nil {
		return 1
	}

	// Both are non-numeric, compare lexically
	if a < b {
		return -1
	}
	if a > b {
		return 1
	}
	return 0
}

// LessThan returns true if v < other
func (v Version) LessThan(other Version) bool {
	return v.Compare(other) < 0
}

// GreaterThan returns true if v > other
func (v Version) GreaterThan(other Version) bool {
	return v.Compare(other) > 0
}

// Equal returns true if v == other
func (v Version) Equal(other Version) bool {
	return v.Compare(other) == 0
}

// LessThanOrEqual returns true if v <= other
func (v Version) LessThanOrEqual(other Version) bool {
	return v.Compare(other) <= 0
}

// GreaterThanOrEqual returns true if v >= other
func (v Version) GreaterThanOrEqual(other Version) bool {
	return v.Compare(other) >= 0
}

// InRange checks if the version is between the lower and upper bounds (inclusive)
func (v Version) InRange(lower, upper Version) bool {
	return v.GreaterThanOrEqual(lower) && v.LessThanOrEqual(upper)
}

// IsValid checks if the version is valid according to the semver specification
func (v Version) IsValid() bool {
	// Use the regex to validate the string representation
	return semverRegex.MatchString(v.String())
}

// WithMajor returns a new version with the major component changed
func (v Version) WithMajor(major uint64) Version {
	newV := v
	newV.Major = major
	return newV
}

// WithMinor returns a new version with the minor component changed
func (v Version) WithMinor(minor uint64) Version {
	newV := v
	newV.Minor = minor
	return newV
}

// WithPatch returns a new version with the patch component changed
func (v Version) WithPatch(patch uint64) Version {
	newV := v
	newV.Patch = patch
	return newV
}

// WithPrerelease returns a new version with the prerelease component changed
func (v Version) WithPrerelease(prerelease string) Version {
	newV := v
	newV.Prerelease = prerelease
	return newV
}

// WithBuild returns a new version with the build component changed
func (v Version) WithBuild(build string) Version {
	newV := v
	newV.Build = build
	return newV
}

// Increment returns a new version with the specified component incremented
func (v Version) Increment(component string) (Version, error) {
	newV := v

	switch strings.ToLower(component) {
	case "major":
		newV.Major++
		newV.Minor = 0
		newV.Patch = 0
	case "minor":
		newV.Minor++
		newV.Patch = 0
	case "patch":
		newV.Patch++
	default:
		return v, fmt.Errorf("invalid component: %s", component)
	}

	// Reset prerelease and build metadata when incrementing
	newV.Prerelease = ""
	newV.Build = ""

	return newV, nil
}

// MajorMinorPatch returns the version without prerelease or build metadata
func (v Version) MajorMinorPatch() string {
	return fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
}

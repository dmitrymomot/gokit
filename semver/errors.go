package semver

import "errors"

// Error definitions for semver package
var (
	// ErrInvalidVersion indicates that the provided string is not a valid semantic version
	ErrInvalidVersion = errors.New("invalid semantic version")
	
	// ErrInvalidPrerelease indicates that the provided prerelease string is invalid
	ErrInvalidPrerelease = errors.New("invalid prerelease identifier")
	
	// ErrInvalidBuild indicates that the provided build metadata is invalid
	ErrInvalidBuild = errors.New("invalid build metadata")
	
	// ErrEmptyVersion indicates an empty version string was provided
	ErrEmptyVersion = errors.New("empty version string")
	
	// ErrNegativeValue indicates a negative number was encountered in version components
	ErrNegativeValue = errors.New("negative value in version")
)

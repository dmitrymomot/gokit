# Feature Flags Package

A flexible and extensible feature flagging system for Go applications.

## Installation

```bash
go get github.com/dmitrymomot/gokit/feature
```

## Overview

The `feature` package provides a robust feature flagging system to control feature rollouts in your application. It supports various strategies including percentage-based rollouts, targeted users/groups, environment-based flags, and composite conditions.

## Features

- Generic `Provider` interface with pluggable storage backends
- In-memory implementation for quick setup and testing
- Flexible rollout strategies:
  - Always on/off flags
  - Targeted user/group flags
  - Percentage-based progressive rollouts
  - Environment-based flags 
  - Composite strategies (AND/OR logic)
- Context-based evaluation for user-specific flags
- Tag-based flag organization
- Thread-safe implementations

## Usage

### Basic Setup

```go
package main

import (
	"context"
	"log"

	"github.com/dmitrymomot/gokit/feature"
)

func main() {
	// Create a provider with initial flags
	provider, err := feature.NewMemoryProvider(
		&feature.Flag{
			Name:        "dark-mode",
			Description: "Enable dark mode UI",
			Enabled:     true,
			Tags:        []string{"ui", "theme"},
		},
		&feature.Flag{
			Name:        "beta-features",
			Description: "Enable beta features",
			Enabled:     true,
			Strategy:    feature.NewTargetedStrategy(feature.TargetCriteria{
				Groups: []string{"beta-users", "internal"},
			}),
		},
	)
	if err != nil {
		log.Fatalf("Failed to create feature provider: %v", err)
	}
	defer provider.Close()

	// Check if a feature is enabled
	ctx := context.Background()
	darkModeEnabled, err := provider.IsEnabled(ctx, "dark-mode")
	if err != nil {
		log.Printf("Error checking flag: %v", err)
	}

	if darkModeEnabled {
		// Enable dark mode UI
		log.Println("Dark mode is enabled!")
	}
}
```

### User-Specific Features

```go
// Add user information to context
ctx := context.Background()
ctx = context.WithValue(ctx, feature.UserIDKey, "user-123")
ctx = context.WithValue(ctx, feature.UserGroupsKey, []string{"beta-users"})

// Check if beta features are enabled for this user
betaEnabled, err := provider.IsEnabled(ctx, "beta-features")
```

### Rollout Strategies

```go
// Always on strategy
alwaysOn := feature.NewAlwaysOnStrategy()

// Environment-based strategy
envStrategy := feature.NewEnvironmentStrategy("dev", "staging")

// Targeted strategy
targetedStrategy := feature.NewTargetedStrategy(feature.TargetCriteria{
	UserIDs:    []string{"user-1", "user-2"},  // Specific users
	Groups:     []string{"beta", "internal"},  // User groups
	Percentage: intPtr(20),                    // 20% of users
	AllowList:  []string{"vip-1"},             // Always enabled
	DenyList:   []string{"banned-user"},       // Never enabled
})

// Composite strategy (both conditions must be true)
compositeStrategy := feature.NewAndStrategy(
	envStrategy,      // Must be in dev/staging
	targetedStrategy, // Must match target criteria
)

// Helper function for percentage
func intPtr(i int) *int {
	return &i
}
```

### Managing Flags

```go
// Create a new flag
newFlag := &feature.Flag{
	Name:        "new-feature",
	Description: "A new experimental feature",
	Enabled:     true,
	Strategy:    feature.NewTargetedStrategy(feature.TargetCriteria{
		Percentage: intPtr(10), // 10% rollout
	}),
	Tags: []string{"experimental"},
}
err = provider.CreateFlag(ctx, newFlag)

// Update an existing flag
flag, err := provider.GetFlag(ctx, "new-feature")
flag.Strategy = feature.NewTargetedStrategy(feature.TargetCriteria{
	Percentage: intPtr(50), // Increase to 50% rollout
})
err = provider.UpdateFlag(ctx, flag)

// Delete a flag
err = provider.DeleteFlag(ctx, "deprecated-feature")
```

### Listing Flags

```go
// List all flags
allFlags, err := provider.ListFlags(ctx)

// List flags by tags (flags with any of these tags)
uiFlags, err := provider.ListFlags(ctx, "ui", "theme")
```

## Custom Providers

Implement the `Provider` interface to create custom storage backends:

```go
type Provider interface {
	IsEnabled(ctx context.Context, flagName string) (bool, error)
	GetFlag(ctx context.Context, flagName string) (*Flag, error)
	ListFlags(ctx context.Context, tags ...string) ([]*Flag, error)
	CreateFlag(ctx context.Context, flag *Flag) error
	UpdateFlag(ctx context.Context, flag *Flag) error
	DeleteFlag(ctx context.Context, flagName string) error
	Close() error
}
```

## Error Handling

```go
enabled, err := provider.IsEnabled(ctx, "my-flag")
if err != nil {
	if errors.Is(err, feature.ErrFlagNotFound) {
		// Handle flag not found case
	} else {
		// Handle other errors
	}
}
```

The package defines the following errors:

- `ErrFlagNotFound`: Flag doesn't exist
- `ErrInvalidFlag`: Invalid flag parameters 
- `ErrProviderNotInitialized`: Provider not ready
- `ErrInvalidContext`: Missing required context values
- `ErrInvalidStrategy`: Invalid rollout strategy
- `ErrOperationFailed`: General operation failure

# Feature Flags Package

This package provides a flexible and extensible feature flagging system for Go applications.

## Features

- Generic `Provider` interface that supports different backend storage options
- In-memory implementation for testing and simple applications
- Powerful rollout strategies:
  - Always on/off strategies
  - Targeted strategies (by user IDs, groups, etc.)
  - Percentage-based rollouts
  - Environment-based flags
  - Composite strategies with AND/OR operators
- Context-based evaluation for request-specific flag resolution
- Support for tagging and metadata
- Thread-safe implementations

## Installation

```bash
go get github.com/dmitrymomot/gokit/feature
```

## Usage

### Basic Setup

Create a feature flag provider and check if features are enabled:

```go
package main

import (
	"context"
	"log"

	"github.com/dmitrymomot/gokit/feature"
)

func main() {
	// Create an in-memory provider with initial flags
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

### Context-Aware Flags

Use context values to provide user-specific data for feature flag evaluation:

```go
// Add user information to context
userID := "user-123"
userGroups := []string{"beta-users", "premium"}

ctx := context.Background()
ctx = context.WithValue(ctx, feature.UserIDKey, userID)
ctx = context.WithValue(ctx, feature.UserGroupsKey, userGroups)

// Check if beta features are enabled for this user
betaEnabled, err := provider.IsEnabled(ctx, "beta-features")
if err != nil {
	log.Printf("Error checking beta flag: %v", err)
}

if betaEnabled {
	// Show beta features to this user
	log.Println("Beta features are enabled for this user!")
}
```

### Creating Rollout Strategies

The package supports various rollout strategies:

```go
// Always on strategy
alwaysOn := feature.NewAlwaysOnStrategy()

// Environment-based strategy (only enabled in dev/staging)
envStrategy := feature.NewEnvironmentStrategy("dev", "staging")

// Targeted strategy for specific users and groups
targetedStrategy := feature.NewTargetedStrategy(feature.TargetCriteria{
	UserIDs:    []string{"user-1", "user-2"},      // Specific users
	Groups:     []string{"beta", "internal"},      // User groups
	Percentage: intPtr(20),                       // 20% of all users
	AllowList:  []string{"vip-1", "admin-user"},  // Always enabled for these users
	DenyList:   []string{"banned-user"},          // Never enabled for these users
})

// Composite strategies
compositeStrategy := feature.NewAndStrategy(
	envStrategy,                // Must be in dev/staging AND
	targetedStrategy,           // Must match targeting criteria
)

// Helper for percentage pointer
func intPtr(i int) *int {
	return &i
}
```

### Managing Flags

Create, update, and delete flags programmatically:

```go
// Create a new flag
newFlag := &feature.Flag{
	Name:        "new-feature",
	Description: "A new experimental feature",
	Enabled:     true,
	Strategy:    feature.NewTargetedStrategy(feature.TargetCriteria{
		Percentage: intPtr(10), // 10% rollout
	}),
	Tags: []string{"experimental", "v2"},
}

err = provider.CreateFlag(ctx, newFlag)
if err != nil {
	log.Printf("Failed to create flag: %v", err)
}

// Update an existing flag
updatedFlag, err := provider.GetFlag(ctx, "new-feature") 
if err != nil {
	log.Printf("Failed to get flag: %v", err)
}

// Increase rollout to 50%
percentage := 50
updatedFlag.Strategy = feature.NewTargetedStrategy(feature.TargetCriteria{
	Percentage: &percentage,
})

err = provider.UpdateFlag(ctx, updatedFlag)
if err != nil {
	log.Printf("Failed to update flag: %v", err)
}

// Delete a flag when no longer needed
err = provider.DeleteFlag(ctx, "deprecated-feature")
if err != nil {
	log.Printf("Failed to delete flag: %v", err)
}
```

### Listing Flags

List all flags or filter by tags:

```go
// List all flags
allFlags, err := provider.ListFlags(ctx)
if err != nil {
	log.Printf("Failed to list flags: %v", err)
}

for _, flag := range allFlags {
	log.Printf("Flag: %s, Enabled: %v", flag.Name, flag.Enabled)
}

// List flags by tags
uiFlags, err := provider.ListFlags(ctx, "ui")
if err != nil {
	log.Printf("Failed to list UI flags: %v", err)
}

// List flags with multiple tags (OR operation)
filteredFlags, err := provider.ListFlags(ctx, "experimental", "beta")
if err != nil {
	log.Printf("Failed to list filtered flags: %v", err)
}
```

## Custom Providers

You can implement your own providers by implementing the `Provider` interface:

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

This allows for implementing providers backed by Redis, databases, or remote services.

## Error Handling

The package defines standard errors:

```go
var (
	ErrFlagNotFound          = errors.New("feature: flag not found")
	ErrInvalidFlag           = errors.New("feature: invalid flag parameters")
	ErrProviderNotInitialized = errors.New("feature: provider not initialized")
	ErrInvalidContext        = errors.New("feature: invalid context")
	ErrInvalidStrategy       = errors.New("feature: invalid rollout strategy")
	ErrOperationFailed       = errors.New("feature: operation failed")
)
```

Use `errors.Is()` to check for specific error types:

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

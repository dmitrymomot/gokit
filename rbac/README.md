# RBAC Package

The RBAC (Role-Based Access Control) package provides a flexible and efficient way to manage role and permission hierarchies within Go applications. It enforces security policies based on roles and permissions with support for multi-level inheritance and various storage backends.

## Overview

This package implements a comprehensive RBAC system with the following features:

- **Role & Permission Management**: Complete CRUD operations for roles and permissions
- **Inheritance Support**: Both roles and permissions can inherit from others, creating complex hierarchies
- **Efficient Permission Checking**: Fast operations to check if a role has specific permissions
- **Storage Abstraction**: Flexible store interface with an included thread-safe in-memory implementation
- **Caching**: Optional caching to improve performance of frequent permission checks
- **Cycle Prevention**: Built-in protection against cyclic inheritance chains

## Core Concepts

### Roles

A role represents a set of responsibilities or capabilities that can be assigned to users. Each role can:

- Inherit from one or more parent roles
- Directly own permissions
- Effectively possess permissions from its entire inheritance chain

### Permissions

A permission represents the ability to perform a specific action. Each permission can:

- Inherit from one or more parent permissions
- Be directly assigned to roles
- Be effectively granted through permission inheritance chains

### Inheritance

Both roles and permissions support multi-level inheritance:

- **Role Inheritance**: When role A inherits from role B, role A automatically gains all permissions associated with role B.
- **Permission Inheritance**: When permission X inherits from permission Y, any role with permission X automatically has permission Y as well.

## Usage Examples

### Basic Setup

```go
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/dmitrymomot/gokit/rbac"
)

func main() {
	// Create a context
	ctx := context.Background()

	// Create an in-memory store
	store := rbac.NewMemoryStore()

	// Create an RBAC service
	rbacService := rbac.NewService(store)

	// Create permissions
	viewPerm := rbac.Permission{
		ID:   "post:read",
		Name: "Can read posts",
	}
	
	writePerm := rbac.Permission{
		ID:        "post:write",
		Name:      "Can write posts",
		ParentIDs: []string{"post:read"}, // Writing includes reading
	}
	
	deletePerm := rbac.Permission{
		ID:        "post:delete",
		Name:      "Can delete posts",
		ParentIDs: []string{"post:write"}, // Deleting includes writing
	}

	// Store the permissions
	if err := store.CreatePermission(ctx, viewPerm); err != nil {
		log.Fatalf("Failed to create view permission: %v", err)
	}
	
	if err := store.CreatePermission(ctx, writePerm); err != nil {
		log.Fatalf("Failed to create write permission: %v", err)
	}
	
	if err := store.CreatePermission(ctx, deletePerm); err != nil {
		log.Fatalf("Failed to create delete permission: %v", err)
	}

	// Create roles
	guestRole := rbac.Role{
		ID:                  "guest",
		Name:                "Guest User",
		DirectPermissionIDs: []string{"post:read"},
	}
	
	editorRole := rbac.Role{
		ID:                  "editor",
		Name:                "Editor User",
		ParentIDs:           []string{"guest"},
		DirectPermissionIDs: []string{"post:write"},
	}
	
	adminRole := rbac.Role{
		ID:                  "admin",
		Name:                "Admin User",
		ParentIDs:           []string{"editor"},
		DirectPermissionIDs: []string{"post:delete"},
	}

	// Store the roles
	if err := store.CreateRole(ctx, guestRole); err != nil {
		log.Fatalf("Failed to create guest role: %v", err)
	}
	
	if err := store.CreateRole(ctx, editorRole); err != nil {
		log.Fatalf("Failed to create editor role: %v", err)
	}
	
	if err := store.CreateRole(ctx, adminRole); err != nil {
		log.Fatalf("Failed to create admin role: %v", err)
	}

	// Check permissions
	checkPermission(ctx, rbacService, "guest", "post:read")
	checkPermission(ctx, rbacService, "guest", "post:write")
	
	checkPermission(ctx, rbacService, "editor", "post:read")
	checkPermission(ctx, rbacService, "editor", "post:write")
	checkPermission(ctx, rbacService, "editor", "post:delete")
	
	checkPermission(ctx, rbacService, "admin", "post:read")
	checkPermission(ctx, rbacService, "admin", "post:write")
	checkPermission(ctx, rbacService, "admin", "post:delete")
}

func checkPermission(ctx context.Context, rbacService *rbac.Service, roleID, permissionID string) {
	hasPermission, err := rbacService.HasPermission(ctx, roleID, permissionID)
	if err != nil {
		log.Printf("Error checking permission: %v", err)
		return
	}
	
	if hasPermission {
		fmt.Printf("Role %s has permission %s\n", roleID, permissionID)
	} else {
		fmt.Printf("Role %s does NOT have permission %s\n", roleID, permissionID)
	}
}
```

### Using Caching

```go
package main

import (
	"context"
	"log"
	"time"

	"github.com/dmitrymomot/gokit/rbac"
)

func main() {
	// Create a context
	ctx := context.Background()

	// Create an in-memory store
	store := rbac.NewMemoryStore()

	// Create an RBAC service with caching enabled (5 minute TTL)
	rbacService := rbac.NewService(store, rbac.WithCaching(5*time.Minute))

	// ... Create roles and permissions as in the previous example

	// Check permissions (results will be cached)
	hasPermission, err := rbacService.HasPermission(ctx, "admin", "post:read")
	if err != nil {
		log.Fatalf("Failed to check permission: %v", err)
	}

	// Invalidate cache for a specific role when role or permissions change
	rbacService.InvalidateCache("admin")

	// Invalidate the entire cache if many changes are made
	rbacService.InvalidateAllCache()
}
```

### Complex Permission Checks

```go
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/dmitrymomot/gokit/rbac"
)

func main() {
	// ... Setup store and service as in previous examples

	ctx := context.Background()
	rbacService := rbac.NewService(rbac.NewMemoryStore())
	
	// ... Create roles and permissions

	// Check if a role has at least one of the specified permissions
	hasAny, err := rbacService.HasAnyPermission(ctx, "editor", "post:delete", "user:manage", "post:write")
	if err != nil {
		log.Fatalf("Failed to check permissions: %v", err)
	}
	
	if hasAny {
		fmt.Println("Editor has at least one of the requested permissions")
	}

	// Check if a role has all of the specified permissions
	hasAll, err := rbacService.HasAllPermissions(ctx, "admin", "post:read", "post:write", "post:delete")
	if err != nil {
		log.Fatalf("Failed to check permissions: %v", err)
	}
	
	if hasAll {
		fmt.Println("Admin has all of the requested permissions")
	}

	// Get all effective permissions for a role
	permissions, err := rbacService.GetEffectivePermissions(ctx, "admin")
	if err != nil {
		log.Fatalf("Failed to get effective permissions: %v", err)
	}
	
	fmt.Println("Admin's effective permissions:")
	for _, p := range permissions {
		fmt.Printf("- %s (%s)\n", p.Name, p.ID)
	}
}
```

### Using with HTTP Middleware

```go
package main

import (
	"context"
	"net/http"

	"github.com/dmitrymomot/gokit/rbac"
)

// RBACMiddleware creates middleware that checks if a user has the required permission
func RBACMiddleware(rbacService *rbac.Service, permissionID string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// In a real application, you would extract user information and role ID from
			// the request context, session, or JWT token
			roleID := getUserRoleFromRequest(r)
			
			// Check if the role has the required permission
			hasPermission, err := rbacService.HasPermission(r.Context(), roleID, permissionID)
			if err != nil || !hasPermission {
				http.Error(w, "Unauthorized", http.StatusForbidden)
				return
			}
			
			// Continue to the next middleware/handler
			next.ServeHTTP(w, r)
		})
	}
}

// Example usage with HTTP routing
func setupRoutes(rbacService *rbac.Service) {
	// Protected endpoint that requires "post:write" permission
	http.Handle("/posts/create", 
		RBACMiddleware(rbacService, "post:write")(
			http.HandlerFunc(handleCreatePost),
		),
	)
	
	// Protected endpoint that requires "post:delete" permission
	http.Handle("/posts/delete", 
		RBACMiddleware(rbacService, "post:delete")(
			http.HandlerFunc(handleDeletePost),
		),
	)
	
	// Public endpoint
	http.Handle("/posts/view", http.HandlerFunc(handleViewPost))
}

// Mock implementation - in a real application, extract from JWT, session, etc.
func getUserRoleFromRequest(r *http.Request) string {
	return r.Header.Get("X-User-Role")
}

func handleCreatePost(w http.ResponseWriter, r *http.Request) {
	// Handler implementation
}

func handleDeletePost(w http.ResponseWriter, r *http.Request) {
	// Handler implementation
}

func handleViewPost(w http.ResponseWriter, r *http.Request) {
	// Handler implementation
}
```

## API Documentation

### Types

#### Role

```go
type Role struct {
	// ID is the unique identifier of the role.
	ID string

	// Name is the display name of the role.
	Name string

	// ParentIDs contains the IDs of the roles this role inherits from.
	ParentIDs []string

	// DirectPermissionIDs contains the IDs of permissions directly assigned to this role.
	DirectPermissionIDs []string
}
```

#### Permission

```go
type Permission struct {
	// ID is the unique identifier of the permission.
	ID string

	// Name is the display name of the permission.
	Name string

	// ParentIDs contains the IDs of the permissions this permission inherits from.
	ParentIDs []string
}
```

### Interfaces

The package provides several interfaces:

- `RBAC`: Core interface for permission checking operations
- `Store`: Combined interface for role and permission storage operations
- `RoleStore`: Interface for role storage operations
- `PermissionStore`: Interface for permission storage operations

### Implementations

- `Service`: Implementation of the RBAC interface
- `MemoryStore`: Thread-safe in-memory implementation of the Store interface

## Performance Considerations

For optimal performance:

1. Use caching when permission checks are frequent
2. Design permission hierarchies carefully to avoid deep inheritance chains
3. Be aware that cyclic inheritance checks can be computationally intensive for large hierarchies

## Thread Safety

All implementations in this package are designed to be thread-safe and can be safely used from multiple goroutines concurrently.

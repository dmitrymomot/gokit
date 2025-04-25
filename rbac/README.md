# RBAC Package

A flexible role-based access control system with multi-level inheritance support for Go applications.

## Installation

```bash
go get github.com/dmitrymomot/gokit/rbac
```

## Overview

The RBAC (Role-Based Access Control) package provides a flexible and efficient way to manage role and permission hierarchies in Go applications. It enforces access control policies based on roles, workspaces, and permissions with support for inheritance, caching, and different storage backends.

## Features

- Multi-workspace support for isolated RBAC environments
- Role and permission hierarchies with multi-level inheritance
- Efficient permission checking algorithms
- Caching for optimized performance
- Thread-safe in-memory implementation included
- Cycle detection and prevention in inheritance chains
- Comprehensive CRUD operations for roles and permissions
- Flexible storage interface for custom implementations

## Usage

### Basic Setup

```go
import (
	"context"
	"fmt"
	"github.com/dmitrymomot/gokit/rbac"
)

func main() {
	// Create a context and in-memory store
	ctx := context.Background()
	store := rbac.NewMemoryStore()
	rbacService := rbac.NewService(store)
	
	// Define workspace ID
	workspaceID := "organization-123"
	
	// Create permissions with inheritance
	viewPerm := rbac.Permission{
		WorkspaceID: workspaceID,
		ID:          "post:read",
		Name:        "Can read posts",
	}
	
	writePerm := rbac.Permission{
		WorkspaceID: workspaceID,
		ID:          "post:write",
		Name:        "Can write posts",
		ParentIDs:   []string{"post:read"}, // Writing includes reading
	}
	
	deletePerm := rbac.Permission{
		WorkspaceID: workspaceID,
		ID:          "post:delete",
		Name:        "Can delete posts",
		ParentIDs:   []string{"post:write"}, // Deleting includes writing
	}
	
	// Store the permissions
	store.CreatePermission(ctx, viewPerm)
	store.CreatePermission(ctx, writePerm)
	store.CreatePermission(ctx, deletePerm)
	
	// Create roles with inheritance
	guestRole := rbac.Role{
		WorkspaceID:         workspaceID,
		ID:                  "guest",
		Name:                "Guest User",
		DirectPermissionIDs: []string{"post:read"},
	}
	
	editorRole := rbac.Role{
		WorkspaceID:         workspaceID,
		ID:                  "editor",
		Name:                "Editor User",
		ParentIDs:           []string{"guest"},
		DirectPermissionIDs: []string{"post:write"},
	}
	
	adminRole := rbac.Role{
		WorkspaceID:         workspaceID,
		ID:                  "admin",
		Name:                "Admin User",
		ParentIDs:           []string{"editor"},
		DirectPermissionIDs: []string{"post:delete"},
	}
	
	// Store the roles
	store.CreateRole(ctx, guestRole)
	store.CreateRole(ctx, editorRole)
	store.CreateRole(ctx, adminRole)
	
	// Check permissions
	hasPermission, _ := rbacService.HasPermission(ctx, workspaceID, "guest", "post:read")
	fmt.Printf("Guest has post:read: %v\n", hasPermission) // true
	
	hasPermission, _ = rbacService.HasPermission(ctx, workspaceID, "guest", "post:write")
	fmt.Printf("Guest has post:write: %v\n", hasPermission) // false
	
	hasPermission, _ = rbacService.HasPermission(ctx, workspaceID, "editor", "post:read")
	fmt.Printf("Editor has post:read: %v\n", hasPermission) // true (inherited from guest)
	
	hasPermission, _ = rbacService.HasPermission(ctx, workspaceID, "admin", "post:delete")
	fmt.Printf("Admin has post:delete: %v\n", hasPermission) // true
}
```

### Using Caching

```go
import (
	"time"
	"github.com/dmitrymomot/gokit/rbac"
)

// Create an RBAC service with caching enabled (5 minute TTL)
rbacService := rbac.NewService(store, rbac.WithCaching(5*time.Minute))

// Check permissions (results will be cached)
hasPermission, _ := rbacService.HasPermission(ctx, workspaceID, "admin", "post:read")

// Invalidate cache for a specific role when role or permissions change
rbacService.InvalidateCache(workspaceID, "admin")

// Invalidate all caches for a specific workspace
rbacService.InvalidateWorkspaceCache(workspaceID)

// Invalidate the entire cache across all workspaces
rbacService.InvalidateAllCache()
```

### Working with Multiple Workspaces

```go
// Define workspace IDs
orgWorkspace := "organization-123"
projectWorkspace := "project-456"

// Create roles and permissions in each workspace
// Even with the same role IDs, permissions are completely isolated

// Check permissions in organization workspace
hasPermission, _ := rbacService.HasPermission(ctx, orgWorkspace, "admin", "system:admin")

// Check permissions in project workspace
hasPermission, _ = rbacService.HasPermission(ctx, projectWorkspace, "admin", "system:admin")
```

### Advanced Permission Checks

```go
// Check if role has ANY of the given permissions
hasAny, _ := rbacService.HasAnyPermission(ctx, workspaceID, "editor", 
	"post:edit", "post:delete", "user:manage")
fmt.Printf("Editor has any of these permissions: %v\n", hasAny)

// Check if role has ALL of the given permissions
hasAll, _ := rbacService.HasAllPermissions(ctx, workspaceID, "admin", 
	"post:read", "post:write", "post:delete")
fmt.Printf("Admin has all these permissions: %v\n", hasAll)

// Get all effective permissions (direct + inherited)
permissions, _ := rbacService.GetEffectivePermissions(ctx, workspaceID, "editor")
fmt.Println("Editor effective permissions:")
for _, perm := range permissions {
	fmt.Printf("- %s (%s)\n", perm.Name, perm.ID)
}
```

## API Reference

### Core Entities

```go
// Role represents a set of responsibilities that can be assigned to users
type Role struct {
	WorkspaceID         string   // Workspace this role belongs to
	ID                  string   // Unique role identifier
	Name                string   // Human-readable name
	ParentIDs           []string // Roles this role inherits from
	DirectPermissionIDs []string // Direct permissions assigned to this role
}

// Permission represents an action that can be performed
type Permission struct {
	WorkspaceID string   // Workspace this permission belongs to
	ID          string   // Unique permission identifier
	Name        string   // Human-readable name
	ParentIDs   []string // Permissions this permission inherits from
}
```

### Main Service

```go
// Create a new RBAC service with options
func NewService(store Store, opts ...Option) *Service

// Permission Checking
func (s *Service) HasPermission(ctx context.Context, workspaceID, roleID, permissionID string) (bool, error)
func (s *Service) HasAnyPermission(ctx context.Context, workspaceID, roleID string, permissionIDs ...string) (bool, error)
func (s *Service) HasAllPermissions(ctx context.Context, workspaceID, roleID string, permissionIDs ...string) (bool, error)
func (s *Service) GetEffectivePermissions(ctx context.Context, workspaceID, roleID string) ([]Permission, error)

// Cache Management
func (s *Service) InvalidateCache(workspaceID, roleID string)
func (s *Service) InvalidateWorkspaceCache(workspaceID string)
func (s *Service) InvalidateAllCache()
```

### Store Interface

```go
// Store provides access to both role and permission storage
type Store interface {
	RoleStore
	PermissionStore
}

// In-memory implementation
func NewMemoryStore() Store
```

### Role Operations

```go
// RoleStore defines operations for managing roles
type RoleStore interface {
	CreateRole(ctx context.Context, role Role) error
	GetRole(ctx context.Context, workspaceID, roleID string) (Role, error)
	GetRoles(ctx context.Context, workspaceID string) ([]Role, error)
	UpdateRole(ctx context.Context, role Role) error
	DeleteRole(ctx context.Context, workspaceID, roleID string) error
	
	// Role hierarchy methods
	AddRoleParent(ctx context.Context, workspaceID, roleID, parentID string) error
	RemoveRoleParent(ctx context.Context, workspaceID, roleID, parentID string) error
	GetRoleParents(ctx context.Context, workspaceID, roleID string) ([]Role, error)
	GetRoleChildren(ctx context.Context, workspaceID, roleID string) ([]Role, error)
}
```

### Permission Operations

```go
// PermissionStore defines operations for managing permissions
type PermissionStore interface {
	CreatePermission(ctx context.Context, permission Permission) error
	GetPermission(ctx context.Context, workspaceID, permissionID string) (Permission, error)
	GetPermissions(ctx context.Context, workspaceID string) ([]Permission, error)
	UpdatePermission(ctx context.Context, permission Permission) error
	DeletePermission(ctx context.Context, workspaceID, permissionID string) error
	
	// Permission hierarchy methods
	AddPermissionParent(ctx context.Context, workspaceID, permissionID, parentID string) error
	RemovePermissionParent(ctx context.Context, workspaceID, permissionID, parentID string) error
	GetPermissionParents(ctx context.Context, workspaceID, permissionID string) ([]Permission, error)
	GetPermissionChildren(ctx context.Context, workspaceID, permissionID string) ([]Permission, error)
	
	// Role-permission assignments
	AddPermissionToRole(ctx context.Context, workspaceID, roleID, permissionID string) error
	RemovePermissionFromRole(ctx context.Context, workspaceID, roleID, permissionID string) error
	GetRolePermissions(ctx context.Context, workspaceID, roleID string) ([]Permission, error)
}
```

## Error Handling

```go
// Common RBAC errors
var (
	ErrRoleNotFound          = errors.New("role not found")
	ErrRoleAlreadyExists     = errors.New("role already exists")
	ErrPermissionNotFound    = errors.New("permission not found")
	ErrPermissionAlreadyExists = errors.New("permission already exists")
	ErrCyclicDependency      = errors.New("cyclic dependency detected")
	ErrInvalidWorkspace      = errors.New("invalid workspace")
)

// Proper error handling
hasPermission, err := rbacService.HasPermission(ctx, workspaceID, roleID, permissionID)
if err != nil {
	switch {
	case errors.Is(err, rbac.ErrRoleNotFound):
		// Handle role not found
	case errors.Is(err, rbac.ErrPermissionNotFound):
		// Handle permission not found
	default:
		// Handle other errors
	}
}
```

## Best Practices

1. **Workspace Isolation**:
   - Use separate workspaces for different organizational boundaries
   - Never share roles or permissions across workspace boundaries
   - Use consistent workspace ID generation schemes

2. **Permission Naming**:
   - Use consistent naming patterns for permissions (e.g., `resource:action`)
   - Group related permissions to enable logical inheritance
   - Document the permission structure for application users

3. **Role Design**:
   - Design role hierarchies to match organizational structures
   - Assign permissions at the appropriate level in the hierarchy
   - Avoid overly complex role structures

4. **Performance**:
   - Use caching for frequently accessed role-permission checks
   - Invalidate caches selectively when roles or permissions change
   - Consider batching permission checks when possible

5. **Security**:
   - Validate all inputs to avoid injection attacks
   - Implement proper authorization checks at all entry points
   - Limit administrative access to role and permission management

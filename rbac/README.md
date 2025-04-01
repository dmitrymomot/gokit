# RBAC Package

The RBAC (Role-Based Access Control) package provides a flexible and efficient way to manage role and permission hierarchies within Go applications. It enforces security policies based on roles and permissions with support for multi-level inheritance, workspace isolation, and various storage backends.

## Overview

This package implements a comprehensive RBAC system with the following features:

- **Role & Permission Management**: Complete CRUD operations for roles and permissions
- **Workspace Isolation**: Support for multiple isolated workspaces, each with their own roles and permissions
- **Inheritance Support**: Both roles and permissions can inherit from others, creating complex hierarchies
- **Efficient Permission Checking**: Fast operations to check if a role has specific permissions
- **Storage Abstraction**: Flexible store interface with an included thread-safe in-memory implementation
- **Caching**: Optional caching to improve performance of frequent permission checks
- **Cycle Prevention**: Built-in protection against cyclic inheritance chains

## Core Concepts

### Workspaces

A workspace represents an isolated RBAC environment where roles and permissions are independently managed. This enables:

- Multi-tenant applications with separate permission systems
- Organizations with distinct role structures
- Projects with isolated access control requirements

Each role and permission belongs to a specific workspace and cannot be accessed from other workspaces.

### Roles

A role represents a set of responsibilities or capabilities that can be assigned to users. Each role can:

- Belong to a specific workspace
- Inherit from one or more parent roles within the same workspace
- Directly own permissions
- Effectively possess permissions from its entire inheritance chain

### Permissions

A permission represents the ability to perform a specific action. Each permission can:

- Belong to a specific workspace
- Inherit from one or more parent permissions within the same workspace
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
	
	// Define workspace ID
	workspaceID := "organization-123"

	// Create permissions
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
	checkPermission(ctx, rbacService, workspaceID, "guest", "post:read")
	checkPermission(ctx, rbacService, workspaceID, "guest", "post:write")
	
	checkPermission(ctx, rbacService, workspaceID, "editor", "post:read")
	checkPermission(ctx, rbacService, workspaceID, "editor", "post:write")
	checkPermission(ctx, rbacService, workspaceID, "editor", "post:delete")
	
	checkPermission(ctx, rbacService, workspaceID, "admin", "post:read")
	checkPermission(ctx, rbacService, workspaceID, "admin", "post:write")
	checkPermission(ctx, rbacService, workspaceID, "admin", "post:delete")
}

func checkPermission(ctx context.Context, rbacService *rbac.Service, workspaceID, roleID, permissionID string) {
	hasPermission, err := rbacService.HasPermission(ctx, workspaceID, roleID, permissionID)
	if err != nil {
		log.Printf("Error checking permission: %v", err)
		return
	}
	
	if hasPermission {
		fmt.Printf("Role %s has permission %s in workspace %s\n", roleID, permissionID, workspaceID)
	} else {
		fmt.Printf("Role %s does NOT have permission %s in workspace %s\n", roleID, permissionID, workspaceID)
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
	
	// Define workspace ID
	workspaceID := "organization-123"

	// ... Create roles and permissions as in the previous example

	// Check permissions (results will be cached)
	hasPermission, err := rbacService.HasPermission(ctx, workspaceID, "admin", "post:read")
	if err != nil {
		log.Fatalf("Failed to check permission: %v", err)
	}

	// Invalidate cache for a specific role in a workspace when role or permissions change
	rbacService.InvalidateCache(workspaceID, "admin")
	
	// Invalidate all caches for a specific workspace
	rbacService.InvalidateWorkspaceCache(workspaceID)

	// Invalidate the entire cache across all workspaces
	rbacService.InvalidateAllCache()
}
```

### Working with Multiple Workspaces

```go
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/dmitrymomot/gokit/rbac"
)

func main() {
	ctx := context.Background()
	store := rbac.NewMemoryStore()
	rbacService := rbac.NewService(store)
	
	// Define workspace IDs
	orgWorkspace := "organization-123"
	projectWorkspace := "project-456"
	
	// Create roles and permissions in first workspace
	createWorkspaceRoles(ctx, store, orgWorkspace)
	
	// Create roles and permissions in second workspace (with same IDs)
	createWorkspaceRoles(ctx, store, projectWorkspace)
	
	// Even though we used the same role IDs in both workspaces,
	// permissions are completely isolated
	
	// Administrator in organization workspace
	hasAdmin, err := rbacService.HasPermission(ctx, orgWorkspace, "admin", "system:admin")
	if err != nil {
		log.Fatalf("Error checking permission: %v", err)
	}
	fmt.Printf("In workspace %s, admin has system:admin permission: %v\n", 
		orgWorkspace, hasAdmin) // true
	
	// Administrator in project workspace (doesn't have the special admin permission)
	hasAdmin, err = rbacService.HasPermission(ctx, projectWorkspace, "admin", "system:admin")
	if err != nil {
		log.Fatalf("Error checking permission: %v", err)
	}
	fmt.Printf("In workspace %s, admin has system:admin permission: %v\n", 
		projectWorkspace, hasAdmin) // false
}

func createWorkspaceRoles(ctx context.Context, store rbac.Store, workspaceID string) {
	// Create basic permissions
	readPerm := rbac.Permission{
		WorkspaceID: workspaceID,
		ID:          "read",
		Name:        "Read Access",
	}
	
	writePerm := rbac.Permission{
		WorkspaceID: workspaceID,
		ID:          "write",
		Name:        "Write Access",
		ParentIDs:   []string{"read"},
	}
	
	// Only add admin permission in organization workspace
	if workspaceID == "organization-123" {
		adminPerm := rbac.Permission{
			WorkspaceID: workspaceID,
			ID:          "system:admin",
			Name:        "System Administration",
		}
		
		if err := store.CreatePermission(ctx, adminPerm); err != nil {
			log.Fatalf("Failed to create admin permission: %v", err)
		}
	}
	
	// Create the other permissions
	if err := store.CreatePermission(ctx, readPerm); err != nil {
		log.Fatalf("Failed to create read permission: %v", err)
	}
	
	if err := store.CreatePermission(ctx, writePerm); err != nil {
		log.Fatalf("Failed to create write permission: %v", err)
	}
	
	// Create basic roles with the same IDs in both workspaces
	guestRole := rbac.Role{
		WorkspaceID:         workspaceID,
		ID:                  "guest",
		Name:                "Guest",
		DirectPermissionIDs: []string{"read"},
	}
	
	adminRole := rbac.Role{
		WorkspaceID: workspaceID,
		ID:          "admin",
		Name:        "Administrator",
		DirectPermissionIDs: []string{"write"},
	}
	
	// Add the system:admin permission only in organization workspace
	if workspaceID == "organization-123" {
		adminRole.DirectPermissionIDs = append(adminRole.DirectPermissionIDs, "system:admin")
	}
	
	// Store the roles
	if err := store.CreateRole(ctx, guestRole); err != nil {
		log.Fatalf("Failed to create guest role: %v", err)
	}
	
	if err := store.CreateRole(ctx, adminRole); err != nil {
		log.Fatalf("Failed to create admin role: %v", err)
	}
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
	ctx := context.Background()
	store := rbac.NewMemoryStore()
	rbacService := rbac.NewService(store)
	
	workspaceID := "organization-123"
	
	// Create permissions with hierarchical structure
	setupPermissions(ctx, store, workspaceID)
	
	// Create roles
	setupRoles(ctx, store, workspaceID)
	
	// Check if role has any of the given permissions
	hasAny, err := rbacService.HasAnyPermission(ctx, workspaceID, "content-creator", 
		"post:edit", "post:delete", "user:manage")
	if err != nil {
		log.Fatalf("Error checking permissions: %v", err)
	}
	fmt.Printf("Content creator has any of the permissions: %v\n", hasAny) // true (has post:edit)
	
	// Check if role has ALL of the given permissions
	hasAll, err := rbacService.HasAllPermissions(ctx, workspaceID, "admin", 
		"post:create", "post:edit", "post:delete")
	if err != nil {
		log.Fatalf("Error checking permissions: %v", err)
	}
	fmt.Printf("Admin has all post permissions: %v\n", hasAll) // true
	
	// Get all effective permissions (direct + inherited)
	perms, err := rbacService.GetEffectivePermissions(ctx, workspaceID, "editor")
	if err != nil {
		log.Fatalf("Error getting effective permissions: %v", err)
	}
	
	fmt.Println("Editor effective permissions:")
	for _, perm := range perms {
		fmt.Printf("- %s (%s)\n", perm.Name, perm.ID)
	}
}

func setupPermissions(ctx context.Context, store rbac.Store, workspaceID string) {
	// Post permission hierarchy
	store.CreatePermission(ctx, rbac.Permission{
		WorkspaceID: workspaceID,
		ID:          "post:view",
		Name:        "View Posts",
	})
	
	store.CreatePermission(ctx, rbac.Permission{
		WorkspaceID: workspaceID,
		ID:          "post:create",
		Name:        "Create Posts",
		ParentIDs:   []string{"post:view"},
	})
	
	store.CreatePermission(ctx, rbac.Permission{
		WorkspaceID: workspaceID,
		ID:          "post:edit",
		Name:        "Edit Posts",
		ParentIDs:   []string{"post:view"},
	})
	
	store.CreatePermission(ctx, rbac.Permission{
		WorkspaceID: workspaceID,
		ID:          "post:delete",
		Name:        "Delete Posts",
		ParentIDs:   []string{"post:edit"},
	})
	
	// User permission hierarchy
	store.CreatePermission(ctx, rbac.Permission{
		WorkspaceID: workspaceID,
		ID:          "user:view",
		Name:        "View Users",
	})
	
	store.CreatePermission(ctx, rbac.Permission{
		WorkspaceID: workspaceID,
		ID:          "user:manage",
		Name:        "Manage Users",
		ParentIDs:   []string{"user:view"},
	})
}

func setupRoles(ctx context.Context, store rbac.Store, workspaceID string) {
	// Guest - can only view content
	store.CreateRole(ctx, rbac.Role{
		WorkspaceID: workspaceID,
		ID:          "guest",
		Name:        "Guest",
		DirectPermissionIDs: []string{"post:view", "user:view"},
	})
	
	// Content Creator - can create and edit posts, inherits from guest
	store.CreateRole(ctx, rbac.Role{
		WorkspaceID: workspaceID,
		ID:          "content-creator",
		Name:        "Content Creator",
		ParentIDs:   []string{"guest"},
		DirectPermissionIDs: []string{"post:create", "post:edit"},
	})
	
	// Editor - can delete posts, inherits from content creator
	store.CreateRole(ctx, rbac.Role{
		WorkspaceID: workspaceID,
		ID:          "editor",
		Name:        "Editor",
		ParentIDs:   []string{"content-creator"},
		DirectPermissionIDs: []string{"post:delete"},
	})
	
	// Admin - can manage users, inherits from editor
	store.CreateRole(ctx, rbac.Role{
		WorkspaceID: workspaceID,
		ID:          "admin",
		Name:        "Administrator",
		ParentIDs:   []string{"editor"},
		DirectPermissionIDs: []string{"user:manage"},
	})
}
```

## API Reference

### Main Interfaces

- `Store`: Provides access to both role and permission storage functionality
- `RoleStore`: Interface for role storage operations
- `PermissionStore`: Interface for permission storage operations

### Core Structs

- `Role`: Represents a role with workspace ID, unique identifier, name, permissions, and parent roles
- `Permission`: Represents a permission with workspace ID, unique identifier, name, and parent permissions
- `Service`: Main implementation of the RBAC functionality with caching capabilities

### Key Methods

#### Service

- `HasPermission(ctx, workspaceID, roleID, permissionID)`: Checks if a role has a specific permission
- `HasAnyPermission(ctx, workspaceID, roleID, permissionIDs...)`: Checks if a role has any of the specified permissions
- `HasAllPermissions(ctx, workspaceID, roleID, permissionIDs...)`: Checks if a role has all of the specified permissions
- `GetEffectivePermissions(ctx, workspaceID, roleID)`: Gets all permissions a role has, including inherited ones
- `UpdateRole(ctx, role)`: Updates a role and invalidates its cache
- `InvalidateCache(workspaceID, roleID)`: Invalidates cache for a specific role in a workspace
- `InvalidateWorkspaceCache(workspaceID)`: Invalidates all caches for a specific workspace
- `InvalidateAllCache()`: Invalidates all caches across all workspaces

#### Store

- `CreateRole(ctx, role)`: Creates a new role
- `GetRole(ctx, workspaceID, roleID)`: Gets a role by ID
- `GetRoles(ctx, workspaceID)`: Gets all roles in a workspace
- `UpdateRole(ctx, role)`: Updates an existing role
- `DeleteRole(ctx, workspaceID, roleID)`: Deletes a role
- `AddRoleParent(ctx, workspaceID, roleID, parentID)`: Adds a parent to a role
- `RemoveRoleParent(ctx, workspaceID, roleID, parentID)`: Removes a parent from a role
- `GetRoleParents(ctx, workspaceID, roleID)`: Gets parent roles of a role
- `GetRoleChildren(ctx, workspaceID, roleID)`: Gets child roles of a role
- `CreatePermission(ctx, permission)`: Creates a new permission
- `GetPermission(ctx, workspaceID, permissionID)`: Gets a permission by ID
- `GetPermissions(ctx, workspaceID)`: Gets all permissions in a workspace
- `UpdatePermission(ctx, permission)`: Updates an existing permission
- `DeletePermission(ctx, workspaceID, permissionID)`: Deletes a permission
- `AddPermissionParent(ctx, workspaceID, permissionID, parentID)`: Adds a parent to a permission
- `RemovePermissionParent(ctx, workspaceID, permissionID, parentID)`: Removes a parent from a permission
- `GetPermissionParents(ctx, workspaceID, permissionID)`: Gets parent permissions of a permission
- `GetPermissionChildren(ctx, workspaceID, permissionID)`: Gets child permissions of a permission
- `AddPermissionToRole(ctx, workspaceID, roleID, permissionID)`: Adds a permission to a role
- `RemovePermissionFromRole(ctx, workspaceID, roleID, permissionID)`: Removes a permission from a role
- `GetRolePermissions(ctx, workspaceID, roleID)`: Gets direct permissions of a role

### Implementations

- `Service`: Implementation of the RBAC interface
- `MemoryStore`: Thread-safe in-memory implementation of the Store interface

package rbac

import "context"

// Role represents a role in the RBAC system.
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

// Permission represents a permission in the RBAC system.
type Permission struct {
	// ID is the unique identifier of the permission.
	ID string

	// Name is the display name of the permission.
	Name string

	// ParentIDs contains the IDs of the permissions this permission inherits from.
	ParentIDs []string
}

// RoleStore defines the interface for role storage operations.
type RoleStore interface {
	// CreateRole creates a new role.
	CreateRole(ctx context.Context, role Role) error

	// GetRole retrieves a role by its ID.
	GetRole(ctx context.Context, roleID string) (Role, error)

	// GetRoles retrieves all roles.
	GetRoles(ctx context.Context) ([]Role, error)

	// UpdateRole updates an existing role.
	UpdateRole(ctx context.Context, role Role) error

	// DeleteRole deletes a role by its ID.
	DeleteRole(ctx context.Context, roleID string) error

	// AddRoleParent adds a parent role to a role.
	AddRoleParent(ctx context.Context, roleID, parentRoleID string) error

	// RemoveRoleParent removes a parent role from a role.
	RemoveRoleParent(ctx context.Context, roleID, parentRoleID string) error

	// GetRoleParents retrieves all parent roles of a role.
	GetRoleParents(ctx context.Context, roleID string) ([]Role, error)

	// GetRoleChildren retrieves all roles that inherit from a role.
	GetRoleChildren(ctx context.Context, roleID string) ([]Role, error)

	// AddPermissionToRole adds a permission to a role.
	AddPermissionToRole(ctx context.Context, roleID, permissionID string) error

	// RemovePermissionFromRole removes a permission from a role.
	RemovePermissionFromRole(ctx context.Context, roleID, permissionID string) error

	// GetRolePermissions retrieves all permissions directly assigned to a role.
	GetRolePermissions(ctx context.Context, roleID string) ([]Permission, error)
}

// PermissionStore defines the interface for permission storage operations.
type PermissionStore interface {
	// CreatePermission creates a new permission.
	CreatePermission(ctx context.Context, permission Permission) error

	// GetPermission retrieves a permission by its ID.
	GetPermission(ctx context.Context, permissionID string) (Permission, error)

	// GetPermissions retrieves all permissions.
	GetPermissions(ctx context.Context) ([]Permission, error)

	// UpdatePermission updates an existing permission.
	UpdatePermission(ctx context.Context, permission Permission) error

	// DeletePermission deletes a permission by its ID.
	DeletePermission(ctx context.Context, permissionID string) error

	// AddPermissionParent adds a parent permission to a permission.
	AddPermissionParent(ctx context.Context, permissionID, parentPermissionID string) error

	// RemovePermissionParent removes a parent permission from a permission.
	RemovePermissionParent(ctx context.Context, permissionID, parentPermissionID string) error

	// GetPermissionParents retrieves all parent permissions of a permission.
	GetPermissionParents(ctx context.Context, permissionID string) ([]Permission, error)

	// GetPermissionChildren retrieves all permissions that inherit from a permission.
	GetPermissionChildren(ctx context.Context, permissionID string) ([]Permission, error)
}

// Store combines RoleStore and PermissionStore interfaces.
type Store interface {
	RoleStore
	PermissionStore
}

// RBAC defines the interface for RBAC operations.
type RBAC interface {
	// HasPermission checks if a role has a specific permission.
	HasPermission(ctx context.Context, roleID, permissionID string) (bool, error)

	// HasAnyPermission checks if a role has at least one of the specified permissions.
	HasAnyPermission(ctx context.Context, roleID string, permissionIDs ...string) (bool, error)

	// HasAllPermissions checks if a role has all of the specified permissions.
	HasAllPermissions(ctx context.Context, roleID string, permissionIDs ...string) (bool, error)

	// GetEffectivePermissions retrieves all permissions a role has, including inherited permissions.
	GetEffectivePermissions(ctx context.Context, roleID string) ([]Permission, error)

	// Store returns the underlying store.
	Store() Store
}

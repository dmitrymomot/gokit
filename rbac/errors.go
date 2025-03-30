package rbac

import "errors"

// Common RBAC errors.
var (
	// ErrRoleNotFound is returned when the role is not found.
	ErrRoleNotFound = errors.New("role not found")
	
	// ErrRoleAlreadyExists is returned when the role already exists.
	ErrRoleAlreadyExists = errors.New("role already exists")
	
	// ErrPermissionNotFound is returned when the permission is not found.
	ErrPermissionNotFound = errors.New("permission not found")
	
	// ErrPermissionAlreadyExists is returned when the permission already exists.
	ErrPermissionAlreadyExists = errors.New("permission already exists")
	
	// ErrInvalidRoleID is returned when the role ID is invalid.
	ErrInvalidRoleID = errors.New("invalid role ID")
	
	// ErrInvalidRoleName is returned when the role name is invalid.
	ErrInvalidRoleName = errors.New("invalid role name")
	
	// ErrInvalidPermissionID is returned when the permission ID is invalid.
	ErrInvalidPermissionID = errors.New("invalid permission ID")
	
	// ErrInvalidPermissionName is returned when the permission name is invalid.
	ErrInvalidPermissionName = errors.New("invalid permission name")
	
	// ErrCyclicInheritance is returned when a cyclic inheritance is detected.
	ErrCyclicInheritance = errors.New("cyclic inheritance detected")
	
	// ErrStoreFailure is returned when a store operation fails.
	ErrStoreFailure = errors.New("store operation failed")
	
	// ErrInvalidArgument is returned when an invalid argument is provided.
	ErrInvalidArgument = errors.New("invalid argument")
	
	// ErrUnauthorized is returned when a user is not authorized to perform an action.
	ErrUnauthorized = errors.New("unauthorized")
)

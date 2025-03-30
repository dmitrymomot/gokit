package rbac

import (
	"context"
	"sync"
	"time"
)

// Service implements the RBAC interface.
type Service struct {
	store          Store
	permissionCache map[string][]string // Maps role ID to effective permission IDs
	cacheMu        sync.RWMutex
	cacheEnabled   bool
	cacheTTL       time.Duration
	cacheLastUpdated map[string]time.Time // Maps role ID to last cache update time
}

// ServiceOption is a function that configures a Service.
type ServiceOption func(*Service)

// WithCaching enables caching of effective permissions with a specified TTL.
func WithCaching(ttl time.Duration) ServiceOption {
	return func(s *Service) {
		s.cacheEnabled = true
		s.cacheTTL = ttl
	}
}

// NewService creates a new RBAC service with the provided store.
func NewService(store Store, options ...ServiceOption) *Service {
	s := &Service{
		store:           store,
		permissionCache: make(map[string][]string),
		cacheLastUpdated: make(map[string]time.Time),
		cacheEnabled:    false,
		cacheTTL:        5 * time.Minute, // Default TTL
	}

	// Apply options
	for _, option := range options {
		option(s)
	}

	return s
}

// HasPermission checks if a role has a specific permission.
func (s *Service) HasPermission(ctx context.Context, roleID, permissionID string) (bool, error) {
	if roleID == "" || permissionID == "" {
		return false, ErrInvalidArgument
	}

	// Get effective permissions for the role
	effectivePermissionIDs, err := s.getEffectivePermissionIDs(ctx, roleID)
	if err != nil {
		return false, err
	}

	// Check if the permission is in the effective permissions
	for _, epID := range effectivePermissionIDs {
		if epID == permissionID {
			return true, nil
		}
	}

	return false, nil
}

// HasAnyPermission checks if a role has at least one of the specified permissions.
func (s *Service) HasAnyPermission(ctx context.Context, roleID string, permissionIDs ...string) (bool, error) {
	if roleID == "" || len(permissionIDs) == 0 {
		return false, ErrInvalidArgument
	}

	// Get effective permissions for the role
	effectivePermissionIDs, err := s.getEffectivePermissionIDs(ctx, roleID)
	if err != nil {
		return false, err
	}

	// Create a map for faster lookups
	effectivePermissionMap := make(map[string]bool, len(effectivePermissionIDs))
	for _, epID := range effectivePermissionIDs {
		effectivePermissionMap[epID] = true
	}

	// Check if any of the permissions are in the effective permissions
	for _, pID := range permissionIDs {
		if effectivePermissionMap[pID] {
			return true, nil
		}
	}

	return false, nil
}

// HasAllPermissions checks if a role has all of the specified permissions.
func (s *Service) HasAllPermissions(ctx context.Context, roleID string, permissionIDs ...string) (bool, error) {
	if roleID == "" || len(permissionIDs) == 0 {
		return false, ErrInvalidArgument
	}

	// Get effective permissions for the role
	effectivePermissionIDs, err := s.getEffectivePermissionIDs(ctx, roleID)
	if err != nil {
		return false, err
	}

	// Create a map for faster lookups
	effectivePermissionMap := make(map[string]bool, len(effectivePermissionIDs))
	for _, epID := range effectivePermissionIDs {
		effectivePermissionMap[epID] = true
	}

	// Check if all of the permissions are in the effective permissions
	for _, pID := range permissionIDs {
		if !effectivePermissionMap[pID] {
			return false, nil
		}
	}

	return true, nil
}

// GetEffectivePermissions retrieves all permissions a role has, including inherited permissions.
func (s *Service) GetEffectivePermissions(ctx context.Context, roleID string) ([]Permission, error) {
	if roleID == "" {
		return nil, ErrInvalidArgument
	}

	// Get the effective permission IDs
	effectivePermissionIDs, err := s.getEffectivePermissionIDs(ctx, roleID)
	if err != nil {
		return nil, err
	}

	// Get the actual permission objects
	permissions := make([]Permission, 0, len(effectivePermissionIDs))
	for _, pID := range effectivePermissionIDs {
		permission, err := s.store.GetPermission(ctx, pID)
		if err != nil {
			return nil, err
		}
		permissions = append(permissions, permission)
	}

	return permissions, nil
}

// Store returns the underlying store.
func (s *Service) Store() Store {
	return s.store
}

// InvalidateCache invalidates the permission cache for a specific role.
func (s *Service) InvalidateCache(roleID string) {
	if !s.cacheEnabled {
		return
	}

	s.cacheMu.Lock()
	defer s.cacheMu.Unlock()

	delete(s.permissionCache, roleID)
	delete(s.cacheLastUpdated, roleID)
}

// InvalidateAllCache invalidates the entire permission cache.
func (s *Service) InvalidateAllCache() {
	if !s.cacheEnabled {
		return
	}

	s.cacheMu.Lock()
	defer s.cacheMu.Unlock()

	s.permissionCache = make(map[string][]string)
	s.cacheLastUpdated = make(map[string]time.Time)
}

// AddPermissionToRole adds a permission to a role and updates the cache.
func (s *Service) AddPermissionToRole(ctx context.Context, roleID, permissionID string) error {
	if roleID == "" || permissionID == "" {
		return ErrInvalidArgument
	}

	// Add permission to role
	err := s.store.AddPermissionToRole(ctx, roleID, permissionID)
	if err != nil {
		return err
	}

	// Invalidate cache for role
	s.InvalidateCache(roleID)

	return nil
}

// getEffectivePermissionIDs retrieves all permission IDs a role has, including inherited permissions.
func (s *Service) getEffectivePermissionIDs(ctx context.Context, roleID string) ([]string, error) {
	// Check cache if enabled
	if s.cacheEnabled {
		s.cacheMu.RLock()
		cachedPermissions, exists := s.permissionCache[roleID]
		lastUpdated := s.cacheLastUpdated[roleID]
		s.cacheMu.RUnlock()

		if exists && time.Since(lastUpdated) < s.cacheTTL {
			return cachedPermissions, nil
		}
	}

	// Get the role
	role, err := s.store.GetRole(ctx, roleID)
	if err != nil {
		return nil, err
	}

	// Use a map to avoid duplicates
	effectivePermissions := make(map[string]bool)

	// Add direct permissions
	for _, pID := range role.DirectPermissionIDs {
		effectivePermissions[pID] = true
		
		// Add inherited permissions
		inheritedPermIDs, err := s.getInheritedPermissionIDs(ctx, pID, make(map[string]bool))
		if err != nil {
			return nil, err
		}
		
		for _, ipID := range inheritedPermIDs {
			effectivePermissions[ipID] = true
		}
	}

	// Add permissions from parent roles
	inheritedRolePermIDs, err := s.getRoleInheritedPermissionIDs(ctx, roleID, make(map[string]bool))
	if err != nil {
		return nil, err
	}
	
	for _, ipID := range inheritedRolePermIDs {
		effectivePermissions[ipID] = true
	}

	// Convert map to slice
	result := make([]string, 0, len(effectivePermissions))
	for pID := range effectivePermissions {
		result = append(result, pID)
	}

	// Update cache if enabled
	if s.cacheEnabled {
		s.cacheMu.Lock()
		s.permissionCache[roleID] = result
		s.cacheLastUpdated[roleID] = time.Now()
		s.cacheMu.Unlock()
	}

	return result, nil
}

// getInheritedPermissionIDs recursively retrieves all permission IDs that a permission inherits,
// using a visited map to avoid cycles.
func (s *Service) getInheritedPermissionIDs(ctx context.Context, permissionID string, visited map[string]bool) ([]string, error) {
	if visited[permissionID] {
		return []string{}, nil
	}
	visited[permissionID] = true

	permission, err := s.store.GetPermission(ctx, permissionID)
	if err != nil {
		return nil, err
	}

	result := []string{}

	// Add parent permission IDs
	for _, parentID := range permission.ParentIDs {
		result = append(result, parentID)

		// Recursively add parents of parents
		parentInheritedPermIDs, err := s.getInheritedPermissionIDs(ctx, parentID, visited)
		if err != nil {
			return nil, err
		}
		result = append(result, parentInheritedPermIDs...)
	}

	return result, nil
}

// getRoleInheritedPermissionIDs recursively retrieves all permission IDs from parent roles,
// using a visited map to avoid cycles.
func (s *Service) getRoleInheritedPermissionIDs(ctx context.Context, roleID string, visited map[string]bool) ([]string, error) {
	if visited[roleID] {
		return []string{}, nil
	}
	visited[roleID] = true

	role, err := s.store.GetRole(ctx, roleID)
	if err != nil {
		return nil, err
	}

	result := make(map[string]bool)

	// Process parent roles
	for _, parentID := range role.ParentIDs {
		// Get parent role
		parentRole, err := s.store.GetRole(ctx, parentID)
		if err != nil {
			return nil, err
		}

		// Add direct permissions from parent
		for _, pID := range parentRole.DirectPermissionIDs {
			result[pID] = true
			
			// Add inherited permissions from each direct permission
			inheritedPermIDs, err := s.getInheritedPermissionIDs(ctx, pID, make(map[string]bool))
			if err != nil {
				return nil, err
			}
			
			for _, ipID := range inheritedPermIDs {
				result[ipID] = true
			}
		}

		// Recursively process parent's parents
		parentInheritedPermIDs, err := s.getRoleInheritedPermissionIDs(ctx, parentID, visited)
		if err != nil {
			return nil, err
		}
		
		for _, ipID := range parentInheritedPermIDs {
			result[ipID] = true
		}
	}

	// Convert map to slice
	permissions := make([]string, 0, len(result))
	for pID := range result {
		permissions = append(permissions, pID)
	}

	return permissions, nil
}

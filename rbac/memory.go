package rbac

import (
	"context"
	"sync"
)

// MemoryStore is an in-memory implementation of the Store interface.
type MemoryStore struct {
	roles       map[string]Role
	permissions map[string]Permission
	mu          sync.RWMutex
}

// NewMemoryStore creates a new in-memory store.
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		roles:       make(map[string]Role),
		permissions: make(map[string]Permission),
	}
}

// CreateRole creates a new role.
func (s *MemoryStore) CreateRole(ctx context.Context, role Role) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if role.ID == "" {
		return ErrInvalidRoleID
	}

	if role.Name == "" {
		return ErrInvalidRoleName
	}

	if _, exists := s.roles[role.ID]; exists {
		return ErrRoleAlreadyExists
	}

	// Validate parent roles exist
	for _, parentID := range role.ParentIDs {
		if _, exists := s.roles[parentID]; !exists {
			return ErrRoleNotFound
		}
	}

	// Validate permissions exist
	for _, permID := range role.DirectPermissionIDs {
		if _, exists := s.permissions[permID]; !exists {
			return ErrPermissionNotFound
		}
	}

	// Check for cyclic inheritance
	if err := s.checkCyclicRoleInheritance(role.ID, role.ParentIDs); err != nil {
		return err
	}

	// Store the role
	s.roles[role.ID] = role
	return nil
}

// GetRole retrieves a role by its ID.
func (s *MemoryStore) GetRole(ctx context.Context, roleID string) (Role, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	role, exists := s.roles[roleID]
	if !exists {
		return Role{}, ErrRoleNotFound
	}

	return role, nil
}

// GetRoles retrieves all roles.
func (s *MemoryStore) GetRoles(ctx context.Context) ([]Role, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	roles := make([]Role, 0, len(s.roles))
	for _, role := range s.roles {
		roles = append(roles, role)
	}

	return roles, nil
}

// UpdateRole updates an existing role.
func (s *MemoryStore) UpdateRole(ctx context.Context, role Role) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if role.ID == "" {
		return ErrInvalidRoleID
	}

	if role.Name == "" {
		return ErrInvalidRoleName
	}

	if _, exists := s.roles[role.ID]; !exists {
		return ErrRoleNotFound
	}

	// Validate parent roles exist
	for _, parentID := range role.ParentIDs {
		if _, exists := s.roles[parentID]; !exists {
			return ErrRoleNotFound
		}
	}

	// Validate permissions exist
	for _, permID := range role.DirectPermissionIDs {
		if _, exists := s.permissions[permID]; !exists {
			return ErrPermissionNotFound
		}
	}

	// Check for cyclic inheritance
	if err := s.checkCyclicRoleInheritance(role.ID, role.ParentIDs); err != nil {
		return err
	}

	// Update the role
	s.roles[role.ID] = role
	return nil
}

// DeleteRole deletes a role by its ID.
func (s *MemoryStore) DeleteRole(ctx context.Context, roleID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.roles[roleID]; !exists {
		return ErrRoleNotFound
	}

	// Check if any role has this role as parent
	for _, role := range s.roles {
		for _, parentID := range role.ParentIDs {
			if parentID == roleID {
				// Remove this parent from the role
				newParents := make([]string, 0, len(role.ParentIDs)-1)
				for _, pid := range role.ParentIDs {
					if pid != roleID {
						newParents = append(newParents, pid)
					}
				}
				role.ParentIDs = newParents
				s.roles[role.ID] = role
			}
		}
	}

	delete(s.roles, roleID)
	return nil
}

// AddRoleParent adds a parent role to a role.
func (s *MemoryStore) AddRoleParent(ctx context.Context, roleID, parentRoleID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	role, exists := s.roles[roleID]
	if !exists {
		return ErrRoleNotFound
	}

	if _, exists := s.roles[parentRoleID]; !exists {
		return ErrRoleNotFound
	}

	// Check if parent already exists
	for _, pid := range role.ParentIDs {
		if pid == parentRoleID {
			return nil // Parent already exists, no need to add again
		}
	}

	// Check for cyclic inheritance
	tempParents := append([]string{}, role.ParentIDs...)
	tempParents = append(tempParents, parentRoleID)
	if err := s.checkCyclicRoleInheritance(roleID, tempParents); err != nil {
		return err
	}

	// Add the parent
	role.ParentIDs = append(role.ParentIDs, parentRoleID)
	s.roles[roleID] = role
	return nil
}

// RemoveRoleParent removes a parent role from a role.
func (s *MemoryStore) RemoveRoleParent(ctx context.Context, roleID, parentRoleID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	role, exists := s.roles[roleID]
	if !exists {
		return ErrRoleNotFound
	}

	// Remove the parent if it exists
	found := false
	newParents := make([]string, 0, len(role.ParentIDs))
	for _, pid := range role.ParentIDs {
		if pid != parentRoleID {
			newParents = append(newParents, pid)
		} else {
			found = true
		}
	}

	if !found {
		return ErrRoleNotFound
	}

	role.ParentIDs = newParents
	s.roles[roleID] = role
	return nil
}

// GetRoleParents retrieves all parent roles of a role.
func (s *MemoryStore) GetRoleParents(ctx context.Context, roleID string) ([]Role, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	role, exists := s.roles[roleID]
	if !exists {
		return nil, ErrRoleNotFound
	}

	parents := make([]Role, 0, len(role.ParentIDs))
	for _, pid := range role.ParentIDs {
		if parent, exists := s.roles[pid]; exists {
			parents = append(parents, parent)
		}
	}

	return parents, nil
}

// GetRoleChildren retrieves all roles that inherit from a role.
func (s *MemoryStore) GetRoleChildren(ctx context.Context, roleID string) ([]Role, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if _, exists := s.roles[roleID]; !exists {
		return nil, ErrRoleNotFound
	}

	children := []Role{}
	for _, role := range s.roles {
		for _, pid := range role.ParentIDs {
			if pid == roleID {
				children = append(children, role)
				break
			}
		}
	}

	return children, nil
}

// AddPermissionToRole adds a permission to a role.
func (s *MemoryStore) AddPermissionToRole(ctx context.Context, roleID, permissionID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	role, exists := s.roles[roleID]
	if !exists {
		return ErrRoleNotFound
	}

	if _, exists := s.permissions[permissionID]; !exists {
		return ErrPermissionNotFound
	}

	// Check if permission already exists
	for _, pid := range role.DirectPermissionIDs {
		if pid == permissionID {
			return nil // Permission already exists, no need to add again
		}
	}

	// Add the permission
	role.DirectPermissionIDs = append(role.DirectPermissionIDs, permissionID)
	s.roles[roleID] = role
	return nil
}

// RemovePermissionFromRole removes a permission from a role.
func (s *MemoryStore) RemovePermissionFromRole(ctx context.Context, roleID, permissionID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	role, exists := s.roles[roleID]
	if !exists {
		return ErrRoleNotFound
	}

	// Remove the permission if it exists
	found := false
	newPermissions := make([]string, 0, len(role.DirectPermissionIDs))
	for _, pid := range role.DirectPermissionIDs {
		if pid != permissionID {
			newPermissions = append(newPermissions, pid)
		} else {
			found = true
		}
	}

	if !found {
		return ErrPermissionNotFound
	}

	role.DirectPermissionIDs = newPermissions
	s.roles[roleID] = role
	return nil
}

// GetRolePermissions retrieves all permissions directly assigned to a role.
func (s *MemoryStore) GetRolePermissions(ctx context.Context, roleID string) ([]Permission, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	role, exists := s.roles[roleID]
	if !exists {
		return nil, ErrRoleNotFound
	}

	permissions := make([]Permission, 0, len(role.DirectPermissionIDs))
	for _, pid := range role.DirectPermissionIDs {
		if perm, exists := s.permissions[pid]; exists {
			permissions = append(permissions, perm)
		}
	}

	return permissions, nil
}

// CreatePermission creates a new permission.
func (s *MemoryStore) CreatePermission(ctx context.Context, permission Permission) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if permission.ID == "" {
		return ErrInvalidPermissionID
	}

	if permission.Name == "" {
		return ErrInvalidPermissionName
	}

	if _, exists := s.permissions[permission.ID]; exists {
		return ErrPermissionAlreadyExists
	}

	// Validate parent permissions exist
	for _, parentID := range permission.ParentIDs {
		if _, exists := s.permissions[parentID]; !exists {
			return ErrPermissionNotFound
		}
	}

	// Check for cyclic inheritance
	if err := s.checkCyclicPermissionInheritance(permission.ID, permission.ParentIDs); err != nil {
		return err
	}

	// Store the permission
	s.permissions[permission.ID] = permission
	return nil
}

// GetPermission retrieves a permission by its ID.
func (s *MemoryStore) GetPermission(ctx context.Context, permissionID string) (Permission, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	permission, exists := s.permissions[permissionID]
	if !exists {
		return Permission{}, ErrPermissionNotFound
	}

	return permission, nil
}

// GetPermissions retrieves all permissions.
func (s *MemoryStore) GetPermissions(ctx context.Context) ([]Permission, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	permissions := make([]Permission, 0, len(s.permissions))
	for _, permission := range s.permissions {
		permissions = append(permissions, permission)
	}

	return permissions, nil
}

// UpdatePermission updates an existing permission.
func (s *MemoryStore) UpdatePermission(ctx context.Context, permission Permission) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if permission.ID == "" {
		return ErrInvalidPermissionID
	}

	if permission.Name == "" {
		return ErrInvalidPermissionName
	}

	if _, exists := s.permissions[permission.ID]; !exists {
		return ErrPermissionNotFound
	}

	// Validate parent permissions exist
	for _, parentID := range permission.ParentIDs {
		if _, exists := s.permissions[parentID]; !exists {
			return ErrPermissionNotFound
		}
	}

	// Check for cyclic inheritance
	if err := s.checkCyclicPermissionInheritance(permission.ID, permission.ParentIDs); err != nil {
		return err
	}

	// Update the permission
	s.permissions[permission.ID] = permission
	return nil
}

// DeletePermission deletes a permission by its ID.
func (s *MemoryStore) DeletePermission(ctx context.Context, permissionID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.permissions[permissionID]; !exists {
		return ErrPermissionNotFound
	}

	// Check if any permission has this permission as parent
	for _, perm := range s.permissions {
		for _, parentID := range perm.ParentIDs {
			if parentID == permissionID {
				// Remove this parent from the permission
				newParents := make([]string, 0, len(perm.ParentIDs)-1)
				for _, pid := range perm.ParentIDs {
					if pid != permissionID {
						newParents = append(newParents, pid)
					}
				}
				perm.ParentIDs = newParents
				s.permissions[perm.ID] = perm
			}
		}
	}

	// Check if any role has this permission
	for _, role := range s.roles {
		for _, pid := range role.DirectPermissionIDs {
			if pid == permissionID {
				// Remove this permission from the role
				newPermissions := make([]string, 0, len(role.DirectPermissionIDs)-1)
				for _, pid := range role.DirectPermissionIDs {
					if pid != permissionID {
						newPermissions = append(newPermissions, pid)
					}
				}
				role.DirectPermissionIDs = newPermissions
				s.roles[role.ID] = role
			}
		}
	}

	delete(s.permissions, permissionID)
	return nil
}

// AddPermissionParent adds a parent permission to a permission.
func (s *MemoryStore) AddPermissionParent(ctx context.Context, permissionID, parentPermissionID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	permission, exists := s.permissions[permissionID]
	if !exists {
		return ErrPermissionNotFound
	}

	if _, exists := s.permissions[parentPermissionID]; !exists {
		return ErrPermissionNotFound
	}

	// Check if parent already exists
	for _, pid := range permission.ParentIDs {
		if pid == parentPermissionID {
			return nil // Parent already exists, no need to add again
		}
	}

	// Check for cyclic inheritance
	tempParents := append([]string{}, permission.ParentIDs...)
	tempParents = append(tempParents, parentPermissionID)
	if err := s.checkCyclicPermissionInheritance(permissionID, tempParents); err != nil {
		return err
	}

	// Add the parent
	permission.ParentIDs = append(permission.ParentIDs, parentPermissionID)
	s.permissions[permissionID] = permission
	return nil
}

// RemovePermissionParent removes a parent permission from a permission.
func (s *MemoryStore) RemovePermissionParent(ctx context.Context, permissionID, parentPermissionID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	permission, exists := s.permissions[permissionID]
	if !exists {
		return ErrPermissionNotFound
	}

	// Remove the parent if it exists
	found := false
	newParents := make([]string, 0, len(permission.ParentIDs))
	for _, pid := range permission.ParentIDs {
		if pid != parentPermissionID {
			newParents = append(newParents, pid)
		} else {
			found = true
		}
	}

	if !found {
		return ErrPermissionNotFound
	}

	permission.ParentIDs = newParents
	s.permissions[permissionID] = permission
	return nil
}

// GetPermissionParents retrieves all parent permissions of a permission.
func (s *MemoryStore) GetPermissionParents(ctx context.Context, permissionID string) ([]Permission, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	permission, exists := s.permissions[permissionID]
	if !exists {
		return nil, ErrPermissionNotFound
	}

	parents := make([]Permission, 0, len(permission.ParentIDs))
	for _, pid := range permission.ParentIDs {
		if parent, exists := s.permissions[pid]; exists {
			parents = append(parents, parent)
		}
	}

	return parents, nil
}

// GetPermissionChildren retrieves all permissions that inherit from a permission.
func (s *MemoryStore) GetPermissionChildren(ctx context.Context, permissionID string) ([]Permission, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if _, exists := s.permissions[permissionID]; !exists {
		return nil, ErrPermissionNotFound
	}

	children := []Permission{}
	for _, perm := range s.permissions {
		for _, pid := range perm.ParentIDs {
			if pid == permissionID {
				children = append(children, perm)
				break
			}
		}
	}

	return children, nil
}

// checkCyclicRoleInheritance checks if adding the given parents to the role would create a cyclic inheritance.
func (s *MemoryStore) checkCyclicRoleInheritance(roleID string, parentIDs []string) error {
	// Create a map of visited roles to detect cycles
	visited := make(map[string]bool)
	
	// Helper function to check for cycles using DFS
	var checkCycle func(currentID string) bool
	checkCycle = func(currentID string) bool {
		if currentID == roleID {
			return true // Cycle detected
		}
		
		if visited[currentID] {
			return false // Already visited, no cycle detected in this path
		}
		
		visited[currentID] = true
		
		// Check all parents of the current role
		currentRole, exists := s.roles[currentID]
		if exists {
			for _, pid := range currentRole.ParentIDs {
				if checkCycle(pid) {
					return true
				}
			}
		}
		
		return false
	}
	
	// Check each parent for cycles
	for _, pid := range parentIDs {
		// Reset visited map for each parent
		visited = make(map[string]bool)
		
		if checkCycle(pid) {
			return ErrCyclicInheritance
		}
	}
	
	return nil
}

// checkCyclicPermissionInheritance checks if adding the given parents to the permission would create a cyclic inheritance.
func (s *MemoryStore) checkCyclicPermissionInheritance(permissionID string, parentIDs []string) error {
	// Create a map of visited permissions to detect cycles
	visited := make(map[string]bool)
	
	// Helper function to check for cycles using DFS
	var checkCycle func(currentID string) bool
	checkCycle = func(currentID string) bool {
		if currentID == permissionID {
			return true // Cycle detected
		}
		
		if visited[currentID] {
			return false // Already visited, no cycle detected in this path
		}
		
		visited[currentID] = true
		
		// Check all parents of the current permission
		currentPermission, exists := s.permissions[currentID]
		if exists {
			for _, pid := range currentPermission.ParentIDs {
				if checkCycle(pid) {
					return true
				}
			}
		}
		
		return false
	}
	
	// Check each parent for cycles
	for _, pid := range parentIDs {
		// Reset visited map for each parent
		visited = make(map[string]bool)
		
		if checkCycle(pid) {
			return ErrCyclicInheritance
		}
	}
	
	return nil
}

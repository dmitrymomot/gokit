package rbac

import (
	"context"
	"fmt"
	"sync"
)

// MemoryStore is an in-memory implementation of the Store interface.
type MemoryStore struct {
	// Maps workspaceID:roleID to Role
	roles       map[string]Role
	// Maps workspaceID:permissionID to Permission
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

// keyFor creates a composite key from workspace ID and entity ID
func (s *MemoryStore) keyFor(workspaceID, entityID string) string {
	return fmt.Sprintf("%s:%s", workspaceID, entityID)
}

// workspaceRoleIDs returns all role IDs for a specific workspace
func (s *MemoryStore) workspaceRoleIDs(workspaceID string) []string {
	result := []string{}
	prefix := workspaceID + ":"
	
	for key, role := range s.roles {
		if len(key) > len(prefix) && key[:len(prefix)] == prefix && role.WorkspaceID == workspaceID {
			result = append(result, role.ID)
		}
	}
	
	return result
}

// workspacePermissionIDs returns all permission IDs for a specific workspace
func (s *MemoryStore) workspacePermissionIDs(workspaceID string) []string {
	result := []string{}
	prefix := workspaceID + ":"
	
	for key, permission := range s.permissions {
		if len(key) > len(prefix) && key[:len(prefix)] == prefix && permission.WorkspaceID == workspaceID {
			result = append(result, permission.ID)
		}
	}
	
	return result
}

// CreateRole creates a new role.
func (s *MemoryStore) CreateRole(ctx context.Context, role Role) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if role.WorkspaceID == "" {
		return ErrInvalidArgument
	}

	if role.ID == "" {
		return ErrInvalidRoleID
	}

	if role.Name == "" {
		return ErrInvalidRoleName
	}

	key := s.keyFor(role.WorkspaceID, role.ID)
	if _, exists := s.roles[key]; exists {
		return ErrRoleAlreadyExists
	}

	// Validate parent roles exist
	for _, parentID := range role.ParentIDs {
		parentKey := s.keyFor(role.WorkspaceID, parentID)
		if _, exists := s.roles[parentKey]; !exists {
			return ErrRoleNotFound
		}
	}

	// Validate permissions exist
	for _, permID := range role.DirectPermissionIDs {
		permKey := s.keyFor(role.WorkspaceID, permID)
		if _, exists := s.permissions[permKey]; !exists {
			return ErrPermissionNotFound
		}
	}

	// Check for cyclic inheritance
	if err := s.checkCyclicRoleInheritance(role.WorkspaceID, role.ID, role.ParentIDs); err != nil {
		return err
	}

	// Store the role
	s.roles[key] = role
	return nil
}

// GetRole retrieves a role by its ID.
func (s *MemoryStore) GetRole(ctx context.Context, workspaceID, roleID string) (Role, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if workspaceID == "" || roleID == "" {
		return Role{}, ErrInvalidArgument
	}

	key := s.keyFor(workspaceID, roleID)
	role, exists := s.roles[key]
	if !exists {
		return Role{}, ErrRoleNotFound
	}

	return role, nil
}

// GetRoles retrieves all roles for a workspace.
func (s *MemoryStore) GetRoles(ctx context.Context, workspaceID string) ([]Role, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if workspaceID == "" {
		return nil, ErrInvalidArgument
	}

	roles := make([]Role, 0)
	prefix := workspaceID + ":"
	
	for key, role := range s.roles {
		if len(key) > len(prefix) && key[:len(prefix)] == prefix && role.WorkspaceID == workspaceID {
			roles = append(roles, role)
		}
	}

	return roles, nil
}

// UpdateRole updates an existing role.
func (s *MemoryStore) UpdateRole(ctx context.Context, role Role) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if role.WorkspaceID == "" {
		return ErrInvalidArgument
	}

	if role.ID == "" {
		return ErrInvalidRoleID
	}

	if role.Name == "" {
		return ErrInvalidRoleName
	}

	key := s.keyFor(role.WorkspaceID, role.ID)
	if _, exists := s.roles[key]; !exists {
		return ErrRoleNotFound
	}

	// Validate parent roles exist
	for _, parentID := range role.ParentIDs {
		parentKey := s.keyFor(role.WorkspaceID, parentID)
		if _, exists := s.roles[parentKey]; !exists {
			return ErrRoleNotFound
		}
	}

	// Validate permissions exist
	for _, permID := range role.DirectPermissionIDs {
		permKey := s.keyFor(role.WorkspaceID, permID)
		if _, exists := s.permissions[permKey]; !exists {
			return ErrPermissionNotFound
		}
	}

	// Check for cyclic inheritance
	if err := s.checkCyclicRoleInheritance(role.WorkspaceID, role.ID, role.ParentIDs); err != nil {
		return err
	}

	// Update the role
	s.roles[key] = role
	return nil
}

// DeleteRole deletes a role by its ID.
func (s *MemoryStore) DeleteRole(ctx context.Context, workspaceID, roleID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if workspaceID == "" || roleID == "" {
		return ErrInvalidArgument
	}

	key := s.keyFor(workspaceID, roleID)
	if _, exists := s.roles[key]; !exists {
		return ErrRoleNotFound
	}

	// Check if any role has this role as parent
	prefix := workspaceID + ":"
	for k, role := range s.roles {
		if len(k) > len(prefix) && k[:len(prefix)] == prefix && role.WorkspaceID == workspaceID {
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
					s.roles[k] = role
				}
			}
		}
	}

	delete(s.roles, key)
	return nil
}

// AddRoleParent adds a parent role to a role.
func (s *MemoryStore) AddRoleParent(ctx context.Context, workspaceID, roleID, parentRoleID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if workspaceID == "" || roleID == "" || parentRoleID == "" {
		return ErrInvalidArgument
	}

	roleKey := s.keyFor(workspaceID, roleID)
	role, exists := s.roles[roleKey]
	if !exists {
		return ErrRoleNotFound
	}

	parentKey := s.keyFor(workspaceID, parentRoleID)
	if _, exists := s.roles[parentKey]; !exists {
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
	if err := s.checkCyclicRoleInheritance(workspaceID, roleID, tempParents); err != nil {
		return err
	}

	// Add the parent
	role.ParentIDs = append(role.ParentIDs, parentRoleID)
	s.roles[roleKey] = role
	return nil
}

// RemoveRoleParent removes a parent role from a role.
func (s *MemoryStore) RemoveRoleParent(ctx context.Context, workspaceID, roleID, parentRoleID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if workspaceID == "" || roleID == "" || parentRoleID == "" {
		return ErrInvalidArgument
	}

	roleKey := s.keyFor(workspaceID, roleID)
	role, exists := s.roles[roleKey]
	if !exists {
		return ErrRoleNotFound
	}

	// Check if parent exists
	parentFound := false
	for _, pid := range role.ParentIDs {
		if pid == parentRoleID {
			parentFound = true
			break
		}
	}

	if !parentFound {
		return nil // Parent not found, nothing to remove
	}

	// Remove the parent
	newParents := make([]string, 0, len(role.ParentIDs)-1)
	for _, pid := range role.ParentIDs {
		if pid != parentRoleID {
			newParents = append(newParents, pid)
		}
	}

	role.ParentIDs = newParents
	s.roles[roleKey] = role
	return nil
}

// GetRoleParents retrieves all parent roles of a role.
func (s *MemoryStore) GetRoleParents(ctx context.Context, workspaceID, roleID string) ([]Role, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if workspaceID == "" || roleID == "" {
		return nil, ErrInvalidArgument
	}

	roleKey := s.keyFor(workspaceID, roleID)
	role, exists := s.roles[roleKey]
	if !exists {
		return nil, ErrRoleNotFound
	}

	parents := make([]Role, 0, len(role.ParentIDs))
	for _, parentID := range role.ParentIDs {
		parentKey := s.keyFor(workspaceID, parentID)
		parent, exists := s.roles[parentKey]
		if exists {
			parents = append(parents, parent)
		}
	}

	return parents, nil
}

// GetRoleChildren retrieves all roles that inherit from a role.
func (s *MemoryStore) GetRoleChildren(ctx context.Context, workspaceID, roleID string) ([]Role, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if workspaceID == "" || roleID == "" {
		return nil, ErrInvalidArgument
	}

	roleKey := s.keyFor(workspaceID, roleID)
	if _, exists := s.roles[roleKey]; !exists {
		return nil, ErrRoleNotFound
	}

	children := make([]Role, 0)
	prefix := workspaceID + ":"
	
	for key, role := range s.roles {
		if len(key) > len(prefix) && key[:len(prefix)] == prefix && role.WorkspaceID == workspaceID {
			for _, parentID := range role.ParentIDs {
				if parentID == roleID {
					children = append(children, role)
					break
				}
			}
		}
	}

	return children, nil
}

// AddPermissionToRole adds a permission to a role.
func (s *MemoryStore) AddPermissionToRole(ctx context.Context, workspaceID, roleID, permissionID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	roleKey := s.keyFor(workspaceID, roleID)
	role, exists := s.roles[roleKey]
	if !exists {
		return ErrRoleNotFound
	}

	permissionKey := s.keyFor(workspaceID, permissionID)
	if _, exists := s.permissions[permissionKey]; !exists {
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
	s.roles[roleKey] = role
	return nil
}

// RemovePermissionFromRole removes a permission from a role.
func (s *MemoryStore) RemovePermissionFromRole(ctx context.Context, workspaceID, roleID, permissionID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	roleKey := s.keyFor(workspaceID, roleID)
	role, exists := s.roles[roleKey]
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
	s.roles[roleKey] = role
	return nil
}

// GetRolePermissions retrieves all permissions directly assigned to a role.
func (s *MemoryStore) GetRolePermissions(ctx context.Context, workspaceID, roleID string) ([]Permission, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	roleKey := s.keyFor(workspaceID, roleID)
	role, exists := s.roles[roleKey]
	if !exists {
		return nil, ErrRoleNotFound
	}

	permissions := make([]Permission, 0, len(role.DirectPermissionIDs))
	for _, pid := range role.DirectPermissionIDs {
		permissionKey := s.keyFor(workspaceID, pid)
		permission, exists := s.permissions[permissionKey]
		if exists {
			permissions = append(permissions, permission)
		}
	}

	return permissions, nil
}

// CreatePermission creates a new permission.
func (s *MemoryStore) CreatePermission(ctx context.Context, permission Permission) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if permission.WorkspaceID == "" {
		return ErrInvalidArgument
	}

	if permission.ID == "" {
		return ErrInvalidPermissionID
	}

	if permission.Name == "" {
		return ErrInvalidPermissionName
	}

	key := s.keyFor(permission.WorkspaceID, permission.ID)
	if _, exists := s.permissions[key]; exists {
		return ErrPermissionAlreadyExists
	}

	// Validate parent permissions exist
	for _, parentID := range permission.ParentIDs {
		parentKey := s.keyFor(permission.WorkspaceID, parentID)
		if _, exists := s.permissions[parentKey]; !exists {
			return ErrPermissionNotFound
		}
	}

	// Check for cyclic inheritance
	if err := s.checkCyclicPermissionInheritance(permission.WorkspaceID, permission.ID, permission.ParentIDs); err != nil {
		return err
	}

	// Store the permission
	s.permissions[key] = permission
	return nil
}

// GetPermission retrieves a permission by its ID.
func (s *MemoryStore) GetPermission(ctx context.Context, workspaceID, permissionID string) (Permission, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if workspaceID == "" || permissionID == "" {
		return Permission{}, ErrInvalidArgument
	}

	key := s.keyFor(workspaceID, permissionID)
	permission, exists := s.permissions[key]
	if !exists {
		return Permission{}, ErrPermissionNotFound
	}

	return permission, nil
}

// GetPermissions retrieves all permissions for a workspace.
func (s *MemoryStore) GetPermissions(ctx context.Context, workspaceID string) ([]Permission, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if workspaceID == "" {
		return nil, ErrInvalidArgument
	}

	permissions := make([]Permission, 0)
	prefix := workspaceID + ":"
	
	for key, permission := range s.permissions {
		if len(key) > len(prefix) && key[:len(prefix)] == prefix && permission.WorkspaceID == workspaceID {
			permissions = append(permissions, permission)
		}
	}

	return permissions, nil
}

// UpdatePermission updates an existing permission.
func (s *MemoryStore) UpdatePermission(ctx context.Context, permission Permission) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if permission.WorkspaceID == "" {
		return ErrInvalidArgument
	}

	if permission.ID == "" {
		return ErrInvalidPermissionID
	}

	if permission.Name == "" {
		return ErrInvalidPermissionName
	}

	key := s.keyFor(permission.WorkspaceID, permission.ID)
	if _, exists := s.permissions[key]; !exists {
		return ErrPermissionNotFound
	}

	// Validate parent permissions exist
	for _, parentID := range permission.ParentIDs {
		parentKey := s.keyFor(permission.WorkspaceID, parentID)
		if _, exists := s.permissions[parentKey]; !exists {
			return ErrPermissionNotFound
		}
	}

	// Check for cyclic inheritance
	if err := s.checkCyclicPermissionInheritance(permission.WorkspaceID, permission.ID, permission.ParentIDs); err != nil {
		return err
	}

	// Update the permission
	s.permissions[key] = permission
	return nil
}

// DeletePermission deletes a permission by its ID.
func (s *MemoryStore) DeletePermission(ctx context.Context, workspaceID, permissionID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if workspaceID == "" || permissionID == "" {
		return ErrInvalidArgument
	}

	key := s.keyFor(workspaceID, permissionID)
	if _, exists := s.permissions[key]; !exists {
		return ErrPermissionNotFound
	}

	// Check if any permission has this permission as parent
	prefix := workspaceID + ":"
	for k, perm := range s.permissions {
		if len(k) > len(prefix) && k[:len(prefix)] == prefix && perm.WorkspaceID == workspaceID {
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
					s.permissions[k] = perm
				}
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
				s.roles[s.keyFor(role.WorkspaceID, role.ID)] = role
			}
		}
	}

	delete(s.permissions, key)
	return nil
}

// AddPermissionParent adds a parent permission to a permission.
func (s *MemoryStore) AddPermissionParent(ctx context.Context, workspaceID, permissionID, parentPermissionID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if workspaceID == "" || permissionID == "" || parentPermissionID == "" {
		return ErrInvalidArgument
	}

	permissionKey := s.keyFor(workspaceID, permissionID)
	permission, exists := s.permissions[permissionKey]
	if !exists {
		return ErrPermissionNotFound
	}

	parentKey := s.keyFor(workspaceID, parentPermissionID)
	if _, exists := s.permissions[parentKey]; !exists {
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
	if err := s.checkCyclicPermissionInheritance(workspaceID, permissionID, tempParents); err != nil {
		return err
	}

	// Add the parent
	permission.ParentIDs = append(permission.ParentIDs, parentPermissionID)
	s.permissions[permissionKey] = permission
	return nil
}

// RemovePermissionParent removes a parent permission from a permission.
func (s *MemoryStore) RemovePermissionParent(ctx context.Context, workspaceID, permissionID, parentPermissionID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if workspaceID == "" || permissionID == "" || parentPermissionID == "" {
		return ErrInvalidArgument
	}

	permissionKey := s.keyFor(workspaceID, permissionID)
	permission, exists := s.permissions[permissionKey]
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
	s.permissions[permissionKey] = permission
	return nil
}

// GetPermissionParents retrieves all parent permissions of a permission.
func (s *MemoryStore) GetPermissionParents(ctx context.Context, workspaceID, permissionID string) ([]Permission, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if workspaceID == "" || permissionID == "" {
		return nil, ErrInvalidArgument
	}

	permissionKey := s.keyFor(workspaceID, permissionID)
	permission, exists := s.permissions[permissionKey]
	if !exists {
		return nil, ErrPermissionNotFound
	}

	parents := make([]Permission, 0, len(permission.ParentIDs))
	for _, parentID := range permission.ParentIDs {
		parentKey := s.keyFor(workspaceID, parentID)
		parent, exists := s.permissions[parentKey]
		if exists {
			parents = append(parents, parent)
		}
	}

	return parents, nil
}

// GetPermissionChildren retrieves all permissions that inherit from a permission.
func (s *MemoryStore) GetPermissionChildren(ctx context.Context, workspaceID, permissionID string) ([]Permission, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if workspaceID == "" || permissionID == "" {
		return nil, ErrInvalidArgument
	}

	permissionKey := s.keyFor(workspaceID, permissionID)
	if _, exists := s.permissions[permissionKey]; !exists {
		return nil, ErrPermissionNotFound
	}

	children := make([]Permission, 0)
	prefix := workspaceID + ":"
	
	for key, perm := range s.permissions {
		if len(key) > len(prefix) && key[:len(prefix)] == prefix && perm.WorkspaceID == workspaceID {
			for _, parentID := range perm.ParentIDs {
				if parentID == permissionID {
					children = append(children, perm)
					break
				}
			}
		}
	}

	return children, nil
}

// checkCyclicRoleInheritance checks if adding the given parents to the role would create a cyclic inheritance.
func (s *MemoryStore) checkCyclicRoleInheritance(workspaceID, roleID string, parentIDs []string) error {
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
		currentRoleKey := s.keyFor(workspaceID, currentID)
		currentRole, exists := s.roles[currentRoleKey]
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
func (s *MemoryStore) checkCyclicPermissionInheritance(workspaceID, permissionID string, parentIDs []string) error {
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
		currentPermissionKey := s.keyFor(workspaceID, currentID)
		currentPermission, exists := s.permissions[currentPermissionKey]
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

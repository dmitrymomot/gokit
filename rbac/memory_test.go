package rbac_test

import (
	"context"
	"testing"

	"github.com/dmitrymomot/gokit/rbac"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMemoryStore_Role(t *testing.T) {
	ctx := context.Background()
	store := rbac.NewMemoryStore()

	// Test role creation
	err := store.CreateRole(ctx, rbac.Role{
		WorkspaceID:         testWorkspaceID,
		ID:                  "role1",
		Name:                "Role 1",
		ParentIDs:           []string{},
		DirectPermissionIDs: []string{},
	})
	require.NoError(t, err)

	// Test getting a role
	role, err := store.GetRole(ctx, testWorkspaceID, "role1")
	require.NoError(t, err)
	assert.Equal(t, "role1", role.ID)
	assert.Equal(t, "Role 1", role.Name)

	// Test updating a role
	role.Name = "Updated Role 1"
	err = store.UpdateRole(ctx, role)
	require.NoError(t, err)

	updatedRole, err := store.GetRole(ctx, testWorkspaceID, "role1")
	require.NoError(t, err)
	assert.Equal(t, "Updated Role 1", updatedRole.Name)

	// Test getting all roles
	roles, err := store.GetRoles(ctx, testWorkspaceID)
	require.NoError(t, err)
	assert.Len(t, roles, 1)

	// Test invalid role ID
	err = store.CreateRole(ctx, rbac.Role{
		WorkspaceID: testWorkspaceID,
		ID:          "",
		Name:        "Invalid Role",
	})
	assert.ErrorIs(t, err, rbac.ErrInvalidRoleID)

	// Test invalid role name
	err = store.CreateRole(ctx, rbac.Role{
		WorkspaceID: testWorkspaceID,
		ID:          "invalid-role",
		Name:        "",
	})
	assert.ErrorIs(t, err, rbac.ErrInvalidRoleName)

	// Test duplicate role
	err = store.CreateRole(ctx, rbac.Role{
		WorkspaceID: testWorkspaceID,
		ID:          "role1",
		Name:        "Duplicate Role",
	})
	assert.ErrorIs(t, err, rbac.ErrRoleAlreadyExists)

	// Test deleting a role
	err = store.DeleteRole(ctx, testWorkspaceID, "role1")
	require.NoError(t, err)

	_, err = store.GetRole(ctx, testWorkspaceID, "role1")
	assert.ErrorIs(t, err, rbac.ErrRoleNotFound)
}

func TestMemoryStore_RoleInheritance(t *testing.T) {
	ctx := context.Background()
	store := rbac.NewMemoryStore()

	// Create parent role
	err := store.CreateRole(ctx, rbac.Role{
		WorkspaceID: testWorkspaceID,
		ID:          "parent",
		Name:        "Parent Role",
	})
	require.NoError(t, err)

	// Create child role with parent
	err = store.CreateRole(ctx, rbac.Role{
		WorkspaceID: testWorkspaceID,
		ID:          "child",
		Name:        "Child Role",
		ParentIDs:   []string{"parent"},
	})
	require.NoError(t, err)

	// Test getting parents
	parents, err := store.GetRoleParents(ctx, testWorkspaceID, "child")
	require.NoError(t, err)
	require.Len(t, parents, 1)
	assert.Equal(t, "parent", parents[0].ID)

	// Test getting children
	children, err := store.GetRoleChildren(ctx, testWorkspaceID, "parent")
	require.NoError(t, err)
	require.Len(t, children, 1)
	assert.Equal(t, "child", children[0].ID)

	// Test adding a parent
	err = store.CreateRole(ctx, rbac.Role{
		WorkspaceID: testWorkspaceID,
		ID:          "grandparent",
		Name:        "Grandparent Role",
	})
	require.NoError(t, err)

	err = store.AddRoleParent(ctx, testWorkspaceID, "child", "grandparent")
	require.NoError(t, err)

	parents, err = store.GetRoleParents(ctx, testWorkspaceID, "child")
	require.NoError(t, err)
	assert.Len(t, parents, 2)

	// Test removing a parent
	err = store.RemoveRoleParent(ctx, testWorkspaceID, "child", "parent")
	require.NoError(t, err)

	parents, err = store.GetRoleParents(ctx, testWorkspaceID, "child")
	require.NoError(t, err)
	assert.Len(t, parents, 1)
	assert.Equal(t, "grandparent", parents[0].ID)

	// Test cyclic inheritance detection
	err = store.AddRoleParent(ctx, testWorkspaceID, "grandparent", "child")
	assert.ErrorIs(t, err, rbac.ErrCyclicInheritance)
}

func TestMemoryStore_Permission(t *testing.T) {
	ctx := context.Background()
	store := rbac.NewMemoryStore()

	// Test permission creation
	err := store.CreatePermission(ctx, rbac.Permission{
		WorkspaceID: testWorkspaceID,
		ID:          "perm1",
		Name:        "Permission 1",
		ParentIDs:   []string{},
	})
	require.NoError(t, err)

	// Test getting a permission
	perm, err := store.GetPermission(ctx, testWorkspaceID, "perm1")
	require.NoError(t, err)
	assert.Equal(t, "perm1", perm.ID)
	assert.Equal(t, "Permission 1", perm.Name)

	// Test updating a permission
	perm.Name = "Updated Permission 1"
	err = store.UpdatePermission(ctx, perm)
	require.NoError(t, err)

	updatedPerm, err := store.GetPermission(ctx, testWorkspaceID, "perm1")
	require.NoError(t, err)
	assert.Equal(t, "Updated Permission 1", updatedPerm.Name)

	// Test getting all permissions
	perms, err := store.GetPermissions(ctx, testWorkspaceID)
	require.NoError(t, err)
	assert.Len(t, perms, 1)

	// Test invalid permission ID
	err = store.CreatePermission(ctx, rbac.Permission{
		WorkspaceID: testWorkspaceID,
		ID:          "",
		Name:        "Invalid Permission",
	})
	assert.ErrorIs(t, err, rbac.ErrInvalidPermissionID)

	// Test invalid permission name
	err = store.CreatePermission(ctx, rbac.Permission{
		WorkspaceID: testWorkspaceID,
		ID:          "invalid-perm",
		Name:        "",
	})
	assert.ErrorIs(t, err, rbac.ErrInvalidPermissionName)

	// Test duplicate permission
	err = store.CreatePermission(ctx, rbac.Permission{
		WorkspaceID: testWorkspaceID,
		ID:          "perm1",
		Name:        "Duplicate Permission",
	})
	assert.ErrorIs(t, err, rbac.ErrPermissionAlreadyExists)

	// Test deleting a permission
	err = store.DeletePermission(ctx, testWorkspaceID, "perm1")
	require.NoError(t, err)

	_, err = store.GetPermission(ctx, testWorkspaceID, "perm1")
	assert.ErrorIs(t, err, rbac.ErrPermissionNotFound)
}

func TestMemoryStore_PermissionInheritance(t *testing.T) {
	ctx := context.Background()
	store := rbac.NewMemoryStore()

	// Create parent permission
	err := store.CreatePermission(ctx, rbac.Permission{
		WorkspaceID: testWorkspaceID,
		ID:          "parent",
		Name:        "Parent Permission",
	})
	require.NoError(t, err)

	// Create child permission with parent
	err = store.CreatePermission(ctx, rbac.Permission{
		WorkspaceID: testWorkspaceID,
		ID:          "child",
		Name:        "Child Permission",
		ParentIDs:   []string{"parent"},
	})
	require.NoError(t, err)

	// Test getting parents
	parents, err := store.GetPermissionParents(ctx, testWorkspaceID, "child")
	require.NoError(t, err)
	require.Len(t, parents, 1)
	assert.Equal(t, "parent", parents[0].ID)

	// Test getting children
	children, err := store.GetPermissionChildren(ctx, testWorkspaceID, "parent")
	require.NoError(t, err)
	require.Len(t, children, 1)
	assert.Equal(t, "child", children[0].ID)

	// Test adding a parent
	err = store.CreatePermission(ctx, rbac.Permission{
		WorkspaceID: testWorkspaceID,
		ID:          "grandparent",
		Name:        "Grandparent Permission",
	})
	require.NoError(t, err)

	err = store.AddPermissionParent(ctx, testWorkspaceID, "child", "grandparent")
	require.NoError(t, err)

	parents, err = store.GetPermissionParents(ctx, testWorkspaceID, "child")
	require.NoError(t, err)
	assert.Len(t, parents, 2)

	// Test removing a parent
	err = store.RemovePermissionParent(ctx, testWorkspaceID, "child", "parent")
	require.NoError(t, err)

	parents, err = store.GetPermissionParents(ctx, testWorkspaceID, "child")
	require.NoError(t, err)
	assert.Len(t, parents, 1)
	assert.Equal(t, "grandparent", parents[0].ID)

	// Test cyclic inheritance detection
	err = store.AddPermissionParent(ctx, testWorkspaceID, "grandparent", "child")
	assert.ErrorIs(t, err, rbac.ErrCyclicInheritance)
}

func TestMemoryStore_RolePermissions(t *testing.T) {
	ctx := context.Background()
	store := rbac.NewMemoryStore()

	// Create a role and permission
	err := store.CreateRole(ctx, rbac.Role{
		WorkspaceID: testWorkspaceID,
		ID:          "role1",
		Name:        "Role 1",
	})
	require.NoError(t, err)

	err = store.CreatePermission(ctx, rbac.Permission{
		WorkspaceID: testWorkspaceID,
		ID:          "perm1",
		Name:        "Permission 1",
	})
	require.NoError(t, err)

	// Test adding a permission to a role
	err = store.AddPermissionToRole(ctx, testWorkspaceID, "role1", "perm1")
	require.NoError(t, err)

	// Test getting role permissions
	perms, err := store.GetRolePermissions(ctx, testWorkspaceID, "role1")
	require.NoError(t, err)
	require.Len(t, perms, 1)
	assert.Equal(t, "perm1", perms[0].ID)

	// Test removing a permission from a role
	err = store.RemovePermissionFromRole(ctx, testWorkspaceID, "role1", "perm1")
	require.NoError(t, err)

	perms, err = store.GetRolePermissions(ctx, testWorkspaceID, "role1")
	require.NoError(t, err)
	assert.Empty(t, perms)

	// Test error cases
	err = store.AddPermissionToRole(ctx, testWorkspaceID, "nonexistent", "perm1")
	assert.ErrorIs(t, err, rbac.ErrRoleNotFound)

	err = store.AddPermissionToRole(ctx, testWorkspaceID, "role1", "nonexistent")
	assert.ErrorIs(t, err, rbac.ErrPermissionNotFound)
}

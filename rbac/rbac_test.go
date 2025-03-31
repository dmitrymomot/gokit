package rbac_test

import (
	"context"
	"testing"
	"time"

	"github.com/dmitrymomot/gokit/rbac"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestEnvironment(t *testing.T) (context.Context, *rbac.Service) {
	ctx := context.Background()
	store := rbac.NewMemoryStore()

	// Create a simple hierarchical structure of roles and permissions
	// Permissions: read -> write -> admin
	require.NoError(t, store.CreatePermission(ctx, rbac.Permission{
		ID:   "read",
		Name: "Read Permission",
	}))

	require.NoError(t, store.CreatePermission(ctx, rbac.Permission{
		ID:        "write",
		Name:      "Write Permission",
		ParentIDs: []string{"read"},
	}))

	require.NoError(t, store.CreatePermission(ctx, rbac.Permission{
		ID:        "admin",
		Name:      "Admin Permission",
		ParentIDs: []string{"write"},
	}))

	// Create a separate permission branch
	require.NoError(t, store.CreatePermission(ctx, rbac.Permission{
		ID:   "view_reports",
		Name: "View Reports Permission",
	}))

	// Roles: guest -> member -> admin
	require.NoError(t, store.CreateRole(ctx, rbac.Role{
		ID:                  "guest",
		Name:                "Guest Role",
		DirectPermissionIDs: []string{"read"},
	}))

	require.NoError(t, store.CreateRole(ctx, rbac.Role{
		ID:                  "member",
		Name:                "Member Role",
		ParentIDs:           []string{"guest"},
		DirectPermissionIDs: []string{"write"},
	}))

	require.NoError(t, store.CreateRole(ctx, rbac.Role{
		ID:                  "admin",
		Name:                "Admin Role",
		ParentIDs:           []string{"member"},
		DirectPermissionIDs: []string{"admin", "view_reports"},
	}))

	// Create a separate role branch
	require.NoError(t, store.CreateRole(ctx, rbac.Role{
		ID:                  "reporter",
		Name:                "Reporter Role",
		DirectPermissionIDs: []string{"view_reports"},
	}))

	return ctx, rbac.NewService(store)
}

func TestRBACService_HasPermission(t *testing.T) {
	ctx, service := setupTestEnvironment(t)

	// Test guest role permissions
	hasRead, err := service.HasPermission(ctx, "guest", "read")
	require.NoError(t, err)
	assert.True(t, hasRead, "Guest should have read permission")

	hasWrite, err := service.HasPermission(ctx, "guest", "write")
	require.NoError(t, err)
	assert.False(t, hasWrite, "Guest should not have write permission")

	// Test member role permissions (inherits from guest)
	hasRead, err = service.HasPermission(ctx, "member", "read")
	require.NoError(t, err)
	assert.True(t, hasRead, "Member should have read permission (inherited from guest)")

	hasWrite, err = service.HasPermission(ctx, "member", "write")
	require.NoError(t, err)
	assert.True(t, hasWrite, "Member should have write permission (direct)")

	hasAdmin, err := service.HasPermission(ctx, "member", "admin")
	require.NoError(t, err)
	assert.False(t, hasAdmin, "Member should not have admin permission")

	// Test admin role permissions (inherits from member, which inherits from guest)
	hasRead, err = service.HasPermission(ctx, "admin", "read")
	require.NoError(t, err)
	assert.True(t, hasRead, "Admin should have read permission (inherited from guest through member)")

	hasWrite, err = service.HasPermission(ctx, "admin", "write")
	require.NoError(t, err)
	assert.True(t, hasWrite, "Admin should have write permission (inherited from member)")

	hasAdmin, err = service.HasPermission(ctx, "admin", "admin")
	require.NoError(t, err)
	assert.True(t, hasAdmin, "Admin should have admin permission (direct)")

	// Test separate branch
	hasViewReports, err := service.HasPermission(ctx, "reporter", "view_reports")
	require.NoError(t, err)
	assert.True(t, hasViewReports, "Reporter should have view_reports permission")

	hasRead, err = service.HasPermission(ctx, "reporter", "read")
	require.NoError(t, err)
	assert.False(t, hasRead, "Reporter should not have read permission")

	// Test error cases
	_, err = service.HasPermission(ctx, "", "read")
	assert.ErrorIs(t, err, rbac.ErrInvalidArgument)

	_, err = service.HasPermission(ctx, "guest", "")
	assert.ErrorIs(t, err, rbac.ErrInvalidArgument)
}

func TestRBACService_HasAnyPermission(t *testing.T) {
	ctx, service := setupTestEnvironment(t)

	// Test guest role
	hasAny, err := service.HasAnyPermission(ctx, "guest", "read", "write")
	require.NoError(t, err)
	assert.True(t, hasAny, "Guest should have at least one of the permissions")

	hasAny, err = service.HasAnyPermission(ctx, "guest", "write", "admin")
	require.NoError(t, err)
	assert.False(t, hasAny, "Guest should not have any of these permissions")

	// Test member role
	hasAny, err = service.HasAnyPermission(ctx, "member", "read", "write")
	require.NoError(t, err)
	assert.True(t, hasAny, "Member should have at least one of the permissions")

	hasAny, err = service.HasAnyPermission(ctx, "member", "admin", "view_reports")
	require.NoError(t, err)
	assert.False(t, hasAny, "Member should not have any of these permissions")

	// Test admin role
	hasAny, err = service.HasAnyPermission(ctx, "admin", "read", "admin")
	require.NoError(t, err)
	assert.True(t, hasAny, "Admin should have at least one of the permissions")

	hasAny, err = service.HasAnyPermission(ctx, "admin", "admin", "view_reports")
	require.NoError(t, err)
	assert.True(t, hasAny, "Admin should have at least one of the permissions")

	// Test error cases
	_, err = service.HasAnyPermission(ctx, "")
	assert.ErrorIs(t, err, rbac.ErrInvalidArgument)

	_, err = service.HasAnyPermission(ctx, "guest")
	assert.ErrorIs(t, err, rbac.ErrInvalidArgument)
}

func TestRBACService_HasAllPermissions(t *testing.T) {
	ctx, service := setupTestEnvironment(t)

	// Test guest role
	hasAll, err := service.HasAllPermissions(ctx, "guest", "read")
	require.NoError(t, err)
	assert.True(t, hasAll, "Guest should have all of the permissions")

	hasAll, err = service.HasAllPermissions(ctx, "guest", "read", "write")
	require.NoError(t, err)
	assert.False(t, hasAll, "Guest should not have all of these permissions")

	// Test member role
	hasAll, err = service.HasAllPermissions(ctx, "member", "read", "write")
	require.NoError(t, err)
	assert.True(t, hasAll, "Member should have all of the permissions")

	hasAll, err = service.HasAllPermissions(ctx, "member", "read", "write", "admin")
	require.NoError(t, err)
	assert.False(t, hasAll, "Member should not have all of these permissions")

	// Test admin role
	hasAll, err = service.HasAllPermissions(ctx, "admin", "read", "write", "admin")
	require.NoError(t, err)
	assert.True(t, hasAll, "Admin should have all of the permissions")

	hasAll, err = service.HasAllPermissions(ctx, "admin", "read", "write", "admin", "non_existent")
	require.NoError(t, err)
	assert.False(t, hasAll, "Admin should not have all of these permissions")

	// Test error cases
	_, err = service.HasAllPermissions(ctx, "")
	assert.ErrorIs(t, err, rbac.ErrInvalidArgument)

	_, err = service.HasAllPermissions(ctx, "guest")
	assert.ErrorIs(t, err, rbac.ErrInvalidArgument)
}

func TestRBACService_GetEffectivePermissions(t *testing.T) {
	ctx, service := setupTestEnvironment(t)

	// Test guest role
	perms, err := service.GetEffectivePermissions(ctx, "guest")
	require.NoError(t, err)
	assert.Len(t, perms, 1)
	assert.Contains(t, permsToIDs(perms), "read")

	// Test member role
	perms, err = service.GetEffectivePermissions(ctx, "member")
	require.NoError(t, err)
	assert.Len(t, perms, 2)
	permIDs := permsToIDs(perms)
	assert.Contains(t, permIDs, "read")
	assert.Contains(t, permIDs, "write")

	// Test admin role
	perms, err = service.GetEffectivePermissions(ctx, "admin")
	require.NoError(t, err)
	assert.Len(t, perms, 4)
	permIDs = permsToIDs(perms)
	assert.Contains(t, permIDs, "read")
	assert.Contains(t, permIDs, "write")
	assert.Contains(t, permIDs, "admin")
	assert.Contains(t, permIDs, "view_reports")

	// Test error cases
	_, err = service.GetEffectivePermissions(ctx, "")
	assert.ErrorIs(t, err, rbac.ErrInvalidArgument)
}

func TestRBACService_Caching(t *testing.T) {
	ctx := context.Background()
	store := rbac.NewMemoryStore()

	// Create a test permission and role
	require.NoError(t, store.CreatePermission(ctx, rbac.Permission{
		ID:   "test_perm",
		Name: "Test Permission",
	}))

	require.NoError(t, store.CreateRole(ctx, rbac.Role{
		ID:                  "test_role",
		Name:                "Test Role",
		DirectPermissionIDs: []string{"test_perm"},
	}))

	// Create service with caching enabled (short TTL for testing)
	service := rbac.NewService(store, rbac.WithCaching(100*time.Millisecond))

	// First check should compute and cache
	hasPermBefore, err := service.HasPermission(ctx, "test_role", "test_perm")
	require.NoError(t, err)
	assert.True(t, hasPermBefore)

	// Modify role directly in store (bypassing the service)
	role, err := store.GetRole(ctx, "test_role")
	require.NoError(t, err)
	role.DirectPermissionIDs = []string{}
	require.NoError(t, store.UpdateRole(ctx, role))

	// Check again - should still have permission due to cache
	hasPermAfter, err := service.HasPermission(ctx, "test_role", "test_perm")
	require.NoError(t, err)
	assert.True(t, hasPermAfter, "Permission check should use cached result")

	// Wait for cache to expire
	time.Sleep(150 * time.Millisecond)

	// Check again - should not have permission now
	hasPermExpired, err := service.HasPermission(ctx, "test_role", "test_perm")
	require.NoError(t, err)
	assert.False(t, hasPermExpired, "Permission check should use fresh result after cache expiry")

	// Test manual cache invalidation
	require.NoError(t, service.AddPermissionToRole(ctx, "test_role", "test_perm"))

	// This should update the cache
	hasPerm, err := service.HasPermission(ctx, "test_role", "test_perm")
	require.NoError(t, err)
	assert.True(t, hasPerm)

	// Modify role directly again
	role, err = store.GetRole(ctx, "test_role")
	require.NoError(t, err)
	role.DirectPermissionIDs = []string{}
	require.NoError(t, store.UpdateRole(ctx, role))

	// Invalidate cache for this role
	service.InvalidateCache("test_role")

	// Should get fresh result
	hasPerm, err = service.HasPermission(ctx, "test_role", "test_perm")
	require.NoError(t, err)
	assert.False(t, hasPerm, "Permission check should use fresh result after manual cache invalidation")
}

func TestRBACService_PermissionInheritance(t *testing.T) {
	ctx := context.Background()
	store := rbac.NewMemoryStore()
	service := rbac.NewService(store)

	// Create a chain of permissions: view -> edit -> delete
	require.NoError(t, store.CreatePermission(ctx, rbac.Permission{
		ID:   "view",
		Name: "View Permission",
	}))

	require.NoError(t, store.CreatePermission(ctx, rbac.Permission{
		ID:        "edit",
		Name:      "Edit Permission",
		ParentIDs: []string{"view"},
	}))

	require.NoError(t, store.CreatePermission(ctx, rbac.Permission{
		ID:        "delete",
		Name:      "Delete Permission",
		ParentIDs: []string{"edit"},
	}))

	// Create a role with only the highest permission
	require.NoError(t, store.CreateRole(ctx, rbac.Role{
		ID:                  "power_user",
		Name:                "Power User",
		DirectPermissionIDs: []string{"delete"},
	}))

	// The role should have all permissions in the chain
	hasView, err := service.HasPermission(ctx, "power_user", "view")
	require.NoError(t, err)
	assert.True(t, hasView, "Role should have view permission (inherited through delete->edit->view chain)")

	hasEdit, err := service.HasPermission(ctx, "power_user", "edit")
	require.NoError(t, err)
	assert.True(t, hasEdit, "Role should have edit permission (inherited through delete->edit chain)")

	hasDelete, err := service.HasPermission(ctx, "power_user", "delete")
	require.NoError(t, err)
	assert.True(t, hasDelete, "Role should have delete permission (direct)")

	// Check effective permissions
	perms, err := service.GetEffectivePermissions(ctx, "power_user")
	require.NoError(t, err)
	assert.Len(t, perms, 3, "Should have all three permissions in the chain")
}

// Helper function to convert []Permission to []string of permission IDs
func permsToIDs(perms []rbac.Permission) []string {
	result := make([]string, len(perms))
	for i, p := range perms {
		result[i] = p.ID
	}
	return result
}

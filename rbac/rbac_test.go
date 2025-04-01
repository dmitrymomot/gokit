package rbac_test

import (
	"context"
	"testing"
	"time"

	"github.com/dmitrymomot/gokit/rbac"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testWorkspaceID = "test-workspace"

func setupTestEnvironment(t *testing.T) (context.Context, *rbac.Service) {
	ctx := context.Background()
	store := rbac.NewMemoryStore()

	// Create a simple hierarchical structure of roles and permissions
	// Permissions: read -> write -> admin
	require.NoError(t, store.CreatePermission(ctx, rbac.Permission{
		WorkspaceID: testWorkspaceID,
		ID:          "read",
		Name:        "Read Permission",
	}))

	require.NoError(t, store.CreatePermission(ctx, rbac.Permission{
		WorkspaceID: testWorkspaceID,
		ID:          "write",
		Name:        "Write Permission",
		ParentIDs:   []string{"read"},
	}))

	require.NoError(t, store.CreatePermission(ctx, rbac.Permission{
		WorkspaceID: testWorkspaceID,
		ID:          "admin",
		Name:        "Admin Permission",
		ParentIDs:   []string{"write"},
	}))

	// Create a separate permission branch
	require.NoError(t, store.CreatePermission(ctx, rbac.Permission{
		WorkspaceID: testWorkspaceID,
		ID:          "view_reports",
		Name:        "View Reports Permission",
	}))

	// Roles: guest -> member -> admin
	require.NoError(t, store.CreateRole(ctx, rbac.Role{
		WorkspaceID:         testWorkspaceID,
		ID:                  "guest",
		Name:                "Guest Role",
		DirectPermissionIDs: []string{"read"},
	}))

	require.NoError(t, store.CreateRole(ctx, rbac.Role{
		WorkspaceID:         testWorkspaceID,
		ID:                  "member",
		Name:                "Member Role",
		ParentIDs:           []string{"guest"},
		DirectPermissionIDs: []string{"write"},
	}))

	require.NoError(t, store.CreateRole(ctx, rbac.Role{
		WorkspaceID:         testWorkspaceID,
		ID:                  "admin",
		Name:                "Admin Role",
		ParentIDs:           []string{"member"},
		DirectPermissionIDs: []string{"admin", "view_reports"},
	}))

	// Create a separate role branch
	require.NoError(t, store.CreateRole(ctx, rbac.Role{
		WorkspaceID:         testWorkspaceID,
		ID:                  "reporter",
		Name:                "Reporter Role",
		DirectPermissionIDs: []string{"view_reports"},
	}))

	return ctx, rbac.NewService(store)
}

func TestRBACService_HasPermission(t *testing.T) {
	ctx, service := setupTestEnvironment(t)

	// Test guest role permissions
	hasRead, err := service.HasPermission(ctx, testWorkspaceID, "guest", "read")
	require.NoError(t, err)
	assert.True(t, hasRead, "Guest should have read permission")

	hasWrite, err := service.HasPermission(ctx, testWorkspaceID, "guest", "write")
	require.NoError(t, err)
	assert.False(t, hasWrite, "Guest should not have write permission")

	// Test member role permissions (inherits from guest)
	hasRead, err = service.HasPermission(ctx, testWorkspaceID, "member", "read")
	require.NoError(t, err)
	assert.True(t, hasRead, "Member should have read permission (inherited from guest)")

	hasWrite, err = service.HasPermission(ctx, testWorkspaceID, "member", "write")
	require.NoError(t, err)
	assert.True(t, hasWrite, "Member should have write permission (direct)")

	hasAdmin, err := service.HasPermission(ctx, testWorkspaceID, "member", "admin")
	require.NoError(t, err)
	assert.False(t, hasAdmin, "Member should not have admin permission")

	// Test admin role permissions (inherits from member, which inherits from guest)
	hasRead, err = service.HasPermission(ctx, testWorkspaceID, "admin", "read")
	require.NoError(t, err)
	assert.True(t, hasRead, "Admin should have read permission (inherited from guest through member)")

	hasWrite, err = service.HasPermission(ctx, testWorkspaceID, "admin", "write")
	require.NoError(t, err)
	assert.True(t, hasWrite, "Admin should have write permission (inherited from member)")

	hasAdmin, err = service.HasPermission(ctx, testWorkspaceID, "admin", "admin")
	require.NoError(t, err)
	assert.True(t, hasAdmin, "Admin should have admin permission (direct)")

	hasViewReports, err := service.HasPermission(ctx, testWorkspaceID, "admin", "view_reports")
	require.NoError(t, err)
	assert.True(t, hasViewReports, "Admin should have view_reports permission (direct)")

	// Test reporter role permissions
	hasViewReports, err = service.HasPermission(ctx, testWorkspaceID, "reporter", "view_reports")
	require.NoError(t, err)
	assert.True(t, hasViewReports, "Reporter should have view_reports permission (direct)")

	hasRead, err = service.HasPermission(ctx, testWorkspaceID, "reporter", "read")
	require.NoError(t, err)
	assert.False(t, hasRead, "Reporter should not have read permission")

	// Test permission inheritance through permission parents
	hasRead, err = service.HasPermission(ctx, testWorkspaceID, "member", "read")
	require.NoError(t, err)
	assert.True(t, hasRead, "Member should have read permission (inherited from guest + write permission inherits from read)")

	// Test error conditions
	_, err = service.HasPermission(ctx, testWorkspaceID, "", "read")
	assert.Error(t, err, "Empty role ID should return an error")

	_, err = service.HasPermission(ctx, testWorkspaceID, "guest", "")
	assert.Error(t, err, "Empty permission ID should return an error")

	_, err = service.HasPermission(ctx, testWorkspaceID, "nonexistent", "read")
	assert.Error(t, err, "Nonexistent role should return an error")

	// Empty workspaceID should return an error
	_, err = service.HasPermission(ctx, "", "guest", "read")
	assert.Error(t, err, "Empty workspace ID should return an error")
}

func TestRBACService_HasAnyPermission(t *testing.T) {
	ctx, service := setupTestEnvironment(t)

	// Test guest role with a permission it has
	hasAny, err := service.HasAnyPermission(ctx, testWorkspaceID, "guest", "read")
	require.NoError(t, err)
	assert.True(t, hasAny, "Guest should have read permission")

	// Test guest role with multiple permissions, one of which it has
	hasAny, err = service.HasAnyPermission(ctx, testWorkspaceID, "guest", "write", "read")
	require.NoError(t, err)
	assert.True(t, hasAny, "Guest should have at least one of the permissions")

	// Test guest role with permissions it doesn't have
	hasAny, err = service.HasAnyPermission(ctx, testWorkspaceID, "guest", "write", "admin")
	require.NoError(t, err)
	assert.False(t, hasAny, "Guest should not have any of the permissions")

	// Test member role with multiple permissions
	hasAny, err = service.HasAnyPermission(ctx, testWorkspaceID, "member", "read", "admin")
	require.NoError(t, err)
	assert.True(t, hasAny, "Member should have at least one of the permissions")

	// Test admin role with all permissions
	hasAny, err = service.HasAnyPermission(ctx, testWorkspaceID, "admin", "read", "write", "admin", "view_reports")
	require.NoError(t, err)
	assert.True(t, hasAny, "Admin should have at least one of the permissions")

	// Test reporter role with multiple permissions
	hasAny, err = service.HasAnyPermission(ctx, testWorkspaceID, "reporter", "read", "write", "view_reports")
	require.NoError(t, err)
	assert.True(t, hasAny, "Reporter should have at least one of the permissions")

	// Test reporter role with permissions it doesn't have
	hasAny, err = service.HasAnyPermission(ctx, testWorkspaceID, "reporter", "read", "write", "admin")
	require.NoError(t, err)
	assert.False(t, hasAny, "Reporter should not have any of the permissions")

	// Test error conditions
	_, err = service.HasAnyPermission(ctx, testWorkspaceID, "", "read")
	assert.Error(t, err, "Empty role ID should return an error")

	_, err = service.HasAnyPermission(ctx, testWorkspaceID, "guest")
	assert.Error(t, err, "Empty permission list should return an error")

	// Empty workspaceID should return an error
	_, err = service.HasAnyPermission(ctx, "", "guest", "read")
	assert.Error(t, err, "Empty workspace ID should return an error")
}

func TestRBACService_HasAllPermissions(t *testing.T) {
	ctx, service := setupTestEnvironment(t)

	// Test guest role with a permission it has
	hasAll, err := service.HasAllPermissions(ctx, testWorkspaceID, "guest", "read")
	require.NoError(t, err)
	assert.True(t, hasAll, "Guest should have read permission")

	// Test guest role with multiple permissions, one of which it does not have
	hasAll, err = service.HasAllPermissions(ctx, testWorkspaceID, "guest", "read", "write")
	require.NoError(t, err)
	assert.False(t, hasAll, "Guest should not have all the permissions")

	// Test member role with permissions it has
	hasAll, err = service.HasAllPermissions(ctx, testWorkspaceID, "member", "read", "write")
	require.NoError(t, err)
	assert.True(t, hasAll, "Member should have all the permissions")

	// Test member role with a permission it doesn't have
	hasAll, err = service.HasAllPermissions(ctx, testWorkspaceID, "member", "read", "write", "admin")
	require.NoError(t, err)
	assert.False(t, hasAll, "Member should not have all the permissions")

	// Test admin role with all permissions
	hasAll, err = service.HasAllPermissions(ctx, testWorkspaceID, "admin", "read", "write", "admin", "view_reports")
	require.NoError(t, err)
	assert.True(t, hasAll, "Admin should have all the permissions")

	// Test admin role with a mix of permissions
	hasAll, err = service.HasAllPermissions(ctx, testWorkspaceID, "admin", "read", "admin")
	require.NoError(t, err)
	assert.True(t, hasAll, "Admin should have all the permissions")

	// Test reporter role with the permission it has
	hasAll, err = service.HasAllPermissions(ctx, testWorkspaceID, "reporter", "view_reports")
	require.NoError(t, err)
	assert.True(t, hasAll, "Reporter should have view_reports permission")

	// Test reporter role with permissions it doesn't have
	hasAll, err = service.HasAllPermissions(ctx, testWorkspaceID, "reporter", "read", "view_reports")
	require.NoError(t, err)
	assert.False(t, hasAll, "Reporter should not have all the permissions")

	// Test error conditions
	_, err = service.HasAllPermissions(ctx, testWorkspaceID, "", "read")
	assert.Error(t, err, "Empty role ID should return an error")

	_, err = service.HasAllPermissions(ctx, testWorkspaceID, "guest")
	assert.Error(t, err, "Empty permission list should return an error")

	// Empty workspaceID should return an error
	_, err = service.HasAllPermissions(ctx, "", "guest", "read")
	assert.Error(t, err, "Empty workspace ID should return an error")
}

func TestRBACService_GetEffectivePermissions(t *testing.T) {
	ctx, service := setupTestEnvironment(t)

	// Test guest role
	guestPermissions, err := service.GetEffectivePermissions(ctx, testWorkspaceID, "guest")
	require.NoError(t, err)
	guestIDs := permsToIDs(guestPermissions)
	assert.Contains(t, guestIDs, "read", "Guest should have read permission")
	assert.Len(t, guestIDs, 1, "Guest should have only one permission")

	// Test member role
	memberPermissions, err := service.GetEffectivePermissions(ctx, testWorkspaceID, "member")
	require.NoError(t, err)
	memberIDs := permsToIDs(memberPermissions)
	assert.Contains(t, memberIDs, "read", "Member should have read permission")
	assert.Contains(t, memberIDs, "write", "Member should have write permission")
	assert.Len(t, memberIDs, 2, "Member should have two permissions")

	// Test admin role
	adminPermissions, err := service.GetEffectivePermissions(ctx, testWorkspaceID, "admin")
	require.NoError(t, err)
	adminIDs := permsToIDs(adminPermissions)
	assert.Contains(t, adminIDs, "read", "Admin should have read permission")
	assert.Contains(t, adminIDs, "write", "Admin should have write permission")
	assert.Contains(t, adminIDs, "admin", "Admin should have admin permission")
	assert.Contains(t, adminIDs, "view_reports", "Admin should have view_reports permission")
	assert.Len(t, adminIDs, 4, "Admin should have four permissions")

	// Test reporter role
	reporterPermissions, err := service.GetEffectivePermissions(ctx, testWorkspaceID, "reporter")
	require.NoError(t, err)
	reporterIDs := permsToIDs(reporterPermissions)
	assert.Contains(t, reporterIDs, "view_reports", "Reporter should have view_reports permission")
	assert.Len(t, reporterIDs, 1, "Reporter should have only one permission")

	// Test error conditions
	_, err = service.GetEffectivePermissions(ctx, testWorkspaceID, "")
	assert.Error(t, err, "Empty role ID should return an error")

	_, err = service.GetEffectivePermissions(ctx, testWorkspaceID, "nonexistent")
	assert.Error(t, err, "Nonexistent role should return an error")

	// Empty workspaceID should return an error
	_, err = service.GetEffectivePermissions(ctx, "", "guest")
	assert.Error(t, err, "Empty workspace ID should return an error")
}

func TestRBACService_Caching(t *testing.T) {
	ctx, nonCachingService := setupTestEnvironment(t)
	store := nonCachingService.Store()

	// Create a caching service with a 1 second TTL
	cachingService := rbac.NewService(store, rbac.WithCaching(1*time.Second))

	// Perform a permission check that should populate the cache
	hasRead, err := cachingService.HasPermission(ctx, testWorkspaceID, "guest", "read")
	require.NoError(t, err)
	assert.True(t, hasRead, "Guest should have read permission")

	// Modify the underlying store directly (without using Service.UpdateRole to avoid cache invalidation)
	role, err := store.GetRole(ctx, testWorkspaceID, "guest")
	require.NoError(t, err)
	role.DirectPermissionIDs = []string{}
	err = store.UpdateRole(ctx, role)
	require.NoError(t, err)

	// Check again, should still return true because of caching
	hasRead, err = cachingService.HasPermission(ctx, testWorkspaceID, "guest", "read")
	require.NoError(t, err)
	assert.True(t, hasRead, "Guest should still have read permission due to caching")

	// Wait for cache to expire
	time.Sleep(1100 * time.Millisecond)

	// Check again, should now return false
	hasRead, err = cachingService.HasPermission(ctx, testWorkspaceID, "guest", "read")
	require.NoError(t, err)
	assert.False(t, hasRead, "Guest should no longer have read permission after cache expiry")

	// Restore the permission using the service's UpdateRole - this should update the cache
	role.DirectPermissionIDs = []string{"read"}
	err = cachingService.UpdateRole(ctx, role)
	require.NoError(t, err)

	// Check again, should return true now
	hasRead, err = cachingService.HasPermission(ctx, testWorkspaceID, "guest", "read")
	require.NoError(t, err)
	assert.True(t, hasRead, "Guest should have read permission again")

	// Manually invalidate the cache
	cachingService.InvalidateCache(testWorkspaceID, "guest")

	// Modify permissions through the service - should automatically invalidate cache
	role.DirectPermissionIDs = []string{}
	err = cachingService.UpdateRole(ctx, role)
	require.NoError(t, err)

	// Check again, should return false because the cache was invalidated
	hasRead, err = cachingService.HasPermission(ctx, testWorkspaceID, "guest", "read")
	require.NoError(t, err)
	assert.False(t, hasRead, "Guest should not have read permission after cache invalidation")

	// Test workspace-wide cache invalidation
	// First, restore permissions and populate cache for both guest and member
	role.DirectPermissionIDs = []string{"read"}
	err = cachingService.UpdateRole(ctx, role)
	require.NoError(t, err)

	hasRead, err = cachingService.HasPermission(ctx, testWorkspaceID, "guest", "read")
	require.NoError(t, err)
	assert.True(t, hasRead, "Guest should have read permission again")

	hasWrite, err := cachingService.HasPermission(ctx, testWorkspaceID, "member", "write")
	require.NoError(t, err)
	assert.True(t, hasWrite, "Member should have write permission")

	// Now invalidate the entire workspace cache
	cachingService.InvalidateWorkspaceCache(testWorkspaceID)

	// Modify both roles
	role.DirectPermissionIDs = []string{}
	err = cachingService.UpdateRole(ctx, role)
	require.NoError(t, err)

	memberRole, err := store.GetRole(ctx, testWorkspaceID, "member")
	require.NoError(t, err)
	memberRole.DirectPermissionIDs = []string{}
	err = cachingService.UpdateRole(ctx, memberRole)
	require.NoError(t, err)

	// Both should now return false
	hasRead, err = cachingService.HasPermission(ctx, testWorkspaceID, "guest", "read")
	require.NoError(t, err)
	assert.False(t, hasRead, "Guest should not have read permission after workspace cache invalidation")

	hasWrite, err = cachingService.HasPermission(ctx, testWorkspaceID, "member", "write")
	require.NoError(t, err)
	assert.False(t, hasWrite, "Member should not have write permission after workspace cache invalidation")
}

func TestRBACService_PermissionInheritance(t *testing.T) {
	ctx := context.Background()
	store := rbac.NewMemoryStore()
	service := rbac.NewService(store)

	// Create a chain of permissions with inheritance
	require.NoError(t, store.CreatePermission(ctx, rbac.Permission{
		WorkspaceID: testWorkspaceID,
		ID:          "level1",
		Name:        "Level 1 Permission",
	}))

	require.NoError(t, store.CreatePermission(ctx, rbac.Permission{
		WorkspaceID: testWorkspaceID,
		ID:          "level2",
		Name:        "Level 2 Permission",
		ParentIDs:   []string{"level1"},
	}))

	require.NoError(t, store.CreatePermission(ctx, rbac.Permission{
		WorkspaceID: testWorkspaceID,
		ID:          "level3",
		Name:        "Level 3 Permission",
		ParentIDs:   []string{"level2"},
	}))

	// Create a role with only the highest level permission
	require.NoError(t, store.CreateRole(ctx, rbac.Role{
		WorkspaceID:         testWorkspaceID,
		ID:                  "power_user",
		Name:                "Power User",
		DirectPermissionIDs: []string{"level3"},
	}))

	// Test that the role has all inherited permissions
	hasLevel3, err := service.HasPermission(ctx, testWorkspaceID, "power_user", "level3")
	require.NoError(t, err)
	assert.True(t, hasLevel3, "Power user should have level3 permission")

	hasLevel2, err := service.HasPermission(ctx, testWorkspaceID, "power_user", "level2")
	require.NoError(t, err)
	assert.True(t, hasLevel2, "Power user should have level2 permission through inheritance")

	hasLevel1, err := service.HasPermission(ctx, testWorkspaceID, "power_user", "level1")
	require.NoError(t, err)
	assert.True(t, hasLevel1, "Power user should have level1 permission through inheritance")

	// Get effective permissions and check
	perms, err := service.GetEffectivePermissions(ctx, testWorkspaceID, "power_user")
	require.NoError(t, err)
	permIDs := permsToIDs(perms)
	assert.Contains(t, permIDs, "level1", "Effective permissions should include level1")
	assert.Contains(t, permIDs, "level2", "Effective permissions should include level2")
	assert.Contains(t, permIDs, "level3", "Effective permissions should include level3")
	assert.Len(t, permIDs, 3, "There should be 3 effective permissions")
}

// Helper function to convert []Permission to []string of permission IDs
func permsToIDs(perms []rbac.Permission) []string {
	ids := make([]string, len(perms))
	for i, p := range perms {
		ids[i] = p.ID
	}
	return ids
}

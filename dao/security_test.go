package dao

import (
	"testing"
	"time"

	"github.com/BOCK-CHAIN/BockChain/crypto"
)

func TestSecurityManager_RoleManagement(t *testing.T) {
	sm := NewSecurityManager()

	// Create test users
	admin := crypto.GeneratePrivateKey().PublicKey()
	user := crypto.GeneratePrivateKey().PublicKey()

	// Grant admin role to admin user
	err := sm.GrantRole(admin, RoleSuperAdmin, admin, 0)
	if err == nil {
		t.Fatal("should not be able to grant role to self without existing permissions")
	}

	// Manually set admin role for testing
	sm.accessControl[admin.String()] = &AccessControlEntry{
		User:        admin,
		Role:        RoleSuperAdmin,
		Permissions: sm.rolePermissions[RoleSuperAdmin],
		GrantedBy:   admin,
		GrantedAt:   time.Now().Unix(),
		ExpiresAt:   0,
		Active:      true,
	}

	// Now admin should be able to grant roles
	err = sm.GrantRole(user, RoleMember, admin, 0)
	if err != nil {
		t.Fatal("admin should be able to grant roles")
	}

	// Check user has member role
	role, exists := sm.GetUserRole(user)
	if !exists {
		t.Fatal("user should have a role")
	}
	if role != RoleMember {
		t.Fatal("user should have member role")
	}

	// Check user has member permissions
	if !sm.HasPermission(user, PermissionViewProposals) {
		t.Fatal("member should have view permission")
	}
	if !sm.HasPermission(user, PermissionCreateProposal) {
		t.Fatal("member should have create permission")
	}
	if sm.HasPermission(user, PermissionManageRoles) {
		t.Fatal("member should not have admin permission")
	}

	// Revoke role
	err = sm.RevokeRole(user, admin)
	if err != nil {
		t.Fatal("admin should be able to revoke roles")
	}

	// Check role is revoked
	if sm.HasPermission(user, PermissionCreateProposal) {
		t.Fatal("revoked user should not have permissions")
	}
}

func TestSecurityManager_EmergencyMode(t *testing.T) {
	sm := NewSecurityManager()

	// Create test users
	admin := crypto.GeneratePrivateKey().PublicKey()
	user := crypto.GeneratePrivateKey().PublicKey()

	// Set up admin with emergency permissions
	sm.accessControl[admin.String()] = &AccessControlEntry{
		User:        admin,
		Role:        RoleSuperAdmin,
		Permissions: sm.rolePermissions[RoleSuperAdmin],
		GrantedBy:   admin,
		GrantedAt:   time.Now().Unix(),
		ExpiresAt:   0,
		Active:      true,
	}

	// Set up regular user
	sm.accessControl[user.String()] = &AccessControlEntry{
		User:        user,
		Role:        RoleMember,
		Permissions: sm.rolePermissions[RoleMember],
		GrantedBy:   admin,
		GrantedAt:   time.Now().Unix(),
		ExpiresAt:   0,
		Active:      true,
	}

	// Test normal operation
	if !sm.HasPermission(user, PermissionCreateProposal) {
		t.Fatal("user should have create permission normally")
	}
	if sm.IsEmergencyActive() {
		t.Fatal("emergency should not be active initially")
	}

	// Activate emergency mode
	err := sm.ActivateEmergency(admin, "Security breach detected", SecurityLevelCritical, []string{"CreateProposal", "Vote"})
	if err != nil {
		t.Fatal("admin should be able to activate emergency")
	}
	if !sm.IsEmergencyActive() {
		t.Fatal("emergency should be active")
	}

	// Test that regular users lose permissions during emergency
	if sm.HasPermission(user, PermissionCreateProposal) {
		t.Fatal("regular user should lose permissions during emergency")
	}
	if !sm.HasPermission(admin, PermissionEmergencyPause) {
		t.Fatal("admin should retain emergency permissions")
	}

	// Test function pausing
	if !sm.IsFunctionPaused("CreateProposal") {
		t.Fatal("CreateProposal should be paused")
	}
	if !sm.IsFunctionPaused("Vote") {
		t.Fatal("Vote should be paused")
	}

	// Deactivate emergency
	err = sm.DeactivateEmergency(admin)
	if err != nil {
		t.Fatal("admin should be able to deactivate emergency")
	}
	if sm.IsEmergencyActive() {
		t.Fatal("emergency should be deactivated")
	}

	// Test that permissions are restored
	if !sm.HasPermission(user, PermissionCreateProposal) {
		t.Fatal("user should regain permissions after emergency")
	}
	if sm.IsFunctionPaused("CreateProposal") {
		t.Fatal("CreateProposal should not be paused")
	}
}

func TestSecurityManager_AuditLogging(t *testing.T) {
	sm := NewSecurityManager()

	// Create test users
	admin := crypto.GeneratePrivateKey().PublicKey()
	user := crypto.GeneratePrivateKey().PublicKey()

	// Set up admin with audit permissions
	sm.accessControl[admin.String()] = &AccessControlEntry{
		User:        admin,
		Role:        RoleAdmin,
		Permissions: sm.rolePermissions[RoleAdmin],
		GrantedBy:   admin,
		GrantedAt:   time.Now().Unix(),
		ExpiresAt:   0,
		Active:      true,
	}

	// Log some events
	sm.LogAuditEvent(user, "CREATE_PROPOSAL", "proposal_123", "SUCCESS",
		map[string]interface{}{"title": "Test Proposal"}, SecurityLevelPublic)

	sm.LogAuditEvent(user, "VOTE", "proposal_123", "SUCCESS",
		map[string]interface{}{"choice": "yes"}, SecurityLevelMember)

	sm.LogAuditEvent(admin, "GRANT_ROLE", user.String(), "SUCCESS",
		map[string]interface{}{"role": "member"}, SecurityLevelSensitive)

	// Retrieve audit log
	entries, err := sm.GetAuditLog(admin, 10, 0, SecurityLevelPublic)
	if err != nil {
		t.Fatal("admin should be able to access audit log")
	}
	if len(entries) != 3 {
		t.Fatalf("should have 3 audit entries, got %d", len(entries))
	}

	// Test filtering by security level
	sensitiveEntries, err := sm.GetAuditLog(admin, 10, 0, SecurityLevelSensitive)
	if err != nil {
		t.Fatal("admin should be able to access sensitive audit log")
	}
	if len(sensitiveEntries) != 1 {
		t.Fatalf("should have 1 sensitive entry, got %d", len(sensitiveEntries))
	}

	// Test unauthorized access
	_, err = sm.GetAuditLog(user, 10, 0, SecurityLevelPublic)
	if err == nil {
		t.Fatal("regular user should not access audit log")
	}
}

func TestSecurityManager_RoleExpiration(t *testing.T) {
	sm := NewSecurityManager()

	// Create test users
	admin := crypto.GeneratePrivateKey().PublicKey()
	user := crypto.GeneratePrivateKey().PublicKey()

	// Set up admin
	sm.accessControl[admin.String()] = &AccessControlEntry{
		User:        admin,
		Role:        RoleAdmin,
		Permissions: sm.rolePermissions[RoleAdmin],
		GrantedBy:   admin,
		GrantedAt:   time.Now().Unix(),
		ExpiresAt:   0,
		Active:      true,
	}

	// Grant temporary role (1 second duration for testing)
	err := sm.GrantRole(user, RoleMember, admin, 1)
	if err != nil {
		t.Fatal("admin should be able to grant temporary role")
	}

	// Check user has permissions
	if !sm.HasPermission(user, PermissionCreateProposal) {
		t.Fatal("user should have permissions initially")
	}

	// Wait for expiration
	time.Sleep(2 * time.Second)

	// Check permissions are revoked
	if sm.HasPermission(user, PermissionCreateProposal) {
		t.Fatal("user should lose permissions after expiration")
	}

	// Check role is marked as expired
	_, exists := sm.GetUserRole(user)
	if exists {
		t.Fatal("expired role should not be active")
	}
}

func TestSecurityManager_UnauthorizedAccess(t *testing.T) {
	sm := NewSecurityManager()

	// Create test users
	unauthorizedUser := crypto.GeneratePrivateKey().PublicKey()
	member := crypto.GeneratePrivateKey().PublicKey()

	// Set up member role
	sm.accessControl[member.String()] = &AccessControlEntry{
		User:        member,
		Role:        RoleMember,
		Permissions: sm.rolePermissions[RoleMember],
		GrantedBy:   member,
		GrantedAt:   time.Now().Unix(),
		ExpiresAt:   0,
		Active:      true,
	}

	// Test unauthorized role granting
	err := sm.GrantRole(unauthorizedUser, RoleAdmin, member, 0)
	if err == nil {
		t.Fatal("member should not be able to grant admin role")
	}

	// Test unauthorized emergency activation
	err = sm.ActivateEmergency(member, "test", SecurityLevelCritical, []string{})
	if err == nil {
		t.Fatal("member should not be able to activate emergency")
	}

	// Test unauthorized audit access
	_, err = sm.GetAuditLog(member, 10, 0, SecurityLevelPublic)
	if err == nil {
		t.Fatal("member should not have audit access")
	}

	// Test unauthorized config access
	_, err = sm.GetSecurityConfig(member)
	if err == nil {
		t.Fatal("member should not access security config")
	}
}

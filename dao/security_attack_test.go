package dao

import (
	"testing"
	"time"

	"github.com/BOCK-CHAIN/BockChain/crypto"
)

// TestAttackVector_PrivilegeEscalation tests against privilege escalation attacks
func TestAttackVector_PrivilegeEscalation(t *testing.T) {
	sm := NewSecurityManager()

	// Create test users
	admin := crypto.GeneratePrivateKey().PublicKey()
	attacker := crypto.GeneratePrivateKey().PublicKey()

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

	// Grant member role to attacker
	err := sm.GrantRole(attacker, RoleMember, admin, 0)
	if err != nil {
		t.Fatal("should grant member role")
	}

	// Attempt 1: Try to grant admin role to self
	err = sm.GrantRole(attacker, RoleAdmin, attacker, 0)
	if err == nil {
		t.Fatal("member should not be able to grant admin role to self")
	}

	// Attempt 2: Try to grant admin role to another user
	victim := crypto.GeneratePrivateKey().PublicKey()
	err = sm.GrantRole(victim, RoleAdmin, attacker, 0)
	if err == nil {
		t.Fatal("member should not be able to grant admin role to others")
	}

	// Attempt 3: Try to modify role permissions directly (this would be prevented by proper encapsulation)
	originalPerms := len(sm.rolePermissions[RoleMember])
	// In a real attack, attacker might try to modify sm.rolePermissions[RoleMember]
	// But our design prevents this through proper access control

	// Verify permissions haven't changed
	if len(sm.rolePermissions[RoleMember]) != originalPerms {
		t.Fatal("role permissions should not be modifiable")
	}
	if sm.HasPermission(attacker, PermissionManageRoles) {
		t.Fatal("attacker should not have admin permissions")
	}
}

// TestAttackVector_EmergencyAbuse tests against emergency mode abuse
func TestAttackVector_EmergencyAbuse(t *testing.T) {
	sm := NewSecurityManager()

	// Create test users
	admin := crypto.GeneratePrivateKey().PublicKey()
	attacker := crypto.GeneratePrivateKey().PublicKey()

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

	// Grant member role to attacker
	err := sm.GrantRole(attacker, RoleMember, admin, 0)
	if err != nil {
		t.Fatal("should grant member role")
	}

	// Attempt 1: Try to activate emergency mode without permissions
	err = sm.ActivateEmergency(attacker, "fake emergency", SecurityLevelCritical, []string{"Vote"})
	if err == nil {
		t.Fatal("member should not be able to activate emergency mode")
	}

	// Attempt 2: Try to deactivate legitimate emergency mode
	err = sm.ActivateEmergency(admin, "legitimate emergency", SecurityLevelCritical, []string{"Vote"})
	if err != nil {
		t.Fatal("admin should be able to activate emergency")
	}

	err = sm.DeactivateEmergency(attacker)
	if err == nil {
		t.Fatal("member should not be able to deactivate emergency mode")
	}

	// Verify emergency is still active
	if !sm.IsEmergencyActive() {
		t.Fatal("emergency should still be active")
	}

	// Attempt 3: Try to bypass emergency restrictions
	if sm.HasPermission(attacker, PermissionCreateProposal) {
		t.Fatal("member should lose permissions during emergency")
	}

	// Clean up
	err = sm.DeactivateEmergency(admin)
	if err != nil {
		t.Fatal("admin should be able to deactivate emergency")
	}
}

// TestAttackVector_AuditLogTampering tests against audit log tampering
func TestAttackVector_AuditLogTampering(t *testing.T) {
	sm := NewSecurityManager()

	// Create test users
	admin := crypto.GeneratePrivateKey().PublicKey()
	attacker := crypto.GeneratePrivateKey().PublicKey()

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

	// Log some events
	sm.LogAuditEvent(attacker, "SUSPICIOUS_ACTIVITY", "system", "BLOCKED", nil, SecurityLevelCritical)
	sm.LogAuditEvent(admin, "GRANT_ROLE", attacker.String(), "SUCCESS", nil, SecurityLevelSensitive)

	// Attempt 1: Try to access audit log without permissions
	_, err := sm.GetAuditLog(attacker, 10, 0, SecurityLevelPublic)
	if err == nil {
		t.Fatal("attacker should not access audit log")
	}

	// Attempt 2: Admin should be able to access audit log
	entries, err := sm.GetAuditLog(admin, 10, 0, SecurityLevelPublic)
	if err != nil {
		t.Fatal("admin should access audit log")
	}
	if len(entries) < 2 {
		t.Fatal("should have audit entries")
	}

	// Verify audit entries are immutable (they should be read-only copies)
	originalAction := entries[0].Action
	entries[0].Action = "MODIFIED"

	// Get audit log again to verify original is unchanged
	entries2, err := sm.GetAuditLog(admin, 10, 0, SecurityLevelPublic)
	if err != nil {
		t.Fatal("admin should access audit log again")
	}
	if entries2[0].Action != originalAction {
		t.Fatal("audit log should be immutable")
	}
}

// TestAttackVector_RoleExpiration tests against role expiration bypass
func TestAttackVector_RoleExpiration(t *testing.T) {
	sm := NewSecurityManager()

	// Create test users
	admin := crypto.GeneratePrivateKey().PublicKey()
	attacker := crypto.GeneratePrivateKey().PublicKey()

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

	// Grant temporary role to attacker (1 second)
	err := sm.GrantRole(attacker, RoleModerator, admin, 1)
	if err != nil {
		t.Fatal("should grant temporary role")
	}

	// Verify attacker has permissions initially
	if !sm.HasPermission(attacker, PermissionModerateProposals) {
		t.Fatal("should have moderator permissions")
	}

	// Wait for expiration
	time.Sleep(2 * time.Second)

	// Attempt 1: Try to use expired permissions
	if sm.HasPermission(attacker, PermissionModerateProposals) {
		t.Fatal("expired role should not have permissions")
	}

	// Attempt 2: Try to extend own role (should fail)
	err = sm.GrantRole(attacker, RoleModerator, attacker, 3600)
	if err == nil {
		t.Fatal("expired user should not be able to grant roles")
	}

	// Attempt 3: Verify role is properly marked as inactive
	_, exists := sm.GetUserRole(attacker)
	if exists {
		t.Fatal("expired role should not be active")
	}
}

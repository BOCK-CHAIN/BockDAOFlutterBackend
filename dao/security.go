package dao

import (
	"fmt"
	"sync"
	"time"

	"github.com/BOCK-CHAIN/BockChain/crypto"
	"github.com/BOCK-CHAIN/BockChain/types"
)

// Role represents different access levels in the DAO
type Role byte

const (
	RoleGuest      Role = 0x00 // Read-only access
	RoleMember     Role = 0x01 // Basic DAO member
	RoleModerator  Role = 0x02 // Can moderate proposals
	RoleAdmin      Role = 0x03 // Administrative privileges
	RoleSuperAdmin Role = 0x04 // Full system access
	RoleEmergency  Role = 0x05 // Emergency response role
)

// Permission represents specific actions that can be performed
type Permission byte

const (
	PermissionViewProposals     Permission = 0x01
	PermissionCreateProposal    Permission = 0x02
	PermissionVote              Permission = 0x03
	PermissionDelegate          Permission = 0x04
	PermissionModerateProposals Permission = 0x05
	PermissionManageTreasury    Permission = 0x06
	PermissionManageRoles       Permission = 0x07
	PermissionEmergencyPause    Permission = 0x08
	PermissionSystemUpgrade     Permission = 0x09
	PermissionAuditAccess       Permission = 0x0A
)

// SecurityLevel represents different security contexts
type SecurityLevel byte

const (
	SecurityLevelPublic    SecurityLevel = 0x01
	SecurityLevelMember    SecurityLevel = 0x02
	SecurityLevelSensitive SecurityLevel = 0x03
	SecurityLevelCritical  SecurityLevel = 0x04
)

// AccessControlEntry represents a single access control rule
type AccessControlEntry struct {
	User        crypto.PublicKey
	Role        Role
	Permissions []Permission
	GrantedBy   crypto.PublicKey
	GrantedAt   int64
	ExpiresAt   int64 // 0 means no expiration
	Active      bool
}

// EmergencyState represents the current emergency status
type EmergencyState struct {
	Active            bool
	ActivatedBy       crypto.PublicKey
	ActivatedAt       int64
	Reason            string
	Level             SecurityLevel
	AffectedFunctions []string
}

// AuditLogEntry represents a single audit log entry
type AuditLogEntry struct {
	ID            types.Hash
	Timestamp     int64
	User          crypto.PublicKey
	Action        string
	Resource      string
	Result        string
	Details       map[string]interface{}
	SecurityLevel SecurityLevel
	IPAddress     string
	UserAgent     string
}

// SecurityManager manages access control and security features
type SecurityManager struct {
	mu                sync.RWMutex
	accessControl     map[string]*AccessControlEntry
	emergencyState    *EmergencyState
	auditLog          []*AuditLogEntry
	rolePermissions   map[Role][]Permission
	securityConfig    *SecurityConfig
	emergencyContacts []crypto.PublicKey
	pausedFunctions   map[string]bool
}

// SecurityConfig holds security-related configuration
type SecurityConfig struct {
	MaxLoginAttempts       int
	LoginLockoutDuration   int64
	SessionTimeout         int64
	RequireMFA             bool
	AuditLogRetention      int64
	EmergencyTimeoutHours  int64
	MaxConcurrentSessions  int
	PasswordMinLength      int
	RequireStrongPasswords bool
	AllowedIPRanges        []string
}

// NewSecurityManager creates a new security manager
func NewSecurityManager() *SecurityManager {
	sm := &SecurityManager{
		accessControl:   make(map[string]*AccessControlEntry),
		emergencyState:  &EmergencyState{Active: false},
		auditLog:        make([]*AuditLogEntry, 0),
		rolePermissions: make(map[Role][]Permission),
		pausedFunctions: make(map[string]bool),
		securityConfig: &SecurityConfig{
			MaxLoginAttempts:       5,
			LoginLockoutDuration:   3600,  // 1 hour
			SessionTimeout:         86400, // 24 hours
			RequireMFA:             false,
			AuditLogRetention:      2592000, // 30 days
			EmergencyTimeoutHours:  72,      // 3 days
			MaxConcurrentSessions:  3,
			PasswordMinLength:      8,
			RequireStrongPasswords: true,
		},
	}

	// Initialize default role permissions
	sm.initializeDefaultRolePermissions()

	return sm
}

// initializeDefaultRolePermissions sets up default permissions for each role
func (sm *SecurityManager) initializeDefaultRolePermissions() {
	sm.rolePermissions[RoleGuest] = []Permission{
		PermissionViewProposals,
	}

	sm.rolePermissions[RoleMember] = []Permission{
		PermissionViewProposals,
		PermissionCreateProposal,
		PermissionVote,
		PermissionDelegate,
	}

	sm.rolePermissions[RoleModerator] = []Permission{
		PermissionViewProposals,
		PermissionCreateProposal,
		PermissionVote,
		PermissionDelegate,
		PermissionModerateProposals,
	}

	sm.rolePermissions[RoleAdmin] = []Permission{
		PermissionViewProposals,
		PermissionCreateProposal,
		PermissionVote,
		PermissionDelegate,
		PermissionModerateProposals,
		PermissionManageTreasury,
		PermissionManageRoles,
		PermissionAuditAccess,
	}

	sm.rolePermissions[RoleSuperAdmin] = []Permission{
		PermissionViewProposals,
		PermissionCreateProposal,
		PermissionVote,
		PermissionDelegate,
		PermissionModerateProposals,
		PermissionManageTreasury,
		PermissionManageRoles,
		PermissionEmergencyPause,
		PermissionSystemUpgrade,
		PermissionAuditAccess,
	}

	sm.rolePermissions[RoleEmergency] = []Permission{
		PermissionEmergencyPause,
		PermissionAuditAccess,
	}
}

// GrantRole grants a role to a user
func (sm *SecurityManager) GrantRole(user crypto.PublicKey, role Role, grantedBy crypto.PublicKey, duration int64) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// Check if granter has permission to manage roles
	if !sm.hasPermissionInternal(grantedBy, PermissionManageRoles) {
		sm.logAuditEvent(grantedBy, "GRANT_ROLE_DENIED", user.String(), "FAILURE",
			map[string]interface{}{"role": role, "reason": "insufficient_permissions"}, SecurityLevelSensitive)
		return NewDAOError(ErrUnauthorized, "insufficient permissions to grant roles", nil)
	}

	expiresAt := int64(0)
	if duration > 0 {
		expiresAt = time.Now().Unix() + duration
	}

	entry := &AccessControlEntry{
		User:        user,
		Role:        role,
		Permissions: sm.rolePermissions[role],
		GrantedBy:   grantedBy,
		GrantedAt:   time.Now().Unix(),
		ExpiresAt:   expiresAt,
		Active:      true,
	}

	sm.accessControl[user.String()] = entry

	sm.logAuditEvent(grantedBy, "GRANT_ROLE", user.String(), "SUCCESS",
		map[string]interface{}{"role": role, "expires_at": expiresAt}, SecurityLevelSensitive)

	return nil
}

// RevokeRole revokes a role from a user
func (sm *SecurityManager) RevokeRole(user crypto.PublicKey, revokedBy crypto.PublicKey) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// Check if revoker has permission to manage roles
	if !sm.hasPermissionInternal(revokedBy, PermissionManageRoles) {
		sm.logAuditEvent(revokedBy, "REVOKE_ROLE_DENIED", user.String(), "FAILURE",
			map[string]interface{}{"reason": "insufficient_permissions"}, SecurityLevelSensitive)
		return NewDAOError(ErrUnauthorized, "insufficient permissions to revoke roles", nil)
	}

	userStr := user.String()
	if entry, exists := sm.accessControl[userStr]; exists {
		entry.Active = false
		sm.logAuditEvent(revokedBy, "REVOKE_ROLE", user.String(), "SUCCESS",
			map[string]interface{}{"role": entry.Role}, SecurityLevelSensitive)
	}

	return nil
}

// HasPermission checks if a user has a specific permission
func (sm *SecurityManager) HasPermission(user crypto.PublicKey, permission Permission) bool {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	return sm.hasPermissionInternal(user, permission)
}

// hasPermissionInternal is the internal permission check (assumes lock is held)
func (sm *SecurityManager) hasPermissionInternal(user crypto.PublicKey, permission Permission) bool {
	// Check if system is in emergency mode and function is paused
	if sm.emergencyState.Active {
		// Only emergency role can act during emergency
		if entry, exists := sm.accessControl[user.String()]; exists {
			if entry.Active && entry.Role != RoleEmergency && entry.Role != RoleSuperAdmin {
				return false
			}
		} else {
			return false
		}
	}

	entry, exists := sm.accessControl[user.String()]
	if !exists || !entry.Active {
		return false
	}

	// Check if role has expired
	if entry.ExpiresAt > 0 && time.Now().Unix() > entry.ExpiresAt {
		entry.Active = false
		return false
	}

	// Check if user has the specific permission
	for _, perm := range entry.Permissions {
		if perm == permission {
			return true
		}
	}

	return false
}

// GetUserRole returns the role of a user
func (sm *SecurityManager) GetUserRole(user crypto.PublicKey) (Role, bool) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	entry, exists := sm.accessControl[user.String()]
	if !exists || !entry.Active {
		return RoleGuest, false
	}

	// Check if role has expired
	if entry.ExpiresAt > 0 && time.Now().Unix() > entry.ExpiresAt {
		entry.Active = false
		return RoleGuest, false
	}

	return entry.Role, true
}

// ActivateEmergency activates emergency mode
func (sm *SecurityManager) ActivateEmergency(activatedBy crypto.PublicKey, reason string, level SecurityLevel, affectedFunctions []string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// Check if user has emergency permissions
	if !sm.hasPermissionInternal(activatedBy, PermissionEmergencyPause) {
		sm.logAuditEvent(activatedBy, "EMERGENCY_ACTIVATION_DENIED", "system", "FAILURE",
			map[string]interface{}{"reason": "insufficient_permissions"}, SecurityLevelCritical)
		return NewDAOError(ErrUnauthorized, "insufficient permissions to activate emergency mode", nil)
	}

	sm.emergencyState = &EmergencyState{
		Active:            true,
		ActivatedBy:       activatedBy,
		ActivatedAt:       time.Now().Unix(),
		Reason:            reason,
		Level:             level,
		AffectedFunctions: affectedFunctions,
	}

	// Pause affected functions
	for _, function := range affectedFunctions {
		sm.pausedFunctions[function] = true
	}

	sm.logAuditEvent(activatedBy, "EMERGENCY_ACTIVATED", "system", "SUCCESS",
		map[string]interface{}{
			"reason":             reason,
			"level":              level,
			"affected_functions": affectedFunctions,
		}, SecurityLevelCritical)

	return nil
}

// DeactivateEmergency deactivates emergency mode
func (sm *SecurityManager) DeactivateEmergency(deactivatedBy crypto.PublicKey) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if !sm.emergencyState.Active {
		return NewDAOError(ErrInvalidProposal, "emergency mode is not active", nil)
	}

	// Check if user has emergency permissions
	if !sm.hasPermissionInternal(deactivatedBy, PermissionEmergencyPause) {
		sm.logAuditEvent(deactivatedBy, "EMERGENCY_DEACTIVATION_DENIED", "system", "FAILURE",
			map[string]interface{}{"reason": "insufficient_permissions"}, SecurityLevelCritical)
		return NewDAOError(ErrUnauthorized, "insufficient permissions to deactivate emergency mode", nil)
	}

	// Clear paused functions
	sm.pausedFunctions = make(map[string]bool)

	sm.emergencyState.Active = false

	sm.logAuditEvent(deactivatedBy, "EMERGENCY_DEACTIVATED", "system", "SUCCESS",
		map[string]interface{}{"duration": time.Now().Unix() - sm.emergencyState.ActivatedAt}, SecurityLevelCritical)

	return nil
}

// IsEmergencyActive returns whether emergency mode is active
func (sm *SecurityManager) IsEmergencyActive() bool {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	return sm.emergencyState.Active
}

// IsFunctionPaused checks if a specific function is paused
func (sm *SecurityManager) IsFunctionPaused(functionName string) bool {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	return sm.pausedFunctions[functionName]
}

// LogAuditEvent logs an audit event
func (sm *SecurityManager) LogAuditEvent(user crypto.PublicKey, action, resource, result string, details map[string]interface{}, level SecurityLevel) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	sm.logAuditEvent(user, action, resource, result, details, level)
}

// logAuditEvent is the internal audit logging function (assumes lock is held)
func (sm *SecurityManager) logAuditEvent(user crypto.PublicKey, action, resource, result string, details map[string]interface{}, level SecurityLevel) {
	entry := &AuditLogEntry{
		ID:            sm.generateAuditID(),
		Timestamp:     time.Now().Unix(),
		User:          user,
		Action:        action,
		Resource:      resource,
		Result:        result,
		Details:       details,
		SecurityLevel: level,
	}

	sm.auditLog = append(sm.auditLog, entry)

	// Clean up old audit logs if necessary
	sm.cleanupAuditLog()
}

// GetAuditLog returns audit log entries with optional filtering
func (sm *SecurityManager) GetAuditLog(user crypto.PublicKey, limit int, offset int, minLevel SecurityLevel) ([]*AuditLogEntry, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	// Check if user has audit access permission
	if !sm.hasPermissionInternal(user, PermissionAuditAccess) {
		return nil, NewDAOError(ErrUnauthorized, "insufficient permissions to access audit log", nil)
	}

	var filteredEntries []*AuditLogEntry
	for _, entry := range sm.auditLog {
		if entry.SecurityLevel >= minLevel {
			filteredEntries = append(filteredEntries, entry)
		}
	}

	// Apply pagination
	start := offset
	if start >= len(filteredEntries) {
		return []*AuditLogEntry{}, nil
	}

	end := start + limit
	if end > len(filteredEntries) {
		end = len(filteredEntries)
	}

	// Return copies to prevent external modification
	result := make([]*AuditLogEntry, end-start)
	for i, entry := range filteredEntries[start:end] {
		// Create a copy of the entry
		entryCopy := *entry
		// Create a copy of the details map
		if entry.Details != nil {
			entryCopy.Details = make(map[string]interface{})
			for k, v := range entry.Details {
				entryCopy.Details[k] = v
			}
		}
		result[i] = &entryCopy
	}

	return result, nil
}

// cleanupAuditLog removes old audit log entries based on retention policy
func (sm *SecurityManager) cleanupAuditLog() {
	if sm.securityConfig.AuditLogRetention <= 0 {
		return
	}

	cutoffTime := time.Now().Unix() - sm.securityConfig.AuditLogRetention
	var filteredLog []*AuditLogEntry

	for _, entry := range sm.auditLog {
		if entry.Timestamp >= cutoffTime {
			filteredLog = append(filteredLog, entry)
		}
	}

	sm.auditLog = filteredLog
}

// generateAuditID generates a unique ID for audit log entries
func (sm *SecurityManager) generateAuditID() types.Hash {
	// Simple hash generation - in production, use proper cryptographic hash
	data := fmt.Sprintf("%d_%d", time.Now().UnixNano(), len(sm.auditLog))
	hash := [32]byte{}
	copy(hash[:], []byte(data))
	return hash
}

// UpdateSecurityConfig updates the security configuration
func (sm *SecurityManager) UpdateSecurityConfig(updatedBy crypto.PublicKey, newConfig *SecurityConfig) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// Check if user has admin permissions
	if !sm.hasPermissionInternal(updatedBy, PermissionManageRoles) {
		sm.logAuditEvent(updatedBy, "SECURITY_CONFIG_UPDATE_DENIED", "system", "FAILURE",
			map[string]interface{}{"reason": "insufficient_permissions"}, SecurityLevelSensitive)
		return NewDAOError(ErrUnauthorized, "insufficient permissions to update security configuration", nil)
	}

	oldConfig := *sm.securityConfig
	sm.securityConfig = newConfig

	sm.logAuditEvent(updatedBy, "SECURITY_CONFIG_UPDATED", "system", "SUCCESS",
		map[string]interface{}{"old_config": oldConfig, "new_config": newConfig}, SecurityLevelSensitive)

	return nil
}

// GetSecurityConfig returns the current security configuration
func (sm *SecurityManager) GetSecurityConfig(requestedBy crypto.PublicKey) (*SecurityConfig, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	// Check if user has audit access permission
	if !sm.hasPermissionInternal(requestedBy, PermissionAuditAccess) {
		return nil, NewDAOError(ErrUnauthorized, "insufficient permissions to view security configuration", nil)
	}

	// Return a copy to prevent external modification
	configCopy := *sm.securityConfig
	return &configCopy, nil
}

// ListActiveRoles returns all active role assignments
func (sm *SecurityManager) ListActiveRoles(requestedBy crypto.PublicKey) (map[string]*AccessControlEntry, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	// Check if user has audit access permission
	if !sm.hasPermissionInternal(requestedBy, PermissionAuditAccess) {
		return nil, NewDAOError(ErrUnauthorized, "insufficient permissions to list roles", nil)
	}

	activeRoles := make(map[string]*AccessControlEntry)
	now := time.Now().Unix()

	for userStr, entry := range sm.accessControl {
		if entry.Active && (entry.ExpiresAt == 0 || entry.ExpiresAt > now) {
			// Return a copy to prevent external modification
			entryCopy := *entry
			activeRoles[userStr] = &entryCopy
		}
	}

	return activeRoles, nil
}

// GetEmergencyState returns the current emergency state
func (sm *SecurityManager) GetEmergencyState(requestedBy crypto.PublicKey) (*EmergencyState, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	// Check if user has audit access permission
	if !sm.hasPermissionInternal(requestedBy, PermissionAuditAccess) {
		return nil, NewDAOError(ErrUnauthorized, "insufficient permissions to view emergency state", nil)
	}

	// Return a copy to prevent external modification
	stateCopy := *sm.emergencyState
	return &stateCopy, nil
}

// ValidateAccess validates access for a specific operation
func (sm *SecurityManager) ValidateAccess(user crypto.PublicKey, operation string, resource string, level SecurityLevel) error {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	// Check if function is paused
	if sm.pausedFunctions[operation] {
		sm.logAuditEvent(user, operation, resource, "BLOCKED",
			map[string]interface{}{"reason": "function_paused"}, level)
		return NewDAOError(ErrUnauthorized, fmt.Sprintf("operation %s is currently paused", operation), nil)
	}

	// Log the access attempt
	sm.logAuditEvent(user, operation, resource, "ALLOWED", nil, level)

	return nil
}

// AddEmergencyContact adds an emergency contact
func (sm *SecurityManager) AddEmergencyContact(contact crypto.PublicKey, addedBy crypto.PublicKey) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// Check if user has admin permissions
	if !sm.hasPermissionInternal(addedBy, PermissionManageRoles) {
		return NewDAOError(ErrUnauthorized, "insufficient permissions to add emergency contact", nil)
	}

	sm.emergencyContacts = append(sm.emergencyContacts, contact)

	sm.logAuditEvent(addedBy, "EMERGENCY_CONTACT_ADDED", contact.String(), "SUCCESS", nil, SecurityLevelSensitive)

	return nil
}

// GetEmergencyContacts returns the list of emergency contacts
func (sm *SecurityManager) GetEmergencyContacts(requestedBy crypto.PublicKey) ([]crypto.PublicKey, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	// Check if user has audit access permission
	if !sm.hasPermissionInternal(requestedBy, PermissionAuditAccess) {
		return nil, NewDAOError(ErrUnauthorized, "insufficient permissions to view emergency contacts", nil)
	}

	// Return a copy to prevent external modification
	contacts := make([]crypto.PublicKey, len(sm.emergencyContacts))
	copy(contacts, sm.emergencyContacts)

	return contacts, nil
}

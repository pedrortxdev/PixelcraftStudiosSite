package service

import (
	"context"
	"time"

	"github.com/pixelcraft/api/internal/models"
	"github.com/pixelcraft/api/internal/repository"
)

// RoleService handles role-related business logic
type RoleService struct {
	roleRepo *repository.RoleRepository
	userRepo *repository.UserRepository
}

// NewRoleService creates a new role service
func NewRoleService(roleRepo *repository.RoleRepository, userRepo *repository.UserRepository) *RoleService {
	return &RoleService{
		roleRepo: roleRepo,
		userRepo: userRepo,
	}
}

// GetUserRoles returns all active roles for a user
func (s *RoleService) GetUserRoles(ctx context.Context, userID string) ([]models.RoleType, error) {
	return s.roleRepo.GetUserRoles(ctx, userID)
}

// GetHighestRole returns the highest role for a user
func (s *RoleService) GetHighestRole(ctx context.Context, userID string) (*models.RoleType, error) {
	roles, err := s.roleRepo.GetUserRoles(ctx, userID)
	if err != nil {
		return nil, err
	}
	return models.GetHighestRole(roles), nil
}

// GetSupportPriority returns the support priority (stars) for a user and category
func (s *RoleService) GetSupportPriority(ctx context.Context, userID string, category models.TicketCategory) (float64, error) {
	roles, err := s.roleRepo.GetUserRoles(ctx, userID)
	if err != nil {
		return 1.0, err
	}
	return models.GetSupportPriority(roles, category), nil
}

// GrantRole adds a role to a user
func (s *RoleService) GrantRole(ctx context.Context, userID string, role models.RoleType, grantedBy *string, expiresAt *time.Time) error {
	return s.roleRepo.AddRole(ctx, userID, role, grantedBy, expiresAt)
}

// RevokeRole removes a role from a user
func (s *RoleService) RevokeRole(ctx context.Context, userID string, role models.RoleType) error {
	return s.roleRepo.RemoveRole(ctx, userID, role)
}

// HasRole checks if a user has a specific role
func (s *RoleService) HasRole(ctx context.Context, userID string, role models.RoleType) (bool, error) {
	return s.roleRepo.HasRole(ctx, userID, role)
}

// HasAnyRole checks if a user has any of the specified roles
func (s *RoleService) HasAnyRole(ctx context.Context, userID string, roles ...models.RoleType) (bool, error) {
	return s.roleRepo.HasAnyRole(ctx, userID, roles...)
}

// CanAccessAdmin checks if a user can access the admin panel
func (s *RoleService) CanAccessAdmin(ctx context.Context, userID string) (bool, error) {
	return s.roleRepo.HasAnyRole(ctx, userID, models.AdminAccessRoles()...)
}

// IsLegacyAdmin checks if a user has the legacy is_admin flag
func (s *RoleService) IsLegacyAdmin(ctx context.Context, userID string) (bool, error) {
	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return false, err
	}
	return user.IsAdmin, nil
}

// CanEditUser checks if source user can edit target user (based on role hierarchy)
func (s *RoleService) CanEditUser(ctx context.Context, sourceUserID, targetUserID string) (bool, error) {
	sourceRoles, err := s.roleRepo.GetUserRoles(ctx, sourceUserID)
	if err != nil {
		return false, err
	}
	
	targetRoles, err := s.roleRepo.GetUserRoles(ctx, targetUserID)
	if err != nil {
		return false, err
	}
	
	return models.CanEditRole(sourceRoles, targetRoles), nil
}

// ===========================================
// AUTOMATIC ROLE PROMOTION/DEMOTION LOGIC
// ===========================================

// OnDeposit handles role changes when a user makes a deposit
// Grants CLIENT role if user doesn't have it
func (s *RoleService) OnDeposit(ctx context.Context, userID string) error {
	hasClient, err := s.roleRepo.HasRole(ctx, userID, models.RoleClient)
	if err != nil {
		return err
	}
	
	// Grant CLIENT role if not already granted
	if !hasClient {
		return s.roleRepo.AddRole(ctx, userID, models.RoleClient, nil, nil)
	}
	
	return nil
}

// OnPurchase handles role changes and spending tracking when a user makes a purchase
// May promote to CLIENT_VIP if monthly spending exceeds R$200
func (s *RoleService) OnPurchase(ctx context.Context, userID string, amount float64) error {
	// Update spending tracking
	if err := s.roleRepo.UpdateUserSpending(ctx, userID, amount); err != nil {
		return err
	}
	
	// Check if user qualifies for CLIENT_VIP
	_, monthlySpent, err := s.roleRepo.GetUserSpending(ctx, userID)
	if err != nil {
		return err
	}
	
	// Promote to CLIENT_VIP if monthly spending >= R$200
	if monthlySpent >= 200.0 {
		hasVIP, err := s.roleRepo.HasRole(ctx, userID, models.RoleClientVIP)
		if err != nil {
			return err
		}
		
		if !hasVIP {
			// Grant permanent CLIENT_VIP (no expiration - earned through spending)
			return s.roleRepo.AddRole(ctx, userID, models.RoleClientVIP, nil, nil)
		}
	}
	
	return nil
}

// OnSubscriptionStart handles role changes when a user starts a subscription
// Grants temporary CLIENT_VIP (expires with subscription)
func (s *RoleService) OnSubscriptionStart(ctx context.Context, userID string, expiresAt time.Time) error {
	// Grant CLIENT_VIP with expiration matching subscription end
	return s.roleRepo.AddRole(ctx, userID, models.RoleClientVIP, nil, &expiresAt)
}

// OnSubscriptionEnd handles role changes when a subscription ends
// CLIENT_VIP is automatically removed by the expires_at check in the repository
// This method can be used for explicit cleanup if needed
func (s *RoleService) OnSubscriptionEnd(ctx context.Context, userID string) error {
	// Check if user has permanent VIP through spending
	_, monthlySpent, err := s.roleRepo.GetUserSpending(ctx, userID)
	if err != nil {
		return err
	}
	
	// If monthly spending < R$200, the expired temporary VIP should be cleaned
	// This is handled automatically by the expires_at check in the repository
	// But we can explicitly remove it here for clarity
	if monthlySpent < 200.0 {
		// The role should already be expired, but remove it explicitly
		return s.roleRepo.RemoveRole(ctx, userID, models.RoleClientVIP)
	}
	
	return nil
}

// OnFullRefund handles role changes when a user gets a full refund and balance becomes 0
// Removes CLIENT role
func (s *RoleService) OnFullRefund(ctx context.Context, userID string) error {
	// Check if user balance is 0
	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return err
	}
	
	// Only remove CLIENT role if balance is exactly 0
	if user.Balance == 0 {
		return s.roleRepo.RemoveRole(ctx, userID, models.RoleClient)
	}
	
	return nil
}

// CleanExpiredRoles cleans up all expired roles (should be called periodically)
func (s *RoleService) CleanExpiredRoles(ctx context.Context) (int64, error) {
	return s.roleRepo.CleanExpiredRoles(ctx)
}

// ===========================================
// ADMIN PANEL PERMISSION HELPERS
// ===========================================

// GetAdminPermissions returns what an admin can access based on their roles
func (s *RoleService) GetAdminPermissions(ctx context.Context, userID string) (*AdminPermissions, error) {
	roles, err := s.roleRepo.GetUserRoles(ctx, userID)
	if err != nil {
		return nil, err
	}
	
	perms := &AdminPermissions{}
	
	for _, role := range roles {
		switch role {
		case models.RoleDirection:
			// Full access
			perms.CanViewSupport = true
			perms.CanViewEmail = true
			perms.CanViewCatalog = true
			perms.CanViewOrders = true
			perms.CanViewUsers = true
			perms.CanViewFinance = true
			perms.CanEditAll = true
			perms.CanEditPasswords = true
			perms.CanEditBalance = true
		case models.RoleEngineering:
			perms.CanViewSupport = true
			perms.CanViewEmail = true
			perms.CanViewCatalog = true
			perms.CanViewOrders = true
			perms.CanViewUsers = true
			perms.CanEditCatalog = true
			perms.CanEditOrders = true
			perms.CanEditEmail = true
			perms.CanEditPasswords = true // Only for lower roles
			perms.CanEditBalance = true   // Only for lower roles
		case models.RoleDevelopment:
			perms.CanViewSupport = true
			perms.CanViewEmail = true
			perms.CanViewCatalog = true
			perms.CanViewOrders = true
			perms.CanViewUsers = true
			perms.CanEditCatalog = true // Plans and products
		case models.RoleAdmin:
			perms.CanViewSupport = true
			perms.CanViewEmail = true
			perms.CanViewCatalog = true
			perms.CanViewOrders = true
			perms.CanViewUsers = true
			perms.CanViewFinance = true
			// No edit permissions
		case models.RoleSupport:
			perms.CanViewSupport = true
			perms.CanViewEmail = true // Only assigned email
			perms.RestrictedEmailAccess = true
		}
	}
	
	return perms, nil
}

// AdminPermissions represents what an admin can do in the admin panel
type AdminPermissions struct {
	CanViewSupport        bool `json:"can_view_support"`
	CanViewEmail          bool `json:"can_view_email"`
	CanViewCatalog        bool `json:"can_view_catalog"`
	CanViewOrders         bool `json:"can_view_orders"`
	CanViewUsers          bool `json:"can_view_users"`
	CanViewFinance        bool `json:"can_view_finance"`
	CanEditAll            bool `json:"can_edit_all"`
	CanEditCatalog        bool `json:"can_edit_catalog"`
	CanEditOrders         bool `json:"can_edit_orders"`
	CanEditEmail          bool `json:"can_edit_email"`
	CanEditPasswords      bool `json:"can_edit_passwords"`
	CanEditBalance        bool `json:"can_edit_balance"`
	RestrictedEmailAccess bool `json:"restricted_email_access"` // Can only see assigned email
}

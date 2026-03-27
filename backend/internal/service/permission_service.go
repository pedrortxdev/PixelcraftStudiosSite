package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/pixelcraft/api/internal/apierrors"
	"github.com/pixelcraft/api/internal/models"
	"github.com/pixelcraft/api/internal/repository"
)

// PermissionService handles permission checks and role management
// REFACTORED: Email logging methods moved to EmailService (SRP)
type PermissionService struct {
	repo *repository.PermissionRepository
}

func NewPermissionService(repo *repository.PermissionRepository) *PermissionService {
	return &PermissionService{repo: repo}
}

// GetUserPermissions returns all permissions for a user (WITH CONTEXT + UUID)
func (s *PermissionService) GetUserPermissions(ctx context.Context, userID uuid.UUID) (*models.UserPermissions, error) {
	return s.repo.GetUserPermissions(ctx, userID.String())
}

// HasPermission checks if a user has a SPECIFIC permission (FAST - uses SELECT EXISTS)
// This is the CORRECT way to check permissions - no memory waste
func (s *PermissionService) HasPermission(ctx context.Context, userID uuid.UUID, resource models.ResourceType, action models.ActionType) (bool, error) {
	// Use repository method that runs SELECT EXISTS - returns single boolean
	// No need to load all permissions into memory
	return s.repo.HasPermission(ctx, userID.String(), string(resource), string(action))
}

// CheckPermission verifies if a user has a specific permission (WITH CONTEXT + UUID)
// DEPRECATED: Use HasPermission instead for better performance
func (s *PermissionService) CheckPermission(ctx context.Context, userID uuid.UUID, resource models.ResourceType, action models.ActionType) (bool, error) {
	// For backward compatibility, but HasPermission is preferred
	return s.HasPermission(ctx, userID, resource, action)
}

// GetRolePermissions returns all permissions for a role (WITH CONTEXT)
func (s *PermissionService) GetRolePermissions(ctx context.Context, role string) ([]models.RolePermission, error) {
	return s.repo.GetRolePermissions(ctx, role)
}

// GetAllRolePermissions returns all permissions for all roles (WITH CONTEXT)
func (s *PermissionService) GetAllRolePermissions(ctx context.Context) (map[string][]models.RolePermission, error) {
	return s.repo.GetAllRolePermissions(ctx)
}

// AddRolePermission adds a permission to a role (WITH CONTEXT + UUID)
func (s *PermissionService) AddRolePermission(ctx context.Context, role string, resource models.ResourceType, action models.ActionType) error {
	// Validate role - includes all system roles
	if !models.RoleType(role).IsValid() {
		return apierrors.ErrInvalidInput
	}

	return s.repo.AddRolePermission(ctx, role, resource, action)
}

// RemoveRolePermission removes a permission from a role (WITH CONTEXT + UUID)
func (s *PermissionService) RemoveRolePermission(ctx context.Context, role string, resource models.ResourceType, action models.ActionType) error {
	// Validate role - consistency with AddRolePermission
	if !models.RoleType(role).IsValid() {
		return apierrors.ErrInvalidInput
	}

	return s.repo.RemoveRolePermission(ctx, role, resource, action)
}

// AssignRoleToUser assigns a role to a user (WITH CONTEXT + UUID)
func (s *PermissionService) AssignRoleToUser(ctx context.Context, userID uuid.UUID, role string) error {
	if !models.RoleType(role).IsValid() {
		return apierrors.ErrInvalidInput
	}

	return s.repo.AssignRoleToUser(ctx, userID.String(), role)
}

// RemoveRoleFromUser removes a role from a user (WITH CONTEXT + UUID)
func (s *PermissionService) RemoveRoleFromUser(ctx context.Context, userID uuid.UUID, role string) error {
	if !models.RoleType(role).IsValid() {
		return apierrors.ErrInvalidInput
	}

	return s.repo.RemoveRoleFromUser(ctx, userID.String(), role)
}

// GetUserRoles returns all roles for a user (WITH CONTEXT + UUID)
func (s *PermissionService) GetUserRoles(ctx context.Context, userID uuid.UUID) ([]string, error) {
	return s.repo.GetUserRoles(ctx, userID.String())
}

// HasRole checks if user has a specific role (WITH CONTEXT + UUID)
func (s *PermissionService) HasRole(ctx context.Context, userID uuid.UUID, role string) (bool, error) {
	roles, err := s.repo.GetUserRoles(ctx, userID.String())
	if err != nil {
		return false, err
	}

	for _, r := range roles {
		if r == role {
			return true, nil
		}
	}
	return false, nil
}

// ValidateRoleHierarchy validates if a role can be assigned by current user's role
func (s *PermissionService) ValidateRoleHierarchy(currentRole, targetRole string) error {
	// Simplified hierarchy check - in production, use proper role hierarchy from models
	adminRoles := map[string]bool{
		"ADMIN":       true,
		"DEVELOPMENT": true,
		"ENGINEERING": true,
		"DIRECTION":   true,
	}

	if !adminRoles[currentRole] && adminRoles[targetRole] {
		return apierrors.ErrForbidden
	}

	return nil
}

// GetAvailableRoles returns roles available for assignment by current user's role
func (s *PermissionService) GetAvailableRoles(currentRole string) []string {
	// Simplified - in production, use proper role hierarchy
	if currentRole == "DIRECTION" {
		return []string{"PARTNER", "CLIENT", "CLIENT_VIP", "SUPPORT", "ADMIN", "DEVELOPMENT", "ENGINEERING"}
	}
	if currentRole == "ENGINEERING" {
		return []string{"PARTNER", "CLIENT", "CLIENT_VIP", "SUPPORT", "ADMIN", "DEVELOPMENT"}
	}
	return []string{}
}

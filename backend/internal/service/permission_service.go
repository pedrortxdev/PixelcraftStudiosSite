package service

import (
	"fmt"

	"github.com/pixelcraft/api/internal/models"
	"github.com/pixelcraft/api/internal/repository"
)

type PermissionService struct {
	repo *repository.PermissionRepository
}

func NewPermissionService(repo *repository.PermissionRepository) *PermissionService {
	return &PermissionService{repo: repo}
}

// GetUserPermissions retorna todas as permissões de um usuário
func (s *PermissionService) GetUserPermissions(userID string) (*models.UserPermissions, error) {
	return s.repo.GetUserPermissions(userID)
}

// CheckPermission verifica se um usuário tem uma permissão específica
func (s *PermissionService) CheckPermission(userID string, resource models.ResourceType, action models.ActionType) (bool, error) {
	perms, err := s.repo.GetUserPermissions(userID)
	if err != nil {
		return false, err
	}

	return perms.HasPermission(resource, action), nil
}

// GetRolePermissions retorna todas as permissões de um cargo
func (s *PermissionService) GetRolePermissions(role string) ([]models.RolePermission, error) {
	return s.repo.GetRolePermissions(role)
}

// GetAllRolePermissions retorna todas as permissões de todos os cargos
func (s *PermissionService) GetAllRolePermissions() (map[string][]models.RolePermission, error) {
	return s.repo.GetAllRolePermissions()
}

// AddRolePermission adiciona uma permissão a um cargo
func (s *PermissionService) AddRolePermission(role string, resource models.ResourceType, action models.ActionType) error {
	// Validar role — inclui todos os cargos do sistema
	if !models.RoleType(role).IsValid() {
		return fmt.Errorf("invalid role: %s", role)
	}

	return s.repo.AddRolePermission(role, resource, action)
}

// RemoveRolePermission remove uma permissão de um cargo
func (s *PermissionService) RemoveRolePermission(role string, resource models.ResourceType, action models.ActionType) error {
	// Validar role — consistência com AddRolePermission
	if !models.RoleType(role).IsValid() {
		return fmt.Errorf("invalid role: %s", role)
	}

	return s.repo.RemoveRolePermission(role, resource, action)
}

// LogEmail registra um email enviado
func (s *PermissionService) LogEmail(log *models.EmailLog) error {
	return s.repo.LogEmail(log)
}

// GetEmailLogs retorna o histórico de emails
func (s *PermissionService) GetEmailLogs(page, limit int, filters map[string]string) ([]models.EmailLog, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	return s.repo.GetEmailLogs(page, limit, filters)
}

// GetEmailLogByID retorna um email específico
func (s *PermissionService) GetEmailLogByID(id string) (*models.EmailLog, error) {
	return s.repo.GetEmailLogByID(id)
}

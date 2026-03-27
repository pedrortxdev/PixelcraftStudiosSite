package models

import "time"

// ResourceType representa os recursos do sistema
type ResourceType string

const (
	ResourceUsers        ResourceType = "USERS"
	ResourceRoles        ResourceType = "ROLES"
	ResourceProducts     ResourceType = "PRODUCTS"
	ResourceOrders       ResourceType = "ORDERS"
	ResourceTransactions ResourceType = "TRANSACTIONS"
	ResourceSupport      ResourceType = "SUPPORT"
	ResourceEmails       ResourceType = "EMAILS"
	ResourceFiles        ResourceType = "FILES"
	ResourceGames        ResourceType = "GAMES"
	ResourceCategories   ResourceType = "CATEGORIES"
	ResourcePlans        ResourceType = "PLANS"
	ResourceDashboard    ResourceType = "DASHBOARD"
	ResourceSettings     ResourceType = "SETTINGS"
	ResourceSystem       ResourceType = "SYSTEM"
	ResourceDiscounts    ResourceType = "DISCOUNTS"
)

// ActionType representa as ações permitidas
type ActionType string

const (
	ActionView    ActionType = "VIEW"
	ActionCreate  ActionType = "CREATE"
	ActionEdit    ActionType = "EDIT"
	ActionDelete  ActionType = "DELETE"
	ActionManage  ActionType = "MANAGE"  // Todas as ações
	ActionViewCPF ActionType = "VIEW_CPF" // Visualizar dados sensíveis (CPF)
)

// RolePermission representa uma permissão de cargo
type RolePermission struct {
	ID        string       `json:"id"`
	Role      string       `json:"role"`
	Resource  ResourceType `json:"resource"`
	Action    ActionType   `json:"action"`
	CreatedAt time.Time    `json:"created_at"`
}

// EmailLog representa um email enviado
type EmailLog struct {
	ID           string                 `json:"id"`
	FromEmail    string                 `json:"from_email"`
	ToEmail      string                 `json:"to_email"`
	Subject      string                 `json:"subject"`
	Body         string                 `json:"body"`
	Status       string                 `json:"status"` // sent, failed, bounced
	ErrorMessage *string                `json:"error_message,omitempty"`
	SentBy       *string                `json:"sent_by,omitempty"`
	SentAt       time.Time              `json:"sent_at"`
	MessageID    *string                `json:"message_id,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// PermissionCheck representa uma verificação de permissão
type PermissionCheck struct {
	Resource ResourceType `json:"resource"`
	Action   ActionType   `json:"action"`
}

// UserPermissions representa todas as permissões de um usuário
type UserPermissions struct {
	UserID      string                        `json:"user_id"`
	Roles       []string                      `json:"roles"`
	Permissions map[ResourceType][]ActionType `json:"permissions"`
}

// HasPermission verifica se o usuário tem uma permissão específica
func (up *UserPermissions) HasPermission(resource ResourceType, action ActionType) bool {
	actions, exists := up.Permissions[resource]
	if !exists {
		return false
	}

	// Se tem MANAGE, tem todas as ações
	for _, a := range actions {
		if a == ActionManage || a == action {
			return true
		}
	}

	return false
}

// CanView verifica se pode visualizar um recurso
func (up *UserPermissions) CanView(resource ResourceType) bool {
	return up.HasPermission(resource, ActionView)
}

// CanCreate verifica se pode criar um recurso
func (up *UserPermissions) CanCreate(resource ResourceType) bool {
	return up.HasPermission(resource, ActionCreate)
}

// CanEdit verifica se pode editar um recurso
func (up *UserPermissions) CanEdit(resource ResourceType) bool {
	return up.HasPermission(resource, ActionEdit)
}

// CanDelete verifica se pode deletar um recurso
func (up *UserPermissions) CanDelete(resource ResourceType) bool {
	return up.HasPermission(resource, ActionDelete)
}

// CanManage verifica se pode gerenciar um recurso (todas as ações)
func (up *UserPermissions) CanManage(resource ResourceType) bool {
	return up.HasPermission(resource, ActionManage)
}

// CanViewCPF verifica se pode ver dados sensíveis (CPF) dos usuários
func (up *UserPermissions) CanViewCPF() bool {
	return up.HasPermission(ResourceUsers, ActionViewCPF) || up.CanManage(ResourceUsers)
}

package models

import "time"

// RoleType represents a user role/cargo in the system
type RoleType string

const (
	RolePartner     RoleType = "PARTNER"      // Parceiro: +1% de lucros em vendas
	RoleClient      RoleType = "CLIENT"       // Cliente: prioridade 2 estrelas, adquirido com depósito
	RoleClientVIP   RoleType = "CLIENT_VIP"   // Cliente VIP: prioridade 3 estrelas, R$200/mês ou assinatura
	RoleSupport     RoleType = "SUPPORT"      // Suporte: acesso restrito admin
	RoleAdmin       RoleType = "ADMIN"        // Administração: visualização total, sem edição
	RoleDevelopment RoleType = "DEVELOPMENT"  // Desenvolvimento: edita planos/produtos
	RoleEngineering RoleType = "ENGINEERING"  // Engenharia: acesso completo exceto Direção
	RoleDirection   RoleType = "DIRECTION"    // Direção: acesso total
)

// AllRoles lists all available roles
var AllRoles = []RoleType{
	RolePartner, RoleClient, RoleClientVIP, RoleSupport,
	RoleAdmin, RoleDevelopment, RoleEngineering, RoleDirection,
}

// IsValid checks if the role type is a valid role
func (r RoleType) IsValid() bool {
	for _, role := range AllRoles {
		if r == role {
			return true
		}
	}
	return false
}

// RoleHierarchy defines the power level of each role (higher = more power)
var RoleHierarchy = map[RoleType]int{
	RolePartner:     1,
	RoleClient:      2,
	RoleClientVIP:   3,
	RoleSupport:     4,
	RoleAdmin:       5,
	RoleDevelopment: 6,
	RoleEngineering: 7,
	RoleDirection:   8,
}

// RoleSupportPriority defines the support ticket priority (in stars) for each role
var RoleSupportPriority = map[RoleType]float64{
	RolePartner:     2.0,
	RoleClient:      3.0,
	RoleClientVIP:   4.0,
	RoleSupport:     3.0,
	RoleAdmin:       4.0,
	RoleDevelopment: 4.0,
	RoleEngineering: 5.0,
	RoleDirection:   5.0,
}

// RoleDisplayInfo contains display information for a role
type RoleDisplayInfo struct {
	Label string // Portuguese label
	Color string // Hex color for badge
}

// RoleDisplayConfig maps roles to their display configuration
var RoleDisplayConfig = map[RoleType]RoleDisplayInfo{
	RolePartner:     {Label: "Parceiro", Color: "#00bd65"},
	RoleClient:      {Label: "Cliente", Color: "#00d415"},
	RoleClientVIP:   {Label: "Cliente VIP", Color: "#00bd13"},
	RoleSupport:     {Label: "Suporte", Color: "#fbff00"},
	RoleAdmin:       {Label: "Administração", Color: "#bd005b"},
	RoleDevelopment: {Label: "Desenvolvimento", Color: "#0047bd"},
	RoleEngineering: {Label: "Engenharia", Color: "#6a00ff"},
	RoleDirection:   {Label: "Direção", Color: "#ff3f00"},
}

// UserRole represents a role assignment to a user
type UserRole struct {
	ID        string     `json:"id" db:"id"`
	UserID    string     `json:"user_id" db:"user_id"`
	Role      RoleType   `json:"role" db:"role"`
	GrantedAt time.Time  `json:"granted_at" db:"granted_at"`
	GrantedBy *string    `json:"granted_by,omitempty" db:"granted_by"`
	ExpiresAt *time.Time `json:"expires_at,omitempty" db:"expires_at"`
}

// GetHighestRole returns the highest role from a list of roles
func GetHighestRole(roles []RoleType) *RoleType {
	if len(roles) == 0 {
		return nil
	}
	
	highest := roles[0]
	highestLevel := RoleHierarchy[highest]
	
	for _, role := range roles[1:] {
		if level := RoleHierarchy[role]; level > highestLevel {
			highest = role
			highestLevel = level
		}
	}
	
	return &highest
}

// GetSupportPriority returns the highest support priority from roles and category
func GetSupportPriority(roles []RoleType, category TicketCategory) float64 {
	priority := 1.0 // Default: sem cargo
	
	hasVIP := false
	
	for _, role := range roles {
		if role == RoleClientVIP {
			hasVIP = true
		}
		if p, ok := RoleSupportPriority[role]; ok && p > priority {
			priority = p
		}
	}
	
	// Special Case: VIP + Finance/Payment = 5 Stars
	if hasVIP && (category == CategoryPayment || category == CategoryBilling) {
		priority = 5.0
	}
	
	return priority
}

// HasRole checks if a role list contains a specific role
func HasRole(roles []RoleType, target RoleType) bool {
	for _, r := range roles {
		if r == target {
			return true
		}
	}
	return false
}

// HasAnyRole checks if a role list contains any of the target roles
func HasAnyRole(roles []RoleType, targets ...RoleType) bool {
	for _, role := range roles {
		for _, target := range targets {
			if role == target {
				return true
			}
		}
	}
	return false
}

// CanEditRole checks if a user with sourceRoles can edit users with targetRoles
// Returns true if source has higher hierarchy than all target roles
func CanEditRole(sourceRoles, targetRoles []RoleType) bool {
	sourceHighest := GetHighestRole(sourceRoles)
	if sourceHighest == nil {
		return false
	}
	sourceLevel := RoleHierarchy[*sourceHighest]
	
	// Direction can edit anyone
	if *sourceHighest == RoleDirection {
		return true
	}
	
	for _, targetRole := range targetRoles {
		if RoleHierarchy[targetRole] >= sourceLevel {
			return false
		}
	}
	
	return true
}

// AdminAccessRoles returns roles that have any admin panel access
func AdminAccessRoles() []RoleType {
	return []RoleType{
		RoleSupport, RoleAdmin, RoleDevelopment, RoleEngineering, RoleDirection,
	}
}

// CanAccessAdminPanel checks if a user can access the admin panel
func CanAccessAdminPanel(roles []RoleType) bool {
	return HasAnyRole(roles, AdminAccessRoles()...)
}

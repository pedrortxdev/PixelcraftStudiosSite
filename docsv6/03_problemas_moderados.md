# Problemas Moderados — Sistema de Cargos e Permissões

**Data da investigação:** 2026-03-04

---

## MODERADO-01: `VIEW_CPF` Definido no Frontend Mas Inexistente no Backend

### Descrição
O frontend define uma ação especial `VIEW_CPF` nas constantes de permissões e no hook `usePermissions`, mas essa ação **não existe no backend** — nem no model Go, nem no enum SQL `action_type`.

### Evidência

**`src/constants/permissions.js`** — linha 31:
```javascript
export const ACTIONS = {
  VIEW: 'VIEW', CREATE: 'CREATE', EDIT: 'EDIT',
  DELETE: 'DELETE', MANAGE: 'MANAGE',
  VIEW_CPF: 'VIEW_CPF',  // ← Não existe no backend!
};
```

**`src/hooks/usePermissions.js`** — linha 87:
```javascript
if (resource === 'view_cpf') return permissions.some(
    p => p.resource === 'USERS' && p.action === 'VIEW_CPF'
);
```

**`backend/internal/models/permission.go`** — enum de actions:
```go
const (
    ActionView   ActionType = "VIEW"
    ActionCreate ActionType = "CREATE"
    ActionEdit   ActionType = "EDIT"
    ActionDelete ActionType = "DELETE"
    ActionManage ActionType = "MANAGE"
    // VIEW_CPF não existe!
)
```

**`migrations/006_create_permissions_system.sql`** — enum `action_type`:
```sql
CREATE TYPE action_type AS ENUM (
    'VIEW', 'CREATE', 'EDIT', 'DELETE', 'MANAGE'
    -- 'VIEW_CPF' ausente!
);
```

### Impacto
- `hasPermission('view_cpf')` no frontend **sempre retorna false**.
- Qualquer funcionalidade que dependa dessa verificação (ex: mostrar/ocultar CPF) está quebrada.
- Inserir `VIEW_CPF` no banco de dados geraria erro de tipo inválido.

### Correção Necessária
- **Opção A**: Adicionar `VIEW_CPF` ao enum SQL e ao model Go.
- **Opção B**: Mapear para `USERS:VIEW` com uma lógica separada e remover `VIEW_CPF` das constantes.

---

## MODERADO-02: `GetAvailableRoles` Retorna Apenas Cargos Admin (Omite Cargos Cliente)

### Descrição
O endpoint `GET /admin/permissions/available-roles` retorna apenas 5 cargos administrativos, ignorando os cargos de cliente (PARTNER, CLIENT, CLIENT_VIP).

### Evidência
**`backend/internal/handlers/permission_handler.go`** — linhas 138-144:
```go
func (h *PermissionHandler) GetAvailableRoles(c *gin.Context) {
    roles := []string{
        "SUPPORT", "ADMIN", "DEVELOPMENT", "ENGINEERING", "DIRECTION",
        // PARTNER, CLIENT, CLIENT_VIP — AUSENTES
    }
```

### Impacto
- O painel de gerenciamento de cargos (`/admin/roles`) exibe apenas os 5 cargos admin.
- Não é possível configurar permissões para cargos de cliente (mesmo que faça sentido para alguns recursos).
- Inconsistente com `models.AllRoles` que inclui todos os 8 cargos.

---

## MODERADO-03: `AddRolePermission` no Service Não Valida Cargos Cliente

### Descrição
O `PermissionService.AddRolePermission` valida explicitamente que o cargo está na lista dos 5 cargos admin. Não há forma de adicionar permissões a `PARTNER`, `CLIENT` ou `CLIENT_VIP`.

### Evidência
**`backend/internal/service/permission_service.go`** — linhas 46-56:
```go
validRoles := []string{"SUPPORT", "ADMIN", "DEVELOPMENT", "ENGINEERING", "DIRECTION"}
isValid := false
for _, r := range validRoles {
    if r == role {
        isValid = true
        break
    }
}
if !isValid {
    return fmt.Errorf("invalid role: %s", role)
}
```

### Impacto
- Inconsistência de design: cargos clientes existem, mas sem forma de gerenciar permissões para eles.
- Se no futuro for necessário dar permissões a `CLIENT_VIP`, requer mudança no código.

---

## MODERADO-04: `CatalogEditMiddleware` Não Verifica Permissões Granulares

### Descrição
O middleware `CatalogEditMiddleware` verifica apenas o cargo (DEVELOPMENT+), mas deveria idealmente verificar a permissão granular `PRODUCTS:MANAGE` (ou similar). Se as permissões de um cargo forem alteradas via `/admin/permissions/roles/:role`, o middleware **não reflete essa mudança**.

### Evidência
**`backend/internal/middleware/role.go`** — linhas 79-85:
```go
func CatalogEditMiddleware(db *sql.DB) gin.HandlerFunc {
    return RoleMiddleware(db,
        models.RoleDevelopment,  // Hardcoded — ignora role_permissions
        models.RoleEngineering,
        models.RoleDirection,
    )
}
```

### Impacto
- O sistema de permissões granulares (`role_permissions`) e os middlewares de cargo (`RoleMiddleware`) são **paralelos e independentes**.
- APIs de catalogar (planos, jogos, categorias) e arquivo no grupo `roleManagement` usam `CatalogEditMiddleware` (cargo), não `RequirePermission` (permissão granular).
- Mudanças via `/admin/permissions/roles/:role` não têm efeito nessas rotas.

---

## MODERADO-05: Rotas de Cargos de Usuários com Middleware Errado

### Descrição
As rotas de adicionar/remover cargos de usuários são protegidas pelo `CatalogEditMiddleware` (requer DEVELOPMENT+), mas a lógica de negócio já implementa verficação de hierarquia. O correto seria `FullAccessMiddleware` (apenas DIRECTION) para modificação de cargos.

### Evidência
**`backend/cmd/api/main.go`** — linhas 411-417:
```go
roleManagement := v1.Group("/admin")
roleManagement.Use(middleware.CatalogEditMiddleware(db.DB)) // DEVELOPMENT, ENGINEERING, DIRECTION
{
    roleManagement.POST("/users/:id/roles", roleHandler.AddUserRole)
    roleManagement.DELETE("/users/:id/roles/:role", roleHandler.RemoveUserRole)
}
```

A documentação `ROLE_MANAGEMENT_SETUP.md` diz:
```
DIRECTION peut éditer : tous os cargos
ENGINEERING peut éditer : DEVELOPMENT, ADMIN, SUPPORT
```

Mas com `CatalogEditMiddleware`, DEVELOPMENT (nível 6) também consegue **tentar** adicionar cargos — só sendo bloqueado pela verificação de hierarquia no handler.

### Impacto
- Violação do princípio de menor privilégio: DEVELOPMENT não deveria ter acesso a rotas de gerenciamento de cargos nem como tentativa inicial.
- Logs e rate limits contabilizam chamadas de DEVELOPMENT às rotas de cargo mesmo que sejam bloqueadas.

---

## MODERADO-06: `RemoveRolePermission` sem Validação do Cargo no Service

### Descrição
Ao contrário do `AddRolePermission`, o método `RemoveRolePermission` no service **não valida** se o cargo é válido antes de executar o DELETE no banco.

### Evidência
**`backend/internal/service/permission_service.go`** — linhas 62-64:
```go
func (s *PermissionService) RemoveRolePermission(role string, resource models.ResourceType, action models.ActionType) error {
    return s.repo.RemoveRolePermission(role, resource, action)  // Sem validação de role!
}
```

Comparado com `AddRolePermission` que valida:
```go
func (s *PermissionService) AddRolePermission(role string, ...) error {
    validRoles := []string{"SUPPORT", "ADMIN", ...}
    // valida role...
```

### Impacto
- Um request malformado com cargo inválido simplesmente não deleta nada (mas também não retorna erro útil).
- Inconsistência de comportamento entre add e remove.

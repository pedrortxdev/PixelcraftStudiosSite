# Plano de Correção — Sistema de Cargos e Permissões

**Data da investigação:** 2026-03-04

---

## Resumo de Todos os Problemas Encontrados

| ID           | Severidade  | Tipo              | Descrição curta                                                    |
|:-------------|:-----------:|:------------------|:-------------------------------------------------------------------|
| CRÍTICO-01   | 🔴 Crítico  | Segurança/Bug     | `admin_auth.go` usa campo `is_admin` legado — DEVELOPMENT/ENGINEERING não podem gerenciar Produtos |
| CRÍTICO-02   | 🔴 Crítico  | Banco de Dados    | SUPPORT tem MANAGE(006) vs VIEW+CREATE+EDIT(007) — comportamento não determinístico |
| CRÍTICO-03   | 🔴 Crítico  | Banco de Dados    | `SYSTEM` resource ausente no enum SQL de 006 — migração 009 pode falhar |
| CRÍTICO-04   | 🔴 Crítico  | Segurança         | Rotas de permissão acessíveis por qualquer staff — SUPPORT pode escalar privilégios |
| CRÍTICO-05   | 🔴 Crítico  | Banco de Dados    | Herança automática em 008 propaga inconsistências das migrações anteriores |
| MODERADO-01  | 🟡 Moderado | Frontend/Backend  | `VIEW_CPF` existe apenas no frontend — verificação sempre retorna false |
| MODERADO-02  | 🟡 Moderado | Backend           | `GetAvailableRoles` omite cargos cliente |
| MODERADO-03  | 🟡 Moderado | Backend           | `AddRolePermission` rejeita cargos cliente |
| MODERADO-04  | 🟡 Moderado | Arquitetura       | Middlewares de cargo ignoram permissões granulares |
| MODERADO-05  | 🟡 Moderado | Segurança         | `CatalogEditMiddleware` errado para rotas de cargo |
| MODERADO-06  | 🟡 Moderado | Backend           | `RemoveRolePermission` sem validação de cargo |
| INC-01       | 🔵 Info     | Documentação      | Hierarquia documentada com dois sistemas de numeração |
| INC-02       | 🟡 Moderado | Backend           | `pq.Array` ausente na query `ANY($1)` |
| INC-03       | 🟡 Moderado | Banco de Dados    | ADMIN com ou sem SETTINGS:VIEW dependendo da migração executada |
| INC-04       | 🟡 Moderado | Banco de Dados    | ENGINEERING com SETTINGS:VIEW(006) vs SETTINGS:MANAGE(007) |
| INC-05       | 🟡 Moderado | Banco de Dados    | DEVELOPMENT sem SETTINGS(006) vs com SETTINGS:VIEW(007) |
| INC-06       | 🔵 Info     | Backend           | `GetUserPermissions` não distingue permissões herdadas de diretas |
| INC-07       | 🔵 Info     | Backend           | `CanView` e similares fazem verificação dupla desnecessária |
| EXTRA-01     | 🔵 Info     | Documentação      | `RoleSupportPriority` PARTNER=2.0 mas comentário diz "2 estrelas"; CLIENT=3.0 mas comentário diz "2 estrelas" |

**Total: 5 críticos, 8 moderados, 5 informativos**

---

## Plano de Correção Priorizado

### Fase 1 — Segurança Imediata (Fazer Agora)

#### 1.1 Proteger Rotas de Permissões (CRÍTICO-04)
**Arquivo:** `backend/cmd/api/main.go`

Adicionar `RequirePermission(ROLES, MANAGE)` nas rotas de modificação:
```go
// Antes:
admin.POST("/permissions/roles/:role", permissionHandler.AddRolePermission)
admin.DELETE("/permissions/roles/:role", permissionHandler.RemoveRolePermission)

// Depois:
admin.POST("/permissions/roles/:role",
    middleware.RequirePermission(permissionService, models.ResourceRoles, models.ActionManage),
    permissionHandler.AddRolePermission)
admin.DELETE("/permissions/roles/:role",
    middleware.RequirePermission(permissionService, models.ResourceRoles, models.ActionManage),
    permissionHandler.RemoveRolePermission)
```

#### 1.2 Corrigir Middleware de Produtos (CRÍTICO-01)
**Arquivo:** `backend/cmd/api/main.go`

```go
// Antes:
products.POST("", middleware.AdminAuthMiddleware(db.DB), productHandler.CreateProduct)
products.PUT("/:id", middleware.AdminAuthMiddleware(db.DB), productHandler.UpdateProduct)
products.DELETE("/:id", middleware.AdminAuthMiddleware(db.DB), productHandler.DeleteProduct)

// Depois:
products.POST("", middleware.CatalogEditMiddleware(db.DB), productHandler.CreateProduct)
products.PUT("/:id", middleware.CatalogEditMiddleware(db.DB), productHandler.UpdateProduct)
products.DELETE("/:id", middleware.CatalogEditMiddleware(db.DB), productHandler.DeleteProduct)
```

#### 1.3 Corrigir Middleware de Rotas de Cargo (MODERADO-05)
**Arquivo:** `backend/cmd/api/main.go`

```go
// Antes:
roleManagement.Use(middleware.CatalogEditMiddleware(db.DB))

// Depois:
roleManagement.Use(middleware.AdminPanelMiddleware(db.DB)) // Hierarquia verificada no handler
```

---

### Fase 2 — Banco de Dados (Migration Fixes)

#### 2.1 Adicionar SYSTEM ao Enum (CRÍTICO-03)
Criar nova migração ou executar diretamente:
```sql
-- Adicionar SYSTEM ao enum resource_type
ALTER TYPE resource_type ADD VALUE IF NOT EXISTS 'SYSTEM';
```

#### 2.2 Consolidar Permissões do SUPPORT (CRÍTICO-02)
Definir estado final correto e garantir consistência:
```sql
-- Remover MANAGE do SUPPORT (que veio do 006) se já aplicado
DELETE FROM role_permissions 
WHERE role = 'SUPPORT' AND resource = 'SUPPORT' AND action = 'MANAGE';

-- Garantir que VIEW, CREATE, EDIT estão presentes
INSERT INTO role_permissions (role, resource, action) VALUES
('SUPPORT', 'SUPPORT', 'VIEW'),
('SUPPORT', 'SUPPORT', 'CREATE'),
('SUPPORT', 'SUPPORT', 'EDIT'),
('SUPPORT', 'EMAILS', 'VIEW'),
('SUPPORT', 'EMAILS', 'CREATE'),
('SUPPORT', 'DASHBOARD', 'VIEW')
ON CONFLICT (role, resource, action) DO NOTHING;
```

#### 2.3 Verificar Estado das Permissões Herdadas (CRÍTICO-05)
Após corrigir 2.2, remover e reaplicar herança para estados limpos:
```sql
-- Verificar permissões herdadas suspeitas
SELECT role, resource, action, inherited_from 
FROM role_permissions 
WHERE is_inherited = TRUE
ORDER BY role, resource;

-- Se necessário, limpar e reaplicar:
-- SELECT remove_inherited_permissions('ADMIN');
-- SELECT remove_inherited_permissions('DEVELOPMENT');
-- SELECT remove_inherited_permissions('ENGINEERING');
-- SELECT remove_inherited_permissions('DIRECTION');
-- Depois reaplicar com estado correto
```

---

### Fase 3 — Backend Code Fixes

#### 3.1 Adicionar VIEW_CPF ao Backend (MODERADO-01)
**Arquivo:** `backend/internal/models/permission.go`
```go
ActionViewCPF ActionType = "VIEW_CPF"
```

**Arquivo:** `migrations/` (nova migração):
```sql
ALTER TYPE action_type ADD VALUE IF NOT EXISTS 'VIEW_CPF';
```

#### 3.2 Corrigir pq.Array na Query (INC-02)
**Arquivo:** `backend/internal/repository/permission_repository.go`
```go
import "github.com/lib/pq"

// Mudar:
permRows, err := r.db.Query(permQuery, roles)
// Para:
permRows, err := r.db.Query(permQuery, pq.Array(roles))
```

#### 3.3 Validação Consistente no RemoveRolePermission (MODERADO-06)
**Arquivo:** `backend/internal/service/permission_service.go`
```go
func (s *PermissionService) RemoveRolePermission(role string, resource models.ResourceType, action models.ActionType) error {
    validRoles := []string{"SUPPORT", "ADMIN", "DEVELOPMENT", "ENGINEERING", "DIRECTION"}
    // Adicionar validação igual ao AddRolePermission
    isValid := false
    for _, r := range validRoles {
        if r == role { isValid = true; break }
    }
    if !isValid {
        return fmt.Errorf("invalid role: %s", role)
    }
    return s.repo.RemoveRolePermission(role, resource, action)
}
```

---

### Fase 4 — Limpeza e Documentação

#### 4.1 Remover `admin_auth.go` Completamente
O `AdminAuthMiddleware` em `admin_auth.go` está completamente obsoleto com o sistema de cargos. Após a Fase 1:
1. Verificar se há outras referências além de produtos
2. Remover o arquivo `admin_auth.go`
3. Garantir que todas as rotas usam os middlewares corretos

#### 4.2 Atualizar ROLE_MANAGEMENT_SETUP.md
Corrigir a inconsistência de numeração de hierarquia (usa 8 no início mas 5 na seção de herança).

#### 4.3 Corrigir Comentários de RoleSupportPriority (EXTRA-01)
**Arquivo:** `backend/internal/models/role.go`
```go
// Antes (comentários inconsistentes):
RolePartner: 2.0,  // "Parceiro: +1% de lucros em vendas"
RoleClient:  3.0,  // "Cliente: prioridade 2 estrelas"

// Verificar e corrigir conforme a regra de negócio real
```

---

## Script SQL de Diagnóstico

Execute este script para verificar o estado atual do banco de dados:

```sql
-- 1. Verificar permissões do SUPPORT
SELECT role, resource, action, is_inherited, inherited_from 
FROM role_permissions 
WHERE role = 'SUPPORT'
ORDER BY resource, action;

-- 2. Verificar se SYSTEM está no enum
SELECT enumlabel FROM pg_enum 
WHERE enumtypid = 'resource_type'::regtype
ORDER BY enumsortorder;

-- 3. Verificar todas as permissões herdadas
SELECT role, COUNT(*) as total, 
       SUM(CASE WHEN is_inherited THEN 1 ELSE 0 END) as herdadas
FROM role_permissions 
GROUP BY role
ORDER BY role;

-- 4. Verificar se action_type tem VIEW_CPF
SELECT enumlabel FROM pg_enum 
WHERE enumtypid = 'action_type'::regtype;

-- 5. Verificar permissões duplicadas
SELECT role, resource, action, COUNT(*) 
FROM role_permissions 
GROUP BY role, resource, action 
HAVING COUNT(*) > 1;
```

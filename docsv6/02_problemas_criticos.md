# Problemas Críticos — Sistema de Cargos e Permissões

**Data da investigação:** 2026-03-04

---

## CRÍTICO-01: Dois Sistemas de Autenticação Admin em Paralelo

### Descrição
Existe um middleware legado `admin_auth.go` que usa **apenas o campo `is_admin`** da tabela `users`, completamente desconectado do sistema de cargos baseado em `user_roles`.

### Evidência
**`backend/internal/middleware/admin_auth.go`** — usado nas rotas de Produtos:
```go
// Verifica apenas is_admin = true na tabela users (legado)
err := db.QueryRow("SELECT is_admin FROM users WHERE id = $1", userID).Scan(&isAdmin)
```

**`backend/cmd/api/main.go`** — linhas 237-239:
```go
products.POST("", middleware.AdminAuthMiddleware(db.DB), ...)
products.PUT("/:id", middleware.AdminAuthMiddleware(db.DB), ...)
products.DELETE("/:id", middleware.AdminAuthMiddleware(db.DB), ...)
```

### Impacto
- Um usuário com cargo `DEVELOPMENT` (que deveria ter permissão total sobre Produtos) **não pode criar/editar/deletar** produtos se não tiver `is_admin = true`.
- Um usuário com `is_admin = true` mas **sem nenhum cargo** consegue modificar produtos, bypassando o sistema de permissões granulares.
- O sistema de cargos e permissões é **ineficaz** para as rotas de Produtos.

### Correção Necessária
Substituir `AdminAuthMiddleware` por `CatalogEditMiddleware` (ou `RequirePermission`) nas rotas de Produtos:
```go
// Errado (atual):
products.POST("", middleware.AdminAuthMiddleware(db.DB), ...)

// Correto:
products.POST("", middleware.CatalogEditMiddleware(db.DB), ...)
```

---

## CRÍTICO-02: Permissões do Cargo SUPPORT Inconsistentes Entre Migrações

### Descrição
A migração `006` (em `migrations/`) e a `007` (em `database/`) definem permissões diferentes para o cargo SUPPORT no recurso SUPPORT.

### Evidência

**Migration 006** (`migrations/006_create_permissions_system.sql`, linha 78):
```sql
('SUPPORT', 'SUPPORT', 'MANAGE'),  -- MANAGE = todas as ações
```

**Migration 007** (`database/007_default_role_permissions.sql`, linhas 9-11):
```sql
('SUPPORT', 'SUPPORT', 'VIEW'),
('SUPPORT', 'SUPPORT', 'CREATE'),
('SUPPORT', 'SUPPORT', 'EDIT'),
-- Sem DELETE, sem MANAGE
```

### Impacto
- Se **apenas** a migração 006 foi executada, SUPPORT tem `MANAGE` (inclui DELETE de tickets).
- Se a migração 007 também foi executada, SUPPORT tem ambos — o que inclui MANAGE + as individuais (redundância).
- O comportamento esperado (sem DELETE) pode ou não estar ativo, dependendo da ordem de execução.
- A migração 006 também dá `DASHBOARD: VIEW` **apenas se a 007 for executada** (a 006 não inclui DASHBOARD para SUPPORT).

### Comparação Completa (SUPPORT):

| Recurso   | Ação   | Migração 006 | Migração 007 |
|-----------|--------|:------------:|:------------:|
| SUPPORT   | MANAGE | ✅           | ❌           |
| SUPPORT   | VIEW   | ❌ (incluso em MANAGE) | ✅ |
| SUPPORT   | CREATE | ❌ (incluso em MANAGE) | ✅ |
| SUPPORT   | EDIT   | ❌ (incluso em MANAGE) | ✅ |
| SUPPORT   | DELETE | ❌ (incluso em MANAGE) | ❌           |
| EMAILS    | VIEW   | ✅           | ✅           |
| EMAILS    | CREATE | ✅           | ✅           |
| DASHBOARD | VIEW   | ❌           | ✅           |

### Correção Necessária
Consolidar em uma única fonte de verdade. Remover `MANAGE` da 006 para SUPPORT:
```sql
-- 006 deveria ter:
('SUPPORT', 'SUPPORT', 'VIEW'),
('SUPPORT', 'SUPPORT', 'CREATE'),
('SUPPORT', 'SUPPORT', 'EDIT'),
('SUPPORT', 'EMAILS', 'VIEW'),
('SUPPORT', 'EMAILS', 'CREATE'),
('SUPPORT', 'DASHBOARD', 'VIEW')
```

---

## CRÍTICO-03: `SYSTEM` Resource Ausente no Enum SQL da Migração 006

### Descrição
O recurso `SYSTEM` existe no model Go (`permission.go`) e no frontend (`permissions.js`), mas **não foi incluído** no enum `resource_type` da migração `006_create_permissions_system.sql`.

### Evidência

**`migrations/006_create_permissions_system.sql`** — enum `resource_type`:
```sql
CREATE TYPE resource_type AS ENUM (
    'USERS', 'ROLES', 'PRODUCTS', 'ORDERS', 'TRANSACTIONS',
    'SUPPORT', 'EMAILS', 'FILES', 'GAMES', 'CATEGORIES',
    'PLANS', 'DASHBOARD', 'SETTINGS'
    -- 'SYSTEM' AUSENTE!
);
```

**`backend/internal/models/permission.go`** — linha 23:
```go
ResourceSystem ResourceType = "SYSTEM"
```

**`database/009_system_resources_permissions.sql`** — tenta inserir permissões para SYSTEM:
```sql
INSERT INTO role_permissions (role, resource, action) VALUES
('ADMIN', 'SYSTEM', 'VIEW')  -- FALHA se o enum não contém 'SYSTEM'
```

### Impacto
- A migração 009 **falha silenciosamente** ou gera erro ao tentar inserir `SYSTEM` em uma coluna do tipo `resource_type` que não contém esse valor.
- A rota `/api/v1/admin/system/metrics` retorna dados, mas qualquer middleware baseado em permissão SYSTEM nunca funcionará corretamente.

### Correção Necessária
Adicionar `SYSTEM` ao enum via `ALTER TYPE`:
```sql
ALTER TYPE resource_type ADD VALUE IF NOT EXISTS 'SYSTEM';
```

---

## CRÍTICO-04: Rotas de Permissões Sem Verificação Adequada de Autorização

### Descrição
As rotas de gerenciamento de permissões (`POST/DELETE /admin/permissions/roles/:role`) estão protegidas apenas pelo `AdminPanelMiddleware` (qualquer staff), não requerendo `DIRECTION` ou a permissão `ROLES:MANAGE`.

### Evidência
**`backend/cmd/api/main.go`** — linhas 380-386:
```go
admin.GET("/permissions/roles", permissionHandler.GetAllRolePermissions)
admin.GET("/permissions/roles/:role", permissionHandler.GetRolePermissions)
admin.POST("/permissions/roles/:role", permissionHandler.AddRolePermission)    // ⚠️ Requer apenas AdminPanel
admin.DELETE("/permissions/roles/:role", permissionHandler.RemoveRolePermission) // ⚠️ Requer apenas AdminPanel
```

O grupo `admin` usa `AdminPanelMiddleware` que aceita: SUPPORT, ADMIN, DEVELOPMENT, ENGINEERING, DIRECTION.

### Impacto
- Um usuário com cargo `SUPPORT` pode **adicionar ou remover permissões de qualquer cargo**, incluindo DIRECTION.
- Isso permite **escalonamento de privilégios**: SUPPORT poderia dar MANAGE em ROLES para si mesmo e tornar-se DIRECTION efetivamente.

### Correção Necessária
```go
// Proteger com ROLES:MANAGE ou ao menos FullAccessMiddleware:
admin.POST("/permissions/roles/:role",
    middleware.RequirePermission(permissionService, models.ResourceRoles, models.ActionManage),
    permissionHandler.AddRolePermission)
admin.DELETE("/permissions/roles/:role",
    middleware.RequirePermission(permissionService, models.ResourceRoles, models.ActionManage),
    permissionHandler.RemoveRolePermission)
```

---

## CRÍTICO-05: Herança de Permissões na Migração 008 Cria Permissões Duplicadas e Contraditórias

### Descrição
A migração `008_permission_enhancements.sql` executa herança automática de permissões usando a função `inherit_permissions_from_role`, mas faz isso sobre um estado potencialmente já errado (dos problemas CRÍTICO-02 e CRÍTICO-03), multiplicando as inconsistências.

### Evidência
**`database/008_permission_enhancements.sql`** — linhas 140-149:
```sql
-- Executa herança imediatamente ao rodar a migração:
SELECT inherit_permissions_from_role('DIRECTION', 'ENGINEERING');
SELECT inherit_permissions_from_role('ENGINEERING', 'DEVELOPMENT');
SELECT inherit_permissions_from_role('DEVELOPMENT', 'ADMIN');
SELECT inherit_permissions_from_role('ADMIN', 'SUPPORT');
```

A herança funciona copiando as permissões existentes com `is_inherited = TRUE`. Se em SUPPORT existe `SUPPORT:MANAGE` (do problema CRÍTICO-02), então:
- ADMIN herda `SUPPORT:MANAGE`
- DEVELOPMENT herda `SUPPORT:MANAGE` (via ADMIN)
- ENGINEERING herda `SUPPORT:MANAGE`
- DIRECTION herda `SUPPORT:MANAGE`

Todos os cargos admin passam a ter MANAGE em SUPPORT, mesmo que não seja a intenção.

### Impacto
- As permissões herdadas são difíceis de rastrear e remover.
- A função `remove_inherited_permissions` pode apagar permissões legítimas se executada acidentalmente.
- O estado real do banco de dados pode diferir significativamente do que está documentado.

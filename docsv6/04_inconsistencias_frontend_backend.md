# Inconsistências Frontend vs. Backend — Sistema de Cargos

**Data da investigação:** 2026-03-04

---

## INC-01: Hierarquia Documentada Inconsistente com o Código

### Descrição
O arquivo `ROLE_MANAGEMENT_SETUP.md` documenta a hierarquia com níveis diferentes dos implementados no código Go e JavaScript.

### Evidência

**ROLE_MANAGEMENT_SETUP.md** (documentação):
```
DIRECTION (8)      → Nível 8
ENGINEERING (7)    → Nível 7
DEVELOPMENT (6)    → Nível 6
ADMIN (5)          → Nível 5
SUPPORT (4)        → Nível 4
CLIENT_VIP (3)     → Nível 3
CLIENT (2)         → Nível 2
PARTNER (1)        → Nível 1
```

**ROLE_MANAGEMENT_SETUP.md** (seção "Hierarquia de Herança"):
```
DIRECTION (Nível 5)
ENGINEERING (Nível 4)
DEVELOPMENT (Nível 3)
ADMIN (Nível 2)
SUPPORT (Nível 1)
```

A documentação usa **dois sistemas de numeração diferentes** (8 níveis vs. 5 níveis) sem explicar a diferença. O código Go (`models/role.go`) usa 8 níveis. A seção de "Hierarquia de Herança" usa apenas os 5 cargos admin numerados de 1 a 5.

### Impacto
- Documentação confusa para desenvolvedores.
- Risco de implementar regras de negócio com base na numeração errada.

---

## INC-02: `GetUserPermissions` no Repository usa `roles` como `[]string` mas Deveria ser `pq.Array`

### Descrição
A query SQL em `GetUserPermissions` passa `roles` (um `[]string`) diretamente como parâmetro `ANY($1)`. Em Go com `database/sql` padrão (sem `lib/pq`), isso **pode falhar** dependendo do driver.

### Evidência
**`backend/internal/repository/permission_repository.go`** — linhas 56-60:
```go
permQuery := `
    SELECT DISTINCT resource, action 
    FROM role_permissions 
    WHERE role = ANY($1)  ← Requer pq.Array ou sintaxe alternativa
`
permRows, err := r.db.Query(permQuery, roles)  ← roles é []string, não pq.Array(roles)
```

### Impacto
- Se o driver PostgreSQL não suportar passar `[]string` diretamente, a query falha silenciosamente.
- O comportamento é dependente da versão do driver e pode quebrar em atualizações.

### Correção Necessária
```go
import "github.com/lib/pq"
permRows, err := r.db.Query(permQuery, pq.Array(roles))
```

---

## INC-03: Permissões do ADMIN Diferentes Entre Migração 006 e 007

### Descrição
O cargo ADMIN recebe `SETTINGS:VIEW` apenas na migração 007, não na 006. Isso significa que ADMIN só vê configurações se a 007 foi executada.

### Evidência

**Migração 006** — ADMIN não tem SETTINGS:
```sql
('ADMIN', 'USERS', 'VIEW'),
('ADMIN', 'PRODUCTS', 'VIEW'),
-- ... demais VIEW
('ADMIN', 'DASHBOARD', 'VIEW')
-- SETTINGS ausente
```

**Migração 007** — ADMIN tem SETTINGS:VIEW:
```sql
('ADMIN', 'SETTINGS', 'VIEW')  ← Adicionado apenas aqui
```

### Impacto
- Se apenas a migração 006 foi executada (sem a 007), ADMIN não pode ver Configurações.
- Inconsistência entre os dois arquivos sobre o que ADMIN deveria ter acesso.

---

## INC-04: ENGINEERING Tem `SETTINGS:VIEW` em 006 Mas `SETTINGS:MANAGE` em 007

### Descrição
O cargo ENGINEERING recebe permissão diferente no recurso SETTINGS dependendo de qual migração foi executada.

### Evidência

**Migração 006** (`migrations/`) — ENGINEERING:
```sql
('ENGINEERING', 'SETTINGS', 'VIEW')  ← VIEW apenas
```

**Migração 007** (`database/`) — ENGINEERING:
```sql
('ENGINEERING', 'SETTINGS', 'MANAGE')  ← MANAGE (todas as ações)
```

### Impacto
- ENGINEERING pode ou não editar configurações do sistema dependendo da ordem de execução das migrações.
- Comportamento não determinístico.

---

## INC-05: DEVELOPMENT Tem `SETTINGS:VIEW` em 007 Mas Nada em 006

### Descrição
DEVELOPMENT recebe `SETTINGS:VIEW` na migração 007, mas não estava previsto na migração 006.

### Evidência

**Migração 006** — DEVELOPMENT não tem SETTINGS:
```sql
('DEVELOPMENT', 'USERS', 'VIEW'),
('DEVELOPMENT', 'PRODUCTS', 'MANAGE'),
-- ... outros
('DEVELOPMENT', 'DASHBOARD', 'VIEW')
-- SETTINGS ausente
```

**Migração 007** — DEVELOPMENT tem SETTINGS:VIEW:
```sql
('DEVELOPMENT', 'SETTINGS', 'VIEW')  ← Adicionado apenas aqui
```

### Impacto
- DEVELOPMENT vê ou não as configurações dependendo de quais migrações foram rodadas.

---

## INC-06: `permQuery` no `GetUserPermissions` Não Usa Colunas de Herança

### Descrição
A query de busca de permissões de usuário não filtra por `is_inherited`, portanto inclui permissões herdadas e permissões diretas igualmente. Não há distinção no resultado.

### Evidência
**`backend/internal/repository/permission_repository.go`** — linhas 54-58:
```go
permQuery := `
    SELECT DISTINCT resource, action 
    FROM role_permissions 
    WHERE role = ANY($1)
    -- Não filtra is_inherited
`
```

### Impacto
- Permissões herdadas (marcadas com `is_inherited = TRUE`) são tratadas como permissões diretas no retorno da API.
- A distinção entre permissão herdada e direta (usada no frontend para indicação visual) é baseada apenas nos dados da tabela, mas a API `GetUserPermissions` retorna ambas sem distinção.

---

## INC-07: `CanView` Implementa Verificação Dupla Desnecessária

### Descrição
Os métodos helper `CanView`, `CanCreate`, etc. em `permission.go` verificam tanto a ação específica quanto `MANAGE`, mas `HasPermission` já faz essa verificação internamente.

### Evidência
**`backend/internal/models/permission.go`** — linhas 90-92:
```go
func (up *UserPermissions) CanView(resource ResourceType) bool {
    return up.HasPermission(resource, ActionView) || up.HasPermission(resource, ActionManage)
    //     ↑ HasPermission já verifica MANAGE internamente!
}
```

**`HasPermission`** — linhas 81-84:
```go
for _, a := range actions {
    if a == ActionManage || a == action {  // Já verifica MANAGE!
        return true
    }
}
```

### Impacto
- `CanView` chama `HasPermission` duas vezes desnecessariamente (performance mínima, mas incorreto conceitualmente).
- Código confuso para mantenedores.

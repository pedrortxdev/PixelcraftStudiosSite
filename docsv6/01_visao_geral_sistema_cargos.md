# Sistema de Cargos e Permissões — Visão Geral

**Data da investigação:** 2026-03-04

## Arquitetura do Sistema

O sistema de cargos (roles) e permissões do Pixelcraft Studio é composto por camadas distintas:

```
Frontend (React)                     Backend (Go)
──────────────────────────────────   ──────────────────────────────────
src/constants/roles.js               internal/models/role.go
src/constants/permissions.js         internal/models/permission.go
src/utils/roleUtils.js               internal/middleware/role.go
src/hooks/usePermissions.js          internal/middleware/permission.go
src/hooks/useRoleHierarchy.js        internal/middleware/admin_auth.go
src/hooks/useRoles.js                internal/handlers/role_handler.go
                                     internal/handlers/permission_handler.go
                                     internal/service/permission_service.go
                                     internal/repository/permission_repository.go
                                     cmd/api/main.go (rotas)
```

## Hierarquia de Cargos

| Cargo       | Nível | Tipo     | Descrição                                        |
|-------------|-------|----------|--------------------------------------------------|
| DIRECTION   | 8     | Admin    | Acesso total a todos os recursos                 |
| ENGINEERING | 7     | Admin    | Acesso completo exceto cargos                    |
| DEVELOPMENT | 6     | Admin    | Edita planos/produtos/jogos/categorias/arquivos  |
| ADMIN       | 5     | Admin    | Visualização total, sem edição                   |
| SUPPORT     | 4     | Admin    | Suporte + acesso ao próprio email                |
| CLIENT_VIP  | 3     | Cliente  | Cliente VIP com prioridade de suporte            |
| CLIENT      | 2     | Cliente  | Cliente padrão                                   |
| PARTNER     | 1     | Cliente  | Parceiro com 1% de lucros em vendas              |

## Tabelas do Banco de Dados

| Tabela                    | Propósito                                        |
|---------------------------|--------------------------------------------------|
| `user_roles`              | Associação usuário↔cargo (N:N com expiração)     |
| `role_permissions`        | Permissões por cargo (recurso + ação)             |
| `permission_audit_log`    | Histórico de mudanças de permissão               |
| `custom_roles`            | Cargos customizados criados pelo admin           |
| `permission_templates`    | Templates de permissões exportáveis              |
| `permission_notifications`| Notificações de mudanças de permissão           |
| `email_logs`              | Histórico de emails enviados via AWS SES         |

## Recursos Disponíveis (Resources)

| Recurso      | Categoria     |
|--------------|---------------|
| USERS        | Administração |
| ROLES        | Administração |
| SETTINGS     | Administração |
| PRODUCTS     | Conteúdo      |
| GAMES        | Conteúdo      |
| CATEGORIES   | Conteúdo      |
| PLANS        | Conteúdo      |
| FILES        | Conteúdo      |
| ORDERS       | Financeiro    |
| TRANSACTIONS | Financeiro    |
| SUPPORT      | Suporte       |
| EMAILS       | Suporte       |
| DASHBOARD    | Sistema       |
| SYSTEM       | Sistema       |

## Ações Disponíveis (Actions)

| Ação   | Descrição                                  |
|--------|--------------------------------------------|
| VIEW   | Visualizar                                 |
| CREATE | Criar                                      |
| EDIT   | Editar                                     |
| DELETE | Deletar                                    |
| MANAGE | Todas as ações (VIEW + CREATE + EDIT + DELETE) |

## Arquivos de Migração (ordem de execução)

| Arquivo | Local        | Propósito                                                |
|---------|--------------|----------------------------------------------------------|
| `004_create_roles_system.sql`        | migrations/  | Cria enum `role_type` e tabela `user_roles`     |
| `006_create_permissions_system.sql`  | migrations/  | Cria enums `resource_type`/`action_type`, tabela `role_permissions` + permissões padrão |
| `007_default_role_permissions.sql`   | database/    | Re-insere permissões padrão (sobrepõe 006)      |
| `008_permission_enhancements.sql`    | database/    | Herança, audit log, custom roles, templates, notificações |
| `009_system_resources_permissions.sql` | database/  | Adiciona permissões para recurso SYSTEM         |

> ⚠️ **Atenção**: Existem dois diretórios com SQLs — `migrations/` (executado pelo RunMigrations automático) e `database/` (deve ser executado manualmente). Veja o documento de problemas para detalhes.

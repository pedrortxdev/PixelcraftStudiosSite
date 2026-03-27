# Changelog — Atualizações do Sistema (docsv5 → docsv6)

**Data:** 2026-03-04

---

## Novas Funcionalidades Adicionadas

### 1. Sistema de Descontos com Restrições Granulares

**Referência:** `docsv5/01_sistema_descontos.md`

**O que foi adicionado:**
- Migração `017_discount_restrictions.sql` — adiciona `restriction_type` (ALL, ITEM_CATEGORY, GAME, PRODUCT) e `target_ids` (UUID[]) à tabela de descontos
- `internal/models/discount.go` — novos campos e tipos de restrição
- `internal/repository/discount_repository.go` — CRUD completo para descontos
- `internal/service/checkout_service.go` — validação de elegibilidade por cupom (desconto só sobre itens elegíveis)
- `internal/handlers/discount_handler.go` — handlers para CRUD admin
- Rotas registradas em `main.go`:
  - `GET /api/v1/admin/discounts`
  - `GET /api/v1/admin/discounts/:id`
  - `POST /api/v1/admin/discounts`
  - `PUT /api/v1/admin/discounts/:id`
  - `DELETE /api/v1/admin/discounts/:id`

**Status de Permissões:**
> ⚠️ As rotas de descontos usam apenas `AdminPanelMiddleware` (SUPPORT+). Recomendado adicionar `CatalogEditMiddleware` para operações de escrita (DEVELOPMENT+), pois é um recurso de catálogo.

---

### 2. Campo `adjustment_type` em Transações Admin

**O que foi adicionado no `admin_handler.go`:**
- Novo campo `AdjustmentType *string` em `AdminUpdateUserRequest`
- Permite distinguir ajustes entre `"Teste"` e `"Pix Direto"` ao atualizar usuário via admin
- O `admin_service.go` já trata e remove o campo antes de atualizar a tabela `users`

**Impacto em permissões:** sem mudança — continua restrito aos cargos com acesso a TRANSACTIONS/USERS MANAGE.

---

## Correções Aplicadas

### 3. Correção de CORS

**Referência:** `docsv5/02_correcoes_e_melhorias.md`

**O que foi corrigido no `main.go`:**
- Origem de produção `https://pixelcraft-studio.store` garantida explicitamente mesmo que ausente do `.env`
- `TrimSpace` em todas as origens para evitar espaços no `.env`
- `MaxAge: 12h` para reduzir requisições OPTIONS desnecessárias

### 4. Sidebar Admin — Bug de Aba Ativa

- `AdminLayout.jsx` atualizado para usar `motion.button` + `useMemo` + `location.pathname.startsWith`
- Aba ativa agora reseta corretamente ao trocar de página sem precisar de refresh

### 5. Dashboard Admin — Simplificação de Métricas

- Removidas métricas de crescimento percentual dos cards principais
- O backend ainda calcula `revenueGrowth`, `usersGrowth`, `salesGrowth` — apenas o UI foi simplificado

### 6. Alinhamento de Colunas na Página de Usuários

- `src/pages/admin/Users.jsx` — células `<td>` reordenadas para alinhar corretamente com os cabeçalhos (CPF, Email, Cargo)

---

## Estado das Correções do Sistema de Cargos (docsv6)

| Problema | Status |
|----------|--------|
| CRÍTICO-01: `AdminAuthMiddleware` legado em Produtos | ✅ Substituído por `CatalogEditMiddleware` |
| CRÍTICO-02: SUPPORT com MANAGE incorreto | ✅ Corrigido via migração 010 |
| CRÍTICO-03: SYSTEM ausente no enum SQL | ✅ Adicionado via migração 010 |
| CRÍTICO-04: Rotas de permissão sem proteção | ✅ Protegidas com `RequirePermission(ROLES, MANAGE)` |
| CRÍTICO-05: Herança incorreta propagada | ✅ Limpeza de permissões herdadas via migração 010 |
| MODERADO-01: VIEW_CPF ausente no backend | ✅ Adicionado ao model Go + migração 010 + permissões para ENGINEERING/DIRECTION |
| MODERADO-02: GetAvailableRoles incompleto | ✅ Agora retorna todos os 8 cargos com `level` e `is_admin` |
| MODERADO-03: Validação aceita só 5 cargos | ✅ Usa `IsValid()` — aceita todos os 8 |
| MODERADO-05: CatalogEditMiddleware errado em rotas de cargo | ✅ Substituído por `AdminPanelMiddleware` |
| MODERADO-06: RemoveRolePermission sem validação | ✅ Adicionada validação `IsValid()` |
| INC-02: pq.Array ausente | ✅ Corrigido no `permission_repository.go` |
| INC-03/04/05: SETTINGS/DASHBOARD inconsistentes | ✅ Uniformizados via migração 010 |
| INC-07: CanView com dupla verificação | ✅ Simplificado |

---

## Pendências Novas (pós-docsv5)

| Item | Prioridade | Descrição |
|------|-----------|-----------|
| DISCOUNTS no enum SQL | Média | Adicionar `'DISCOUNTS'` ao `resource_type` e configurar permissões por cargo |
| CatalogEditMiddleware em Discounts | Média | POST/PUT/DELETE de descontos deveriam exigir DEVELOPMENT+ |

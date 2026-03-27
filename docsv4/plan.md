# Technical Implementation Plan - V4 Update

Este documento descreve o fluxo técnico detalhado para a implementação das mudanças solicitadas, focando em segurança, integridade de dados e performance.

---

## 1. Backend: Banco de Dados (PostgreSQL)

### 1.1 Migração de Preferências
Adicionar coluna JSONB para evitar múltiplas migrações no futuro caso surjam novas configurações de UI.
```sql
ALTER TABLE users ADD COLUMN IF NOT EXISTS preferences JSONB DEFAULT '{"density": "comfortable", "font": "modern"}'::jsonb;
```

### 1.2 Sistema de Permissões
Garantir que a permissão `view_cpf` exista e esteja vinculada aos papéis corretos.
```sql
INSERT INTO permissions (name, description, slug) 
VALUES ('Ver CPF', 'Permite visualizar o CPF descriptografado de usuários', 'view_cpf')
ON CONFLICT (slug) DO NOTHING;
```

### 1.3 Correção de Enums (Suporte e Assinaturas)
Revisar os tipos customizados se existirem no Postgres (ex: `ticket_status`, `subscription_stage`).
- Adicionar o estágio `ENTREGA` se estiver faltando.
- Adicionar o status `WAITING_RESPONSE` caso o erro 500 no suporte seja por cast inválido.

---

## 2. Backend: API (Golang)

### 2.1 Refatoração de Handlers de Suporte
- Local: `backend/internal/handlers/support_handler.go`.
- Mudança: O método `ListAllTickets` deve parsear os múltiplos filtros da query string e passá-los para o repositório.
- Validação: Se `status` e `priority` forem vazios, não aplicar a cláusula WHERE para esse campo específico.

### 2.2 Endpoint de Transações Admin
- Local: `backend/internal/handlers/admin_transaction_handler.go` (Novo).
- Objetivo: Retornar faturas com status `PAID` ou `COMPLETED`.
- Query: `SELECT * FROM transactions WHERE status IN ('PAID', 'COMPLETED') ORDER BY created_at DESC LIMIT 20;`

### 2.3 Lógica de Download de Arquivos
- Local: `backend/internal/handlers/file_handler.go`.
- Ajuste: No método `DownloadFile`, setar explicitamente o header `Content-Type: application/octet-stream` e capturar a extensão original do arquivo salva no banco para compor o `filename` no `Content-Disposition`.

---

## 3. Frontend: Estrutura de Rotas e Navegação

### 3.1 Mapeamento de Rotas no React
Alterar `src/App.jsx` para refletir os novos slugs amigáveis.
- /loja e /shop sao URLs diferentes, /loja é a loja real, /shop é a loja de vitrine

### 3.2 Contexto de Preferências
- Integrar `preferences` no `AuthContext.jsx`.
- Criar um hook `useUserPreferences` que retorna as classes CSS e tokens baseados no estado do banco de dados.

---

## 4. Frontend: Implementação de Componentes Admin

### 4.1 Widget de Receita Total
- Lógica: Consumir do novo endpoint de transações.
- Soma: Iterar sobre as transações pagas do mês atual.

### 4.2 Modal Draggable (Múltiplos Itens)
- Tecnologia: `framer-motion`.
- Estrutura:
  ```jsx
  <motion.div drag dragConstraints={...}>
    <header>Detalhes do Pedido</header>
    <main> {item.products.map(...)} </main>
  </motion.div>
  ```

---

## 5. Fluxo de Debug (Erros 500)

### 5.1 Admin Subscription (Status Change)
1. Ativar logs detalhados no backend.
2. Monitorar a transição de `status` no `subscription_service.go`.
3. Verificar se há algum trigger no banco de dados que falha ao tentar inserir um log de auditoria com dados nulos ou tipos incompatíveis quando o status é `COMPLETED`.

### 5.2 Filtros de Suporte
1. Capturar o SQL gerado pelo repositório quando múltiplos filtros são aplicados.
2. Provável erro de sintaxe SQL ou número incorreto de argumentos (`$1, $2...`) na query dinâmica.

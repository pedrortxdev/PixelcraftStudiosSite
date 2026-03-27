# Sistema de Gerenciamento de Cargos - Pixelcraft

## 📋 Visão Geral

Sistema completo de gerenciamento de cargos e permissões para o painel administrativo da Pixelcraft. Permite que administradores visualizem e configurem permissões granulares para cada cargo do sistema.

## 🎯 Funcionalidades Implementadas

### Frontend (React)
- ✅ Página de gerenciamento de cargos (`/admin/roles`)
- ✅ Lista de cargos com hierarquia visual
- ✅ Matriz de permissões (recursos × ações)
- ✅ Edição de permissões com validação de hierarquia
- ✅ Busca e filtros
- ✅ Layout responsivo (desktop, tablet, mobile)
- ✅ Notificações de sucesso/erro
- ✅ Atualizações otimistas com rollback

### Backend (Go)
- ✅ Endpoints REST para gerenciamento de permissões
- ✅ Validação de hierarquia de cargos
- ✅ Permissões padrão configuradas no banco
- ✅ Sistema de roles já existente integrado

## 🏗️ Estrutura de Arquivos Criados

### Frontend
```
src/
├── pages/admin/
│   └── AdminRoles.jsx                    # Página principal
├── components/admin/roles/
│   ├── RoleCard.jsx                      # Card de cargo
│   ├── RolesList.jsx                     # Lista de cargos
│   ├── PermissionsMatrix.jsx             # Matriz de permissões
│   ├── RoleDetailPanel.jsx               # Painel de detalhes
│   └── NotificationToast.jsx             # Notificações
├── hooks/
│   ├── useRoles.js                       # Hook de cargos
│   ├── usePermissions.js                 # Hook de permissões
│   └── useRoleHierarchy.js               # Hook de hierarquia
├── constants/
│   ├── roles.js                          # Constantes de cargos
│   └── permissions.js                    # Constantes de permissões
├── utils/
│   └── roleUtils.js                      # Funções utilitárias
└── services/
    └── api.js                            # API service (atualizado)
```

### Backend
```
backend/
└── database/
    └── 007_default_role_permissions.sql  # Migração de permissões padrão
```

## 🔐 Hierarquia de Cargos

```
DIRECTION (8)      → Acesso total (incluindo gerenciar cargos)
ENGINEERING (7)    → Acesso completo exceto cargos
DEVELOPMENT (6)    → Edita planos/produtos/jogos/categorias
ADMIN (5)          → Visualização total, sem edição
SUPPORT (4)        → Acesso restrito (tickets + emails próprios)
CLIENT_VIP (3)     → Cliente VIP (área do cliente)
CLIENT (2)         → Cliente padrão (área do cliente)
PARTNER (1)        → Parceiro (área do cliente)
```

## 📊 Recursos e Ações

### Recursos Disponíveis
- **Administração**: USERS, ROLES, SETTINGS
- **Conteúdo**: PRODUCTS, GAMES, CATEGORIES, PLANS, FILES
- **Financeiro**: ORDERS, TRANSACTIONS
- **Suporte**: SUPPORT, EMAILS
- **Sistema**: DASHBOARD

### Ações Disponíveis
- **VIEW**: Visualizar informações
- **CREATE**: Criar novos registros
- **EDIT**: Editar registros existentes
- **DELETE**: Deletar registros
- **MANAGE**: Todas as ações (equivale a VIEW + CREATE + EDIT + DELETE)

## 🎨 Permissões Padrão Configuradas

### SUPPORT
- SUPPORT: VIEW, CREATE, EDIT
- EMAILS: VIEW, CREATE
- DASHBOARD: VIEW

### ADMIN
- Todos os recursos: VIEW (somente visualização)

### DEVELOPMENT
- Todos os recursos: VIEW
- PRODUCTS, PLANS, GAMES, CATEGORIES, FILES: CREATE, EDIT, DELETE

### ENGINEERING
- Todos os recursos (exceto ROLES): MANAGE

### DIRECTION
- Todos os recursos (incluindo ROLES): MANAGE

## 🚀 Como Usar

### Acessar a Página
1. Faça login como administrador
2. Acesse o menu lateral do painel admin
3. Clique em "Cargos" (ícone UserCog)
4. URL: `/admin/roles`

### Visualizar Permissões
1. Selecione um cargo na lista lateral
2. Veja a matriz de permissões no painel principal
3. Permissões ativas aparecem com checkboxes marcados

### Editar Permissões
1. Selecione um cargo que você pode editar (hierarquia inferior)
2. Clique em "Editar Permissões"
3. Marque/desmarque checkboxes para adicionar/remover permissões
4. Mudanças são salvas automaticamente

### Regras de Hierarquia
- Apenas cargos superiores podem editar cargos inferiores
- DIRECTION pode editar todos os cargos
- Cargos bloqueados aparecem com ícone de cadeado

## 🔧 Endpoints da API

### Gerenciamento de Permissões
```
GET    /api/v1/admin/permissions/available-roles  # Lista todos os cargos
GET    /api/v1/admin/permissions/roles/:role      # Permissões de um cargo
POST   /api/v1/admin/permissions/roles/:role      # Adicionar permissão
DELETE /api/v1/admin/permissions/roles/:role      # Remover permissão
GET    /api/v1/admin/permissions/resources        # Recursos disponíveis
GET    /api/v1/admin/permissions/actions          # Ações disponíveis
GET    /api/v1/permissions/me                     # Permissões do usuário logado
```

### Gerenciamento de Cargos de Usuários
```
POST   /api/v1/admin/users/:id/roles          # Adicionar cargo a usuário
DELETE /api/v1/admin/users/:id/roles/:role    # Remover cargo de usuário
GET    /api/v1/admin/users/:id/roles          # Listar cargos de usuário
```

## 📱 Responsividade

### Desktop (> 768px)
- Layout em duas colunas (sidebar + painel principal)
- Matriz de permissões completa

### Tablet/Mobile (≤ 768px)
- Layout em coluna única
- Sidebar colapsável
- Matriz scrollável horizontalmente

## 🎯 Usuários Sem Cargo

Usuários sem cargo administrativo podem:
- ✅ Usar a área do cliente normalmente
- ✅ Fazer depósitos
- ✅ Comprar produtos
- ✅ Abrir tickets de suporte
- ✅ Gerenciar assinaturas
- ❌ Não têm acesso ao painel admin

## 🔒 Segurança

- Validação de hierarquia no backend e frontend
- JWT obrigatório em todas as requisições
- Atualizações otimistas com rollback em caso de erro
- Logs de todas as modificações de permissões
- Proteção contra modificação de cargos superiores

## 🧪 Testes

Para testar o sistema:

1. **Como DIRECTION**:
   - Pode editar todos os cargos
   - Pode adicionar/remover qualquer permissão

2. **Como ENGINEERING**:
   - Pode editar DEVELOPMENT, ADMIN, SUPPORT
   - Não pode editar DIRECTION

3. **Como DEVELOPMENT**:
   - Pode editar ADMIN, SUPPORT
   - Não pode editar ENGINEERING ou DIRECTION

## 📝 Próximos Passos (Opcional)

- [ ] Adicionar histórico de mudanças de permissões
- [ ] Criar cargos customizados (além dos padrão)
- [ ] Exportar/importar configurações de permissões
- [ ] Dashboard de auditoria de permissões
- [ ] Notificações quando permissões são alteradas

## 🐛 Troubleshooting

### Erro "Você não pode editar este cargo"
- Verifique se seu cargo tem hierarquia superior ao cargo que está tentando editar
- Apenas DIRECTION pode editar todos os cargos

### Permissões não aparecem
- Verifique se a migração `007_default_role_permissions.sql` foi executada
- Verifique se o backend está rodando corretamente

### Erro 401/403
- Verifique se está logado
- Verifique se tem permissões administrativas
- Limpe o cache e faça login novamente

## 📞 Suporte

Para dúvidas ou problemas, entre em contato com a equipe de desenvolvimento.

---

**Desenvolvido para Pixelcraft Studio** 🎮


---

## 🆕 Funcionalidades Avançadas Implementadas

### 1. Herança de Permissões ✅
- Cargos podem herdar todas as permissões de cargos inferiores
- Herança com um clique
- Remoção de permissões herdadas
- Indicadores visuais para permissões herdadas vs. diretas

**Como usar:**
1. Selecione um cargo que pode herdar (DIRECTION, ENGINEERING, DEVELOPMENT, ADMIN)
2. Na seção "Herança de Permissões", selecione um cargo de origem
3. Clique em "Herdar" para copiar todas as permissões
4. Use "Remover Permissões Herdadas" para desfazer

### 2. Histórico de Mudanças (Audit Log) ✅
- Histórico completo de todas as mudanças de permissões
- Rastreamento de quem fez as mudanças e quando
- Filtro por cargo e data
- Paginação para grandes volumes de dados
- Tipos de operação: ADD, REMOVE, INHERIT, REMOVE_INHERITED

**Como usar:**
1. Clique no botão "Histórico" na barra superior
2. Filtre por cargo se necessário
3. Visualize todas as mudanças com timestamps e informações do usuário

### 3. Exportar/Importar Configurações ✅
- Exportar configurações de permissões como JSON
- Importar configurações de arquivos JSON
- Opções de sobrescrever ou mesclar
- Funcionalidade de copiar para área de transferência
- Download como arquivo

**Como usar:**
1. Clique no botão "Exportar/Importar"
2. **Exportar**: Selecione cargos, clique em exportar, copie ou baixe o JSON
3. **Importar**: Cole o JSON, escolha a opção de sobrescrever, clique em importar

### 4. Cargos Customizados ✅
- Criar cargos customizados com nomes e níveis de hierarquia personalizados
- Atribuir cores e descrições customizadas
- Funcionalidade de exclusão suave (flag is_active)
- Operações CRUD completas via API
- Interface visual completa na aba "Cargos Customizados"

**Como usar:**
1. Acesse a aba "Cargos Customizados"
2. Clique em "Novo Cargo"
3. Preencha nome, descrição, cor e nível de hierarquia
4. Clique em "Criar Cargo"

### 5. Templates de Permissões ✅
- Salvar configurações de permissões como templates reutilizáveis
- Visibilidade pública/privada de templates
- Biblioteca de templates para configuração rápida
- Aplicar templates com um clique
- Download de templates como JSON

**Como usar:**
1. Acesse a aba "Templates"
2. Para criar: Clique em "Salvar Template", selecione cargos, preencha informações
3. Para aplicar: Clique em "Aplicar" no template desejado
4. Para baixar: Clique no ícone de download

### 6. Dashboard de Auditoria ✅
- Estatísticas visuais no painel de detalhes do cargo
- Total de permissões, permissões diretas e herdadas
- Cards informativos com ícones e cores
- Atualização em tempo real ao modificar permissões

### 7. Notificações ✅
- Painel de notificações na aba "Notificações"
- Feed de notificações com status lido/não lido
- Marcar como lida individualmente ou todas de uma vez
- Indicadores visuais por tipo de notificação
- Timestamps relativos (há X minutos/horas/dias)

**Como usar:**
1. Acesse a aba "Notificações"
2. Clique em uma notificação não lida para marcá-la como lida
3. Use "Marcar todas como lidas" para limpar todas de uma vez

## 🗄️ Schema do Banco de Dados

### Novas Tabelas (Migração 008)
1. `permission_audit_log` - Rastreia todas as mudanças de permissões
2. `custom_roles` - Armazena definições de cargos customizados
3. `permission_templates` - Armazena configurações de permissões reutilizáveis
4. `permission_notifications` - Notificações de usuário para mudanças de permissões

### Tabelas Aprimoradas
- `role_permissions` - Adicionadas colunas `is_inherited` e `inherited_from`

### Funções Criadas
- `inherit_permissions_from_role(target_role, source_role)` - Copia permissões
- `remove_inherited_permissions(role)` - Remove permissões herdadas
- `log_permission_change()` - Trigger de log de auditoria automático

## 🔌 Novos Endpoints da API

### Funcionalidades Avançadas
```
GET    /api/v1/admin/permissions/audit-log                    # Histórico de mudanças
POST   /api/v1/admin/permissions/roles/:role/inherit          # Herdar permissões
DELETE /api/v1/admin/permissions/roles/:role/inherited        # Remover herdadas
GET    /api/v1/admin/permissions/custom-roles                 # Listar cargos customizados
POST   /api/v1/admin/permissions/custom-roles                 # Criar cargo customizado
DELETE /api/v1/admin/permissions/custom-roles/:id             # Deletar cargo customizado
GET    /api/v1/admin/permissions/export                       # Exportar permissões
POST   /api/v1/admin/permissions/import                       # Importar permissões
GET    /api/v1/admin/permissions/templates                    # Listar templates
POST   /api/v1/admin/permissions/templates                    # Salvar template
GET    /api/v1/admin/permissions/dashboard                    # Estatísticas do dashboard
GET    /api/v1/permissions/notifications                      # Notificações do usuário
PUT    /api/v1/permissions/notifications/:id/read             # Marcar como lido
```

## 📦 Novos Componentes Frontend

### Componentes Criados
- `src/components/admin/roles/AuditLogViewer.jsx` - ✅ Visualizador de histórico
- `src/components/admin/roles/PermissionInheritance.jsx` - ✅ UI de herança
- `src/components/admin/roles/ExportImportModal.jsx` - ✅ Modal de exportar/importar
- `src/components/admin/roles/CustomRolesManager.jsx` - ✅ Gerenciador de cargos customizados
- `src/components/admin/roles/NotificationsPanel.jsx` - ✅ Painel de notificações
- `src/components/admin/roles/TemplatesLibrary.jsx` - ✅ Biblioteca de templates
- `src/components/admin/roles/RoleDetailPanel.jsx` - ✅ Melhorado com estatísticas visuais

### Serviços Backend
- `backend/internal/handlers/permission_advanced_handler.go` - ✅ Handler de funcionalidades avançadas
- `backend/internal/service/permission_advanced_service.go` - ✅ Lógica avançada

## 🔄 Hierarquia de Herança

```
DIRECTION (Nível 5)
  ├─ Pode editar: ENGINEERING, DEVELOPMENT, ADMIN, SUPPORT
  └─ Pode herdar de: ENGINEERING, DEVELOPMENT, ADMIN, SUPPORT

ENGINEERING (Nível 4)
  ├─ Pode editar: DEVELOPMENT, ADMIN, SUPPORT
  └─ Pode herdar de: DEVELOPMENT, ADMIN, SUPPORT

DEVELOPMENT (Nível 3)
  ├─ Pode editar: ADMIN, SUPPORT
  └─ Pode herdar de: ADMIN, SUPPORT

ADMIN (Nível 2)
  ├─ Pode editar: SUPPORT
  └─ Pode herdar de: SUPPORT

SUPPORT (Nível 1)
  ├─ Pode editar: Nenhum
  └─ Pode herdar de: Nenhum
```

## 🔐 Considerações de Segurança Adicionais

1. **Aplicação de Hierarquia**: Usuários só podem editar cargos abaixo do seu nível
2. **Trilha de Auditoria**: Todas as mudanças são registradas com ID do usuário e timestamp
3. **Validação de Permissões**: Backend valida todas as mudanças de permissões
4. **Acesso Baseado em Cargo**: Middleware aplica requisitos de cargo
5. **Sem Auto-Modificação**: Usuários não podem modificar as permissões do próprio cargo

## 🧪 Testando as Novas Funcionalidades

### Teste de Herança de Permissões
1. Faça login como DIRECTION
2. Selecione o cargo ENGINEERING
3. Herde permissões de DEVELOPMENT
4. Verifique que as permissões foram copiadas
5. Remova as permissões herdadas

### Teste de Histórico
1. Faça algumas mudanças de permissões
2. Clique em "Histórico"
3. Verifique que todas as mudanças estão registradas
4. Filtre por cargo específico

### Teste de Exportar/Importar
1. Exporte permissões de ADMIN e SUPPORT
2. Salve o JSON
3. Faça mudanças nas permissões
4. Importe o JSON salvo
5. Verifique que as permissões foram restauradas

## 📊 Migrações do Banco de Dados

### Ordem de Execução
```bash
# 1. Permissões padrão
psql -U your_user -d your_database -f backend/database/007_default_role_permissions.sql

# 2. Funcionalidades avançadas
psql -U your_user -d your_database -f backend/database/008_permission_enhancements.sql
```

## 🚀 Build e Deploy

### Backend
```bash
cd backend
go mod tidy
go build -o api ./cmd/api
./api
```

### Frontend
```bash
npm install
npm run build
npm run dev
```

## 🐛 Troubleshooting Adicional

### Herança não funciona
- Verifique se a migração 008 foi executada
- Verifique se o cargo pode herdar (hierarquia)
- Cheque os logs do backend para erros

### Histórico não aparece
- Verifique se a tabela `permission_audit_log` existe
- Verifique se o trigger está ativo
- Faça uma mudança de permissão para testar

### Exportar/Importar falha
- Verifique o formato do JSON
- Certifique-se de que o JSON contém a chave "permissions"
- Verifique os logs do backend para erros de validação

## 📈 Melhorias Futuras

### Funcionalidades Implementadas
1. ✅ Herança de permissões - IMPLEMENTADO
2. ✅ Log de auditoria - IMPLEMENTADO
3. ✅ UI de cargos customizados - IMPLEMENTADO
4. ✅ Exportar/Importar - IMPLEMENTADO
5. ✅ Dashboard de auditoria - IMPLEMENTADO (cards visuais)
6. ✅ UI de notificações - IMPLEMENTADO

### Melhorias Potenciais
- Operações de permissões em massa
- Clonagem de cargos
- Busca e filtragem de permissões
- Filtragem avançada do log de auditoria
- Atualizações de permissões em tempo real via WebSocket
- Comparação de permissões entre cargos
- Estatísticas de uso de cargos
- Gráficos e visualizações de dados

---

**Última atualização:** Sistema completo implementado com todas as funcionalidades avançadas ✅

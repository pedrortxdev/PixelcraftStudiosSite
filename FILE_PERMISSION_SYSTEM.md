# Sistema de Gerenciamento de Arquivos com Permissões - Pixelcraft

## 📋 Visão Geral

Sistema completo de gerenciamento de arquivos com controle de acesso baseado em:
- **Tipo de Acesso** (Público, Privado, Cargo)
- **Cargos** (Roles) específicos
- **Produtos** comprados
- **Links públicos** com token e expiração
- **Limites de download**

## 🎯 Funcionalidades Implementadas

### Backend (Go)

#### 1. Controle de Acesso por Tipo
- **PUBLIC**: Qualquer pessoa com o link pode baixar
- **PRIVATE**: Apenas quem comprou produtos vinculados pode baixar
- **ROLE**: Apenas usuários com cargos específicos podem baixar

#### 2. Permissões por Cargo
- Múltiplos cargos permitidos por arquivo
- Integração com sistema de roles existente
- Hierarquia de cargos respeitada

#### 3. Permissões por Produto
- Vincula arquivo a produtos da loja
- Usuário precisa ter comprado pelo menos um produto vinculado
- Múltiplos produtos podem ser vinculados

#### 4. Links Públicos
- Token único para compartilhamento
- Expiração configurável
- Limite de downloads configurável
- Regeneração de token a qualquer momento

#### 5. Audit Log
- Registro de todos os acessos (view, download, attempted)
- IP address e user agent
- Motivo de concessão/negação de acesso

### Frontend (React)

#### Página Admin Files Aprimorada
- Visualização de todos os arquivos
- Tipo de acesso com ícones coloridos
- Contador de downloads
- Modal de permissões completo
- Seleção de cargos e produtos
- Geração e cópia de links públicos

## 🗄️ Schema do Banco de Dados

### Tabela `files` (Atualizada)
```sql
-- Novas colunas adicionadas
access_type              VARCHAR(20)      -- PUBLIC, PRIVATE, ROLE
required_role            VARCHAR(50)      -- Cargo único requerido
allowed_roles            JSONB            -- Array de cargos permitidos
required_product_id      UUID             -- Produto único requerido
allowed_product_ids      JSONB            -- Array de produtos permitidos
public_link_token        UUID             -- Token para link público
public_link_expires_at   TIMESTAMP        -- Expiração do link público
download_count           INTEGER          -- Contador de downloads
max_downloads            INTEGER          -- Limite máximo de downloads
```

### Tabelas Criadas
```sql
file_access_logs         -- Log de auditoria de acessos
file_role_permissions    -- Permissões de cargos (normalizado)
file_product_permissions -- Permissões de produtos (normalizado)
```

### Funções do Banco
```sql
check_file_access(file_id, user_id)  -- Verifica se usuário tem acesso
log_file_access(...)                 -- Loga tentativa de acesso
```

## 🔐 Tipos de Acesso

### 1. PÚBLICO (PUBLIC)
- **Quem pode baixar**: Qualquer pessoa com o link
- **Requer autenticação**: Não (para download via token público)
- **Configurações**:
  - Link público com token único
  - Expiração opcional
  - Limite de downloads opcional

**Use cases**:
- Arquivos de demonstração
- Recursos gratuitos
- Documentos públicos

### 2. PRIVADO (PRIVATE)
- **Quem pode baixar**: Apenas compradores de produtos vinculados
- **Requer autenticação**: Sim
- **Configurações**:
  - Produtos vinculados (um ou múltiplos)
  - Usuário precisa ter comprado pelo menos um

**Use cases**:
- Downloads de produtos pagos
- Conteúdo exclusivo para clientes
- Arquivos de suporte técnico

### 3. CARGO (ROLE)
- **Quem pode baixar**: Usuários com cargos específicos
- **Requer autenticação**: Sim
- **Configurações**:
  - Cargos permitidos (um ou múltiplos)
  - Respeita hierarquia de cargos

**Use cases**:
- Documentos internos da equipe
- Recursos para staff
- Arquivos administrativos

## 🚀 Endpoints da API

### Upload e Gerenciamento Básico
```
POST   /api/v1/files                           # Upload de arquivo (admin)
GET    /api/v1/files                           # Lista arquivos do usuário
GET    /api/v1/files/:id                       # Detalhes do arquivo
GET    /api/v1/files/:id/download              # Download do arquivo
PUT    /api/v1/files/:id                       # Atualizar nome (admin)
DELETE /api/v1/files/:id                       # Excluir arquivo (admin)
GET    /api/v1/admin/files                     # Lista todos arquivos (admin)
GET    /api/v1/files/selection                 # Seleção para produtos
```

### Permissões
```
GET    /api/v1/files/:id/permissions           # Obter permissões
PUT    /api/v1/files/:id/permissions           # Atualizar permissões (admin)
POST   /api/v1/files/:id/permissions/roles     # Adicionar cargo (admin)
DELETE /api/v1/files/:id/permissions/roles/:role  # Remover cargo (admin)
POST   /api/v1/files/:id/permissions/products  # Adicionar produto (admin)
DELETE /api/v1/files/:id/permissions/products/:id  # Remover produto (admin)
```

### Links Públicos e Logs
```
POST   /api/v1/files/:id/regenerate-public-link  # Regenerar link (admin)
GET    /api/v1/files/:id/access-logs          # Logs de acesso (admin)
GET    /api/v1/files/public/:token/download   # Download público (sem auth)
```

## 📦 Estrutura de Arquivos

### Backend
```
backend/
├── cmd/api/main.go                        # Rotas atualizadas
├── internal/
│   ├── handlers/file_handler.go           # Handlers completos
│   ├── service/file_service.go            # Lógica de negócio
│   ├── repository/file_repository.go      # Acesso ao banco
│   └── models/file.go                     # Modelos atualizados
└── database/
    └── 011_file_access_control.sql        # Migração completa
```

### Frontend
```
src/
├── pages/admin/AdminFiles.jsx             # Página aprimorada
├── services/api.js                        # Endpoints atualizados
└── constants/permissions.js               # Permissões (já existe)
```

## 🎨 Interface do Usuário

### Tabela de Arquivos
- **Nome**: Nome do arquivo com ícone
- **Tipo**: JAR, ZIP, PNG, etc.
- **Tamanho**: Formatado (B, KB, MB)
- **Acesso**: Ícone + cor por tipo (Globe/Verde, Shield/Roxo, Lock/Vermelho)
- **Downloads**: Contador de downloads
- **Usuário**: Quem fez upload
- **Ações**: Permissões, Download, Excluir

### Modal de Permissões
1. **Tipo de Acesso**: 3 botões (Público, Cargos, Privado)
2. **Seleção de Cargos**: Lista de cargos com cores
3. **Seleção de Produtos**: Lista de produtos da loja
4. **Link Público**: Campo com URL, copiar e regenerar
5. **Limite de Downloads**: Input numérico opcional
6. **Expiração**: Date-time picker opcional

## 🔧 Como Usar

### 1. Upload de Arquivo
```javascript
// Frontend
const file = document.querySelector('input[type="file"]').files[0];
const formData = new FormData();
formData.append('file', file);
formData.append('name', 'Meu Arquivo');

await filesAPI.upload(file, 'Meu Arquivo');
```

### 2. Configurar Permissões por Cargo
```javascript
// Admin define que apenas DEVELOPMENT e ADMIN podem baixar
await filesAPI.updatePermissions(fileId, {
  access_type: 'ROLE',
  allowed_roles: ['DEVELOPMENT', 'ADMIN'],
  required_role: 'DEVELOPMENT' // primário
});
```

### 3. Configurar Permissões por Produto
```javascript
// Admin vincula arquivo a produtos
await filesAPI.updatePermissions(fileId, {
  access_type: 'PRIVATE',
  allowed_product_ids: [productId1, productId2]
});
```

### 4. Gerar Link Público
```javascript
// Admin configura acesso público com expiração
await filesAPI.updatePermissions(fileId, {
  access_type: 'PUBLIC',
  public_link_expires_at: '2026-12-31T23:59:59Z',
  max_downloads: 100
});

// Regenerar link (invalida anterior)
const response = await filesAPI.regeneratePublicLink(fileId);
console.log(response.public_link_url);
```

### 5. Download Público
```
// Usuário não autenticado pode baixar com token
GET https://api.pixelcraft-studio.store/api/v1/files/public/{token}/download
```

### 6. Ver Logs de Acesso
```javascript
// Admin visualiza quem acessou o arquivo
const logs = await filesAPI.getAccessLogs(fileId, {
  page: 1,
  page_size: 20
});
```

## 🔒 Segurança

### Validações
- Apenas dono do arquivo ou admin pode modificar permissões
- Admin panel middleware requer cargo administrativo
- Token público é UUID único
- Expiração é verificada no banco de dados
- Limite de downloads é atômico

### Audit Trail
- Todos os acessos são logados
- IP address e user agent capturados
- Motivo de concessão/negação registrado
- Logs podem ser visualizados por admin

## 📊 Exemplos de Uso

### Exemplo 1: Produto Pago
```javascript
// 1. Admin faz upload do arquivo
const uploaded = await filesAPI.upload(file, 'Plugin Premium');

// 2. Vincula ao produto
await filesAPI.updatePermissions(uploaded.id, {
  access_type: 'PRIVATE',
  allowed_product_ids: [produtoPremiumId]
});

// 3. Atualiza produto com file_id
await productsAPI.update(produtoPremiumId, {
  file_id: uploaded.id
});
```

### Exemplo 2: Documento Interno
```javascript
// Admin faz upload de documento interno
const doc = await filesAPI.upload(file, 'Documentação Interna');

// Restringe a cargos específicos
await filesAPI.updatePermissions(doc.id, {
  access_type: 'ROLE',
  allowed_roles: ['DIRECTION', 'ENGINEERING', 'ADMIN']
});
```

### Exemplo 3: Recurso Gratuito
```javascript
// Admin faz upload de recurso gratuito
const free = await filesAPI.upload(file, 'Texture Pack Free');

// Torna público com limite
await filesAPI.updatePermissions(free.id, {
  access_type: 'PUBLIC',
  max_downloads: 1000,
  public_link_expires_at: '2027-12-31T23:59:59Z'
});
```

## 🧪 Testes

### Testar Acesso por Cargo
1. Faça login como usuário sem cargo
2. Tente acessar arquivo ROLE
3. Deve receber erro 403

4. Faça login como usuário com cargo permitido
5. Tente acessar arquivo ROLE
6. Download deve funcionar

### Testar Acesso por Produto
1. Compre um produto
2. Tente acessar arquivo vinculado a esse produto
3. Download deve funcionar

4. Cancele a compra (ou use usuário sem compra)
5. Tente acessar arquivo
6. Deve receber erro 403

### Testar Link Público
1. Configure arquivo como PUBLIC
2. Copie o link público
3. Abra em janela anônima
4. Download deve funcionar

5. Espere expirar (ou defina expiração no passado)
6. Tente acessar link
7. Deve receber erro 404

## 🐛 Troubleshooting

### Erro "Access denied"
- Verifique se usuário tem cargo necessário
- Verifique se usuário comprou produto necessário
- Verifique se arquivo não está expirado
- Verifique se limite de downloads não foi atingido

### Link público não funciona
- Verifique se token está correto
- Verifique se link não expirou
- Verifique se max_downloads não foi atingido
- Tente regenerar o link

### Permissões não salvam
- Verifique se é admin
- Verifique se é dono do arquivo
- Verifique se dados estão no formato correto
- Verifique logs do backend

## 📈 Melhorias Futuras

### Funcionalidades Implementadas
1. ✅ Controle de acesso por tipo (PUBLIC/PRIVATE/ROLE)
2. ✅ Permissões por cargo
3. ✅ Permissões por produto
4. ✅ Links públicos com token
5. ✅ Expiração de links
6. ✅ Limite de downloads
7. ✅ Audit log completo
8. ✅ Interface administrativa

### Melhorias Potenciais
- Upload em massa
- Download em massa
- Visualizador de arquivos (preview)
- Estatísticas de uso
- Webhooks para downloads
- Integração com CDN
- Anti-hotlink protection
- Rate limiting por IP

## 📝 Migração

### Executar Migração
```bash
# No banco de dados
psql -U pixelcraft_user -d pixelcraft -f backend/database/011_file_access_control.sql
```

### Verificar Migração
```sql
-- Verificar colunas
\d files

-- Verificar funções
\df check_file_access
\df log_file_access
```

## 🔑 Chaves de Permissão

### Access Types
- `PUBLIC` - Qualquer pessoa com link
- `PRIVATE` - Apenas compradores
- `ROLE` - Apenas cargos específicos

### Roles Disponíveis
- `DIRECTION` - Nível mais alto
- `ENGINEERING` - Engenharia
- `DEVELOPMENT` - Desenvolvimento
- `ADMIN` - Administração
- `SUPPORT` - Suporte

### Actions (Audit Log)
- `VIEW` - Visualizou detalhes
- `DOWNLOAD` - Baixou arquivo
- `ATTEMPTED` - Tentou acessar

---

**Desenvolvido para Pixelcraft Studio** 🎮

*Sistema completo de gerenciamento de arquivos com controle de acesso granular*

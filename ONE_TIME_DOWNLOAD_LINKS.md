# One-Time Download Links - Sistema de Links Temporários

## Visão Geral

O sistema de **One-Time Download Links** permite gerar links públicos temporários e únicos para download de arquivos privados. Cada link pode ser configurado para expirar após um determinado tempo ou número de usos.

## Como Funciona

1. **Geração do Link**: Um administrador seleciona um arquivo privado e gera um link one-time
2. **Configuração**: Define o tempo de expiração (em minutos) e o número máximo de downloads
3. **Distribuição**: O link é copiado e enviado para o destinatário
4. **Download**: O destinatário acessa o link e baixa o arquivo
5. **Invalidação**: Após o uso ou expiração, o link é automaticamente invalidado

## Endpoints da API

### 1. Gerar Link One-Time
```http
POST /api/v1/files/:id/generate-one-time-link
Authorization: Bearer <token>
```

**Request Body:**
```json
{
  "expires_in_minutes": 15,
  "max_downloads": 1
}
```

**Response:**
```json
{
  "token_id": "uuid",
  "file_id": "uuid",
  "file_name": "nome_do_arquivo.pdf",
  "download_url": "https://api.pixelcraft-studio.store/api/v1/files/one-time/<token>/download",
  "expires_at": "2026-02-28T15:30:00Z",
  "max_downloads": 1,
  "created_at": "2026-02-28T15:15:00Z"
}
```

### 2. Download via Link One-Time
```http
GET /api/v1/files/one-time/:token/download
```

**Sem autenticação necessária** - o token é a autenticação.

## Características de Segurança

- **Token Único**: Cada link possui um token UUID único
- **Expiração Automática**: Links expiram após o tempo configurado
- **Limite de Downloads**: Limite máximo de downloads por link
- **Rastreamento**: IP e User Agent são registrados no primeiro uso
- **Invalidação Automática**: Tokens usados são marcados e removidos após 24h

## Casos de Uso

### 1. Envio de Arquivo para Cliente
```
- Admin gera link com 1 download e 15 minutos de expiração
- Envia o link para o cliente
- Cliente baixa o arquivo
- Link é automaticamente invalidado
```

### 2. Compartilhamento Temporário
```
- Admin gera link com 5 downloads e 60 minutos de expiração
- Compartilha com equipe
- Cada membro pode baixar uma vez
- Link expira após 1 hora ou 5 downloads
```

### 3. Distribuição Controlada
```
- Admin gera link com 100 downloads e 24 horas de expiração
- Publica o link em canal restrito
- Usuários autorizados podem baixar
- Link expira após 24h ou 100 downloads
```

## Banco de Dados

### Tabela: `one_time_download_tokens`

| Coluna | Tipo | Descrição |
|--------|------|-----------|
| id | UUID | ID do registro |
| file_id | UUID | ID do arquivo |
| user_id | UUID | ID do usuário que gerou |
| token | UUID | Token único do link |
| created_at | TIMESTAMP | Data de criação |
| expires_at | TIMESTAMP | Data de expiração |
| used_at | TIMESTAMP | Data do primeiro uso |
| is_used | BOOLEAN | Se foi totalmente consumido |
| download_count | INTEGER | Contador de downloads |
| max_downloads | INTEGER | Limite máximo de downloads |
| ip_address | INET | IP do primeiro uso |
| user_agent | TEXT | User Agent do primeiro uso |

### Funções do Banco

- `validate_and_use_download_token(token, ip, user_agent)`: Valida e incrementa uso
- `cleanup_expired_download_tokens()`: Remove tokens expirados/usados

## Frontend (Admin)

No painel administrativo de arquivos, um novo botão com ícone de **chave** 🔑 permite gerar links one-time:

1. Clique no botão 🔑 ao lado do arquivo desejado
2. Configure:
   - Tempo de expiração (1-1440 minutos)
   - Máximo de downloads (1-100)
3. Clique em "Gerar Link"
4. Copie o link gerado
5. Envie para o destinatário

## Migração

Para aplicar a migração no banco de dados:

```bash
cd /Pixelcraft-studio-website
psql -U pixelcraft_user -d pixelcraft -h localhost -f backend/database/012_one_time_download_links.sql
```

Ou adicione ao script de migrações existente.

## Diferenças: Link Público vs One-Time

| Característica | Link Público | One-Time Link |
|----------------|--------------|---------------|
| Duração | Até expiração configurada | Até uso ou expiração |
| Reutilização | Múltiplos downloads | Limitado/configurado |
| Geração | Manual (regenerável) | Automática (sob demanda) |
| Ideal para | Arquivos sempre disponíveis | Compartilhamento pontual |
| Segurança | Média | Alta |

## Exemplo de Uso com cURL

```bash
# 1. Gerar link one-time
curl -X POST "https://api.pixelcraft-studio.store/api/v1/files/<file_id>/generate-one-time-link" \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"expires_in_minutes": 30, "max_downloads": 1}'

# 2. Baixar arquivo (sem autenticação)
curl -L "https://api.pixelcraft-studio.store/api/v1/files/one-time/<token>/download" -o arquivo.pdf
```

## Implementação

- **Backend**: Go (Gin)
- **Banco**: PostgreSQL
- **Frontend**: React
- **Migração**: `backend/database/012_one_time_download_links.sql`

## Arquivos Modificados

### Backend
- `backend/database/012_one_time_download_links.sql` - Nova migração
- `backend/internal/models/file.go` - Novos modelos
- `backend/internal/repository/file_repository.go` - Métodos de repositório
- `backend/internal/service/file_service.go` - Lógica de negócio
- `backend/internal/handlers/file_handler.go` - Handlers HTTP
- `backend/cmd/api/main.go` - Novas rotas

### Frontend
- `src/services/api.js` - Método `generateOneTimeLink`
- `src/pages/admin/AdminFiles.jsx` - UI de geração de links

---

**Status**: ✅ Implementado e testado
**Data**: 2026-02-28

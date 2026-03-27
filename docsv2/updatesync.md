# Contratos de Sincronia de Backend (docsv2/updatesync.md)

Este apêndice especifica rigorosamente as alterações mecânicas e lógicas esperadas do desenvolvedor de Backend. São as definições de endpoint, estrutura de dados e tratativas obrigatórias para se conciliar perfeitamente e seguramente com as exigências da Fase 2 de front-end. O não seguimento destas cláusulas resultará em falhas de comunicação por quebras de protocolo REST.

---

## 1. Módulo Auth: Redefinição de Senha (Novo Protocolo)

### Arquitetura de Tabela (DB Migration Obrigatória)
Criar infraestrutura temporária para persistência do ticket de verificação. Exemplo em DDL Postgres:
```sql
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE password_resets (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    token_hash VARCHAR(255) NOT NULL UNIQUE,  -- Recomendado armazenar Hash e não token raw se for extrema segurança (opcional)
    token_url VARCHAR(255) NOT NULL UNIQUE,   -- Disparado na URL
    verification_code VARCHAR(8) NOT NULL,    -- Random Alphanum Uppercase
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    used BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
```

### Contrato API: Solicitação de Reset
**Endpoint:** `POST /api/v1/auth/forgot-password`
**Payload Entrada:** `{ "email": "usr@exemplo.com" }`
**Comportamento:**
1. Buscar User via e-mail (ignore case).
2. Se não achar, retornar `200 OK` silent (MITIGAR Enumeração de e-mail por atacantes).
3. Se achar: Anular tickets passados dele (`UPDATE password_resets SET used=true WHERE user_id = X`).
4. Gerar: 
   - `Token`: String pseudo-random crypto-segura de 64 bytes hexed.
   - `Code`: `Math/Random` seguro -> string 8 de length (ex: G9PL2XM8).
   - `TTL`: Data UTC + 15 min.
5. Inserir na Tabela, ativar fila/serviço assíncrono do SMTP/E-mail de Serviço.
6. Disparar o envio HTML interpolando link e texto em negrito do código.

### Contrato API: Consumo e Resgate
**Endpoint:** `POST /api/v1/auth/reset-password`
**Payload Entrada:** `{ "token": "abc...xyz", "code": "G9PL2XM8", "new_password": "..." }`
**Comportamento:**
1. Check Tabela `where token_url = ? AND used = false`.
2. Se não existe ou *TTL expirou*: `400 Bad Request` "Token inválido ou expirado".
3. Validar se Code fornecido é IDENTICO ao da base (Timing Safe Compare e Sensitive Case).
4. Proceder validação de força de senha (`new_password`), encriptamento BCrypt e associar a `users`.
5. Marcar como `used = true`. Retornar `200 OK`.

---

## 2. Módulo Mídia: Payload Content-Disposition Strict

A causa raiz da ausência de extensões reside na montagem negligente dos HTTP response headers em streams do backend na biblioteca REST servida. Sem ser explicita com o MIME e o Name do attachment, bibliotecas de browser (Chrome, Firefox) renegam propriedades dependendo de aspas cruas.

### Ação Necessária (Response Handling)
Ao montar resposta do arquivo baixável (`GET /api/v1/files/:id/download` ou via `libraryAPI`), utilizar `utf-8` filename fallback nativamente em Golang/Node (Dependendo da stack original) + o fallback aspas estritas.
**Padrão Imperativo Recomendado Header HTTP:**
```http
Content-Type: application/zip
Content-Disposition: attachment; filename="projeto_pixelcraft_final.zip"; filename*=UTF-8''projeto_pixelcraft_final.zip
```
*(Nota Crítica: Ao servir nomes de arquivos com espaços ou acentos, o Go/Backend deve utilizar `url.PathEscape` ou apropriado no `filename*=UTF-8''` para garantir que o cliente não falhe por nomes complexos, evitando arquivos "sem extensão".)*
Note que se o nome for gerado dinamicamente, garanta via DB Query que `.type` ou metadados forneçam o suffix (`.rar`, `.jar` etc) correto. Retornar `arquivo_id_45` seco ou sem suffix forçará o S.O (Windows/macOS) do cliente a indagar com qual app abrir o arquivo indefinido.

---

## 3. Módulo Segurança ABAC (CPF Privativo)

### Infiltração em DTOs/Marshalling Models
Atualmente, se os JSON responses que mapeiam de tabelas enviam CPF em bruto a todas as responses via Reflect genéricos, isso se configura **Vazamento de Dados Primário**. 

**Regras de Mascaramento em Endpoints Admin:**
Endpoint Catcher: Exemplo: Listagem de Usuasrios `GET /admin/users`
Implementar lógica interceptadora no handler:
1. Obter cargo / roles atrelados do Access Token do admin solicitante.
2. Interrogar `HasPermission(role_id, "users", "view_cpf")`.
3. Procedimento de Serialização JSON: 
   - Se HasPermission == true: `jsonList[i].cpf = dbUser.cpf;`
   - Se HasPermission == false: `jsonList[i].cpf = null;` (Ou string omitida `***.***.***-**`).
Essa lógica reflete o "Zero Trust". Não delegar ao Javascript do JSX do Frontend de cuidar de privacidade por renderização condicional se os Bytes via Network Network Tooling vierem abertos e preenchidos no inspector.

---

## 4. Módulo Financeiro: Sync Payment Gateways

### Desacoplamento do Cliente via Webhooks Reais
O frontend fará poll em URL direta e limpa para aguardar Pix.
- **Ação 1:** Ao criar Payload na `depositAPI`, o Backend tem obrigação de devolver a transação criada no DB.
  Response: `{ "transaction_id": "tx_xyz123", "qr_code_base64": "...", "qr_code": "..." }`
- **Ação 2:** Expor Endpoint de Ping/Polling seguro.
  `GET /api/v1/wallet/transactions/:id/status` (Apenas owner token = PENDENTE ou PAGO etc).
- **Ação 3:** Listener de Webhook Bancário `POST /api/v1/webhooks/payment` validando HashCryptográfica payload, marcando Status Transaction -> COMPLETED; Add DB balance.

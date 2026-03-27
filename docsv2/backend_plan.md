# Design e Arquitetura do Backend (docsv2/backend_plan.md)

Este documento detalha exaustivamente a estrutura técnica e os novos protocolos REST do Backend para dar suporte às implementações rigorosas da Fase 2 do Frontend. O descumprimento destes contratos gerará falhas de comunicação e quebra da aplicação.

---

## 1. Fluxo de Autenticação Segura: Redefinição de Senha (Password Reset Process)

### Contexto e Problema
A arquitetura de backend atual falha criticamente em enviar a senha pura e hardcoded pelo sistema de e-mails, sem garantir mecanismos preventivos como confirmações e expiração segura. O Frontend agora suporta "Esqueci a Senha" de forma moderna, o que demanda o fluxo Duplo (Token em Link + Código de Autorização).

### Arquitetura da Solução (Backend)

#### A. Entidade "PasswordReset" (DB Migration)
Deve-se criar uma estrutura temporária em banco relacional:
- `id` (UUID, PK)
- `user_id` (UUID, FK constraint)
- `token_url` (String, UK): Token seguro e extenso (Ex: Hash SHA-2 ou random crypto de 64 chars) que roteará o cliente.
- `verification_code` (String, UK): Código randômico e curto (ex: 8 dígitos Uppercase AlphaNumérico). Submetido pelo preenchimento da UI.
- `expires_at` (Timestamp): Mandatório (TTL de cerca de 15 minutos).
- `used` (Boolean): Flag de idempotência.

#### B. Diagrama de Estados e Endpoint
1. **Solicitação do Reset (`POST /api/v1/auth/forgot-password`)**:
   - Recebe `{ "email": "..." }`. Se o e-mail não existir, responder estritamente `200 OK` (Prevenir enumeração).
   - Se existir: Inválida e apaga tokens antigos (`used=true` / `delete`), insere um novo token, gera o `verification_code` e despacha para o Mail Service o template HTML.
2. **Confirmação e Resgate (`POST /api/v1/auth/reset-password`)**:
   - Recebe `{ "token": "...", "code": "...", "new_password": "..." }`.
   - Backend busca na DB o `token_url`. Valida se `code` bate (utilizar *timing safe compare*), valida expiração e a flag `used`.
   - Gera Hash (Bcrypt) de `new_password`, salva na tabela de Users, e anula a row atual marcando `used=true`.

---

## 2. Controle de Acesso ABAC (Zero Trust em Rede e CPF)

### Contexto e Problema
O Backend tem enviado o DTO bruto de todos os campos primários (particularmente o CPF) em endpoints de listagem de usuários do Admin. Delegar à UI o ato de "esconder visualmente" é um crime contra leis de proteção de dados.

### Arquitetura de Filtro Interceptador (Response Masking)
Os endpoints `GET /admin/users` e de detalhamento `GET /admin/users/:id` precisam se atentar à Role (Cargo) contida pelo solicitante JWT.
- A função interceptadora avalia `HasPermission(role_id, "users", "view_cpf")` do banco ou Cache Local.
- Caso a validação seja Falsa, o processo de JSON Marshalling da API deve forçar: `List[i].cpf = null;`. A chave pode ser suprimida ou enviar nulo/mascarado (`***.***.***-**`), mas os dados não descem no *Network Tool* do navegador.

---

## 3. Dinâmica Financeira e Históricos (Pix Asynchronous Flow)

### Contexto e Problema
O Frontend foi ajustado para parar a rotina bizarra de observar o saldo global da conta do usuário em loops intervalados (isso gerava a falsa percepção de conta paga quando na verdade o saldo poderia ter recebido uma alteração independente).

### Arquitetura da Solução: Webhooks e Resource Polling
Toda verificação de Pix deve operar unicamente à luz do *Id da Transação* (`transaction_id`).
1. **Payload Generation**: Rota Criadora de transações financeiras não retorna mais só QR Code. Retorna o QR Code e o ID Único da transação no BD Local com estado `PENDING`.
2. **Ping Server Engine**: Construir Rota explícita `GET /api/v1/wallet/transactions/:id/status`. Ela atende apenas o `owner` logado e entrega o estado ('PENDING', 'COMPLETED', etc).
3. **Webhooks Locks**: Quando a Gateway despachar a completude no Endpoint `/webhooks/pix/`, a execução Backend exige Locks rígidos (`SELECT FOR UPDATE`), altera pra `COMPLETED` e insere saldo simultaneamente (ACID) em apenas uma operação no Database.

---

## 4. Fixação de Mídias e Extensões Ausentes (HTTP Strict Files)

### O Problema Tecnico (Content-Disposition Parsing)
Quando os administradores sobem conteúdos nomeados como "Versão Final Patch 1.1.zip", e esses nomes têm barras ou acentos, o Web Browser quebra e entrega os bytes crus, fazendo o computador não saber qual programa abre o arquivo. 

### Modelagem da Resolução:
Em toda rota que sirva Downloads (ex: `GET /files/:id/download`):
- O backend injetará e concatenará compulsoriamente os caracteres exigíveis de `utf-8` na configuração de Disposition do pacote Stream REST de volta.
- Regra de Ouro HTTP Golang/Node Header:
  ```http
  Content-Disposition: attachment; filename="projeto_pixelcraft_final.zip"; filename*=UTF-8''projeto_pixelcraft_final.zip
  ```
  Isso protege 100% dos browsers de inferirem erroneamente arquivos corrompidos. O Backend deve utilizar `PathEscape` ou similar Encoding no trecho literal do arquivo originado em banco, inclusive suas terminações originadas (`.jar`, `.zip`).

# Check-list Fila de Tarefas do Backend (docsv2/backend_tasks.md)

Este documento acompanha as etapas de programação sistemáticas e refatorações que deverão ser implementadas a nível de Servidor para conciliar perfeitamente com a infraestrutura descrita em `backend_plan.md`.

---

## Fase 1: DB Migrations e Estruturas Essenciais
- [x] Construir Table Schema / Migration File para `password_resets` (`id`, `user_id`, `token_url`, `verification_code`, `expires_at`, `used`).
- [x] Checar estrutura de `transactions`: Necessita forçosamente da coluna `transaction_id_gateway` (string), `status` (ENUM type ou string indexada) e `type` (Pix, Crypto).

## Fase 2: Módulo Auth Seguro (2FA Password Reset)
- [x] Atualizar Endpoint `POST /api/v1/auth/forgot-password`:
  - Encontrar conta por e-mail, limpar tokens anteriores ativos.
  - Gerar Random String Hex (Token) de 64 bytes e Código alfanumérico UPPERCASE (A-Z0-9) de 8 chars. Salvar referências na base com TTL (15 min).
  - Executar fila de e-mail integrando template HTML (incluindo URL Param Token e Code). Remover indicação de erro se o user não for achado, enviando always 200 OK silenciosamente.
- [x] Criar Endpoint `POST /api/v1/auth/reset-password`:
  - Receber os 3 elementos (token, code, new_password). Validação de DB (onde `used` é falso e dentro do limite `expires_at`). Trancar tentativa caso code erre. Resetar bcrypt de Users se sucesso, anulando o token via `used=true`.

## Fase 3: Segurança e Privacidade (Data Masking)
- [x] Construir arquitetura helper (ex: `CheckActionPermissionType(token, "users", "view_cpf")`) caso não exista nativo no middleware.
- [x] Refatorar respostas DTO de Endpoints Adms: `GET /admin/users` e `GET /admin/users/:id`. Percorrer `user.cpf`. Se o JWT não contiver permissão `view_cpf`, apagar `cpf` (atribuir "null") para evitar Data Bleeding para o Frontend React.

## Fase 4: Async Pix Flow Engine
- [x] Ajustar a controller de depósito pix retornando o objeto DTO modificado, agora forçosamente com a chave em anexo `{ transaction_id: "tx_..." }`.
- [x] Adicionar Nova Rota Guardada de consulta rápida: `GET /api/v1/wallet/transactions/:id/status`. Deve responder 403 se o solicitante JWT não ditar a ID dona da transação e devolver `{ status: "COMPLETED" | "PENDING" }` no payload.
- [x] Refinar rota Webhook de escuta do Provedor de pagamento que finaliza as faturas para usar Proteção Pessimista (`FOR UPDATE` do BD) evitando acúmulos multiplicados por race condition.

## Fase 5: Strict Headers em Downloads Stream 
- [x] Atualizar as lógicas HTTP de escrita de Header nas controllers de File Serving (onde ocorre a injeção do Mime e File Name). Injetar instrução RFC 5987 e encoding url. E.g: `Header("Content-Disposition", "attachment; filename=\"file.ext\"; filename*=UTF-8''file.ext")`.

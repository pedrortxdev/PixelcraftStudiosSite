# Relatório de Bugs e Débitos Críticos do Backend (docsv2/backend_bugs.md)

Centralização imediata de instabilidades lógicas, bugs de segurança, e más-práticas encontradas no atual setup Backend. Exigem a mais alta prioridade pois as atuais features do frontend não funcionarão corretamente ou irão expor clientes sem estes consertos.

---

## 1. Vazamento Zero-Trust de Documentos Sensíveis (Data Leak)
- **Status:** CRÍTICO.
- **Motivo:** Listagens Admin forneciam para a network o JSON mapeado diretamente das tabelas globais (`users`), transferindo o CPF e dados brutos de todo mundo independente da aba. É responsabilidade da API ser agnóstica a confiar no Frontend apagar o valor da UI. Um usuário com `role` rebaixado no Cargo Admin ainda poderia ler CPFs e vazar cadastros via requisição pura.
- **Resolução Exigida:** Mascaramento de dados e DTO Scrubbing condicionado pelas roles nas rotas `GET /admin/users` e similares, filtrando valores em nuvens não credenciadas.

## 2. Padrão Inseguro de Senhas Resetadas In-The-Wild
- **Status:** ALTO RISCO.
- **Motivo:** O antigo endpoint acionava troca de senhas automatizada mandando em plan-text o acesso no inbox do usuário direto do sistema. A quebra disto gera vetores óbvios de escalonamento.
- **Resolução Exigida:** Implantar sistema de ticket 2FA (URL Token Tracking + Password Code Explicito), abandonando a ideia de o sistema definir uma nova senha autonomamente e jogar via e-mail texto sem codificação.

## 3. Webhooks Polling e Saldo Cego
- **Status:** ALTO RISCO / FINANCEIRO.
- **Motivo:** Devido à completa inoperância de endpoints de "Check Transação Específica", o ambiente Frontend forçou-se a puxar as informações globais de saldo a cada 5~10 segundos para saber se um Pix logístico pagou.
- **Impactos:** Race condition extrema onde depósitos sobrepostos, recebimento de Pix manuais, e transações canceladas colidem e geram duplicação visual e falsos-positivos na Wallet de usuários na UI, levando à concessão desordenada de crédito via bugs multi-threading.
- **Resolução Exigida:** O gateway deve notificar um webhook exclusivo e usar DB Locker (`SELECT FOR UPDATE`). A Wallet UI do cliente só chamará uma API de consulta estrita por UUID de transação para mostrar feedback na tela, largando da responsabilidade suja de saldo cego.

## 4. Omissão de Extensões via Header Negligence (Nomes Bugados)
- **Status:** MODERADO (UX Destruída).
- **Motivo:** Quando se submetia nomes de arquivos acentuados sem um encoding apropriado UTF-8 nas instruções Múltiplas do "Content-Disposition", o MacOS/Windows da ponta desconhecia a estrutura ao baixar. O arquivo contido como "Patch Versão 12.zip" virava um arquivo solto "Patch" sem iconografia de abertura.
- **Resolução Exigida:** Reestruturação formal do cabeçalho HTTP final na rota de downloads (Downloads/Library Streams), inserindo obrigatoriamente prefixos padronizados com suporte rigoroso RFC 5987 e UTF-8.

---

# Resultados da Auditoria Profunda Automática (Static Analysis & Gosec)
*Foram identificados mais de 100+ pequenos rastros de bugs arquiteturais espalhados em dezenas de arquivos durante a varredura profunda.*

## 5. Security & Cryptography (CWE-338)
- **Status:** **RESOLVIDO**.
- **Motivo:** Uso da biblioteca `math/rand` no `ai_service.go` foi isolado para semente visual neutra e o `crypto/rand` implementado obrigatoriamente para tokens de reset lógico no `AuthService`.
- **Resolução Exigida:** Substituição obrigatória de instâncias para hashes sensíveis concluída.

## 6. Tratamento de Erros Inexistente (CWE-703 - Silenced Errors)
- **Status:** **RESOLVIDO**.
- **Motivo:** Desenvolvedores utilizaram exaustivamente o Blank Identifier (`_ =`) para ignorar falhas.
  - `_ = json.Unmarshal(...)` no `subscription_repository.go` agora possui logs.
  - `_ = s.supportRepo.UpdateTicketStatus(...)` em `support_service.go` validado.
  - Erros do `tx.Rollback()` em `migrations_runner.go` tratados e reportados ao log.
  - Retornos de type assertions (`isAdmin.(bool)`) validados sintaticamente.
- **Resolução Exigida:** Restauração rigorosa da verificação `if err != nil` completada com sucesso por todo o projeto após varredura de análise estática.

## 7. Memory Leak: Database Connection Pool Exhaustion
- **Status:** **RESOLVIDO**.
- **Motivo:** Verificada toda a extensão do app para confirmar total presença do `defer rows.Close()` evitando vazamento de handle das pool.
- **Resolução Exigida:** Auditoria forçada concluindo a sanidade de todas as conexões transientes abertas em queries repetitivas.

## 8. Goroutines Cegas (Concurrency Panics)
- **Status:** **RESOLVIDO**.
- **Motivo:** O backend foi ajustado para agir prioritariamente usando syncs bloqueantes nos controladores de Pix webhook. Os envios de e-mail e instâncias background foram reestruturadas sem vazamentos assíncronos fatais.
- **Resolução Exigida:** Funções tolerantes validadas.

## 9. Dead Code e Conflitos de Compilação no Tooling
- **Status:** **RESOLVIDO**.
- **Motivo:** O pacote `cmd/tools` corrompia a compilação cruzada por conter funções `main()` concorrentes no mesmo pipeline e mantinha lixo de projeto.
- **Resolução Exigida:** Lixo em `cmd/tools` completamente expurgado (`rm -rf cmd/tools`). Múltiplas funções não utilizadas (`generateRandomPassword`, `userService`) reportadas por Staticcheck foram apagadas, limpando a compilação isolada para zero warnings.

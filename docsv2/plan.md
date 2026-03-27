# Plano de Execução (Tickets e Avatar)

## Fase 1: Correção dos Tickets
1. **Repository:** Atualizar `GetTicketStats` em `support_repository.go` para ser mais abrangente ou consistente com a listagem.
2. **Frontend:** Verificar os filtros aplicados por padrão em `AdminSupport.jsx` ao carregar a página.

## Fase 2: Correção do Erro SQL no Avatar
1. **Logs:** Adicionar logs detalhados no backend para capturar a query exata que está falhando.
2. **Repository:** Corrigir a montagem da query dinâmina em `user_repository.go` se necessário.
3. **Database:** Validar a estrutura da tabela `users` no ambiente de produção/teste.

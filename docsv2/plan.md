# Plano de Recuperação

## Fase 1: Estabilização do Suporte (Urgente)
1. **Repository:** Corrigir as queries de `ListTickets` e `GetTicketByID` em `support_repository.go` para lidar com valores `NULL` usando `COALESCE`. Isso deve eliminar o erro 500 e fazer os tickets aparecerem.

## Fase 2: Investigação de Vendas
1. **Logs:** Verificar logs de transações falhas no backend.
2. **Checkout:** Testar o fluxo de criação de assinatura e validar se os planos estão sendo carregados corretamente para o checkout.

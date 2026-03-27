# Lista de Tarefas (Urgente)

## 🐛 Bug Fixes
- [x] Corrigir erro de `Scan` (NULL string) na listagem de tickets do admin.
- [x] Sincronizar contagem de tickets com listagem (JOIN com users).
- [x] Corrigir erro de `Scan` ao visualizar detalhes e mensagens do ticket.
- [x] Garantir que tickets de usuários deletados sejam visíveis (LEFT JOIN).
- [x] Implementar auto-atribuição ao responder ticket (melhoria UX).
- [x] Implementar escolha de método de pagamento no Checkout (Saldo vs Mercado Pago).
- [x] Integrar `DepositService` ao `CheckoutService` para gerar cobranças diretas.

## 🧪 Validação
- [x] Confirmar se a listagem de tickets carregou (Status 200).
- [x] Testar checkout com saldo insuficiente gerando link de pagamento.
- [x] Simular um checkout de assinatura via PIX.

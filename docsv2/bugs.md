# Bugs Identificados (Críticos)

## 1. Falha no Carregamento de Tickets (Erro 500)
- **Sintoma:** O admin vê "5 tickets abertos", mas a lista está vazia e o console mostra erro 500 ao acessar `/api/v1/admin/support/tickets`.
- **Análise Técnica:** 
    - A query de listagem faz `JOIN users u ON t.user_id = u.id` e tenta ler `u.full_name` e `u.avatar_url`.
    - No banco de dados, `full_name` e `avatar_url` podem ser `NULL`.
    - O código Go está tentando dar um `Scan` desses valores diretamente em campos do tipo `string`. O Go não permite converter `NULL` do SQL para `string` diretamente, resultando em erro.
- **Solução:** Usar `COALESCE(u.full_name, '')` na query SQL ou utilizar `sql.NullString` no código Go para capturar os valores.

## 2. Falha na Compra de Assinatura
- **Sintoma:** Cliente tentou assinar um serviço ontem e não conseguiu concluir a operação.
- **Análise Técnica:** 
    - O sistema de checkout (`Checkout.jsx` e `CheckoutService.go`) está configurado para aceitar apenas pagamentos via **Saldo da Carteira** (`use_balance: true`).
    - Se o cliente tentar finalizar a compra sem ter depositado saldo previamente via PIX/Link na aba "Carteira", o backend retorna erro de saldo insuficiente.
    - Atualmente, não existe a opção de "Pagar Agora" via PIX diretamente no checkout; o fluxo obriga o usuário a fazer dois passos (Depositar -> Comprar). Isso confunde o cliente e causa falhas na conversão.
- **Solução:** Implementar a opção de pagamento direto via Mercado Pago (PIX/Link) no Checkout, integrando o `DepositService` ao fluxo de finalização de compra.


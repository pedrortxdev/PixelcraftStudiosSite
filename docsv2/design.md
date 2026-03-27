# Melhorias de Design e Lógica

## 1. Dashboard de Suporte
- **Contagem Consistente:** As métricas de "Aberto", "Em Progresso" e "Aguardando" devem ser claras e corresponder exatamente aos itens filtráveis na lista lateral.
- **Indicadores de Status:** Adicionar badges coloridos na listagem para diferenciar visualmente tickets 'OPEN' de 'WAITING_RESPONSE'.

## 2. Feedback de Erro (Avatar)
- **Mensagens Amigáveis:** Em vez de exibir o erro SQL bruto, o sistema deve validar o upload no backend e retornar uma mensagem clara caso o banco de dados falhe.

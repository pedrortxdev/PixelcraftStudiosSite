# Correções e Melhorias Gerais

Este documento registra as correções de bugs e melhorias de interface realizadas recentemente.

## 1. Interface Administrativa (Sidebar)
- **Bug:** A cor da aba ativa na sidebar do admin permanecia marcada mesmo após trocar de aba, resetando apenas com o refresh da página.
- **Correção:**
    - Atualizado `AdminLayout.jsx` para utilizar `motion.button` em vez de `motion.div`.
    - Implementado `useMemo` para os estilos estáticos.
    - Otimizada a lógica de detecção de rota ativa utilizando `location.pathname.startsWith`.

## 2. Dashboard Admin
- **Alteração:** Removidas as métricas de crescimento percentual ("X% este mês") dos cards principais.
- **Motivo:** Simplificação da interface a pedido do usuário, focando nos totais reais e contagens absolutas.
- **Implementação:** Removidos os elementos de UI que exibiam `revenueGrowth`, `userGrowth` e `salesGrowth` no arquivo `Dashboard.jsx`.

## 3. Segurança e Conectividade (CORS)
- **Problema:** Requisições do frontend de produção (`https://pixelcraft-studio.store`) estavam sendo bloqueadas pela política de CORS do backend.
- **Correção:**
    - Atualizado `backend/cmd/api/main.go` para incluir explicitamente a origem de produção.
    - Atualizado `backend/internal/config/config.go` para realizar o `TrimSpace` em todas as origens carregadas via variáveis de ambiente, prevenindo erros causados por espaços em branco no arquivo `.env`.
    - Ajustado o `MaxAge` do CORS para 12 horas para reduzir o tráfego de requisições OPTIONS.

## 4. Alinhamento de Dados (Página de Usuários)
- **Bug:** As colunas de CPF, Email e Cargo estavam com os dados deslocados (ex: email aparecendo na coluna de CPF).
- **Correção:** Reordenadas as células `<td>` no arquivo `src/pages/admin/Users.jsx` para alinhar corretamente com os cabeçalhos `<th>` definidos.

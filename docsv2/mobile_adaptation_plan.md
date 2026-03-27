# Plano de Adaptação Mobile Nativa (Dashboard & UX)

Este documento detalha a estratégia arquitetural e de design para refatorar a Dashboard e a UI geral da Pixelcraft Studios, transformando a experiência web em algo idêntico a um aplicativo nativo (PWA / Mobile-First).

## 1. Experiência de App Nativo (Mobile-First)
- **Viewport e Fullscreen:** Adicionar meta tags para forçar o web app a rodar em tela cheia (`apple-mobile-web-app-capable`, `mobile-web-app-capable`) e travar o zoom do usuário (`user-scalable=no, maximum-scale=1.0`).
- **Navegação (Bottom Navigation Bar):** Remover a Sidebar tradicional no mobile e implementar uma *Bottom Tab Navigation* fixa na parte inferior da tela, contendo ícones de rápido acesso (Início, Loja, Wallet, Suporte, Perfil).
- **Gestos de Swipe:** Implementar detecção de gestos (via `framer-motion` ou `react-use-gesture`) para deslizar entre abas da dashboard, fechar modais arrastando para baixo (Bottom Sheets) e abrir menus laterais.

## 2. Refatoração da Dashboard (Client & Admin)
### Client Dashboard
- **Cards Compactos:** Transformar as estatísticas (Saldo, Gastos, Planos Ativos) em cards horizontalmente scroláveis (Carousel tipo banco/fintech) para otimizar espaço vertical.
- **Histórico de Transações:** Listas de transações devem usar o padrão de *List Item* nativo (Ícone à esquerda, Título/Data no meio, Valor à direita), ao invés de tabelas complexas que quebram a tela.
- **Bottom Sheets:** Formulários de Ação (Adicionar Saldo, Novo Ticket) abrirão de baixo para cima como *Bottom Sheets* curvos, com efeito de blur de fundo (Glassmorphism), evitando redirecionamentos de página que quebram o fluxo no mobile.

### Admin Dashboard
- A interface de Admin é densa em dados. Para responsividade:
- **Tabelas Responsivas:** Tabelas serão convertidas em "Stacked Cards" no mobile. Cada linha da tabela vira um card expansível que mostra os metadados ao tocar.
- **Filtros Ocultos:** Controles de paginação, busca e filtro vão ficar agrupados em um ícone de "Filtro" que abre um modal, economizando espaço no topo.

## 3. Melhorias Visuais e de Performance (UI/UX)
- **Safe Areas:** Utilizar variáveis CSS (`env(safe-area-inset-bottom)`) para garantir que a navegação não fique por baixo do indicador de home do iPhone.
- **Micro-interações:** Adicionar feedback tátil e visual ao clicar em botões (ripple effect/scale down leve com `framer-motion`).
- **Tipografia:** Aumentar o espaçamento base (line-height) e o touch-target global para no mínimo `44x44px` ou `48x48px` para seguir as diretrizes da Apple/Google.
- **Pull-to-Refresh:** Adicionar puxar para atualizar nas páginas de listagem nativamente (já suportado parcialmente por mobile browsers, mas com loading spinner customizado da Pixelcraft).

## 4. Integração PWA
- Configurar o `manifest.json` adequadamente com cores de tema do sistema (Dark Mode integration via `theme-color` meta tag).
- Registrar um Service Worker básico para cache de assets estáticos, garantindo carregamentos quase instantâneos na Dashboard após o primeiro acesso.

## Resumo das Fases de Execução
1. **Fase 1:** Atualização das meta tags globais, index.css (Safe areas e touch-targets), e substituição da Navbar/Sidebar por Bottom Navigation no Mobile.
2. **Fase 2:** Refatoração da Home da UI do Cliente (`Dashboard / Wallet / Library`) para usar Scroll Horizontal, Cards e Bottom Sheets.
3. **Fase 3:** Refatoração da Área de Admin (`Orders / Users / System`) usando Padrão de Stacked Cards para as tabelas.
4. **Fase 4:** PWA, Gestos Core (Swipe) e Refinamento de Animações.

# Plano de Implementação (V3) - O Motor de Vendas Mobile

Este roteiro estratégico descreve as intervenções no código e UX para implementar a arquitetura de Vendas Diretas e Experiência Mobile Nativa mapeada em `design.md`.

---

## Fase 1: Fundação Mobile Responsiva e Desempenho
1. **Auditoria `App.css` e Ajustes de Layout**
   - Refinamento agressivo dos media queries `< 768px`.
   - Garantir que a dobra principal (Hero) carregue a CTA (call to action) sem precisar scrollar no mobile (escondendo assets gigantes via classe `.desktop-only`).
   - Adaptação das fontes e gaps globais (reduzir margens soltas em telas pequenas para economizar espaço).

2. **Criação do Sistema "Native Feel" (Componentes Estruturais)**
   - Construir `<MobileDrawer>`: Substituir os Dropdowns/Alertas por "Swipe-up Sheets" fluídas para quem acessa pelo celular.
   - Construir `<Skeleton>`: Framework unificado e leve de carregamento pre-bound (ocupando o height exato no celular p/ evitar layout shift).

## Fase 2: O Carrinho Vendedor e Interação de Compra
1. **O Novo Carrinho "Swipe/Drawer"**
   - Eliminar o botãozinho flutuante obsoleto do grid. Criar um painel expansível ou Bottom Sheet com resumo claro da compra, total de Fatura e botão massivo de Checkout.
   - Micro-interações instantâneas de "Adicionado" sem travar o usuário (remoção do alerta de bloqueio por toast suave no topo do celular).

2. **Navegação Cirúrgica**
   - Transformar as Pílulas do `<StealthRouter>` em carrossel horizontal contínuo no mobile para poupar height, permitindo deslizar o dedo horizontalmente entre os Game Ecosystems sem quebrar as colunas em 100% _width_ e perdendo foco.

## Fase 3: Reformulação Dashboard / Área do Cliente
1. **Dashboard (Layout Assimétrico / Bento Box Focado)**
   - Substituição da página fria atual de histórico por um dashboard útil: 
     - 1 Banner com atalho de Ação (Comprar fundo, Último arquivo pendente).
     - Cards condensados com totalizadores rápidos em cima, histórico compactado em baixo.

## Fase 4: O Momento do Dinheiro (Pix & Checkout Sem Dor)
1. **Recorte Responsivo do Pix**
   - Centralizar o Flow de Pix, encolhendo os passos desnecessários. QR Code focado, botão de "Copiar Inteligente" que pega a tela toda de baixo pra cima em Smartphones, acendendo verde instantâneo com um pequeno tremor tátil se tiver lib de haptics disponiveis via WebAPI (`navigator.vibrate`).
2. **Confirmação Visual Exuberante (Apenas Pós-Conversão)**
   - Deixar a chuva de confetes (`canvas-confetti`) restrito unicamente à fração de segundo em que o Pagamento WebSocket/Pooling der `COMPLETED`, coroando a jornada nativa com estilo extremo.

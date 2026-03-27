# Checklist de Execução V3 - Motor de Vendas Mobile
Rastreador de progresso para a fase focada em Conversões, Experiência Nativa no Celular e Clareza de Catálogo.

## 1. Responsividade e Performance Mobile (First-Hit)
- [x] Ocultar Assets Gulosos na Primeira Dobra Mobile (Desktop-only na Imagem 3D do Hero).
- [x] Alterar Cobertura / Copywriting da Home para "Oferta Direta & Monetização".
- [x] Implementar a "Morte da Tipografia": Adicionar Teko/Bebas Neue p/ Títulos e Fira Code p/ Dados Numéricos (Monospace).
- [x] Auditar CSS e Forçar o Fix do `clamp()` globalmente para evitar vazar a largura da tela Mobile (Bug dos 80% da tela quebrada).

## 2. A Fundação Brutalista e Texturas
- [x] Aplicar Overlay de Ruído (CSS Grain / Noise pseudo-element `::before`) global fixo no site.
- [x] Inverter Border-radius de tudo: Converter borders redondas de cards e produtos para quadrados 0px ou recortes `clip-path` chanfrados.
- [x] Sobreposições: Alterar os grids da Home para fazer elementos invadirem colunas e sobreporem imagens.

## 3. O Novo Funil da Home (Venda Direta Agressiva)
- [x] Eliminar "padding/buraco negro" abusivo pós-Hero.
- [x] Substituir o `StealthRouter` por um `<HomeBentoCategories>`: Grid assimétrico brutalista de navegação.
- [x] Ajustar proporções do `<HomeBentoCategories>`: Colocar "Plugins e Mapas" e "Ragnarok" lado a lado exato (`col-span-6`).
- [x] Garantir que o clique nos itens do Express Showcase direcione p/ Checkout de Convidado (ou Login) ou filtre a Loja corretamente para todos os blocos do Bento.
- [x] Criar `<HomeExpressShowcase>`: Grid seco de 4 a 8 "Mais Vendidos".
- [x] Rebaixar a `<PricingSection>` para o fim do funil. 

## 4. O Funil Final e UX
- [x] Construir a Dashboard do Usuário usando estrutura hibrida Asimétrica para facilitar leitura (Bento-style Cards de Vendas/Uso).
- [x] Limpar os Feedbacks "Toasts" transformando a pilha de notificações no padrão de apps modernos (Pequenas faixas superiores ou estilo Sonner inferior).
- [x] O Estado de Sucesso Pix: Renderizar feedback instantâneo e tátil + mini disparo visual (Confetti) no webhook success. Embelezar e compactar render do QR-Code.

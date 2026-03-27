# Pixelcraft Studio - Cyber Brutalism & Vendas (V3)

## Visão Geral
O objetivo absoluto da V3 é transformar o site em uma máquina de vendas, mas com uma Estética que destrói o modelo genérico atual (limpo e esterilizado) e assume um **Sci-Fi Cyber Brutalism**. 
Nós vendemos Engenharia, Infraestrutura e Servidores: o site tem que parecer um Terminal Aberto, irregular, corajoso, com uma experiência Mobile inquebrável.

## Pilares do Design TOP

### 1. Quebra de Simetria (Overlaps & Asymmetry)
O site perde o aspecto "caixinha da Apple".
- **Bento Grid Desconstruído**: Cards de proporções drásticas lutando por atenção. 1 Card gigante cortando o grid principal, e minis-cards espremidos do lado.
- **Sobreposições Livres**: Imagens invadindo espaços de texto. Tags saltando para fora do card propositalmente. Sem obediência cega ao padding perfeito.

### 2. A Morte da Tipografia Genérica
Chega de Inter para everything.
- **Títulos Display Agressivos**: Fontes grossas e condensadas (Teko, Bebas Neue, Clash Display) para dominar a tela.
- **Micro-dados em Monospace**: Elementos técnicos (IDs de servidor, preços, slots, ram) devem usar fontes Monospaced de IDE (JetBrains Mono, Fira Code) passando a sensação de console/terminal para nosso público tech.

### 3. Sci-Fi Borders & CRT Brutalism
- **Bordas Afiadas (Sharp Edges)**: Eliminar os "cantinhos arredondados fofinhos". Aplicar bordas duras `0px` ou Chanfros (Chamfered CSS paths) simulando peças industriais cortadas.
- **Textura e Noise**: Sobrepor um leve filtro de estática (grain effect overlay) por todo o site, matando o aspecto digital puro e parecendo tela de hardware real operando.
- **Grid Lines Infinitas**: Background construído com linhas de grades arquitetônicas discretas.

### 4. Responsividade "Anti-Quebra" (Mobile Survival)
- Títulos elásticos (`clamp`) com tamanhos mínimos muito baixos (2rem) para garantir que palavras gigantes ("INFRAESTRUTURA") nunca vazem o celular a 80% do viewport, envelopados na classe brutalista. Em mobile o layout sobreposto ganha prioridade vertical estrita sem margins mortas.
## Arquitetura Agressiva da Home (O Funil Brutalista)
O fluxo da página principal deve refletir quem vende o produto rápido e fácil antes de empurrar assinaturas caras. Ticket Baixo em Cima, Ticket Alto Embaixo.

### 1. Hero Section (A Isca)
Copy agressiva ("TUDO O QUE VOCÊ PRECISA PARA MONETIZAR"), números técnicos visíveis (Monospace) e imagem 3D. 100% focado no soco visual sem scroll inútil.

### 2. O Roteador Brutalista (O "Vendemos")
Fim dos espaços vazios gigantes (padding bizarro). Logo abaixo do Hero, entra um **Bento Grid Assimétrico** com blocos escuros e tipografia Display gigante apontando para as categorias de conversão rápida (Ex: [SERVIDORES PRONTOS], [SCRIPTS FIVEM], [PLUGINS & MAPAS]). Nada de visual "lojinha arrumada", é um arsenal.

### 3. A Vitrine Expressa (O Gatilho de Compra)
Abaixo do Bento Grid, uma grade brutal e simples (4 a 8 itens) com os 'Mais Vendidos'. Imagem com overlay escuro, Título chamativo e botão de "Adicionar", puxando a conversão pela emoção sem precisar ir para a página da loja explorar.

### 4. A Engenharia Premium (O "Fazemos")
As assinaturas / Planos Caros (R$ 150 - R$ 300) descem para o fim da página. 
Mudança na Narrativa: "Não quer ter trabalho? Nós construímos para você." Foco exclusivo em quem quer terceirizar o problema e tem dinheiro.

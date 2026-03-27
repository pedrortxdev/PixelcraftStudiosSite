# Design & UX Specification - Pixelcraft V4

Este documento detalha a arquitetura visual, tokens de design e comportamento de componentes para a atualização V4. Ele serve como a única fonte de verdade para a implementação da interface.

## 1. Sistema de Tipografia e Identidade Visual

### 1.1 Nova Fonte Padrão
Substituiremos a fonte atual (Square-ish) por uma família sans-serif de alta legibilidade. - Nota: nao encostar na fonte minecraft.
- **Fonte Primária:** `Geist Sans` (Vercel) ou `Inter`.
- **Fonte Monospace:** `Geist Mono` (para dados financeiros e IDs).
- **Fonte Legada (Opcional):** A fonte Minecraft atual será mantida como "Classic Pixel", ativável via preferências.

### 1.2 Tokens de Cores (Refinamento)
```css
:root {
  --bg-primary: #0A0E1A;
  --bg-secondary: #0F1219;
  --bg-card: rgba(15, 18, 25, 0.6);
  --accent-primary: #583AFF; /* Indigo */
  --accent-secondary: #1AD2FF; /* Cyan */
  --accent-red: #E01A4F; /* Crimson */
  --text-primary: #F8F9FA;
  --text-secondary: #B8BDC7;
  --text-muted: #6C727F;
  --border-subtle: rgba(255, 255, 255, 0.08);
  --border-active: rgba(88, 58, 255, 0.4);
}
```

---

## 2. Refatoração do Dashboard: Densidade e Minimalismo

O usuário poderá alternar entre os modos **Confortável** e **Minimalista**. Essa preferência será persistida no banco de dados e aplicada via atributo `data-density` no `<body>`.

### 2.1 Cards de Estatísticas (Admin & Client) Nota: essa parte deve ser focada principalmente no /dashboard do cliente, alem o que mais foi pedido, claro

#### Modo Confortável (Atual)
- **Padding:** `2rem`
- **Ícone:** `48px` com background circular/quadrado suave.
- **Título:** `var(--title-h3)` (Bold).
- **Gap:** `1.5rem`.

#### Modo Minimalista (Novo)
- **Padding:** `1.25rem`.
- **Ícone:** `24px` sem background, apenas a cor do ícone.
- **Título:** `1.1rem` (Semi-bold).
- **Gap:** `0.75rem`.
- **Borda:** `0.5px solid var(--border-subtle)`.
- **Efeito:** Remoção de sombras pesadas; uso de um `glassmorphism` mais sutil.

### 2.2 Seção de Preferências do Usuário
Localizada em `/configuracoes`, dentro de uma nova aba "Interface".
- **Toggle de Densidade:** Botões segmentados [Minimalista | Confortável].
- **Toggle de Fonte:** Select [Moderna (Geist) | Pixel (Classic)].
- **Preview em Tempo Real:** As mudanças devem ser refletidas no componente de preview ao lado do formulário antes de salvar.
- **Toggle de Background:** Atualmente o background tem um "filtro" que deixa meio pixelado/fosco, o usuario deve poder ativar isso ou nao

---

## 3. Loja Pública (`/loja`) - Estrutura "Netflix Style"

A página `/shop` pública deve ser otimizada para conversão, apresentando o catálogo de forma visualmente rica, mas limitada.

### 3.1 Comportamento das Prateleiras (Rows)
Cada jogo (ex: Minecraft, GTA V) terá sua própria linha horizontal.
- **Título da Linha:** Nome do Jogo + Ícone.
- **Limite de Itens:** Máximo de 2 produtos por jogo exibidos na home da loja.
- **Card de Produto:** Versão compacta do `ProductCard` atual.
  - Imagem de destaque com overlay gradiente.
  - Preço em destaque com fonte Mono.
  - Badges de categoria (ex: Plugin, Script).

### 3.2 O Botão "Ver Mais"
No final de cada linha horizontal ou como um card especial:
- **Design:** Card com bordas tracejadas e ícone de `ArrowRight`.
- **Ação:** Redireciona para `/register`.
- **Contexto:** Deve haver um tooltip ou texto abaixo: "Crie uma conta para visualizar todos os +50 produtos desta categoria".

- A loja real da area do cliente é o /loja, essa é o /shop
---

## 4. Admin: Histórico de Transações e Fluxo Financeiro

### 4.1 Novo Widget de Transações
Substitui "Produtos Mais Vendidos" no dashboard admin.
- **Colunas:**
  - `Usuário`: Nome + Avatar (clicável para `/admin/users/:id`).
  - `Valor`: Formato `R$ XX,XX` (cor verde se aprovado, cinza se pendente).
  - `Método`: Pix, Saldo, etc.
  - `Status`: Badge estilizado [PAGO | PENDENTE | CANCELADO].
- **Lógica de Dados:** Deve listar as últimas 10 transações reais de entrada de saldo (faturas pagas).

### 4.2 Modal Flutuante de Itens (Pedido)
Quando um pedido tem "Multiple Items":
- **Trigger:** Link azul sutil "Múltiplos Itens (+X)".
- **Comportamento:** Janela flutuante (`position: absolute` ou `fixed`) com `z-index: 9999`.
- **Interatividade:**
  - Drag-and-drop usando `framer-motion` (prop `drag`).
  - Lista de itens com ícone, nome e subtotal.
  - Botão de fechar (X) no canto superior direito.
  - Blur de fundo leve para focar no modal.

---

## 5. Admin: Detalhes do Usuário e Permissões de CPF

### 5.1 Visibilidade de Dados Sensíveis
- **Permissão:** `view_cpf`.
- **Comportamento UI:**
  - Se `false`: O campo CPF exibe `***.***.***-**` e o botão de copiar é desabilitado.
  - Se `true`: Exibe o CPF real e permite edição.
- **Localização:** `src/pages/admin/UserDetail.jsx`.

---

## 6. Padronização de Rotas (Slugs)

| Antigo Path | Novo Path | Descrição |
| :--- | :--- | :--- |
| `/shop` |  | Nova roda para nova pagina publica |
| `/dashboard/support` | `/suporte` | Central de ajuda do cliente |
| `/dashboard/downloads` | `/downloads` | Área de arquivos comprados |
| `/dashboard/projects` | `/projetos` | Gestão de assinaturas ativas |
| `/dashboard/settings` | `/configuracoes` | Perfil e Preferências |
| `/dashboard/history` | `/history` | Logs de compras e transações |
| `/wallet` | `/carteira` | Gestão de saldo e depósitos |

Vale lembrar que essa alteração é para a pagina principal.

---

## 7. Sistema de Downloads e Extensões

### 7.1 Correção de MIME Types
O backend deve garantir que o header `Content-Disposition` inclua o nome do arquivo original com extensão.
- **Exemplo:** `attachment; filename="plugin_v1.jar"`.
- **Frontend:** O link de download deve usar o atributo `download` do HTML5 se possível, ou confiar no stream do backend.

---

## 8. Tratamento de Erros e Estados Globais

### 8.1 Erros 500 em Assinaturas
- **Causa Provável:** Falta de tratamento para o estado `Entrega` no enum de status do banco ou na lógica de transição do Go.
- **UX:** Implementar um `Toast` de erro detalhado para administradores, informando se o erro foi de DB ou de validação de regra de negócio.

### 8.2 Filtros de Suporte
- **Comportamento:** Os filtros de Status e Prioridade devem ser cumulativos (AND logic).
- **UI:** Dropdowns múltiplos que atualizam a URL via `SearchParams` para permitir compartilhamento do link filtrado.

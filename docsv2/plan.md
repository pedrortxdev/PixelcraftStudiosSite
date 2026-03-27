# 🔍 Plano de Investigação de Bugs & Inconsistências — Frontend Pixelcraft

## Objetivo

Investigação profunda e sistemática de **todos os bugs e inconsistências** no frontend do site Pixelcraft Studio (React + Vite + Tailwind). O foco é encontrar problemas de funcionamento, inconsistências visuais entre páginas, padrões divergentes, funcionalidades que parecem funcionar mas não funcionam, e tudo que prejudica a qualidade da experiência do usuário.

## Escopo

- **APENAS FRONTEND** — Sem alterações ou investigação de backend.
- Todas as 13 páginas do usuário, 9 páginas admin, 14+ componentes, 2 contextos, 1 serviço API, 2 arquivos CSS, 3 hooks, 2 constantes, 1 utils, 1 layout.

## Metodologia

### 1. Análise Estática (Código Fonte)
- Ler e auditar **cada arquivo** do diretório `src/`
- Comparar estilos, padrões e convenções entre páginas
- Identificar código duplicado, inconsistente ou morto
- Verificar tratamento de erros, loading e estados vazios

### 2. Categorias de Bugs Investigadas

| # | Categoria | Descrição |
|---|-----------|-----------|
| 1 | **Arquitetura** | Padrões misturados, código duplicado, acoplamento |
| 2 | **Navegação** | Links mortos, âncoras inexistentes, rotas faltantes |
| 3 | **Estilos** | Inconsistências visuais entre páginas A e B |
| 4 | **API/Dados** | Duplicação de endpoints, erros silenciosos, dados não tratados |
| 5 | **UX/Interação** | Botões não-funcionais, falta de feedback, estados perdidos |
| 6 | **Responsividade** | Layout quebrado em mobile, elementos escondidos |
| 7 | **Acessibilidade** | Falta de labels, keyboard nav, contraste |
| 8 | **Performance** | Re-renders desnecessários, memory leaks, importações pesadas |
| 9 | **Segurança** | Tokens expostos, validações faltantes, XSS |
| 10 | **Funcionalidade** | Coisas que parecem funcionar mas não funcionam |

### 3. Arquivos de Saída

| Arquivo | Conteúdo |
|---------|----------|
| `docsv2/bugs.md` | Lista completa e categorizada de todos os bugs e inconsistências |
| `docsv2/inconsistencias_visuais.md` | Inconsistências visuais específicas entre páginas |
| `docsv2/tasks.md` | Checklist de tarefas para correção |
| `docsv2/plan.md` | Este documento de planejamento |

## Arquivos Investigados

### Páginas do Usuário (13)
- `App.jsx` (Home/Landing)
- `Login.jsx`
- `Register.jsx`
- `Dashboard.jsx`
- `Shop.jsx`
- `ProductDetails.jsx`
- `Checkout.jsx`
- `MyProjects.jsx`
- `Downloads.jsx`
- `History.jsx`
- `Wallet.jsx`
- `Billing.jsx`
- `Settings.jsx`
- `Support.jsx`

### Páginas Admin (9)
- `admin/Dashboard.jsx`
- `admin/Orders.jsx`
- `admin/AdminCatalog.jsx`
- `admin/AdminFiles.jsx`
- `admin/Users.jsx`
- `admin/UserDetail.jsx`
- `admin/AdminSupport.jsx`
- `admin/AdminRoles.jsx`
- `admin/SubscriptionDetail.jsx`

### Componentes (14+)
- `DashboardLayout.jsx`
- `AdminLayout.jsx`
- `ProtectedRoute.jsx`
- `AdminRoute.jsx`
- `Footer.jsx`
- `HeroIllustration.jsx`
- `PricingSection.jsx`
- `ProductsSection.jsx`
- `PartnersSection.jsx`
- `RoleBadge.jsx`
- `ErrorBoundary.jsx`
- `SubscriptionChat.jsx`
- `shop/ProductCard.jsx`
- `shop/FloatingCart.jsx`
- `dashboard/*` (5 componentes)

### Core
- `main.jsx` (Rotas)
- `context/AuthContext.jsx`
- `context/CartContext.jsx`
- `services/api.js`
- `App.css`
- `index.css`
- `hooks/*` (3 hooks)
- `constants/*` (2 constantes)
- `utils/roleUtils.js`

## Status

✅ **Investigação Completa** — Todos os arquivos foram lidos e analisados.

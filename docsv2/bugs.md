# Relatório de Bugs do Frontend (docsv2/bugs.md)

Este documento centraliza os bugs, débitos técnicos e potenciais problemas de segurança encontrados na investigação do frontend.

## 1. Roteamento (main.jsx)
- **Rotas Duplicadas**: Existem definições de rotas duplicadas para `/dashboard`, `/shop`, `/checkout`, `/history`, `/downloads` e `/support` dentro do `<Route path="/" element={<AppLayout />}>`. Isso pode causar renderizações inesperadas ou problemas de performance caso o React Router monte múltiplas instâncias.
- **Redirecionamento de Auth**: Usuários logados tentando acessar `/login` deveriam ser redirecionados para `/dashboard` mas as regras de redirecionamento preventivo nem sempre estão claras no nível do roteador.

## 2. Débitos de API e Hardcoded URLs
- **URL Chumbada (AdminFiles.jsx)**: A função `handleDownload` faz um bypass completo do ambiente e utiliza uma URL fixa `https://api.pixelcraft-studio.store/api/v1/files/${fileId}/download`. Isso quebra o app em ambientes de desenvolvimento ou staging.
- **Uso do Fetch Direto (AdminSupport.jsx)**: Ao invés de usar a instância padronizada do axios/fetch em `services/api.js`, esta página realiza `fetch()` direto com construção manual de headers (`Authorization: Bearer ${token}`). Isso duplica a lógica de auth, ignora interceptores (como redirect no 401) e suja o componente.
- **Construção de Avatar (Múltiplos Arquivos)**: Em `Settings.jsx`, `Users.jsx`, `UserDetail.jsx` e `AdminSupport.jsx`, existe uma lógica repetida e frágil: `import.meta.env.VITE_API_URL?.replace('/api/v1', '') || 'https://api.pixelcraft-studio.store'`. Se a rota da API mudar, todos os avatares quebram. Deve ser abstraído para um utilitário.
- **Upload de Arquivos (api.js)**: Funções como `filesAPI.upload` evitam a função genérica `apiRequest` porque usam `FormData`.

## 3. Segurança e Estado
- **Armazenamento de Token**: O JWT é salvo no `localStorage`, o que é uma vulnerabilidade clássica de XSS. A verificação de expiração também é fundamentalmente feita no client-side em algumas áreas.
- **Race Condition em Pix (Wallet.jsx)**: A verificação de pagamento Pix usa um polling (`setInterval`) de 5 segundos que compara se `userResp.balance > initialBalance`. Se o usuário receber fundos de outra origem (ou houver um atraso na captura do `initialBalance`), o sistema pode interpretar erroneamente o Pix como pago.
- **Race Conditions de Contextos**: `CartContext` usa debounce no salvamento no `localStorage`. Se o usuário navegar rápido ou fechar a aba logo após adicionar algo no carrinho, a alteração pode não ser salva.

## 4. WebSockets
- **Conexão de Suporte (AdminSupport.jsx)**: Possui uma dependência forte com substituição de string `API_URL.replace(/^http/, 'ws')` para conectar ao WebSocket. Isso é frágil e pode construir URLs inválidas dependendo da configuração do proxy reverso.

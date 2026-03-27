# Inconsistências Visuais e de UI (docsv2/inconsistencias_visuais.md)

Este documento lista problemas e oportunidades de melhoria na experiência do usuário e consistência visual na aplicação frontend.

## 1. Estilos Mistos e Inline
- O projeto usa uma mistura severa de Tailwind CSS (classes utilitárias), CSS regular (`App.css`, `index.css`), e muitos objetos de estilos inline nos componentes (ex: `App.jsx`, `AdminDashboard.jsx`, etc). Isso torna a manutenção caótica e impede um tema verdadeiramente unificado.
- Exemplo: `AdminCatalog.jsx` e `AdminFiles.jsx` têm enormes atributos de estilo (`const styles = { ... }`) com mais de 100 linhas que deveriam ser extraídos para CSS Modules ou classes do Tailwind.

## 2. Estados de Carregamento (Loading States)
- Vários arquivos têm estados de loading duplicados (como `<div className="flex justify-center items-center h-screen">...</div>`), como visto em `Login.jsx`, `Register.jsx`, `Dashboard.jsx`, e `Shop.jsx`.
- Há também trechos em componentes de Admin que retornam `Carregando detalhes do usuário...` como texto solto, que quebram a consistência elegante da UI que usa framer-motion em outros lugares. É necessário criar um componente global `<PageLoader />`.

## 3. Tipografia Dispersa
- A fonte padrão do projeto deveria ser controlada via `index.css` na tag `body` ou através da configuração de tema do Tailwind (font-family global). 
- No entanto, propriedades como `fontFamily: "'Inter', sans-serif"` são repetidas inline em vários containers de alto nível, por exemplo, no `AdminFiles.jsx`.

## 4. Sombras e Efeitos de Glassmorphism Inconsistentes
- A aplicação faz uso acentuado de `backdrop-filter: blur(10px)` ou `blur(20px)`. Contudo, as cores de fundo RGBA não seguem padronização.
- Encontradas variações arbitrárias nos "Cards" como: `background: 'rgba(21, 26, 38, 0.6)'`, `background: 'rgba(15, 18, 25, 0.6)'` e `background: 'rgba(15, 20, 35, 0.8)'` ao redor de `Orders.jsx`, `AdminDashboard.jsx`, e `AdminSupport.jsx`. Isso gera incompatibilidade de "tons de preto/azul transcúcido" ao navegar pelo painel administrativo.

## 5. Código Comentado e Lixo Residual
- Componentes chave como `Shop.jsx`, `ProductDetails.jsx`, e `Checkout.jsx` contêm blocos extensos de código HTML/JSX comentado e `console.log` residuais de debugging, que poluem a base de código do produto final.

## 6. Feedback de Usuário
- Em `Downloads.jsx`, após um download, a função utiliza alertas nativos ou toasts simples, que poderiam não estar 100% integrados à estética principal da plataforma. Melhorar o design do feedabck usando o `ToastContext` existente consistentemente.

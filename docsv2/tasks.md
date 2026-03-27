# Check-list Fila de Tarefas Exaustiva (docsv2/tasks.md)

Este apanhado documenta o RoadMap exato de implementação cobrindo relatórios prévios (Fase 1) e as arquiteturas novas de design rigoroso de protocolos solicitadas na revisão arquitetural (Fase 2). Este é o documento dinâmico em que os DEV's se guiarão com a vida de tracking.

---

## Fase 1: Débitos Técnicos Críticos e Cleanup Geral (Code Smells & Basics)

**[ Roteamento e Setup Base ]**
- [x] Limpar definições ambíguas/duplicadas em `src/main.jsx`. Evitar rotas como `/shop` acopladas repetidas vezes no escopo AppLayout.
- [x] Aplicar redirect programático protetor para o Router em `/login` via Hooks quando logged in.

**[ Refatorações Estruturais de Variáveis & Endpoints Globais ]**
- [x] Abstrair url base `import.meta.env.VITE_API_URL`. Nunca gerar referências cegas a "https://api.pixelcraft...".
- [x] Alterar o fetch manual sujo de `AdminSupport.jsx` refatorando-o sob chamadas engessadas e protegidas via headers no wrapper comum contido em `src/services/api.js`.
- [x] Centralização de Lógica Estética Avatares: Construir helper global `src/utils/formatAvatarUrl.js` abolindo a replicação constante feita ao longo de toda a hierarquia de `Pages/Admin`.
- [x] Eliminar dependências/loops de cache infinito ou duplas chamadas causadas pelo hook `useRoles.js`.

**[ Consistência Visual Uniforme & Qualidade UI ]**
- [x] Consolidar `.glass-card-element` no index styles unificando `backdrop-filter` transparentes da UI pra prevenir paletas discordantes em Orders/Dashboards/Support.
- [x] Varredura exaustiva removendo strings literárias inline de tipografia CSS (ex: Inter, Helvetica), forçando global no Tailwind default-theme block in config.
- [x] Substituir o `<div class="loading">Carregando detalhes...</div>` literal para o component visual spinner global interativo, aprimorando transições frame-motion.
- [x] Lixar o código solto (console logs esquecidos, lixo de marcação HTML comentado) que restaram em paginas de client-facing como Shop e Products.

---

## Fase 2: Implantação de Novos Fluxos de Negócios e Criptografia (Architecture Update)

**[ A. Fluxo Duplo 2FA: Redefinição de Senha (Reset Pass Code + Link) ]**
- [x] Configuração de Modal UI no `Login.jsx`: Modificar o modal para conter apenas 1 step (Aviso "Instruções Foram para E-mail Envolvendo seu Código Pessoal").
- [x] Criação `src/pages/ResetPassword.jsx`. Template de form completo com validação custom (Regex password length) blindada.
- [x] Instalar interceptador de Token Parameter URL via React-Router parameters extraction.
- [x] Expandir o SDK interno `src/services/api.js` contendo o trigger de payload `{token, validation_code, password}` via post object.

**[ B. Segurança de Manipulação de Anexos Stream / Binary Pipeline ]**
- [x] Criação de Regex util em arquivo apart (regex decoder) testando strings formatadas RFC 5987 / e aspas clássicas pra Content Dispo.
- [x] Ligar Helper às Pages de Baixar arquivos (`Downloads.jsx` do User Client side e `AdminFiles.jsx` do Administrative View).
- [x] Ligar force flag em blob click-event listener de JS document com nome padronizado.

**[ C. Polidesamento de Permissões ABAC Model + Masking ]**
- [x] No painel administrativo em `AdminRoles.jsx`: Atualize o dropdown/visual arrays component para aceitar inputs das checkboxes nomeadas de `view_cpf`, diferenciando-a com tags label UI da permissão crua.
- [x] Teste lógico em `UserDetail.jsx` envolta por conditional wrap: Condicionar preloading, labels and edits text masks do Field CPF, a basear-se no token hook permissão ativa.
- [x] `Settings.jsx` (Lado do Cliente final): Polimento do forms do cliente, habilitando formatters customizados para CPF BR display friendly na view stateful React, sempre mostrando info ao verdadeiro dono.

**[ D. Pix Flow Engine - Assíncronia em Pagamentos ]**
- [x] No `Wallet.jsx`: Romper infraestrutura atual de "Check Current Profile Total Balance Every 5 Secs" (polling leigo).
- [x] Injetar mecanismo the `setInterval` direcionado unicamente ao Status Resource Check Endpoit: Refatoração que aguarda objeto `{status: COMPLETED}` em polling específico do id transacionado pela `creationMethod()`.
- [x] Disparo UX aprimorado e limpo da modal Success Modal Animated State. Remodelação em grid na tela `History` garantindo visual clear differentiation de fundos versus itens Library.

---

## Fase 3: UX/UI Redesign: Roteamento Stealth e Funil da Loja

**[ A. A Home: Roteamento "Stealth" Elegante ]**
- [x] Refatoração da Dobra Principal (Hero): Alterar textos de conversão e substituir 3D Genérico pela imagem principal requerida.
- [x] Criação do "Roteador": Remover os cards velhos coloridos (ProductsSection) e criar o componente de Pílulas Minimalistas.
- [x] Implementar as animações Stealth do hover color & background. Lincar as rotas diretas `/shop?game=[id]`.

**[ B. A Loja: O Fim da "Aba Todos" & Layout Netflix ]**
- [x] Exterminar aba e opção mista "Todos os Jogos".
- [x] Adicionar suporte a scroll horizontal isolado "Prateleiras" (Estilo Netflix) para visualização global.
- [x] Ao rotear com game pre-selecionado (`?game=fivem`), reestruturar os header pills com categorias estritas daquele ecossistema (Sem poluição).
- [x] Uniformizar as Miniaturas (Thumbnails): CSS Global aplicando *Vignette Overlay* escuro (10/15%) em todas as capas de produtos globais.

# Actionable Task List - Pixelcraft V4

## 1. Infraestrutura e Banco de Dados
- [ ] Criar migração SQL para adicionar `users.preferences`.
- [ ] Criar migração SQL para adicionar a permissão `view_cpf`.
- [ ] Executar script de limpeza e correção de enums para `ticket_status` e `subscription_stage`.
- [ ] Backup do banco de dados antes das alterações de estrutura.

## 2. Backend (Golang)
- [ ] **Suporte:** Corrigir a query dinâmica no `support_repository.go` para suportar filtros combinados.
- [ ] **Assinaturas:** Corrigir o erro 500 na mudança de status para `COMPLETED`. Verificar logs de erro e constraints de DB.
- [ ] **Transações:** Criar o endpoint `GET /api/v1/admin/transactions`.
- [ ] **Stats:** Atualizar o endpoint de estatísticas para calcular receita baseada em transações reais e formatar crescimento de usuários como string `+X`.
- [ ] **Arquivos:** Corrigir headers de download no `file_handler.go`. Garantir que `Content-Disposition` use o nome real com extensão.

## 3. Frontend - Rotas e Navegação
- [ ] Atualizar `App.jsx` com as novas rotas em português:
    - `/loja`, `/suporte`, `/downloads`, `/projetos`, `/configuracoes`, `/carteira`, `/history`. Nota: garanta a diferença: /loja é a loja real, /shop é a loja de vitrine sem login. (nota extra: /shop é uma pagina inexistente e deve ser feita como pagina publica no qual os cards já existentes na pagina principal são usados para levar aos produtos da /shop)
- [ ] Atualizar `Footer.jsx` com os novos links.
- [ ] Atualizar links no `DashboardLayout.jsx` e `Sidebar`.

## 4. Frontend - Loja Pública
- [ ] Criar a página `PublicShop.jsx` (ou renomear a atual e adicionar lógica condicional).
- [ ] Implementar as prateleiras horizontais com limite de 2 produtos por jogo.
- [ ] Adicionar o card/botão "Ver Mais" que leva para `/register`.

## 5. Frontend - Dashboard e Preferências
- [ ] Adicionar aba "Interface" em `Settings.jsx`.
- [ ] Criar UI para selecionar Densidade e Fonte.
- [ ] Atualizar `AuthContext` para carregar e salvar as preferências.
- [ ] Aplicar classes condicionais nos cards do Dashboard baseadas na densidade.
- [ ] Trocar fonte global do projeto e adicionar suporte a fontes dinâmicas via CSS Variables.

## 6. Frontend - Admin Dashboard
- [ ] Remover modal "Produtos Mais Vendidos".
- [ ] Criar componente de "Histórico de Transações" no dashboard admin.
- [ ] Implementar modal flutuante/arrastável para "Múltiplos Itens" nos pedidos recentes.
- [ ] Atualizar cartões de estatísticas (Receita Total e Usuários Totais).

## 7. Admin - Detalhes do Usuário
- [ ] Corrigir link quebrado de `/admin/subscription/:id` para `/admin/subscriptions/:id`.
- [ ] Implementar a lógica de permissão `view_cpf` para exibir/ocultar o campo.

## 8. Testes e Validação
- [ ] Validar fluxo completo de compra de saldo -> compra de produto.
- [ ] Validar fluxo de download de arquivo (testar .jar, .zip, .rar).
- [ ] Validar transição de status de assinatura por um administrador.
- [ ] Testar persistência de preferências em diferentes navegadores.
- [ ] Verificar se todos os erros 500 reportados foram eliminados.

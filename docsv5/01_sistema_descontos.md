# Sistema de Descontos Restritos

Este documento detalha as atualizações realizadas no sistema de cupons de desconto para permitir restrições granulares.

## 1. Alterações no Banco de Dados
Foi criada a migração `017_discount_restrictions.sql`:
- Adicionada coluna `restriction_type` (VARCHAR) com valor padrão 'ALL'.
- Adicionada coluna `target_ids` (UUID ARRAY) para armazenar os IDs de jogos, categorias ou produtos.

## 2. Backend (Go)
- **Modelos:** Atualizado `internal/models/discount.go` com os novos campos e tipos de restrição (`ALL`, `ITEM_CATEGORY`, `GAME`, `PRODUCT`).
- **Repositório:** `internal/repository/discount_repository.go` agora suporta CRUD completo para descontos e manipulação de arrays UUID do PostgreSQL.
- **Serviço de Checkout:** `internal/service/checkout_service.go` atualizado para validar as restrições:
    - O sistema identifica quais itens no carrinho são elegíveis para o cupom.
    - O desconto é calculado apenas sobre a soma dos valores dos itens elegíveis.
    - Se nenhum item no carrinho for elegível, o cupom é rejeitado.
- **Novos Endpoints:** Criado `internal/handlers/discount_handler.go` e registrado rotas administrativas em `cmd/api/main.go`:
    - `GET /api/v1/admin/discounts` - Listar todos.
    - `POST /api/v1/admin/discounts` - Criar novo.
    - `PUT /api/v1/admin/discounts/:id` - Atualizar.
    - `DELETE /api/v1/admin/discounts/:id` - Excluir.

## 3. Frontend (React)
- **API Service:** Adicionados métodos de gerenciamento de descontos ao `adminAPI`.
- **Nova Página Admin:** Criada `src/pages/admin/AdminDiscounts.jsx`:
    - Interface para listagem, criação e edição de cupons.
    - Seleção dinâmica de restrições (Tudo, Por Jogo, Por Categoria ou Itens Específicos).
    - Buscador integrado e filtros visuais.
- **Navegação:** Adicionado item "Descontos" na Sidebar do Admin e registrada nova rota no `main.jsx`.

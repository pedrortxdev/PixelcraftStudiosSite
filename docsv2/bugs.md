# Bugs Identificados (Novos)

## 1. Discrepância na Contagem de Tickets
- **Sintoma:** O sistema indica 5 tickets abertos, mas nenhum é visível na lista.
- **Análise Técnica:** 
    - O frontend em `AdminSupport.jsx` usa o endpoint `/api/v1/admin/support/stats` para obter a contagem.
    - No repositório Go (`support_repository.go`), a função `GetTicketStats` conta apenas tickets com status exatamente igual a `'OPEN'`.
    - No entanto, o sistema de tickets possui outros status ativos como `'WAITING_RESPONSE'` e `'IN_PROGRESS'`.
    - Se a lista de tickets estiver filtrada por padrão para mostrar apenas "Abertos" (OPEN), mas a contagem do dashboard somar outros status ou se houver um erro de sincronização entre a contagem e a query de listagem (que usa JOIN com users), tickets de usuários deletados ou com dados incompletos podem ser contados mas não listados.
- **Solução:** Normalizar as queries de contagem e listagem. Ajustar a contagem para refletir o que o admin realmente vê ou corrigir o filtro padrão.

## 2. Erro SQL na Foto de Perfil (Avatar)
- **Sintoma:** Erro SQL ao tentar atualizar a foto de perfil.
- **Análise Técnica:** 
    - O repositório `user_repository.go` monta a query de UPDATE dinamicamente.
    - Se a coluna `avatar_url` não estiver presente em todas as tabelas relacionadas ou se houver um erro de sintaxe na cláusula SET quando apenas o avatar é enviado.
    - Verificado que a migração `002_add_avatar_url.sql` adiciona a coluna, mas o erro pode ser um mismatch entre o `db:"avatar_url"` no modelo e o nome real na tabela se houver ambiguidades em JOINs.
- **Solução:** Revisar a query SQL gerada em `UpdateUser` e validar se o campo `avatar_url` está sendo mapeado corretamente.

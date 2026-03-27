-- Script para identificar e corrigir IDs de produtos inválidos no banco de dados

-- Primeiro, vamos identificar os IDs inválidos
-- Isso inclui IDs que não tem 36 caracteres (tamanho padrão de UUID) ou que não estão no formato correto

-- Identificar IDs inválidos na tabela user_purchases
SELECT 
    up.id as purchase_id,
    up.user_id,
    up.product_id as invalid_product_id,
    LENGTH(up.product_id::text) as id_length,
    up.purchased_at
FROM user_purchases up
WHERE 
    LENGTH(up.product_id::text) != 36  -- UUID deve ter 36 caracteres
    OR up.product_id::text !~ '^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$'  -- formato UUID
ORDER BY up.purchased_at DESC;

-- Identificar IDs inválidos na tabela library (caso esteja sendo usada)
SELECT 
    l.id as library_id,
    l.user_id,
    l.product_id as invalid_product_id,
    LENGTH(l.product_id::text) as id_length,
    l.purchased_at
FROM library l
WHERE 
    LENGTH(l.product_id::text) != 36
    OR l.product_id::text !~ '^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$'
ORDER BY l.purchased_at DESC;

-- Se encontrar IDs inválidos, podemos tomar algumas ações:
-- Opção 1: Excluir registros com IDs inválidos (se não forem importantes)
-- DELETE FROM user_purchases WHERE LENGTH(product_id::text) != 36 OR product_id::text !~ '^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$';

-- Opção 2: Atualizar com UUIDs válidos para produtos existentes (mais complexo, depende do contexto)
-- Isso seria feito se você soubesse qual produto deveria estar associado

-- Opção 3: Verificar se os IDs inválidos estão relacionados a produtos reais
-- Primeiro vamos ver quais são os IDs inválidos
SELECT DISTINCT product_id::text as invalid_id 
FROM user_purchases 
WHERE LENGTH(product_id::text) != 36 OR product_id::text !~ '^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$';

-- E agora verificar se esses IDs aparecem em algum outro lugar ou são completamente inválidos
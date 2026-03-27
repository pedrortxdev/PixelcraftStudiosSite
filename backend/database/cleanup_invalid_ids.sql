-- Migration para identificar e remover registros com IDs de produto inválidos

-- Criar backup das tabelas afetadas antes de qualquer modificação
CREATE TABLE user_purchases_backup AS SELECT * FROM user_purchases;

-- Verificar quais produtos têm IDs inválidos (não UUIDs válidos)
-- Isso pode ajudar a identificar a causa raiz do problema
SELECT 
    up.id as purchase_id,
    up.user_id,
    up.product_id as invalid_product_id,
    LENGTH(up.product_id::text) as id_length,
    up.purchased_at
FROM user_purchases up
WHERE 
    LENGTH(up.product_id::text) != 36  -- UUID deve ter 36 caracteres
    OR up.product_id::text !~ '^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$';  -- formato UUID

-- Contar quantos registros inválidos existem
SELECT COUNT(*) as invalid_records_count
FROM user_purchases
WHERE 
    LENGTH(product_id::text) != 36
    OR product_id::text !~ '^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$';

-- Remover registros com IDs de produto inválidos
-- ATENÇÃO: Esta operação removerá dados permanentemente
-- Comente esta linha se quiser apenas identificar os registros sem removê-los
DELETE FROM user_purchases 
WHERE 
    LENGTH(product_id::text) != 36
    OR product_id::text !~ '^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$';

-- Fazer o mesmo para a tabela library, se existir registros inválidos
CREATE TABLE library_backup AS SELECT * FROM library;

DELETE FROM library 
WHERE 
    LENGTH(product_id::text) != 36
    OR product_id::text !~ '^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$';

-- Após a limpeza, verificar se ainda existem registros inválidos
SELECT COUNT(*) as remaining_invalid_records
FROM user_purchases
WHERE 
    LENGTH(product_id::text) != 36
    OR product_id::text !~ '^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$';

-- Adicionar uma verificação adicional para garantir a integridade referencial
-- Verificar se todos os product_ids na tabela user_purchases existem na tabela products
DELETE FROM user_purchases 
WHERE NOT EXISTS (
    SELECT 1 FROM products p WHERE p.id = user_purchases.product_id
);

-- Adicionar uma constraint para impedir IDs inválidos no futuro (opcional)
-- Isso pode ser feito como uma trigger ou constraint de verificação
-- CREATE OR REPLACE FUNCTION validate_uuid_format()
-- RETURNS TRIGGER AS $$
-- BEGIN
--     IF LENGTH(NEW.product_id::text) != 36 OR NEW.product_id::text !~ '^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$' THEN
--         RAISE EXCEPTION 'Invalid UUID format for product_id: %', NEW.product_id;
--     END IF;
--     RETURN NEW;
-- END;
-- $$ LANGUAGE plpgsql;

-- CREATE TRIGGER validate_user_purchases_uuid
-- BEFORE INSERT OR UPDATE ON user_purchases
-- FOR EACH ROW EXECUTE FUNCTION validate_uuid_format();
-- ================================================================
-- Migration 010: Fix Permissions System
-- Fixes: CRÍTICO-02, CRÍTICO-03, INC-03, INC-04, INC-05
-- 2026-03-04
-- ================================================================

-- ----------------------------------------------------------------
-- 1. Adicionar 'SYSTEM' ao enum resource_type (CRÍTICO-03)
-- ----------------------------------------------------------------
DO $$ BEGIN
    ALTER TYPE resource_type ADD VALUE IF NOT EXISTS 'SYSTEM';
EXCEPTION WHEN others THEN
    -- SYSTEM já existe, ignorar
    NULL;
END $$;

-- ----------------------------------------------------------------
-- 2. Adicionar 'VIEW_CPF' ao enum action_type (MODERADO-01)
-- ----------------------------------------------------------------
DO $$ BEGIN
    ALTER TYPE action_type ADD VALUE IF NOT EXISTS 'VIEW_CPF';
EXCEPTION WHEN others THEN
    NULL;
END $$;

-- Aguardar que os novos valores de enum estejam disponíveis
-- (necessário em alguns contextos de transação)
COMMIT;
BEGIN;

-- ----------------------------------------------------------------
-- 3. Corrigir permissões do SUPPORT (CRÍTICO-02)
-- Remover MANAGE (que não deveria existir) e garantir apenas VIEW/CREATE/EDIT
-- ----------------------------------------------------------------

-- Remover MANAGE incorreto do SUPPORT no recurso SUPPORT
DELETE FROM role_permissions
WHERE role = 'SUPPORT' AND resource = 'SUPPORT' AND action = 'MANAGE';

-- Garantir permissões corretas para SUPPORT
INSERT INTO role_permissions (role, resource, action) VALUES
    ('SUPPORT', 'SUPPORT',   'VIEW'),
    ('SUPPORT', 'SUPPORT',   'CREATE'),
    ('SUPPORT', 'SUPPORT',   'EDIT'),
    ('SUPPORT', 'EMAILS',    'VIEW'),
    ('SUPPORT', 'EMAILS',    'CREATE'),
    ('SUPPORT', 'DASHBOARD', 'VIEW')
ON CONFLICT (role, resource, action) DO NOTHING;

-- ----------------------------------------------------------------
-- 4. Corrigir permissões do ADMIN — adicionar SETTINGS:VIEW (INC-03)
-- ----------------------------------------------------------------
INSERT INTO role_permissions (role, resource, action) VALUES
    ('ADMIN', 'SETTINGS', 'VIEW'),
    ('ADMIN', 'SYSTEM',   'VIEW')
ON CONFLICT (role, resource, action) DO NOTHING;

-- ----------------------------------------------------------------
-- 5. Corrigir DEVELOPMENT — adicionar SETTINGS:VIEW + SYSTEM (INC-05)
-- ----------------------------------------------------------------
INSERT INTO role_permissions (role, resource, action) VALUES
    ('DEVELOPMENT', 'SETTINGS', 'VIEW'),
    ('DEVELOPMENT', 'SYSTEM',   'VIEW'),
    ('DEVELOPMENT', 'SYSTEM',   'MANAGE')
ON CONFLICT (role, resource, action) DO NOTHING;

-- ----------------------------------------------------------------
-- 6. Corrigir ENGINEERING — SETTINGS:MANAGE + DASHBOARD:MANAGE + SYSTEM:MANAGE (INC-04)
-- ----------------------------------------------------------------

-- Remover VIEW de SETTINGS e DASHBOARD para ENGINEERING (substituir por MANAGE)
-- (mantemos MANAGE que já cobre VIEW)
INSERT INTO role_permissions (role, resource, action) VALUES
    ('ENGINEERING', 'SETTINGS',  'MANAGE'),
    ('ENGINEERING', 'DASHBOARD', 'MANAGE'),
    ('ENGINEERING', 'SYSTEM',    'MANAGE')
ON CONFLICT (role, resource, action) DO NOTHING;

-- Remover entradas duplicadas VIEW de SETTINGS/DASHBOARD para ENGINEERING (agora tem MANAGE)
DELETE FROM role_permissions
WHERE role = 'ENGINEERING' AND resource = 'SETTINGS' AND action = 'VIEW';
DELETE FROM role_permissions
WHERE role = 'ENGINEERING' AND resource = 'DASHBOARD' AND action = 'VIEW';

-- ----------------------------------------------------------------
-- 7. Garantir que DIRECTION tem tudo (incluindo SYSTEM) (CRÍTICO-02)
-- ----------------------------------------------------------------
INSERT INTO role_permissions (role, resource, action) VALUES
    ('DIRECTION', 'SYSTEM', 'MANAGE')
ON CONFLICT (role, resource, action) DO NOTHING;

-- ----------------------------------------------------------------
-- 8. Permissão VIEW_CPF para cargos que podem ver dados sensíveis
-- Apenas ENGINEERING e DIRECTION têm acesso ao CPF
-- ----------------------------------------------------------------
INSERT INTO role_permissions (role, resource, action) VALUES
    ('ENGINEERING', 'USERS', 'VIEW_CPF'),
    ('DIRECTION',   'USERS', 'VIEW_CPF')
ON CONFLICT (role, resource, action) DO NOTHING;

-- ----------------------------------------------------------------
-- 9. Remover herança incorreta causada por permissões erradas do SUPPORT
-- e reaplicar corretamente (CRÍTICO-05)
-- ----------------------------------------------------------------

-- Remover permissões herdadas que foram propagadas com base no estado errado
-- (apenas as adicionadas via inherit_permissions_from_role, flag is_inherited=TRUE)
DELETE FROM role_permissions
WHERE is_inherited = TRUE;

-- Reaplicar herança correta após correção das permissões base

-- ADMIN herda de SUPPORT (VIEW-only de suporte e emails, mas ADMIN já tem mais)
-- Não reaplicar herança automática para evitar confusão — ADMIN já tem suas permissões diretas

-- Garantir que DEVELOPMENT e acima NÃO herdem lixo do SUPPORT
-- As permissões corretas já estão inseridas acima diretamente

-- ----------------------------------------------------------------
-- 10. Log da migração
-- ----------------------------------------------------------------
DO $$
BEGIN
    RAISE NOTICE 'Migration 010: Fix Permissions System applied successfully at %', NOW();
END $$;

-- Verificação rápida
SELECT
    role,
    COUNT(*) AS total_permissions,
    STRING_AGG(resource || ':' || action, ', ' ORDER BY resource, action) AS permissions
FROM role_permissions
WHERE role IN ('SUPPORT', 'ADMIN', 'DEVELOPMENT', 'ENGINEERING', 'DIRECTION')
GROUP BY role
ORDER BY role;

-- Migration: Create Granular Permissions System
-- Description: Sistema de permissões granulares para controle de acesso baseado em recursos e ações

-- Enum para recursos do sistema
DO $$ BEGIN
    CREATE TYPE resource_type AS ENUM (
        'USERS',           -- Gerenciamento de usuários
        'ROLES',           -- Gerenciamento de cargos
        'PRODUCTS',        -- Catálogo de produtos
        'ORDERS',          -- Pedidos e assinaturas
        'TRANSACTIONS',    -- Transações financeiras
        'SUPPORT',         -- Sistema de suporte/tickets
        'EMAILS',          -- Gerenciamento de emails
        'FILES',           -- Gerenciamento de arquivos
        'GAMES',           -- Gerenciamento de jogos
        'CATEGORIES',      -- Categorias de jogos
        'PLANS',           -- Planos de assinatura
        'DASHBOARD',       -- Dashboard administrativo
        'SETTINGS'         -- Configurações do sistema
    );
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

-- Enum para ações permitidas
DO $$ BEGIN
    CREATE TYPE action_type AS ENUM (
        'VIEW',      -- Visualizar
        'CREATE',    -- Criar
        'EDIT',      -- Editar
        'DELETE',    -- Deletar
        'MANAGE'     -- Gerenciar (todas as ações)
    );
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

-- Tabela de permissões por cargo
CREATE TABLE IF NOT EXISTS role_permissions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    role role_type NOT NULL,
    resource resource_type NOT NULL,
    action action_type NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(role, resource, action)
);

-- Índices para performance
CREATE INDEX IF NOT EXISTS idx_role_permissions_role ON role_permissions(role);
CREATE INDEX IF NOT EXISTS idx_role_permissions_resource ON role_permissions(resource);

-- Tabela de histórico de emails enviados (para gerenciamento)
CREATE TABLE IF NOT EXISTS email_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    from_email VARCHAR(255) NOT NULL,
    to_email VARCHAR(255) NOT NULL,
    subject VARCHAR(500) NOT NULL,
    body TEXT NOT NULL,
    status VARCHAR(50) DEFAULT 'sent', -- sent, failed, bounced
    error_message TEXT,
    sent_by UUID REFERENCES users(id),
    sent_at TIMESTAMPTZ DEFAULT NOW(),
    message_id VARCHAR(255), -- AWS SES Message ID
    metadata JSONB -- Dados adicionais (tags, attachments, etc)
);

-- Índices para email_logs
CREATE INDEX IF NOT EXISTS idx_email_logs_from ON email_logs(from_email);
CREATE INDEX IF NOT EXISTS idx_email_logs_to ON email_logs(to_email);
CREATE INDEX IF NOT EXISTS idx_email_logs_sent_at ON email_logs(sent_at DESC);
CREATE INDEX IF NOT EXISTS idx_email_logs_sent_by ON email_logs(sent_by);
CREATE INDEX IF NOT EXISTS idx_email_logs_status ON email_logs(status);

-- Inserir permissões padrão para cada cargo

-- SUPPORT: Apenas suporte e email próprio
INSERT INTO role_permissions (role, resource, action) VALUES
    ('SUPPORT', 'SUPPORT', 'MANAGE'),
    ('SUPPORT', 'EMAILS', 'VIEW'),
    ('SUPPORT', 'EMAILS', 'CREATE')
ON CONFLICT DO NOTHING;

-- ADMIN: Visualização total, sem edição
INSERT INTO role_permissions (role, resource, action) VALUES
    ('ADMIN', 'USERS', 'VIEW'),
    ('ADMIN', 'PRODUCTS', 'VIEW'),
    ('ADMIN', 'ORDERS', 'VIEW'),
    ('ADMIN', 'TRANSACTIONS', 'VIEW'),
    ('ADMIN', 'SUPPORT', 'VIEW'),
    ('ADMIN', 'EMAILS', 'VIEW'),
    ('ADMIN', 'FILES', 'VIEW'),
    ('ADMIN', 'GAMES', 'VIEW'),
    ('ADMIN', 'CATEGORIES', 'VIEW'),
    ('ADMIN', 'PLANS', 'VIEW'),
    ('ADMIN', 'DASHBOARD', 'VIEW')
ON CONFLICT DO NOTHING;

-- DEVELOPMENT: Edita planos e produtos
INSERT INTO role_permissions (role, resource, action) VALUES
    ('DEVELOPMENT', 'USERS', 'VIEW'),
    ('DEVELOPMENT', 'PRODUCTS', 'MANAGE'),
    ('DEVELOPMENT', 'ORDERS', 'VIEW'),
    ('DEVELOPMENT', 'TRANSACTIONS', 'VIEW'),
    ('DEVELOPMENT', 'SUPPORT', 'VIEW'),
    ('DEVELOPMENT', 'EMAILS', 'VIEW'),
    ('DEVELOPMENT', 'FILES', 'MANAGE'),
    ('DEVELOPMENT', 'GAMES', 'MANAGE'),
    ('DEVELOPMENT', 'CATEGORIES', 'MANAGE'),
    ('DEVELOPMENT', 'PLANS', 'MANAGE'),
    ('DEVELOPMENT', 'DASHBOARD', 'VIEW')
ON CONFLICT DO NOTHING;

-- ENGINEERING: Emails, catálogo, pedidos, editar senhas/saldo
INSERT INTO role_permissions (role, resource, action) VALUES
    ('ENGINEERING', 'USERS', 'MANAGE'),
    ('ENGINEERING', 'PRODUCTS', 'MANAGE'),
    ('ENGINEERING', 'ORDERS', 'MANAGE'),
    ('ENGINEERING', 'TRANSACTIONS', 'MANAGE'),
    ('ENGINEERING', 'SUPPORT', 'MANAGE'),
    ('ENGINEERING', 'EMAILS', 'MANAGE'),
    ('ENGINEERING', 'FILES', 'MANAGE'),
    ('ENGINEERING', 'GAMES', 'MANAGE'),
    ('ENGINEERING', 'CATEGORIES', 'MANAGE'),
    ('ENGINEERING', 'PLANS', 'MANAGE'),
    ('ENGINEERING', 'DASHBOARD', 'VIEW'),
    ('ENGINEERING', 'SETTINGS', 'VIEW')
ON CONFLICT DO NOTHING;

-- DIRECTION: Acesso total
INSERT INTO role_permissions (role, resource, action) VALUES
    ('DIRECTION', 'USERS', 'MANAGE'),
    ('DIRECTION', 'ROLES', 'MANAGE'),
    ('DIRECTION', 'PRODUCTS', 'MANAGE'),
    ('DIRECTION', 'ORDERS', 'MANAGE'),
    ('DIRECTION', 'TRANSACTIONS', 'MANAGE'),
    ('DIRECTION', 'SUPPORT', 'MANAGE'),
    ('DIRECTION', 'EMAILS', 'MANAGE'),
    ('DIRECTION', 'FILES', 'MANAGE'),
    ('DIRECTION', 'GAMES', 'MANAGE'),
    ('DIRECTION', 'CATEGORIES', 'MANAGE'),
    ('DIRECTION', 'PLANS', 'MANAGE'),
    ('DIRECTION', 'DASHBOARD', 'MANAGE'),
    ('DIRECTION', 'SETTINGS', 'MANAGE')
ON CONFLICT DO NOTHING;

-- Comentários para documentação
COMMENT ON TABLE role_permissions IS 'Permissões granulares por cargo. Define o que cada cargo pode fazer em cada recurso.';
COMMENT ON TABLE email_logs IS 'Histórico de emails enviados pelo sistema através do AWS SES.';
COMMENT ON COLUMN email_logs.metadata IS 'Dados adicionais em JSON: tags, attachments, reply_to, etc.';


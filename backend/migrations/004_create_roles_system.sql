-- Migration: Create Roles System
-- Description: Replaces is_admin boolean with a comprehensive role-based permission system

-- Enum para tipos de cargo
DO $$ BEGIN
    CREATE TYPE role_type AS ENUM (
        'PARTNER',      -- Parceiro: +1% de lucros em vendas
        'CLIENT',       -- Cliente: prioridade 2 estrelas, adquirido com depósito
        'CLIENT_VIP',   -- Cliente VIP: prioridade 3 estrelas, R$200/mês ou assinatura
        'SUPPORT',      -- Suporte: acesso restrito admin (Atendimento + Email próprio)
        'ADMIN',        -- Administração: visualização total, sem edição
        'DEVELOPMENT',  -- Desenvolvimento: edita planos/produtos
        'ENGINEERING',  -- Engenharia: emails, catálogo, pedidos, editar senhas/saldo (cargos inferiores)
        'DIRECTION'     -- Direção: acesso total
    );
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

-- Tabela de cargos do usuário (muitos-para-muitos)
CREATE TABLE IF NOT EXISTS user_roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role role_type NOT NULL,
    granted_at TIMESTAMPTZ DEFAULT NOW(),
    granted_by UUID REFERENCES users(id),
    expires_at TIMESTAMPTZ, -- NULL = permanente, data = temporário (ex: VIP por assinatura)
    UNIQUE(user_id, role)
);

-- Índices para performance
CREATE INDEX IF NOT EXISTS idx_user_roles_user ON user_roles(user_id);
CREATE INDEX IF NOT EXISTS idx_user_roles_role ON user_roles(role);
CREATE INDEX IF NOT EXISTS idx_user_roles_expires ON user_roles(expires_at) WHERE expires_at IS NOT NULL;

-- Adicionar campos de tracking de gastos no users
ALTER TABLE users ADD COLUMN IF NOT EXISTS total_spent NUMERIC(10,2) DEFAULT 0;
ALTER TABLE users ADD COLUMN IF NOT EXISTS monthly_spent NUMERIC(10,2) DEFAULT 0;
ALTER TABLE users ADD COLUMN IF NOT EXISTS monthly_spent_reset_at TIMESTAMPTZ DEFAULT NOW();

-- Adicionar campo para email atribuído ao staff de suporte
ALTER TABLE users ADD COLUMN IF NOT EXISTS assigned_email VARCHAR(255);

-- Migrar is_admin existentes para cargo DIRECTION
INSERT INTO user_roles (user_id, role)
SELECT id, 'DIRECTION'::role_type FROM users WHERE is_admin = true
ON CONFLICT (user_id, role) DO NOTHING;

-- Comentários para documentação
COMMENT ON TABLE user_roles IS 'Associação de cargos aos usuários. Um usuário pode ter múltiplos cargos.';
COMMENT ON COLUMN user_roles.expires_at IS 'Data de expiração do cargo. NULL = permanente. Usado para VIP por assinatura.';
COMMENT ON COLUMN users.total_spent IS 'Total gasto pelo usuário em compras (não inclui depósitos).';
COMMENT ON COLUMN users.monthly_spent IS 'Gasto no mês atual. Resetado mensalmente.';
COMMENT ON COLUMN users.assigned_email IS 'Email atribuído ao staff de suporte para acesso.';

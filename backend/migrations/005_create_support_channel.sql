-- Migration: Create Support Channel System
-- Description: Support ticket system with priority based on user roles

-- Status do ticket
DO $$ BEGIN
    CREATE TYPE ticket_status AS ENUM (
        'OPEN',              -- Novo, aguardando atendimento
        'IN_PROGRESS',       -- Em atendimento
        'WAITING_RESPONSE',  -- Aguardando resposta do cliente
        'RESOLVED',          -- Resolvido
        'CLOSED'             -- Fechado
    );
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

-- Categoria do ticket
DO $$ BEGIN
    CREATE TYPE ticket_category AS ENUM (
        'GENERAL',           -- Dúvida geral
        'SUBSCRIPTION',      -- Relacionado a assinatura
        'PAYMENT',           -- Problema com pagamento
        'TECHNICAL',         -- Suporte técnico
        'BILLING',           -- Faturamento
        'OTHER'              -- Outros
    );
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

-- Tabela de tickets de suporte
CREATE TABLE IF NOT EXISTS support_tickets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    subject VARCHAR(255) NOT NULL,
    category ticket_category DEFAULT 'GENERAL',
    priority NUMERIC(2,1) DEFAULT 1, -- 1, 1.5, 2, 3, 4, 5 estrelas baseado no cargo
    status ticket_status DEFAULT 'OPEN',
    assigned_to UUID REFERENCES users(id), -- staff atribuído
    subscription_id UUID REFERENCES subscriptions(id), -- NULL se não for relacionado a assinatura
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    resolved_at TIMESTAMPTZ,
    closed_at TIMESTAMPTZ
);

-- Mensagens do ticket
CREATE TABLE IF NOT EXISTS support_messages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    ticket_id UUID NOT NULL REFERENCES support_tickets(id) ON DELETE CASCADE,
    sender_id UUID NOT NULL REFERENCES users(id),
    content TEXT NOT NULL,
    is_staff BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Índices para performance
CREATE INDEX IF NOT EXISTS idx_support_tickets_user ON support_tickets(user_id);
CREATE INDEX IF NOT EXISTS idx_support_tickets_status ON support_tickets(status);
CREATE INDEX IF NOT EXISTS idx_support_tickets_priority ON support_tickets(priority DESC);
CREATE INDEX IF NOT EXISTS idx_support_tickets_assigned ON support_tickets(assigned_to);
CREATE INDEX IF NOT EXISTS idx_support_tickets_created ON support_tickets(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_support_messages_ticket ON support_messages(ticket_id);
CREATE INDEX IF NOT EXISTS idx_support_messages_created ON support_messages(created_at);

-- Trigger para atualizar updated_at
CREATE OR REPLACE FUNCTION update_support_ticket_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trigger_support_ticket_updated_at ON support_tickets;
CREATE TRIGGER trigger_support_ticket_updated_at
    BEFORE UPDATE ON support_tickets
    FOR EACH ROW
    EXECUTE FUNCTION update_support_ticket_updated_at();

-- Comentários para documentação
COMMENT ON TABLE support_tickets IS 'Tickets de suporte com prioridade baseada no cargo do usuário.';
COMMENT ON COLUMN support_tickets.priority IS 'Prioridade em estrelas: 1 (sem cargo), 1.5 (Parceiro), 2 (Cliente), 3 (VIP/Admin), 4 (Eng), 5 (Direção).';
COMMENT ON COLUMN support_tickets.assigned_to IS 'ID do staff atribuído ao ticket. NULL = não atribuído.';
COMMENT ON TABLE support_messages IS 'Mensagens dentro de um ticket de suporte.';
COMMENT ON COLUMN support_messages.is_staff IS 'TRUE se a mensagem foi enviada por um membro da equipe.';

-- Create plans table
CREATE TABLE IF NOT EXISTS public.plans (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL PRIMARY KEY,
    name character varying(100) NOT NULL,
    description text,
    price numeric(10,2) NOT NULL,
    is_active boolean DEFAULT true,
    features jsonb,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT plans_price_check CHECK ((price >= (0)::numeric))
);

-- Seed default plans
INSERT INTO public.plans (name, description, price, features) VALUES
('Plano Network', 'Solução completa para grandes redes de servidores.', 499.90, '["Suporte Prioritário", "Plugins Exclusivos", "Otimização de Performance"]'),
('Plano Dev Pro', 'Para desenvolvedores que buscam ferramentas avançadas.', 199.90, '["Acesso ao Git", "CI/CD Pipeline", "Ambiente de Staging"]');

-- Alter subscriptions table
ALTER TABLE public.subscriptions 
ADD COLUMN IF NOT EXISTS plan_id uuid REFERENCES public.plans(id),
ADD COLUMN IF NOT EXISTS project_stage character varying(50) DEFAULT 'Planejamento',
ADD COLUMN IF NOT EXISTS agreed_price numeric(10,2);

-- Create project_logs table
CREATE TABLE IF NOT EXISTS public.project_logs (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL PRIMARY KEY,
    subscription_id uuid NOT NULL REFERENCES public.subscriptions(id),
    message text NOT NULL,
    created_by_user_id uuid, -- Optional: to track who added the log (admin or system)
    created_at timestamp with time zone DEFAULT now() NOT NULL
);

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_subscriptions_plan_id ON public.subscriptions(plan_id);
CREATE INDEX IF NOT EXISTS idx_project_logs_subscription_id ON public.project_logs(subscription_id);

-- Create files table
CREATE TABLE IF NOT EXISTS public.files (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL PRIMARY KEY,
    name character varying(255) NOT NULL, -- User-friendly name
    file_name character varying(255) NOT NULL, -- Internal UUID-based name
    file_type character varying(20), -- File extension/type
    file_path character varying(500), -- Path to file in filesystem
    size bigint, -- File size in bytes
    created_by uuid NOT NULL REFERENCES public.users(id), -- Who uploaded
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    is_deleted boolean DEFAULT false
);

-- Create indexes for files
CREATE INDEX IF NOT EXISTS idx_files_created_by ON public.files(created_by);
CREATE INDEX IF NOT EXISTS idx_files_is_deleted ON public.files(is_deleted);
CREATE INDEX IF NOT EXISTS idx_files_created_by_deleted ON public.files(created_by, is_deleted);

-- Add payment_id column to user_purchases table (if not exists)
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns
                   WHERE table_name = 'user_purchases' AND column_name = 'payment_id') THEN
        ALTER TABLE user_purchases ADD COLUMN payment_id uuid REFERENCES public.payments(id);
    END IF;
END $$;

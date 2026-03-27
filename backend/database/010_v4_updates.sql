-- Migration: Pixelcraft V4 Infrastructure Updates
-- Description: Adds user preferences, subscription status 'COMPLETED', and 'VIEW_CPF' permission.

-- 1. Add preferences column to users if not exists
-- density: comfortable | minimalist
-- font: modern | classic
-- backgroundFilter: boolean
ALTER TABLE users ADD COLUMN IF NOT EXISTS preferences JSONB DEFAULT '{"density": "comfortable", "font": "modern", "backgroundFilter": true}'::jsonb;

-- 2. Update subscription_status enum to include 'COMPLETED'
-- We wrap this in a DO block to prevent errors if it already exists, 
-- though ALTER TYPE ADD VALUE cannot be executed in a transaction block in some PG versions.
-- In some environments, we might need to run this outside a transaction.
ALTER TYPE public.subscription_status ADD VALUE IF NOT EXISTS 'COMPLETED';

-- 3. Update action_type enum to include 'VIEW_CPF'
ALTER TYPE public.action_type ADD VALUE IF NOT EXISTS 'VIEW_CPF';

-- 4. Grant view_cpf permission to DIRECTION and ENGINEERING roles
-- resource: USERS, action: VIEW_CPF
INSERT INTO role_permissions (role, resource, action)
VALUES 
    ('DIRECTION', 'USERS', 'VIEW_CPF'),
    ('ENGINEERING', 'USERS', 'VIEW_CPF')
ON CONFLICT (role, resource, action) DO NOTHING;

-- 5. Add 'WAITING_RESPONSE' to ticket_status (just in case it's missing on some environments)
ALTER TYPE public.ticket_status ADD VALUE IF NOT EXISTS 'WAITING_RESPONSE';

-- 6. Add 'CLOSED' to ticket_status (just in case)
ALTER TYPE public.ticket_status ADD VALUE IF NOT EXISTS 'CLOSED';

-- 7. Add 'OPEN' to ticket_status (just in case)
ALTER TYPE public.ticket_status ADD VALUE IF NOT EXISTS 'OPEN';

-- 8. Add 'IN_PROGRESS' to ticket_status (just in case)
ALTER TYPE public.ticket_status ADD VALUE IF NOT EXISTS 'IN_PROGRESS';

-- 9. Add 'RESOLVED' to ticket_status (just in case)
ALTER TYPE public.ticket_status ADD VALUE IF NOT EXISTS 'RESOLVED';

-- 10. Add users_growth_count to admin_analytics_snapshot
ALTER TABLE admin_analytics_snapshot ADD COLUMN IF NOT EXISTS users_growth_count INTEGER DEFAULT 0;

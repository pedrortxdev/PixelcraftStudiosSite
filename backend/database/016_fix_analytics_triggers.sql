-- 1. Ensure columns are correct
ALTER TABLE admin_analytics_snapshot ADD COLUMN IF NOT EXISTS users_growth_count INT DEFAULT 0;

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='admin_analytics_snapshot' AND column_name='last_updated') THEN
        IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='admin_analytics_snapshot' AND column_name='updated_at') THEN
            ALTER TABLE admin_analytics_snapshot RENAME COLUMN updated_at TO last_updated;
        ELSE
            ALTER TABLE admin_analytics_snapshot ADD COLUMN last_updated TIMESTAMP DEFAULT NOW();
        END IF;
    END IF;
END $$;

-- 2. Update Triggers to use last_updated
CREATE OR REPLACE FUNCTION update_user_stats() RETURNS TRIGGER AS $$
BEGIN
    IF (TG_OP = 'INSERT') THEN
        UPDATE admin_analytics_snapshot
        SET total_users = total_users + 1,
            last_updated = NOW()
        WHERE id = 1;
    ELSIF (TG_OP = 'DELETE') THEN
        UPDATE admin_analytics_snapshot
        SET total_users = total_users - 1,
            last_updated = NOW()
        WHERE id = 1;
    END IF;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION update_sales_stats() RETURNS TRIGGER AS $$
BEGIN
    IF (NEW.status = 'COMPLETED' AND (OLD.status IS NULL OR OLD.status != 'COMPLETED')) THEN
        -- Only count as a sale if NOT a test payment
        IF (NEW.is_test = FALSE) THEN
            UPDATE admin_analytics_snapshot
            SET 
                total_sales = total_sales + 1,
                last_updated = NOW()
            WHERE id = 1;
        END IF;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Trigger for Revenue (based on real deposits)
CREATE OR REPLACE FUNCTION update_revenue_stats() RETURNS TRIGGER AS $$
BEGIN
    -- Only count if status becomes COMPLETED
    IF (NEW.status = 'completed' AND (OLD.status IS NULL OR OLD.status != 'completed')) THEN
        -- Only count real deposits or "Pix Direto" adjustments
        IF (NEW.type = 'deposit' OR (NEW.type = 'admin_adjustment' AND NEW.adjustment_type = 'Pix Direto')) THEN
            UPDATE admin_analytics_snapshot
            SET 
                total_revenue = total_revenue + NEW.amount,
                last_updated = NOW()
            WHERE id = 1;
        END IF;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION update_product_stats() RETURNS TRIGGER AS $$
BEGIN
    IF (TG_OP = 'INSERT') THEN
        UPDATE admin_analytics_snapshot SET active_products = active_products + 1, last_updated = NOW() WHERE id = 1;
    ELSIF (TG_OP = 'DELETE') THEN
        UPDATE admin_analytics_snapshot SET active_products = active_products - 1, last_updated = NOW() WHERE id = 1;
    END IF;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

-- Migration to add adjustment tracking and fix revenue/sales metrics
-- 1. Add columns
ALTER TABLE transactions ADD COLUMN IF NOT EXISTS adjustment_type VARCHAR(50);
ALTER TABLE payments ADD COLUMN IF NOT EXISTS is_test BOOLEAN DEFAULT FALSE;

-- 2. Update update_sales_stats trigger to respect is_test for sales count
-- Note: We will move revenue calculation to a separate trigger on transactions
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

-- 3. Create a new trigger for Revenue (based on real deposits)
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

DROP TRIGGER IF EXISTS trg_update_revenue_stats ON transactions;
CREATE TRIGGER trg_update_revenue_stats
AFTER INSERT OR UPDATE ON transactions
FOR EACH ROW EXECUTE FUNCTION update_revenue_stats();

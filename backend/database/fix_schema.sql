SET search_path = public;

CREATE TABLE IF NOT EXISTS admin_analytics_snapshot (
    id SERIAL PRIMARY KEY,
    total_revenue DECIMAL(15, 2) DEFAULT 0,
    total_users INT DEFAULT 0,
    active_products INT DEFAULT 0,
    total_sales INT DEFAULT 0,
    revenue_growth_pct DECIMAL(5, 2) DEFAULT 0,
    users_growth_pct DECIMAL(5, 2) DEFAULT 0,
    products_status VARCHAR(50) DEFAULT 'Estável',
    sales_growth_pct DECIMAL(5, 2) DEFAULT 0,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Delete existing if any to avoid duplication error on insert
DELETE FROM admin_analytics_snapshot WHERE id = 1;
INSERT INTO admin_analytics_snapshot (id, total_revenue, total_users, active_products, total_sales)
VALUES (1, 0, 0, 0, 0);

CREATE OR REPLACE FUNCTION update_user_stats() RETURNS TRIGGER AS $$
BEGIN
    IF (TG_OP = 'INSERT') THEN
        UPDATE admin_analytics_snapshot
        SET total_users = total_users + 1,
            updated_at = NOW()
        WHERE id = 1;
    ELSIF (TG_OP = 'DELETE') THEN
        UPDATE admin_analytics_snapshot
        SET total_users = total_users - 1,
            updated_at = NOW()
        WHERE id = 1;
    END IF;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_update_user_stats ON users;
CREATE TRIGGER trg_update_user_stats
AFTER INSERT OR DELETE ON users
FOR EACH ROW EXECUTE FUNCTION update_user_stats();

CREATE OR REPLACE FUNCTION update_sales_stats() RETURNS TRIGGER AS $$
BEGIN
    IF (NEW.status = 'COMPLETED') THEN
        UPDATE admin_analytics_snapshot
        SET 
            total_sales = total_sales + 1,
            total_revenue = total_revenue + NEW.final_amount,
            updated_at = NOW()
        WHERE id = 1;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_update_sales_stats ON payments;
CREATE TRIGGER trg_update_sales_stats
AFTER INSERT OR UPDATE ON payments
FOR EACH ROW EXECUTE FUNCTION update_sales_stats();

CREATE OR REPLACE FUNCTION update_product_stats() RETURNS TRIGGER AS $$
BEGIN
    IF (TG_OP = 'INSERT') THEN
        UPDATE admin_analytics_snapshot SET active_products = active_products + 1 WHERE id = 1;
    ELSIF (TG_OP = 'DELETE') THEN
        UPDATE admin_analytics_snapshot SET active_products = active_products - 1 WHERE id = 1;
    END IF;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_update_product_stats ON products;
CREATE TRIGGER trg_update_product_stats
AFTER INSERT OR DELETE ON products
FOR EACH ROW EXECUTE FUNCTION update_product_stats();

-- Check if column name needs adjusting (the app seems to expect updated_at or last_updated)
-- Looking at the dump it was renamed to last_updated at the end.
-- I will keep it as updated_at first and see. 
-- Wait, the tail showed: RENAME COLUMN updated_at TO last_updated;
ALTER TABLE admin_analytics_snapshot RENAME COLUMN updated_at TO last_updated;

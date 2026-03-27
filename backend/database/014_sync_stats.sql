-- Recalculate stats excluding test data
UPDATE admin_analytics_snapshot
SET 
    total_revenue = (
        SELECT COALESCE(SUM(amount), 0) 
        FROM transactions 
        WHERE status IN ('completed', 'approved') 
        AND (type = 'deposit' OR (type = 'admin_adjustment' AND adjustment_type = 'Pix Direto'))
    ),
    total_sales = (
        SELECT COUNT(*) 
        FROM payments 
        WHERE status = 'COMPLETED' AND is_test = FALSE
    ),
    last_updated = NOW()
WHERE id = 1;

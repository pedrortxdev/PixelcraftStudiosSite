CREATE OR REPLACE FUNCTION update_sales_stats() RETURNS TRIGGER AS $$
BEGIN
    -- Só conta se o pagamento for confirmado
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

CREATE OR REPLACE FUNCTION update_user_stats() RETURNS TRIGGER AS $$
BEGIN
    -- Se for INSERT (Novo usuário)
    IF (TG_OP = 'INSERT') THEN
        UPDATE admin_analytics_snapshot
        SET total_users = total_users + 1,
            last_updated = NOW()
        WHERE id = 1;
    -- Se for DELETE (Caso você permita deletar usuário)
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
    -- Só conta se o pagamento for confirmado
    IF (NEW.status = 'COMPLETED') THEN
        UPDATE admin_analytics_snapshot
        SET 
            total_sales = total_sales + 1,
            total_revenue = total_revenue + NEW.final_amount,
            last_updated = NOW()
        WHERE id = 1;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

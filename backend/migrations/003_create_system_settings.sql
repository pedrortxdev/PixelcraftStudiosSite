CREATE TABLE IF NOT EXISTS system_settings (
    key VARCHAR(255) PRIMARY KEY,
    value TEXT NOT NULL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO system_settings (key, value) VALUES 
('smtp_host', 'mail.pixelcraft-studio.store'),
('smtp_port', '587'),
('smtp_email', 'suporte@pixelcraft-studio.store'),
('smtp_password', ''), -- Password must be set by admin
('smtp_from', 'Pixelcraft Studio <suporte@pixelcraft-studio.store>')
ON CONFLICT (key) DO NOTHING;


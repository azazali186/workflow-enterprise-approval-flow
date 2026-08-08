-- Migration 000007: Seed default roles

-- Insert default roles (using ON CONFLICT to be idempotent)
INSERT INTO roles (name, description, is_default) VALUES
    ('admin', 'System administrator with full access', false),
    ('user', 'Regular user with basic access', true),
    ('viewer', 'Read-only access', false)
ON CONFLICT (name) DO NOTHING;

-- NOTE: The default administrator account is intentionally NOT seeded here.
-- The initial admin is bootstrapped at startup from the ADMIN_EMAIL and
-- ADMIN_PASSWORD environment variables (see Server.bootstrapAdmin), so no
-- well-known credentials ever ship with the application.
-- Migration 000009 removes any legacy default admin accounts created by
-- earlier versions of this migration.

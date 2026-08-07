-- Migration 000007: Seed default roles

-- Insert default roles (using ON CONFLICT to be idempotent)
INSERT INTO roles (name, description, is_default) VALUES
    ('admin', 'System administrator with full access', false),
    ('user', 'Regular user with basic access', true),
    ('viewer', 'Read-only access', false)
ON CONFLICT (name) DO NOTHING;

-- Insert default admin user (password: Admin123!)
-- bcrypt hash of Admin123!
INSERT INTO users (email, password, name, status) VALUES
    ('admin@approval-flow.com', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', 'System Admin', 'active')
ON CONFLICT (email) DO NOTHING;

-- Assign admin role to admin user
INSERT INTO user_roles (user_id, role_id)
SELECT u.id, r.id
FROM users u, roles r
WHERE u.email = 'admin@approval-flow.com' AND r.name = 'admin'
ON CONFLICT DO NOTHING;

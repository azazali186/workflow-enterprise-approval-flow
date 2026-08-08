-- Migration 000009 (down): restore the legacy default admin account
-- NOTE: This intentionally restores the previous behavior for rollback
-- completeness. Production deployments should rely on ADMIN_EMAIL /
-- ADMIN_PASSWORD bootstrap instead and change this password immediately.

INSERT INTO users (email, password, name, status) VALUES
    ('admin@approval-flow.com', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', 'System Admin', 'active')
ON CONFLICT (email) DO NOTHING;

INSERT INTO user_roles (user_id, role_id)
SELECT u.id, r.id
FROM users u, roles r
WHERE u.email = 'admin@approval-flow.com' AND r.name = 'admin'
ON CONFLICT DO NOTHING;

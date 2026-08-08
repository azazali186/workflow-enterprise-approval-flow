-- Migration 000009: Remove legacy default admin account
--
-- Earlier versions of migration 000007 seeded a well-known administrator
-- account (admin@approval-flow.com with a published password). This migration
-- removes that account and its role assignments from any database where it
-- was already created. The initial admin is now bootstrapped from
-- ADMIN_EMAIL / ADMIN_PASSWORD environment variables at startup.

-- Remove role assignments for the legacy default admin account only.
-- Both statements filter on the exact seeded name ('System Admin') so a
-- legitimate user who later registered the same email is never affected.
DELETE FROM user_roles
WHERE user_id IN (
    SELECT id FROM users
    WHERE email = 'admin@approval-flow.com'
      AND name = 'System Admin'
);

-- Remove the legacy default admin account
DELETE FROM users
WHERE email = 'admin@approval-flow.com'
  AND name = 'System Admin';

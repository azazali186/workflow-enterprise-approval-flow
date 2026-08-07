-- Migration 000007: Remove seed data

-- Remove seed user
DELETE FROM user_roles WHERE user_id IN (SELECT id FROM users WHERE email = 'admin@approval-flow.com');
DELETE FROM users WHERE email = 'admin@approval-flow.com';

-- Remove default roles
DELETE FROM roles WHERE name IN ('admin', 'user', 'viewer');

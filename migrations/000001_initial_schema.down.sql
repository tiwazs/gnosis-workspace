-- drops in reverse order to respect foreign key constraints
SET search_path TO workspaces;
DROP TABLE IF EXISTS registration_tokens;
DROP TABLE IF EXISTS active_sessions;
DROP TABLE IF EXISTS usage_events;
DROP TABLE IF EXISTS workspace_quotas;
DROP TABLE IF EXISTS workspace_entitlements;
DROP TABLE IF EXISTS workspace_members;
DROP TABLE IF EXISTS role_permissions;
DROP TABLE IF EXISTS roles;
DROP TABLE IF EXISTS permissions;
DROP TABLE IF EXISTS workspaces;
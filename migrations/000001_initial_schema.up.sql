-- ─────────────────────────────────────────────
-- WORKSPACE SERVICE — Initial Schema
-- ─────────────────────────────────────────────
-- Covers:
--   workspaces
--   roles + permissions      (RBAC)
--   workspace_members
--   workspace_entitlements   (which services are enabled)
--   workspace_quotas         (usage limits)
--   usage_events             (usage tracking)
--   active_sessions          (concurrent session tracking)
--   registration_tokens      (device provisioning)
-- ─────────────────────────────────────────────

-- enable UUID generation
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE SCHEMA IF NOT EXISTS workspaces;
SET search_path TO workspaces;

-- ─────────────────────────────────────────────
-- WORKSPACES
-- ─────────────────────────────────────────────

CREATE TABLE workspaces (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    name        TEXT        NOT NULL,
    owner_id    TEXT        NOT NULL,   -- main-service user id (cuid)
    created_at  TIMESTAMPTZ DEFAULT NOW(),
    updated_at  TIMESTAMPTZ DEFAULT NOW()
);


-- ─────────────────────────────────────────────
-- RBAC — permissions, roles, role_permissions
-- ─────────────────────────────────────────────

-- atomic permissions: one row per service+action combination
CREATE TABLE permissions (
    id          UUID    PRIMARY KEY DEFAULT gen_random_uuid(),
    service     TEXT    NOT NULL,   -- "iot" | "recognition" | "workspaces" | "commands" | "telemetry"
    action      TEXT    NOT NULL,   -- "read" | "write" | "manage" | "admin"
    description TEXT,
    UNIQUE (service, action)
);

-- roles: system-wide defaults (workspace_id IS NULL) or custom per workspace
CREATE TABLE roles (
    id           UUID    PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID    REFERENCES workspaces(id) ON DELETE CASCADE,
    name         TEXT    NOT NULL,
    description  TEXT,
    is_system    BOOLEAN DEFAULT FALSE,   -- system roles cannot be deleted or modified
    created_at   TIMESTAMPTZ DEFAULT NOW()
);

-- which permissions belong to which role
CREATE TABLE role_permissions (
    role_id       UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    permission_id UUID NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,
    PRIMARY KEY (role_id, permission_id)
);


-- ─────────────────────────────────────────────
-- WORKSPACE MEMBERS
-- ─────────────────────────────────────────────

CREATE TABLE workspace_members (
    id           UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID        NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    user_id      TEXT        NOT NULL,   -- main-service user id (cuid)
    role_id      UUID        NOT NULL REFERENCES roles(id),
    created_at   TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE (workspace_id, user_id)       -- user can only have one role per workspace
);


-- ─────────────────────────────────────────────
-- ENTITLEMENTS — which services a workspace can use
-- ─────────────────────────────────────────────

CREATE TABLE workspace_entitlements (
    workspace_id  UUID    NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    service       TEXT    NOT NULL,   -- "iot" | "recognition" | "analytics"
    enabled       BOOLEAN DEFAULT FALSE,
    enabled_at    TIMESTAMPTZ,
    PRIMARY KEY (workspace_id, service)
);


-- ─────────────────────────────────────────────
-- QUOTAS — usage limits per workspace per service
-- ─────────────────────────────────────────────

CREATE TABLE workspace_quotas (
    workspace_id              UUID    PRIMARY KEY REFERENCES workspaces(id) ON DELETE CASCADE,
    -- recognition
    max_concurrent_sessions   INT     DEFAULT 1,
    max_monthly_minutes       INT     DEFAULT 60,
    -- iot
    max_devices               INT     DEFAULT 10,
    max_daily_commands        INT     DEFAULT 1000,
    -- general
    max_daily_requests        INT     DEFAULT 500,
    updated_at                TIMESTAMPTZ DEFAULT NOW()
);


-- ─────────────────────────────────────────────
-- USAGE EVENTS — append-only log of consumption
-- ─────────────────────────────────────────────

CREATE TABLE usage_events (
    id            UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id  UUID        NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    service       TEXT        NOT NULL,   -- "recognition" | "iot" | "commands"
    event_type    TEXT        NOT NULL,   -- "session_started" | "session_ended" | "command_sent" | "request"
    quantity      FLOAT       DEFAULT 1,  -- minutes, count, bytes, etc
    metadata      JSONB,                  -- any extra context (device_id, session_id, etc)
    recorded_at   TIMESTAMPTZ DEFAULT NOW()
);

-- index for fast monthly usage queries
CREATE INDEX idx_usage_events_workspace_service_date
    ON usage_events (workspace_id, service, recorded_at);


-- ─────────────────────────────────────────────
-- ACTIVE SESSIONS — for concurrent limit checks
-- ─────────────────────────────────────────────

CREATE TABLE active_sessions (
    id            UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id  UUID        NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    user_id       UUID        NOT NULL,
    service       TEXT        NOT NULL,   -- "recognition" | "iot"
    device_id     UUID,                   -- if session is tied to a device
    started_at    TIMESTAMPTZ DEFAULT NOW()
);

-- index for fast concurrent session count
CREATE INDEX idx_active_sessions_workspace_service
    ON active_sessions (workspace_id, service);


-- ─────────────────────────────────────────────
-- REGISTRATION TOKENS — device provisioning
-- ─────────────────────────────────────────────

CREATE TABLE registration_tokens (
    token             TEXT        PRIMARY KEY,
    workspace_id      UUID        NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    created_by        TEXT        NOT NULL,   -- main-service user id (cuid)
    expires_at        TIMESTAMPTZ NOT NULL,
    used              BOOLEAN     DEFAULT FALSE,
    used_at           TIMESTAMPTZ,
    used_by_device    UUID                    -- device_id from device service
);

-- index for fast token lookup on registration
CREATE INDEX idx_registration_tokens_token_used
    ON registration_tokens (token, used);


-- ─────────────────────────────────────────────
-- SEED — system permissions
-- ─────────────────────────────────────────────

INSERT INTO permissions (service, action, description) VALUES
-- iot service
('iot', 'read',   'View devices and their status'),
('iot', 'write',  'Register devices and send commands'),
('iot', 'manage', 'Deprovision devices and manage device settings'),

-- commands
('commands', 'read',  'View command history'),
('commands', 'write', 'Send commands to devices'),

-- telemetry
('telemetry', 'read',  'View telemetry data and readings'),
('telemetry', 'write', 'Publish telemetry (for service accounts)'),

-- recognition service
('recognition', 'read',   'View recognition sessions and results'),
('recognition', 'write',  'Start and stop recognition sessions'),
('recognition', 'manage', 'Manage groups and profiles used for recognition'),

-- workspace management
('workspaces', 'read',   'View workspace info and member list'),
('workspaces', 'write',  'Edit workspace name and settings'),
('workspaces', 'manage', 'Invite and remove members, assign roles'),
('workspaces', 'admin',  'Delete workspace, manage custom roles, manage entitlements');


-- ─────────────────────────────────────────────
-- SEED — system roles (workspace_id NULL = global defaults)
-- ─────────────────────────────────────────────

INSERT INTO roles (id, name, description, is_system) VALUES
('00000000-0000-0000-0000-000000000001', 'Owner',  'Full access to everything in the workspace', TRUE),
('00000000-0000-0000-0000-000000000002', 'Editor', 'Can operate services but not manage workspace', TRUE),
('00000000-0000-0000-0000-000000000003', 'Viewer', 'Read only access across all services', TRUE);

-- Owner gets every permission
INSERT INTO role_permissions (role_id, permission_id)
SELECT '00000000-0000-0000-0000-000000000001', id FROM permissions;

-- Editor gets operational permissions (no workspace admin)
INSERT INTO role_permissions (role_id, permission_id)
SELECT '00000000-0000-0000-0000-000000000002', id
FROM permissions
WHERE (service, action) IN (
    ('iot',         'read'),
    ('iot',         'write'),
    ('commands',    'read'),
    ('commands',    'write'),
    ('telemetry',   'read'),
    ('recognition', 'read'),
    ('recognition', 'write'),
    ('workspaces',  'read')
);

-- Viewer gets read only
INSERT INTO role_permissions (role_id, permission_id)
SELECT '00000000-0000-0000-0000-000000000003', id
FROM permissions
WHERE action = 'read';
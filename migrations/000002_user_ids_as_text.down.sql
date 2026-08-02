-- revert user reference columns back to UUID (only valid if values are UUIDs)
ALTER TABLE registration_tokens
    ALTER COLUMN created_by TYPE UUID USING created_by::uuid;

ALTER TABLE workspace_members
    ALTER COLUMN user_id TYPE UUID USING user_id::uuid;

ALTER TABLE workspaces
    ALTER COLUMN owner_id TYPE UUID USING owner_id::uuid;

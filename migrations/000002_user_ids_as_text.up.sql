-- main-service users use cuid strings, not UUIDs
ALTER TABLE workspaces
    ALTER COLUMN owner_id TYPE TEXT USING owner_id::text;

ALTER TABLE workspace_members
    ALTER COLUMN user_id TYPE TEXT USING user_id::text;

ALTER TABLE registration_tokens
    ALTER COLUMN created_by TYPE TEXT USING created_by::text;

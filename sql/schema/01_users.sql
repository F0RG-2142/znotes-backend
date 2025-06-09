-- +goose Up
CREATE TABLE IF NOT EXISTS Users (
    id UUID PRIMARY KEY,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    email TEXT NOT NULL,
    hashed_password TEXT NOT NULL DEFAULT 'unset',
    has_notes_premium BOOL NOT NULL DEFAULT 'false'
);
CREATE TABLE IF NOT EXISTS Notes (
    id UUID PRIMARY KEY,
    name TEXT NOT NULL DEFAULT 'unset',
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    body TEXT NOT NULL,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE
);
CREATE TABLE IF NOT EXISTS Teams (
    id UUID PRIMARY KEY,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    team_name TEXT NOT NULL,
    created_by UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    is_private BOOLEAN NOT NULL DEFAULT true
);
CREATE TABLE IF NOT EXISTS User_Teams (
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    team_id UUID NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
    role TEXT NOT NULL DEFAULT 'viewer' CHECK (role IN ('admin', 'editor', 'viewer')),
    joined_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (user_id, team_id)
);
CREATE TABLE IF NOT EXISTS Note_Teams (
    note_id UUID NOT NULL REFERENCES notes(id) ON DELETE CASCADE,
    team_id UUID NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
    shared_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (note_id, team_id)
);
CREATE TABLE IF NOT EXISTS Refresh_Tokens (
    token TEXT PRIMARY KEY,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    expires_at  TIMESTAMP NOT NULL,
    revoked_at TIMESTAMP
);
-- Indexes for performance
CREATE INDEX idx_notes_user_id ON Notes (user_id);
CREATE INDEX idx_user_teams_team_id ON User_Teams (team_id);
CREATE INDEX idx_note_teams_team_id ON Note_Teams (team_id);
CREATE INDEX idx_refresh_tokens_user_id ON Refresh_Tokens (user_id);

-- +goose Down
DROP TABLE users CASCADE;
DROP TABLE notes CASCADE;
DROP TABLE refresh_tokens;
DROP TABLE teams CASCADE;
DROP TABLE user_teams;
DROP TABLE note_teams;
-- name: NewGroup :exec
INSERT INTO teams (id, created_at, updated_at, team_name, created_by, is_private)
VALUES(
    gen_random_uuid (),
    NOW(),
    NOW(),
    $1,
    $2,
    $3
);

-- name: GetAllGroups :many
SELECT * FROM teams ORDER BY created_at ASC;

-- name: GetGroupById :one
SELECT * FROM teams WHERE id = $1;

-- name: DeleteGroup :exec
DELETE FROM teams WHERE id = $1;

-- name: AddToTeam :exec
INSERT INTO user_teams (user_id, team_id, role, joined_at)
VALUES (
    $1,
    $2,
    $3,
    NOW()
);

-- name: RemoveUser :exec
DELETE FROM user_teams WHERE user_id = $1;

-- name: GetTeamMembers :many
SELECT * FROM teams WHERE id = $1;
-- name: NewTeam :exec
INSERT INTO teams (id, created_at, updated_at, team_name, created_by, is_private)
VALUES(
    gen_random_uuid (),
    NOW(),
    NOW(),
    $1,
    $2,
    $3
);

-- name: GetAllTeams :many
SELECT t.*
FROM Teams t
INNER JOIN User_Teams ut ON t.id = ut.team_id
WHERE ut.user_id = $1;

-- name: GetTeamById :one
SELECT t.*
FROM Teams t
INNER JOIN User_Teams ut ON t.id = ut.team_id
WHERE ut.user_id = $1 AND ut.team_id = $2;

-- name: DeleteTeam :exec
DELETE FROM Teams t
USING User_Teams ut
WHERE t.id = ut.team_id
AND ut.user_id = $1
AND ut.role = 'admin'
AND t.id = $2;

-- name: AddToTeam :exec
INSERT INTO user_teams (user_id, team_id, role, joined_at)
VALUES (
    $1,
    $2,
    $3,
    NOW()
);

-- name: RemoveUser :exec
DELETE FROM user_teams WHERE user_id = $1 AND team_id = $2;

-- name: GetTeamMembers :many
SELECT * FROM teams WHERE id = $1;

-- name: GetTeamMember :one
SELECT * FROM User_Teams WHERE user_id = $1 AND team_id = $2;
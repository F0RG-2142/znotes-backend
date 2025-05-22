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

-- name: AddUserToTeam :exec
INSERT INTO user_teams (user_id, team_id, role, joined_at)
VALUES (
    $1,
    $2,
    $3,
    NOW()
);

-- name: RemoveUserFromTeam :exec
DELETE FROM user_teams WHERE user_id = $1 AND team_id = $2;

-- name: GetTeamMembers :many
SELECT * FROM teams WHERE id = $1;

-- name: GetTeamMember :one
SELECT * FROM User_Teams WHERE user_id = $1 AND team_id = $2;

-- name: AddNoteToTeam :exec
INSERT INTO Note_Teams (note_id, team_id, shared_at)
SELECT $1, t.id, NOW()
FROM Teams t
JOIN User_Teams ut ON t.id = ut.team_id
WHERE t.id = $2
AND ut.user_id = $3
AND ut.role IN ('admin', 'editor');

-- name: RemoveNoteFromTeam :exec
DELETE FROM Notes n
USING Note_Teams nt
JOIN User_Teams ut ON nt.team_id = ut.team_id
WHERE n.id = nt.note_id
AND nt.note_id = $1
AND ut.user_id = $2
AND ut.role = 'admin';

-- name: GetTeamNote :one
SELECT n.*
FROM Notes n
JOIN Note_Teams nt ON n.id = nt.note_id
JOIN User_Teams ut ON nt.team_id = ut.team_id
WHERE n.id = $1
AND nt.team_id = $2
AND ut.user_id = $3;

-- name: GetTeamNotes :many
SELECT n.*
FROM Notes n
JOIN Note_Teams nt ON n.id = nt.note_id
JOIN User_Teams ut ON nt.team_id = ut.team_id
WHERE nt.team_id = $1
AND ut.user_id = $2;

-- name: UpdateTeamNote :exec
UPDATE Notes n
SET body = $1, updated_at = NOW()
FROM Note_Teams nt
JOIN User_Teams ut ON nt.team_id = ut.team_id
WHERE n.id = nt.note_id
AND n.id = $2
AND nt.team_id = $3
AND ut.user_id = $4;
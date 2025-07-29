-- name: NewNote :one
INSERT INTO notes (id, created_at, updated_at, body, user_id)
VALUES (
    gen_random_uuid (),
    NOW(),
    NOW(),
    $1,
    $2 
)
RETURNING id;

-- name: GetAllNotes :many
SELECT * FROM notes WHERE user_id = $1 ORDER BY created_at ASC ;

-- name: GetNoteByID :one
SELECT * FROM notes WHERE id = $1 AND user_id = $2;

-- name: DeleteNote :exec
DELETE FROM notes WHERE id = $1 AND user_id = $2;

-- name: UpdateNote :exec
UPDATE notes
SET 
    updated_at = NOW(),
    body = $1,
    name = $2
WHERE 
    id = $3;

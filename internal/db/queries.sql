-- name: ListCandidatesByWheel :many
SELECT * FROM candidates WHERE wheel_id = ?;

-- name: DeclareWinnerForWheel :exec
UPDATE wheels SET winner_id = ?, status = 'winner_declared' WHERE id = ?;

-- name: ListWheels :many
SELECT * FROM wheels;

-- name: CountWheels :one
SELECT COUNT(*) FROM wheels;

-- name: GetWheelByID :one
SELECT * FROM wheels WHERE id = ?;

-- name: CreateWheel :execlastid
INSERT INTO wheels (name, description) VALUES (?, ?);

-- name: SetWheelStatus :exec
UPDATE wheels SET status = ? WHERE id = ?;

-- name: CreateCandidate :exec
INSERT INTO candidates (username, wheel_id, creator_id) VALUES (?, ?, ?);

-- name: GetDuplicateCandidatesForWheel :one
SELECT * FROM candidates WHERE wheel_id = ? AND (username = ? OR creator_id = ?);

-- name: GetCandidateByCreatorIDAndWheelID :one
SELECT * FROM candidates WHERE creator_id = ? AND wheel_id = ?;

-- name: GetCandidateBySessionIDAndWheelID :one
SELECT * FROM candidates WHERE creator_id = ? AND wheel_id = ?;

-- name: GetCandidateByID :one
SELECT * FROM candidates WHERE id = ? AND wheel_id = ?;

-- name: DeleteCandidateByID :exec
DELETE FROM candidates WHERE id = ? AND wheel_id = ?;

-- name: DeleteWheelByID :exec
DELETE FROM wheels WHERE id = ?;

-- name: ListCandidatesByWheel :many
SELECT * FROM candidates WHERE wheel_id = ?;

-- name: ListWheels :many
SELECT * FROM wheels;

-- name: GetWheelByID :one
SELECT * FROM wheels WHERE id = ?;

-- name: AddCandidateToWheel :exec
INSERT INTO candidates (username, wheel_id, creator_id) VALUES (?, ?, ?);

-- name: GetDuplicateCandidatesForWheel :one
SELECT * FROM candidates WHERE wheel_id = ? AND (username = ? OR creator_id = ?);

-- name: GetCandidateByCreatorIdAndWheelId :one
SELECT * FROM candidates WHERE creator_id = ? AND wheel_id = ?;

-- name: GetCandidateBySessionIdAndWheelId :one
SELECT * FROM candidates WHERE creator_id = ? AND wheel_id = ?;

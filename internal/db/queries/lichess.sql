-- name: InsertLichessData :exec
INSERT INTO lichess(lichess_id, username) VALUES ($1, $2);

-- name: GetLichessData :many
SELECT lichess_id from lichess;
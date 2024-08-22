-- name: InsertMember :exec
INSERT INTO lichess(lichess_id, username) VALUES ($1, $2);

-- name: GetLichessMembers :many
SELECT lichess_id from lichess;
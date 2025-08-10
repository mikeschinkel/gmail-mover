-- Metadata operations for database identification and sync state

-- name: SelectMetadata :one
SELECT value FROM metadata WHERE name = ?;

-- name: UpsertMetadata :exec
INSERT OR REPLACE INTO metadata (name, value) VALUES (?, ?);

-- name: DeleteMetadata :exec
DELETE FROM metadata WHERE name = ?;

-- name: ListMetadata :many
SELECT key, value FROM metadata ORDER BY name;
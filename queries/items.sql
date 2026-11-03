-- name: ListItems :many
SELECT id, name, created_at FROM items ORDER BY created_at DESC;

-- name: CreateItem :exec
INSERT INTO items (id, name, created_at)
VALUES ($1, $2, $3);

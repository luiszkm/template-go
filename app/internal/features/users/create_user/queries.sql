-- name: InsertUser :one
INSERT INTO users (email, name, password_hash) VALUES ($1, $2, $3)
ON CONFLICT (email) DO NOTHING
RETURNING id, email, name, created_at;

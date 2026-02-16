-- name: CreateUser :one
INSERT INTO users(name, email, hashed_pass, created_at, updated_at)
VALUES (
    $1,
    $2,
    $3,
    NOW(),
    NOW()
)
RETURNING *;

-- name: GetUserByEmail :one
SELECT * FROM users 
WHERE email = $1;

-- name: GetUserByID :one
SELECT * FROM users
WHERE id = $1;
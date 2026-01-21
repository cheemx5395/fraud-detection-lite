-- name: CreateUser :one
INSERT INTO users(name, email, mobile_number, hashed_pass, created_at, updated_at)
VALUES (
    $1,
    $2,
    $3,
    $4,
    NOW(),
    NOW()
)
RETURNING *;

-- name: GetUserByEmail :one
SELECT * FROM users 
WHERE email = $1;
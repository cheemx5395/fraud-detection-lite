-- name: CountRecentTransactions :one
SELECT COUNT(*)
FROM transactions
WHERE user_id = $1
AND created_at > NOW() - make_interval(secs => $2);

-- name: CreateTransaction :one
INSERT INTO transactions (
    user_id,
    amount,
    type,
    mode,
    risk_score,
    triggered_factors,
    decision,
    created_at,
    updated_at
) VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), NOW())
RETURNING *;

-- name: CountTodaysTransactions :one
SELECT COUNT(*)
FROM transactions
WHERE user_id = $1
AND created_at >= CURRENT_DATE;

-- name: GetTransactions :many
SELECT * FROM transactions
LIMIT 20 OFFSET 40;

-- name: GetAllTransactionsByUserID :many
SELECT * FROM transactions
WHERE user_id = $1
LIMIT 20 OFFSET 40;

-- name: GetDailyTransactionStats :one
SELECT
    COUNT(*)                        AS total_transactions,
    COALESCE(AVG(amount), 0)::INT   AS average_amount,
    MIN(created_at)                 AS first_transaction_time,
    MAX(created_at)                 AS last_transaction_time
FROM transactions
WHERE user_id = $1
  AND created_at >= CURRENT_DATE;

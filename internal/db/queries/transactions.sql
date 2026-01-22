-- name: CountRecentTransactions :one
SELECT COUNT(*)
FROM transactions
WHERE user_id = $1
AND created_at > NOW() - make_interval(secs => $2);

-- name: CreateTransaction :one
INSERT INTO transactions (
    user_id,
    amount,
    mode,
    risk_score,
    triggered_factors,
    decision,
    created_at,
    updated_at
) VALUES (
    $1,
    $2,
    $3,
    $4,
    $5::text[]::trigger_factors[],   
    $6,
    NOW(),
    NOW()
)
RETURNING
    id,
    user_id,
    amount,
    mode,
    risk_score,
    triggered_factors::text[] AS triggered_factors, 
    decision,
    created_at,
    updated_at;

-- name: CountTodaysTransactions :one
SELECT COUNT(*)
FROM transactions
WHERE user_id = $1
AND created_at >= CURRENT_DATE;

-- name: GetAllTransactionsByUserID :many
SELECT * FROM transactions
WHERE user_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: GetDailyTransactionStats :one
SELECT
    COUNT(*)                        AS total_transactions,
    COALESCE(AVG(amount), 0)::INT   AS average_amount,
    MIN(created_at)                 AS first_transaction_time,
    MAX(created_at)                 AS last_transaction_time
FROM transactions
WHERE user_id = $1
  AND created_at >= CURRENT_DATE;

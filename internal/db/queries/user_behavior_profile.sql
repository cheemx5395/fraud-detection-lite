-- name: UpsertUserProfileByUserID :exec
INSERT INTO user_profile_behavior (
    user_id,
    average_transaction_amount,
    std_dev_transaction_amount,
    max_transaction_amount_seen,
    average_number_of_transactions_per_day,
    registered_payment_modes,
    usual_transaction_start_hour,
    usual_transaction_end_hour,
    total_transactions,
    allowed_transactions,
    updated_at
)
SELECT
    u.id AS user_id,

    COALESCE(
        AVG(t.amount) FILTER (WHERE t.decision IN ('ALLOW', 'FLAG')),
        0
    )::INTEGER AS average_transaction_amount,

    COALESCE(
        STDDEV(t.amount) FILTER (WHERE t.decision IN ('ALLOW', 'FLAG')),
        0
    )::INTEGER AS std_dev_transaction_amount,

    COALESCE(MAX(t.amount), 0)::INTEGER AS max_transaction_amount_seen,

    LEAST(
        COUNT(*) FILTER (WHERE t.decision IN ('ALLOW', 'FLAG')),
        50
    )::INTEGER AS average_number_of_transactions_per_day,

    COALESCE(
        ARRAY_AGG(DISTINCT t.mode) FILTER (WHERE t.decision IN ('ALLOW', 'FLAG')),
        ARRAY[]::mode[]
    ) AS registered_payment_modes,

    MIN(t.created_at)
        FILTER (WHERE t.decision IN ('ALLOW', 'FLAG')) AS usual_transaction_start_hour,

    MAX(t.created_at)
        FILTER (WHERE t.decision IN ('ALLOW', 'FLAG')) AS usual_transaction_end_hour,

    COUNT(t.id) AS total_transactions,

    COUNT(t.id) FILTER (WHERE t.decision IN ('ALLOW', 'FLAG')) AS allowed_transactions,

    NOW() AS updated_at
FROM users u
LEFT JOIN transactions t
    ON t.user_id = u.id
    AND t.created_at < CURRENT_DATE
WHERE u.id = $1
GROUP BY u.id

ON CONFLICT (user_id) DO UPDATE SET
    average_transaction_amount = EXCLUDED.average_transaction_amount,
    std_dev_transaction_amount = EXCLUDED.std_dev_transaction_amount,
    max_transaction_amount_seen = EXCLUDED.max_transaction_amount_seen,
    average_number_of_transactions_per_day = EXCLUDED.average_number_of_transactions_per_day,
    registered_payment_modes = EXCLUDED.registered_payment_modes,
    usual_transaction_start_hour = EXCLUDED.usual_transaction_start_hour,
    usual_transaction_end_hour = EXCLUDED.usual_transaction_end_hour,
    total_transactions = EXCLUDED.total_transactions,
    allowed_transactions = EXCLUDED.allowed_transactions,
    updated_at = EXCLUDED.updated_at;


-- name: RebuildAllUserProfiles :exec
INSERT INTO user_profile_behavior (
    user_id,
    average_transaction_amount,
    std_dev_transaction_amount,
    max_transaction_amount_seen,
    average_number_of_transactions_per_day,
    registered_payment_modes,
    usual_transaction_start_hour,
    usual_transaction_end_hour,
    updated_at
)
SELECT
    t.user_id,

    COALESCE(
        AVG(t.amount) FILTER (WHERE t.decision IN ('ALLOW', 'FLAG')),
        0
    )::INTEGER AS average_transaction_amount,

    COALESCE(
        STDDEV(t.amount) FILTER (WHERE t.decision IN ('ALLOW', 'FLAG')),
        0
    )::INTEGER AS std_dev_transaction_amount,

    COALESCE(MAX(t.amount), 0)::INTEGER AS max_transaction_amount_seen,

    LEAST(
        COUNT(*) FILTER (WHERE t.decision IN ('ALLOW', 'FLAG')),
        50
    )::INTEGER AS average_number_of_transactions_per_day,

    COALESCE(
        (ARRAY_AGG(DISTINCT t.mode) 
            FILTER (WHERE t.decision IN ('ALLOW', 'FLAG'))),
        ARRAY[]::mode[]
    ) AS registered_payment_modes,

    MIN(t.created_at)
        FILTER (WHERE t.decision IN ('ALLOW', 'FLAG')) AS usual_transaction_start_hour,

    MAX(t.created_at)
        FILTER (WHERE t.decision IN ('ALLOW', 'FLAG')) AS usual_transaction_end_hour,

    NOW() AS updated_at

FROM transactions t
WHERE t.created_at < CURRENT_DATE
GROUP BY t.user_id

ON CONFLICT (user_id) DO UPDATE SET
    average_transaction_amount = EXCLUDED.average_transaction_amount,
    std_dev_transaction_amount = EXCLUDED.std_dev_transaction_amount,
    max_transaction_amount_seen = EXCLUDED.max_transaction_amount_seen,
    average_number_of_transactions_per_day = EXCLUDED.average_number_of_transactions_per_day,
    registered_payment_modes = EXCLUDED.registered_payment_modes,
    usual_transaction_start_hour = EXCLUDED.usual_transaction_start_hour,
    usual_transaction_end_hour = EXCLUDED.usual_transaction_end_hour,
    updated_at = EXCLUDED.updated_at;

-- name: RecalculateUserProfile :exec
INSERT INTO user_profile_behavior (
    user_id,
    average_transaction_amount,
    std_dev_transaction_amount,
    max_transaction_amount_seen,
    average_number_of_transactions_per_day,
    registered_payment_modes,
    usual_transaction_start_hour,
    usual_transaction_end_hour,
    total_transactions,
    allowed_transactions,
    updated_at
)
SELECT
    u.id AS user_id,

    COALESCE(
        AVG(t.amount) FILTER (WHERE t.decision IN ('ALLOW', 'FLAG')),
        0
    )::INTEGER AS average_transaction_amount,

    COALESCE(
        STDDEV(t.amount) FILTER (WHERE t.decision IN ('ALLOW', 'FLAG')),
        0
    )::INTEGER AS std_dev_transaction_amount,

    COALESCE(MAX(t.amount), 0)::INTEGER AS max_transaction_amount_seen,

    LEAST(
        COUNT(*) FILTER (WHERE t.decision IN ('ALLOW', 'FLAG')) / GREATEST(DATE_PART('day', NOW() - MIN(t.created_at)), 1),
        50
    )::INTEGER AS average_number_of_transactions_per_day,

    COALESCE(
        ARRAY_AGG(DISTINCT t.mode) FILTER (WHERE t.decision IN ('ALLOW', 'FLAG')),
        ARRAY[]::mode[]
    ) AS registered_payment_modes,

    MIN(t.created_at)
        FILTER (WHERE t.decision IN ('ALLOW', 'FLAG')) AS usual_transaction_start_hour,

    MAX(t.created_at)
        FILTER (WHERE t.decision IN ('ALLOW', 'FLAG')) AS usual_transaction_end_hour,

    COUNT(t.id) AS total_transactions,

    COUNT(t.id) FILTER (WHERE t.decision IN ('ALLOW', 'FLAG')) AS allowed_transactions,

    NOW() AS updated_at
FROM users u
LEFT JOIN transactions t
    ON t.user_id = u.id
WHERE u.id = $1
GROUP BY u.id

ON CONFLICT (user_id) DO UPDATE SET
    average_transaction_amount = EXCLUDED.average_transaction_amount,
    std_dev_transaction_amount = EXCLUDED.std_dev_transaction_amount,
    max_transaction_amount_seen = EXCLUDED.max_transaction_amount_seen,
    average_number_of_transactions_per_day = EXCLUDED.average_number_of_transactions_per_day,
    registered_payment_modes = EXCLUDED.registered_payment_modes,
    usual_transaction_start_hour = EXCLUDED.usual_transaction_start_hour,
    usual_transaction_end_hour = EXCLUDED.usual_transaction_end_hour,
    total_transactions = EXCLUDED.total_transactions,
    allowed_transactions = EXCLUDED.allowed_transactions,
    updated_at = EXCLUDED.updated_at;

-- name: GetUserProfileByUserID :one
SELECT
    user_id,
    average_transaction_amount,
    std_dev_transaction_amount,
    max_transaction_amount_seen,
    average_number_of_transactions_per_day,
    registered_payment_modes::text[] AS registered_payment_modes,
    usual_transaction_start_hour,
    usual_transaction_end_hour,
    total_transactions,
    allowed_transactions,
    updated_at
FROM user_profile_behavior
WHERE user_id = $1;

-- name: UpsertUserProfileFromProfile :exec
INSERT INTO user_profile_behavior (
    user_id,
    average_transaction_amount,
    std_dev_transaction_amount,
    max_transaction_amount_seen,
    average_number_of_transactions_per_day,
    registered_payment_modes,
    usual_transaction_start_hour,
    usual_transaction_end_hour,
    total_transactions,
    allowed_transactions,
    updated_at
)
VALUES (
    $1,  -- user_id
    $2,  -- average_transaction_amount
    $3,  -- std_dev_transaction_amount
    $4,  -- max_transaction_amount_seen
    $5,  -- average_number_of_transactions_per_day
    $6,  -- registered_payment_modes
    $7,  -- usual_transaction_start_hour
    $8,  -- usual_transaction_end_hour
    $9,  -- total_transactions
    $10, -- allowed_transactions
    NOW()
)
ON CONFLICT (user_id) DO UPDATE SET
    average_transaction_amount = EXCLUDED.average_transaction_amount,
    std_dev_transaction_amount = EXCLUDED.std_dev_transaction_amount,
    max_transaction_amount_seen = EXCLUDED.max_transaction_amount_seen,
    average_number_of_transactions_per_day = EXCLUDED.average_number_of_transactions_per_day,
    registered_payment_modes = EXCLUDED.registered_payment_modes,
    usual_transaction_start_hour = EXCLUDED.usual_transaction_start_hour,
    usual_transaction_end_hour = EXCLUDED.usual_transaction_end_hour,
    total_transactions = EXCLUDED.total_transactions,
    allowed_transactions = EXCLUDED.allowed_transactions,
    updated_at = NOW();

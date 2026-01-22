-- name: SeedUserBehaviorProfiles :exec
INSERT INTO user_profile_behavior (
    user_id,
    average_transaction_amount,
    max_transaction_amount_seen,
    average_number_of_transactions_per_day,
    registered_payment_modes,
    usual_transaction_start_hour,
    usual_transaction_end_hour,
    updated_at
)
SELECT
    u.id,
    0,
    0,
    0,
    '{}',
    NULL,
    NULL,
    NOW()
FROM users u
ON CONFLICT (user_id) DO NOTHING;

-- name: RebuildAllUserProfiles :exec
BEGIN;

INSERT INTO user_profile_behavior (
    user_id,
    average_transaction_amount,
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

    COALESCE(MAX(t.amount), 0)::INTEGER AS max_transaction_amount_seen,

    LEAST(
        COUNT(*) FILTER (WHERE t.decision IN ('ALLOW', 'FLAG')),
        50
    )::INTEGER AS average_number_of_transactions_per_day,

    ARRAY_AGG(DISTINCT t.mode)
        FILTER (WHERE t.decision IN ('ALLOW', 'FLAG')) AS registered_payment_modes,

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
    max_transaction_amount_seen = EXCLUDED.max_transaction_amount_seen,
    average_number_of_transactions_per_day = EXCLUDED.average_number_of_transactions_per_day,
    registered_payment_modes = EXCLUDED.registered_payment_modes,
    usual_transaction_start_hour = EXCLUDED.usual_transaction_start_hour,
    usual_transaction_end_hour = EXCLUDED.usual_transaction_end_hour,
    updated_at = EXCLUDED.updated_at;

COMMIT;

-- name: RebuildUserProfileByID :exec
INSERT INTO user_profile_behavior (
    user_id,
    average_transaction_amount,
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
    $1::INTEGER AS user_id,

    COALESCE(
        AVG(amount) FILTER (WHERE decision IN ('ALLOW', 'FLAG')),
        0
    )::INTEGER AS average_transaction_amount,

    COALESCE(MAX(amount), 0)::INTEGER AS max_transaction_amount_seen,

    LEAST(
        COUNT(*) FILTER (WHERE decision IN ('ALLOW', 'FLAG')),
        50
    )::INTEGER AS average_number_of_transactions_per_day,

    ARRAY_AGG(DISTINCT mode)
        FILTER (WHERE decision IN ('ALLOW', 'FLAG')) AS registered_payment_modes,

    MIN(created_at)
        FILTER (WHERE decision IN ('ALLOW', 'FLAG')) AS usual_transaction_start_hour,

    MAX(created_at)
        FILTER (WHERE decision IN ('ALLOW', 'FLAG')) AS usual_transaction_end_hour,

    COUNT(*) AS total_transactions,

    COUNT(*) FILTER (WHERE decision IN ('ALLOW', 'FLAG')) AS allowed_transactions,

    NOW() AS updated_at

FROM transactions
WHERE user_id = $1
  AND created_at < CURRENT_DATE

ON CONFLICT (user_id) DO UPDATE SET 
    average_transaction_amount = EXCLUDED.average_transaction_amount,
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


-- name: RebuildProfilesForSeededUsers :exec
SELECT RebuildAllUserProfiles();

-- name: UserProfileExists :one
SELECT EXISTS (
    SELECT 1
    FROM user_profile_behavior
    WHERE user_id = $1
);

-- name: CreateEmptyUserProfile :exec
INSERT INTO user_profile_behavior (
    user_id,
    average_transaction_amount,
    max_transaction_amount_seen,
    average_number_of_transactions_per_day,
    registered_payment_modes,
    usual_transaction_start_hour,
    usual_transaction_end_hour,
    updated_at
)
VALUES (
    $1,
    0,
    0,
    0,
    '{}',
    NULL,
    NULL,
    NOW()
)
ON CONFLICT (user_id) DO NOTHING;
